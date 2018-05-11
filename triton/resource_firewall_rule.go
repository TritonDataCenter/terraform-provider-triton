package triton

import (
	"context"

	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/network"
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
				StateFunc: func(v interface{}) string {
					switch v.(type) {
					case string:
						return strings.TrimSpace(v.(string))
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
	n, err := client.Network()
	if err != nil {
		return err
	}

	rule, err := n.Firewall().CreateRule(context.Background(), &network.CreateRuleInput{
		Rule:        d.Get("rule").(string),
		Enabled:     d.Get("enabled").(bool),
		Description: d.Get("description").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(rule.ID)

	return resourceFirewallRuleRead(d, meta)
}

func resourceFirewallRuleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	n, err := client.Network()
	if err != nil {
		return false, err
	}

	return resourceExists(n.Firewall().GetRule(context.Background(), &network.GetRuleInput{
		ID: d.Id(),
	}))
}

func resourceFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	n, err := client.Network()
	if err != nil {
		return err
	}

	rule, err := n.Firewall().GetRule(context.Background(), &network.GetRuleInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	d.SetId(rule.ID)
	d.Set("rule", rule.Rule)
	d.Set("enabled", rule.Enabled)
	d.Set("global", rule.Global)
	d.Set("description", rule.Description)

	return nil
}

func resourceFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	n, err := client.Network()
	if err != nil {
		return err
	}

	_, err = n.Firewall().UpdateRule(context.Background(), &network.UpdateRuleInput{
		ID:          d.Id(),
		Rule:        d.Get("rule").(string),
		Enabled:     d.Get("enabled").(bool),
		Description: d.Get("description").(string),
	})
	if err != nil {
		return err
	}

	return resourceFirewallRuleRead(d, meta)
}

func resourceFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	n, err := client.Network()
	if err != nil {
		return err
	}

	return n.Firewall().DeleteRule(context.Background(), &network.DeleteRuleInput{
		ID: d.Id(),
	})
}
