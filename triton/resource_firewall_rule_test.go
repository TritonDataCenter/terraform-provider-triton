package triton

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

	resp, err := client.API().ListFirewallRulesWithResponse(context.Background(), client.Account())
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error listing firewall rules: unexpected status %d", resp.StatusCode())
	}

	rules := *resp.JSON200
	log.Printf("[DEBUG] Found %d firewall rules", len(rules))

	for _, v := range rules {
		desc := ""
		if v.Description != nil {
			desc = *v.Description
		}
		if strings.HasPrefix(desc, "Test-Firewall-Rule") {
			log.Printf("Destroying firewall rule %q", desc)

			delResp, err := client.API().DeleteFirewallRuleWithResponse(context.Background(), client.Account(), v.ID)
			if err != nil {
				return err
			}
			if delResp.StatusCode() >= 400 && !isNotFound(delResp.StatusCode()) {
				return fmt.Errorf("error deleting firewall rule: status %d", delResp.StatusCode())
			}
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
			{
				ResourceName:      "triton_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
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

func TestAccTritonFirewallRule_heredoc(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccTritonFirewallRule_heredoc,
				ExpectNonEmptyPlan: false,
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

		ruleID, err := parseUUID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Bad: invalid firewall rule ID: %s", err)
		}

		resp, err := conn.API().GetFirewallRuleWithResponse(context.Background(), conn.Account(), ruleID)
		if err != nil {
			return fmt.Errorf("Bad: Check Firewall Rule Exists: %s", err)
		}
		if isNotFound(resp.StatusCode()) {
			return fmt.Errorf("Bad: Check Firewall Rule Exists: not found")
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("Bad: Firewall Rule %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckTritonFirewallRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "triton_firewall_rule" {
			continue
		}

		ruleID, err := parseUUID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid firewall rule ID: %s", err)
		}

		resp, err := conn.API().GetFirewallRuleWithResponse(context.Background(), conn.Account(), ruleID)
		if err != nil {
			return err
		}
		if isNotFound(resp.StatusCode()) {
			return nil
		}

		if resp.JSON200 != nil {
			return fmt.Errorf("Bad: Firewall Rule %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

var testAccTritonFirewallRule_basic = `
resource "triton_firewall_rule" "test" {
	rule = "FROM any TO tag \"www\" ALLOW tcp PORT 80"
	enabled = false
	description = "Test-Firewall-Rule"
}
`

var testAccTritonFirewallRule_update = `
resource "triton_firewall_rule" "test" {
	rule = "FROM any TO tag \"www\" BLOCK tcp PORT 80"
	enabled = true
	description = "Test-Firewall-Rule"
}
`

var testAccTritonFirewallRule_enable = `
resource "triton_firewall_rule" "test" {
	rule = "FROM any TO tag \"www\" ALLOW tcp PORT 80"
	enabled = true
	description = "Test-Firewall-Rule"
}
`

var testAccTritonFirewallRule_heredoc = `
resource "triton_firewall_rule" "test" {
	rule = <<EOS
FROM any TO tag "www" ALLOW tcp PORT 80
EOS
	enabled = true
	description = "Test-Firewall-Rule"
}
`
