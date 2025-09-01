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

	triton "github.com/TritonDataCenter/triton-go"
	"github.com/TritonDataCenter/triton-go/authentication"
	"github.com/TritonDataCenter/triton-go/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"account": {
				Description: "This is the name of the Triton account. It can also be provided via the `SDC_ACCOUNT` or `TRITON_ACCOUNT` environment variables.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_ACCOUNT", "SDC_ACCOUNT"}, ""),
			},

			"user": {
				Description: "This is the username to interact with the Triton API. It can be provided via the `SDC_USER` or `TRITON_USER` environment variables.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_USER", "SDC_USER"}, ""),
			},

			"url": {
				Description: "This is the URL to the Triton API endpoint. It is required if using a private installation of Triton. The default is to use the MNX.io public cloud `us-central-1` endpoint. It can be provided via the `SDC_URL` or `TRITON_URL` environment variables.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_URL", "SDC_URL"}, "https://us-central-1.api.mnx.io"),
			},

			"key_material": {
				Description: "This is the private key of an SSH key associated with the Triton account to be used. If this is not set, the private key corresponding to the fingerprint in `key_id` must be available via an SSH Agent. It can be provided via the `SDC_KEY_MATERIAL` or `TRITON_KEY_MATERIAL` environment variables.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_KEY_MATERIAL", "SDC_KEY_MATERIAL"}, ""),
			},

			"key_id": {
				Description: "This is the fingerprint of the public key matching the key specified in `key_path`. It can be obtained via the command `ssh-keygen -l -E md5 -f /path/to/key`. It can be provided via the `SDC_KEY_ID` or `TRITON_KEY_ID` environment variables.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"TRITON_KEY_ID", "SDC_KEY_ID"}, ""),
			},

			"insecure_skip_tls_verify": {
				Description: "This allows skipping TLS verification of the Triton endpoint. It is useful when connecting to a temporary Triton installation such as Cloud-On-A-Laptop which does not generally use a certificate signed by a trusted root CA.",
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TRITON_SKIP_TLS_VERIFY", false),
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
			"triton_volume":         dataSourceVolume(),
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
			"triton_volume":            resourceVolume(),
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
			return nil, fmt.Errorf("Error Creating SSH Agent Signer: %s", err)
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
			return nil, fmt.Errorf("Error Creating SSH Private Key Signer: %s", err)
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

// Default polling interval - how long to wait between subsequent resource
// checks.
const defaultPollInterval = 3 * time.Second
