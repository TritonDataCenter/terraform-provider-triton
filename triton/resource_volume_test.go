package triton

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("triton_volume", &resource.Sweeper{
		Name: "triton_volume",
		F:    testSweepVolumes,
	})
}

func testSweepVolumes(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*Client)

	resp, err := client.API().ListVolumesWithResponse(context.Background(), client.Account())
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error listing volumes: unexpected status %d", resp.StatusCode())
	}

	volumes := *resp.JSON200
	log.Printf("[DEBUG] Found %d volumes", len(volumes))

	for _, v := range volumes {
		if strings.HasPrefix(v.Name, "acctest-") {
			log.Printf("Destroying volume %s", v.Name)

			delResp, err := client.API().DeleteVolumeWithResponse(context.Background(), client.Account(), v.ID)
			if err != nil {
				return err
			}
			if delResp.StatusCode() >= 400 && !isNotFound(delResp.StatusCode()) {
				return fmt.Errorf("error deleting volume %s: status %d", v.Name, delResp.StatusCode())
			}
		}
	}

	return nil
}

func TestAccTritonVolume_basic(t *testing.T) {
	volumeName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	config := fmt.Sprintf(`
		resource "triton_volume" "test" {
			name = "%s"
		}
	`, volumeName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonVolumeExists("triton_volume.test"),
					resource.TestCheckResourceAttrSet("triton_volume.test", "size"),
					resource.TestCheckResourceAttrSet("triton_volume.test", "filesystem_path"),
					resource.TestCheckResourceAttr("triton_volume.test", "type", "tritonnfs"),
					resource.TestCheckResourceAttr("triton_volume.test", "state", volumeStateReady),
				),
			},
			{
				ResourceName:      "triton_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTritonVolume_singleNetwork(t *testing.T) {
	networkName := testAccConfig(t, "test_network_name")
	volumeName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	config := fmt.Sprintf(`
		data "triton_network" "test" {
			name = "%s"
		}

		resource "triton_volume" "test" {
			name = "%s-volume"
			tags = {
				test = "Test"
			}
			networks = ["${data.triton_network.test.id}"]
		}
	`, networkName, volumeName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonVolumeExists("triton_volume.test"),
					resource.TestCheckResourceAttr("triton_volume.test", "networks.#", "1"),
				),
			},
		},
	})
}

func testCheckTritonVolumeExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		conn := testAccProvider.Meta().(*Client)

		volID, err := parseUUID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Bad: invalid volume ID: %s", err)
		}

		resp, err := conn.API().GetVolumeWithResponse(context.Background(), conn.Account(), volID)
		if err != nil {
			return fmt.Errorf("Bad: Check Volume Exists: %s", err)
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("Bad: Volume %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckTritonVolumeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "triton_volume" {
			continue
		}

		volID, err := parseUUID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid volume ID: %s", err)
		}

		resp, err := conn.API().GetVolumeWithResponse(context.Background(), conn.Account(), volID)
		if err != nil {
			return err
		}

		if isNotFound(resp.StatusCode()) {
			return nil
		}

		if resp.JSON200 != nil && string(resp.JSON200.State) != volumeStateDeleted {
			return fmt.Errorf("Bad: Volume %q still exists", rs.Primary.ID)
		}
	}

	return nil
}
