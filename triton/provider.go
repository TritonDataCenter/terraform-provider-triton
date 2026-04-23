/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2021 Joyent, Inc.
 * Copyright 2022 MNX Cloud, Inc.
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
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
			"triton_fabric":        resourceFabric(),
			"triton_firewall_rule": resourceFirewallRule(),
			"triton_key":           resourceKey(),
			"triton_machine":       resourceMachine(),
			"triton_snapshot":      resourceSnapshot(),
			"triton_vlan":          resourceVLAN(),
			"triton_volume":        resourceVolume(),
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
		err = multierror.Append(err, stderrors.New("url must be configured for the triton provider"))
	}
	if c.KeyID == "" {
		err = multierror.Append(err, stderrors.New("key id must be configured for the triton provider"))
	}
	if c.Account == "" {
		err = multierror.Append(err, stderrors.New("account must be configured for the triton provider"))
	}

	return err.ErrorOrNil()
}

func (c Config) newClient() (*Client, error) {
	signer, err := cloudapi.LoadSignerFromEnvWithOptions(cloudapi.SignerFromEnvOptions{
		AccountName: c.Account,
		Username:    c.Username,
		KeyID:       c.KeyID,
		KeyMaterial: c.KeyMaterial,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating signer: %s", err)
	}

	authOpts := cloudapi.SignatureAuthOptions{
		Signer: signer,
	}

	var opts []cloudapi.ClientOption
	if c.InsecureSkipTLSVerify {
		opts = append(opts, cloudapi.WithTLSInsecure())
	}

	api, err := cloudapi.NewAuthenticatedClientWithResponses(c.URL, authOpts, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating cloudapi client: %s", err)
	}

	return &Client{
		api:          api,
		account:      c.Account,
		url:          c.URL,
		affinityLock: &sync.RWMutex{},
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

// isNotFound returns true if the HTTP status code indicates the resource
// does not exist (404 Not Found or 410 Gone).
func isNotFound(statusCode int) bool {
	return statusCode == http.StatusNotFound || statusCode == http.StatusGone
}

// formatAPIError extracts a human-readable error from a CloudAPI response body.
// Falls back to reporting the status code if the body cannot be parsed.
func formatAPIError(statusCode int, body []byte) string {
	var apiErr struct {
		Code    string  `json:"code"`
		Message *string `json:"message,omitempty"`
	}
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Code != "" {
		if apiErr.Message != nil && *apiErr.Message != "" {
			return fmt.Sprintf("%s: %s (HTTP %d)", apiErr.Code, *apiErr.Message, statusCode)
		}
		return fmt.Sprintf("%s (HTTP %d)", apiErr.Code, statusCode)
	}
	return fmt.Sprintf("unexpected status %d", statusCode)
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
