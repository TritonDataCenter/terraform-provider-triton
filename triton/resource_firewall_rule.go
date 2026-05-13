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
	"fmt"
	"strings"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallRuleCreate,
		Exists: resourceFirewallRuleExists,
		Read:   resourceFirewallRuleRead,
		Update: resourceFirewallRuleUpdate,
		Delete: resourceFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"rule": {
				Description: "firewall rule text",
				Type:        schema.TypeString,
				Required:    true,
				// Regression guard (#105): trim trailing whitespace so
				// heredoc-defined rules do not cause perpetual diffs.
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return strings.TrimSpace(v)
					default:
						return ""
					}
				},
			},
			"enabled": {
				Description: "Indicates if the rule is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"description": {
				Description: "Human-readable description of the rule",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"global": {
				Description: "Indicates whether or not the rule is global",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func resourceFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	enabled := d.Get("enabled").(bool)
	description := d.Get("description").(string)

	resp, err := client.API().CreateFirewallRuleWithResponse(context.Background(), client.Account(),
		cloudapi.CreateFirewallRuleJSONRequestBody{
			Rule:        d.Get("rule").(string),
			Enabled:     &enabled,
			Description: &description,
		})
	if err != nil {
		return fmt.Errorf("error creating firewall rule: %s", err)
	}
	if resp.JSON201 == nil {
		return fmt.Errorf("error creating firewall rule: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	d.SetId(uuidString(resp.JSON201.ID))

	return resourceFirewallRuleRead(d, meta)
}

func resourceFirewallRuleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)

	id, err := parseUUID(d.Id())
	if err != nil {
		return false, fmt.Errorf("invalid firewall rule ID: %s", err)
	}

	resp, err := client.API().GetFirewallRuleWithResponse(context.Background(), client.Account(), id)
	if err != nil {
		return false, err
	}
	if isNotFound(resp.StatusCode()) {
		return false, nil
	}
	if resp.JSON200 == nil {
		return false, fmt.Errorf("error checking firewall rule: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return true, nil
}

func resourceFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	id, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid firewall rule ID: %s", err)
	}

	resp, err := client.API().GetFirewallRuleWithResponse(context.Background(), client.Account(), id)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error reading firewall rule: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	rule := resp.JSON200
	d.SetId(uuidString(rule.ID))
	d.Set("rule", rule.Rule)
	d.Set("enabled", rule.Enabled)
	d.Set("global", derefBool(rule.Global))
	d.Set("description", derefString(rule.Description))

	return nil
}

func resourceFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	id, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid firewall rule ID: %s", err)
	}

	rule := d.Get("rule").(string)
	enabled := d.Get("enabled").(bool)
	description := d.Get("description").(string)

	resp, err := client.API().UpdateFirewallRuleWithResponse(context.Background(), client.Account(), id,
		cloudapi.UpdateFirewallRuleJSONRequestBody{
			Rule:        &rule,
			Enabled:     &enabled,
			Description: &description,
		})
	if err != nil {
		return fmt.Errorf("error updating firewall rule: %s", err)
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error updating firewall rule: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return resourceFirewallRuleRead(d, meta)
}

func resourceFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	id, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid firewall rule ID: %s", err)
	}

	resp, err := client.API().DeleteFirewallRuleWithResponse(context.Background(), client.Account(), id)
	if err != nil {
		return fmt.Errorf("error deleting firewall rule: %s", err)
	}
	if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
		return fmt.Errorf("error deleting firewall rule: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return nil
}
