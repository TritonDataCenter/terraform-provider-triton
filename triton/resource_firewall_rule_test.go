package triton

import (
	"context"
	"fmt"
	"testing"

	"log"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
)

func init() {
	resource.AddTestSweepers("triton_firewall_rule", &resource.Sweeper{
		Name: "triton_firewall_rule",
		F:    testSweepFirewallRules,
	})

}

func testSweepFirewallRules(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*Client)
	a, err := client.Network()
	if err != nil {
		return err
	}

	rules, err := a.Firewall().ListRules(context.Background(), &network.ListRulesInput{})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Found %d rules to sweep", len(rules))

	for _, v := range rules {
		log.Printf("Destroying rule %q", v.Description)

		if err := a.Firewall().DeleteRule(context.Background(), &network.DeleteRuleInput{
			ID: v.ID,
		}); err != nil {
			return err
		}

	}

	return nil
}

func TestAccTritonFirewallRule_basic(t *testing.T) {
	config := testAccTritonFirewallRule_basic

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonFirewallRuleExists("triton_firewall_rule.test"),
				),
			},
		},
	})
}

func TestAccTritonFirewallRule_update(t *testing.T) {
	preConfig := testAccTritonFirewallRule_basic
	postConfig := testAccTritonFirewallRule_update

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonFirewallRuleExists("triton_firewall_rule.test"),
					resource.TestCheckResourceAttr("triton_firewall_rule.test", "rule", "FROM any TO tag \"www\" ALLOW tcp PORT 80"),
					resource.TestCheckResourceAttr("triton_firewall_rule.test", "enabled", "false"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonFirewallRuleExists("triton_firewall_rule.test"),
					resource.TestCheckResourceAttr("triton_firewall_rule.test", "rule", "FROM any TO tag \"www\" BLOCK tcp PORT 80"),
					resource.TestCheckResourceAttr("triton_firewall_rule.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccTritonFirewallRule_enable(t *testing.T) {
	preConfig := testAccTritonFirewallRule_basic
	postConfig := testAccTritonFirewallRule_enable

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonFirewallRuleExists("triton_firewall_rule.test"),
					resource.TestCheckResourceAttr("triton_firewall_rule.test", "rule", "FROM any TO tag \"www\" ALLOW tcp PORT 80"),
					resource.TestCheckResourceAttr("triton_firewall_rule.test", "enabled", "false"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonFirewallRuleExists("triton_firewall_rule.test"),
					resource.TestCheckResourceAttr("triton_firewall_rule.test", "rule", "FROM any TO tag \"www\" ALLOW tcp PORT 80"),
					resource.TestCheckResourceAttr("triton_firewall_rule.test", "enabled", "true"),
				),
			},
		},
	})
}

func testCheckTritonFirewallRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		conn := testAccProvider.Meta().(*Client)
		n, err := conn.Network()
		if err != nil {
			return err
		}

		resp, err := n.Firewall().GetRule(context.Background(), &network.GetRuleInput{
			ID: rs.Primary.ID,
		})
		if err != nil && compute.IsResourceNotFound(err) {
			return fmt.Errorf("Bad: Check Firewall Rule Exists: %s", err)
		} else if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("Bad: Firewall Rule %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckTritonFirewallRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)
	n, err := conn.Network()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "triton_firewall_rule" {
			continue
		}

		resp, err := n.Firewall().GetRule(context.Background(), &network.GetRuleInput{
			ID: rs.Primary.ID,
		})
		if compute.IsResourceNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("Bad: Firewall Rule %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

var testAccTritonFirewallRule_basic = `
resource "triton_firewall_rule" "test" {
	rule = "FROM any TO tag \"www\" ALLOW tcp PORT 80"
	enabled = false
}
`

var testAccTritonFirewallRule_update = `
resource "triton_firewall_rule" "test" {
	rule = "FROM any TO tag \"www\" BLOCK tcp PORT 80"
	enabled = true
}
`

var testAccTritonFirewallRule_enable = `
resource "triton_firewall_rule" "test" {
	rule = "FROM any TO tag \"www\" ALLOW tcp PORT 80"
	enabled = true
}
`
