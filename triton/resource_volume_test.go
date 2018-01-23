package triton

import (
	"context"
	"fmt"
	"testing"

	"strings"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/joyent/triton-go/compute"
)

func TestAccTritonVolume_basic(t *testing.T) {
	rInt := acctest.RandInt()
	var volume compute.Volume

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonVolumeConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonVolumeExists("triton_volume.test", &volume),
					resource.TestCheckResourceAttr("triton_volume.test", "name", fmt.Sprintf("test-volume-%d", rInt)),
					resource.TestCheckResourceAttr("triton_volume.test", "networks.#", "1"),
					resource.TestCheckResourceAttr("triton_volume.test", "size", "10240"),
					resource.TestCheckResourceAttr("triton_volume.test", "type", "tritonnfs"),
					resource.TestCheckResourceAttrSet("triton_volume.test", "filesystem_path"),
					resource.TestCheckResourceAttrSet("triton_volume.test", "owner"),
				),
			},
		},
	})
}

func TestAccTritonVolume_updateName(t *testing.T) {
	before := acctest.RandInt()
	after := acctest.RandInt()
	var beforeVolume, afterVolume compute.Volume

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonVolumeConfig(before),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonVolumeExists("triton_volume.test", &beforeVolume),
					resource.TestCheckResourceAttr("triton_volume.test", "name", fmt.Sprintf("test-volume-%d", before)),
				),
			},
			{
				Config: testAccTritonVolumeConfig(after),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonVolumeExists("triton_volume.test", &afterVolume),
					resource.TestCheckResourceAttr("triton_volume.test", "name", fmt.Sprintf("test-volume-%d", after)),
					testCheckTritonVolumeNotRecreated(t, &beforeVolume, &afterVolume),
				),
			},
		},
	})
}

func TestAccTritonVolume_networkChanges(t *testing.T) {
	before := acctest.RandInt()
	after := acctest.RandInt()
	vlan := acctest.RandIntRange(3, 4095)
	var beforeVolume, afterVolume compute.Volume

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonVolumeConfig(before),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonVolumeExists("triton_volume.test", &beforeVolume),
					resource.TestCheckResourceAttr("triton_volume.test", "networks.#", "1"),
				),
			},
			{
				Config: testAccTritonVolumeConfigNetworkChange(after, vlan),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonVolumeExists("triton_volume.test", &afterVolume),
					resource.TestCheckResourceAttr("triton_volume.test", "networks.#", "2"),
					testCheckTritonVolumeRecreated(t, &beforeVolume, &afterVolume),
				),
			},
		},
	})
}

func testCheckTritonVolumeRecreated(t *testing.T,
	before, after *compute.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID == after.ID {
			t.Fatalf("Expected volume to be recreated, but both have ID of %s", before.ID)
		}
		return nil
	}
}

func testCheckTritonVolumeNotRecreated(t *testing.T,
	before, after *compute.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID != after.ID {
			t.Fatalf("Expected volume to be the same, but found IDs of before: %q, after: %q", before.ID, after.ID)
		}
		return nil
	}
}

func testCheckTritonVolumeExists(name string, volume *compute.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		conn := testAccProvider.Meta().(*Client)
		c, err := conn.Compute()
		if err != nil {
			return err
		}

		resp, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
			ID: rs.Primary.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "VolumeNotFound") {
				return fmt.Errorf("Bad: Check Volume Exists: %v", err)
			}
			return err
		}

		if resp == nil {
			return fmt.Errorf("Bad: Volume %q does not exist", rs.Primary.ID)
		}

		*volume = *resp

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
			if strings.Contains(err.Error(), "VolumeNotFound") {
				return nil
			}
			return err
		}

		if resp != nil {
			return fmt.Errorf("Bad: Volume %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTritonVolumeConfig(rInt int) string {
	return fmt.Sprintf(`
data "triton_network" "my_fabric" {
  name = "My-Fabric-Network"
}
resource "triton_volume" "test" {
  name = "test-volume-%d"
  networks = ["${data.triton_network.my_fabric.id}"]
}
`, rInt)
}

func testAccTritonVolumeConfigNetworkChange(rInt int, vlan int) string {
	return fmt.Sprintf(`
data "triton_network" "my_fabric" {
  name = "My-Fabric-Network"
}
resource "triton_volume" "test" {
  name = "test-volume-%d"
  networks = ["${triton_fabric.test.id}","${data.triton_network.my_fabric.id}"]
}

resource "triton_vlan" "test" {
  vlan_id = "%d"
  name = "my-vlan-%d"
  description = "testAccTritonFabric_basic"
}

resource "triton_fabric" "test" {
  name = "my-fabric-%d"
  description = "test network"
  vlan_id = "${triton_vlan.test.id}"

  subnet = "172.23.52.0/24"
  gateway = "172.23.52.1"
  provision_start_ip = "172.23.52.10"
  provision_end_ip = "172.23.52.250"

  resolvers = ["8.8.8.8", "8.8.4.4"]
}
`, rInt, vlan, rInt, rInt)
}
