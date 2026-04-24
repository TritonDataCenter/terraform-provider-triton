package triton

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("triton_key", &resource.Sweeper{
		Name: "triton_key",
		F:    testSweepKeys,
	})
}

func testSweepKeys(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*Client)

	resp, err := client.API().ListKeysWithResponse(context.Background(), client.Account())
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error listing keys: unexpected status %d", resp.StatusCode())
	}

	keys := *resp.JSON200
	log.Printf("[DEBUG] Found %d keys", len(keys))

	for _, v := range keys {
		if strings.HasPrefix(v.Name, "acctest-") {
			log.Printf("Destroying key %s", v.Name)

			delResp, err := client.API().DeleteKeyWithResponse(context.Background(), client.Account(), v.Name)
			if err != nil {
				return err
			}
			if delResp.StatusCode() >= 400 && !isNotFound(delResp.StatusCode()) {
				return fmt.Errorf("error deleting key %s: status %d", v.Name, delResp.StatusCode())
			}
		}
	}

	return nil
}

func TestAccTritonKey_basic(t *testing.T) {
	keyName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("TestAccTritonKey_basic@terraform")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}
	config := testAccTritonKey_basic(keyName, publicKeyMaterial)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonKeyExists("triton_key.test"),
					resource.TestCheckResourceAttr("triton_key.test", "name", keyName),
					resource.TestCheckResourceAttr("triton_key.test", "key", publicKeyMaterial),
					func(*terraform.State) error {
						time.Sleep(10 * time.Second)
						return nil
					},
				),
			},
			{
				ResourceName:      "triton_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTritonKey_noKeyName(t *testing.T) {
	keyComment := fmt.Sprintf("acctest-%d@terraform", acctest.RandInt())
	keyMaterial, _, err := acctest.RandSSHKeyPair(keyComment)
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}
	config := testAccTritonKey_noKeyName(keyMaterial)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonKeyExists("triton_key.test"),
					resource.TestCheckResourceAttr("triton_key.test", "name", keyComment),
					resource.TestCheckResourceAttr("triton_key.test", "key", keyMaterial),
					func(*terraform.State) error {
						time.Sleep(10 * time.Second)
						return nil
					},
				),
			},
		},
	})
}

func TestAccTritonKey_nameWithSpace(t *testing.T) {
	keyName := fmt.Sprintf("acctest- space key %d", acctest.RandInt())
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("TestAccTritonKey_nameWithSpace@terraform")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}
	config := testAccTritonKey_basic(keyName, publicKeyMaterial)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonKeyExists("triton_key.test"),
					resource.TestCheckResourceAttr("triton_key.test", "name", keyName),
					resource.TestCheckResourceAttr("triton_key.test", "key", publicKeyMaterial),
					func(*terraform.State) error {
						time.Sleep(10 * time.Second)
						return nil
					},
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonKeyExists("triton_key.test"),
					resource.TestCheckResourceAttr("triton_key.test", "name", keyName),
				),
			},
			{
				ResourceName:      "triton_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckTritonKeyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		conn := testAccProvider.Meta().(*Client)

		resp, err := conn.API().GetKeyWithResponse(context.Background(), conn.Account(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Bad: Check Key Exists: %s", err)
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("Bad: Key %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckTritonKeyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)

	return retry.Retry(1*time.Minute, func() *retry.RetryError {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "triton_key" {
				continue
			}

			resp, err := conn.API().GetKeyWithResponse(context.Background(), conn.Account(), rs.Primary.ID)
			if err != nil {
				return nil
			}

			if resp.JSON200 != nil {
				return retry.RetryableError(fmt.Errorf("Bad: Key %q still exists", rs.Primary.ID))
			}
		}

		return nil
	})
}

var testAccTritonKey_basic = func(keyName string, keyMaterial string) string {
	return fmt.Sprintf(`resource "triton_key" "test" {
		name = "%s"
		key = "%s"
	}
	`, keyName, keyMaterial)
}

var testAccTritonKey_noKeyName = func(keyMaterial string) string {
	return fmt.Sprintf(`resource "triton_key" "test" {
		key = "%s"
	}
	`, keyMaterial)
}
