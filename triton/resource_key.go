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
	"context"
	"errors"
	"fmt"
	"strings"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKey() *schema.Resource {
	return &schema.Resource{
		Create:   resourceKeyCreate,
		Exists:   resourceKeyExists,
		Read:     resourceKeyRead,
		Delete:   resourceKeyDelete,
		Timeouts: fastResourceTimeout,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		// v0 used the key name as the resource ID; v1 uses the fingerprint.
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourceKeyV0Schema().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceKeyStateUpgradeV0,
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the key (generated from the key comment if not set)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"key": {
				Description: "Content of public key from disk in OpenSSH format",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
					return strings.TrimSpace(oldVal) == strings.TrimSpace(newVal)
				},
			},
		},
	}
}

func resourceKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	if keyName := d.Get("name").(string); keyName == "" {
		parts := strings.SplitN(d.Get("key").(string), " ", 3)
		if len(parts) == 3 {
			d.Set("name", parts[2])
		} else {
			return errors.New("no key name specified, and key material has no comment")
		}
	}

	resp, err := client.API().CreateKeyWithResponse(context.Background(), client.Account(),
		cloudapi.CreateKeyJSONRequestBody{
			Name: d.Get("name").(string),
			Key:  d.Get("key").(string),
		})
	if err != nil {
		return fmt.Errorf("error creating key: %s", err)
	}
	if resp.JSON201 == nil {
		return fmt.Errorf("error creating key: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	d.SetId(resp.JSON201.Fingerprint)

	return resourceKeyRead(d, meta)
}

func resourceKeyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)

	resp, err := client.API().GetKeyWithResponse(context.Background(), client.Account(), d.Id())
	if err != nil {
		return false, err
	}
	if isNotFound(resp.StatusCode()) {
		return false, nil
	}
	if resp.JSON200 == nil {
		return false, fmt.Errorf("error checking key existence: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return true, nil
}

func resourceKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	resp, err := client.API().GetKeyWithResponse(context.Background(), client.Account(), d.Id())
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error reading key: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	key := resp.JSON200
	d.SetId(key.Fingerprint)
	d.Set("name", key.Name)
	d.Set("key", strings.TrimSpace(key.Key))

	return nil
}

func resourceKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	resp, err := client.API().DeleteKeyWithResponse(context.Background(), client.Account(), d.Id())
	if err != nil {
		return fmt.Errorf("error deleting key: %s", err)
	}
	if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
		return fmt.Errorf("error deleting key: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return nil
}

// resourceKeyV0Schema returns the v0 schema, which is identical to v1.
// The only difference is the ID semantics (name vs fingerprint).
func resourceKeyV0Schema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

// resourceKeyStateUpgradeV0 migrates a v0 state (name-based ID) to v1
// (fingerprint-based ID) by looking up the key via CloudAPI.
func resourceKeyStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	client := meta.(*Client)

	id, ok := rawState["id"].(string)
	if !ok || id == "" {
		return rawState, fmt.Errorf("key state upgrade: missing or empty id")
	}

	// If the ID already looks like a fingerprint, no migration needed.
	if strings.Contains(id, ":") {
		return rawState, nil
	}

	resp, err := client.API().GetKeyWithResponse(context.Background(), client.Account(), id)
	if err != nil {
		return rawState, fmt.Errorf("key state upgrade: error looking up key %q: %s", id, err)
	}
	if resp.JSON200 == nil {
		return rawState, fmt.Errorf("key state upgrade: key %q not found (status %d)", id, resp.StatusCode())
	}

	rawState["id"] = resp.JSON200.Fingerprint
	return rawState, nil
}
