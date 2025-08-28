package triton

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/TritonDataCenter/triton-go/compute"
	"github.com/TritonDataCenter/triton-go/errors"
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
	a, err := client.Compute()
	if err != nil {
		return err
	}

	volumes, err := a.Volumes().List(context.Background(), &compute.ListVolumesInput{})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Found %d volumes", len(volumes))

	for _, v := range volumes {
		if strings.HasPrefix(v.Name, "acctest-") {
			log.Printf("Destroying volume %s", v.Name)

			if err := a.Volumes().Delete(context.Background(), &compute.DeleteVolumeInput{
				ID: v.ID,
			}); err != nil {
				return err
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
		c, err := conn.Compute()
		if err != nil {
			return err
		}

		volume, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
			ID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("Bad: Check Volume Exists: %s", err)
		}

		if volume == nil {
			return fmt.Errorf("Bad: Volume %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckTritonVolumeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)
	c, err := conn.Compute()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "triton_volume" {
			continue
		}

		resp, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
			ID: rs.Primary.ID,
		})
		if err != nil {
			if errors.IsSpecificStatusCode(err, http.StatusNotFound) || errors.IsSpecificStatusCode(err, http.StatusGone) {
				return nil
			}
			return err
		}

		if resp != nil && resp.State != volumeStateDeleted {
			return fmt.Errorf("Bad: Volume %q still exists", rs.Primary.ID)
		}
	}

	return nil
}
