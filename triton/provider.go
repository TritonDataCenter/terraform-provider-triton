package triton

import (
	"encoding/pem"
	stderrors "errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/errors"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"account": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_ACCOUNT", "SDC_ACCOUNT"}, ""),
			},

			"user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_USER", "SDC_USER"}, ""),
			},

			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_URL", "SDC_URL"}, "https://us-west-1.api.joyentcloud.com"),
			},

			"key_material": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_KEY_MATERIAL", "SDC_KEY_MATERIAL"}, ""),
			},

			"key_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_KEY_ID", "SDC_KEY_ID"}, ""),
			},

			"insecure_skip_tls_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TRITON_SKIP_TLS_VERIFY", ""),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"triton_account":        dataSourceAccount(),
			"triton_datacenter":     dataSourceDataCenter(),
			"triton_image":          dataSourceImage(),
			"triton_network":        dataSourceNetwork(),
			"triton_package":        dataSourcePackage(),
			"triton_fabric_vlan":    dataSourceFabricVLAN(),
			"triton_fabric_network": dataSourceFabricNetwork(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"triton_fabric":            resourceFabric(),
			"triton_firewall_rule":     resourceFirewallRule(),
			"triton_instance_template": resourceInstanceTemplate(),
			"triton_key":               resourceKey(),
			"triton_machine":           resourceMachine(),
			"triton_service_group":     resourceServiceGroup(),
			"triton_snapshot":          resourceSnapshot(),
			"triton_vlan":              resourceVLAN(),
		},
		ConfigureFunc: providerConfigure,
	}
}

// Config represents this provider's configuration data.
type Config struct {
	Account               string
	Username              string
	KeyMaterial           string
	KeyID                 string
	URL                   string
	InsecureSkipTLSVerify bool
}

func (c Config) validate() error {
	var err *multierror.Error

	if c.URL == "" {
		err = multierror.Append(err, stderrors.New("URL must be configured for the Triton provider"))
	}
	if c.KeyID == "" {
		err = multierror.Append(err, stderrors.New("Key ID must be configured for the Triton provider"))
	}
	if c.Account == "" {
		err = multierror.Append(err, stderrors.New("Account must be configured for the Triton provider"))
	}

	return err.ErrorOrNil()
}

func (c Config) newClient() (*Client, error) {
	var signer authentication.Signer
	var err error

	if c.KeyMaterial == "" {
		signer, err = authentication.NewSSHAgentSigner(authentication.SSHAgentSignerInput{
			KeyID:       c.KeyID,
			AccountName: c.Account,
			Username:    c.Username,
		})
		if err != nil {
			return nil, errwrap.Wrapf("Error Creating SSH Agent Signer: {{err}}", err)
		}
	} else {
		var keyBytes []byte
		if _, err = os.Stat(c.KeyMaterial); err == nil {
			keyBytes, err = ioutil.ReadFile(c.KeyMaterial)
			if err != nil {
				return nil, fmt.Errorf("Error reading key material from %s: %s",
					c.KeyMaterial, err)
			}
			block, _ := pem.Decode(keyBytes)
			if block == nil {
				return nil, fmt.Errorf(
					"Failed to read key material '%s': no key found", c.KeyMaterial)
			}

			if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
				return nil, fmt.Errorf(
					"Failed to read key '%s': password protected keys are\n"+
						"not currently supported. Please decrypt the key prior to use.", c.KeyMaterial)
			}

		} else {
			keyBytes = []byte(c.KeyMaterial)
		}

		signer, err = authentication.NewPrivateKeySigner(authentication.PrivateKeySignerInput{
			KeyID:              c.KeyID,
			PrivateKeyMaterial: keyBytes,
			AccountName:        c.Account,
			Username:           c.Username,
		})
		if err != nil {
			return nil, errwrap.Wrapf("Error Creating SSH Private Key Signer: {{err}}", err)
		}
	}

	config := &triton.ClientConfig{
		TritonURL:   c.URL,
		AccountName: c.Account,
		Username:    c.Username,
		Signers:     []authentication.Signer{signer},
	}

	return &Client{
		config:                config,
		insecureSkipTLSVerify: c.InsecureSkipTLSVerify,
		affinityLock:          &sync.RWMutex{},
	}, nil
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Account: d.Get("account").(string),
		URL:     d.Get("url").(string),
		KeyID:   d.Get("key_id").(string),

		InsecureSkipTLSVerify: d.Get("insecure_skip_tls_verify").(bool),
	}

	if keyMaterial, ok := d.GetOk("key_material"); ok {
		config.KeyMaterial = keyMaterial.(string)
	}

	if user, ok := d.GetOk("user"); ok {
		config.Username = user.(string)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	client, err := config.newClient()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func resourceExists(resource interface{}, err error) (bool, error) {
	if err != nil {
		if errors.IsSpecificStatusCode(err, http.StatusNotFound) ||
			errors.IsSpecificStatusCode(err, http.StatusGone) {
			return false, nil
		}

		return false, err
	}

	return resource != nil, nil
}

var fastResourceTimeout = &schema.ResourceTimeout{
	Create: schema.DefaultTimeout(1 * time.Minute),
	Read:   schema.DefaultTimeout(30 * time.Second),
	Update: schema.DefaultTimeout(1 * time.Minute),
	Delete: schema.DefaultTimeout(1 * time.Minute),
}

var slowResourceTimeout = &schema.ResourceTimeout{
	Create: schema.DefaultTimeout(10 * time.Minute),
	Read:   schema.DefaultTimeout(30 * time.Second),
	Update: schema.DefaultTimeout(10 * time.Minute),
	Delete: schema.DefaultTimeout(10 * time.Minute),
}
