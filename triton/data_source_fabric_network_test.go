package triton

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTritonFabricNetwork_MissingArguments(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTritonFabricNetworkMissingArguments,
				ExpectError: regexp.MustCompile(`.* \\"name\\": .* \\"vlan_id\\": .* field is not set.*`),
			},
		},
	})
}

func TestAccTritonFabricNetwork_BadArguments(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTritonFabricNetworkBadArguments,
				ExpectError: regexp.MustCompile(`.* \\"vlan_id\\" value must be between 0 and 4095`),
			},
		},
	})
}

func TestAccTritonFabricNetwork_NotFound(t *testing.T) {
	vlanID := acctest.RandIntRange(3, 2048)
	resourcesOnly, config := testAccTritonFabricNetworkNotFound(vlanID)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: resourcesOnly,
			},
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`unable to find .* "Bad-Fabric-Network-Name" .* try again`),
			},
		},
	})
}

func TestAccTritonFabricNetwork_Basic(t *testing.T) {
	vlanID := acctest.RandIntRange(3, 2048)
	resourcesOnly, config := testAccTritonFabricNetworkBasic(vlanID)

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: resourcesOnly,
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_network.test",
						"name",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_network.test",
						"subnet",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_network.test",
						"provision_start_ip",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_network.test",
						"provision_end_ip",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_network.test",
						"gateway",
					),
					resource.TestCheckResourceAttr(
						"data.triton_fabric_network.test",
						"resolvers.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.triton_fabric_network.test",
						"vlan_id",
						fmt.Sprintf("%d", vlanID),
					),
					resource.TestCheckResourceAttrPair(
						"data.triton_fabric_network.test",
						"vlan_id",
						"triton_vlan.test",
						"id",
					),
				),
			},
		},
	})
}

var testAccTritonFabricNetworkMissingArguments = `
  data "triton_fabric_network" "test" {}
`

var testAccTritonFabricNetworkBadArguments = `
  data "triton_fabric_network" "test" {
    name    = "Test-Fabric-Network"
    vlan_id = 12345
  }
`

var testAccTritonFabricNetworkNotFound = func(vlanID int) (string, string) {
	resources := fmt.Sprintf(`
  resource "triton_vlan" "test" {
    name    = "Test-Fabric-VLAN-%d"
    vlan_id = %d
  }
`, vlanID, vlanID)

	both := fmt.Sprintf(`%s
  data "triton_fabric_network" "test" {
    name    = "Bad-Fabric-Network-Name"
    vlan_id = %d
  }
`, resources, vlanID)

	return resources, both
}

var testAccTritonFabricNetworkBasic = func(vlanID int) (string, string) {
	resources := fmt.Sprintf(`
  resource "triton_vlan" "test" {
    name    = "Test-Fabric-VLAN-%d"
    vlan_id = %d
  }

  resource "triton_fabric" "test" {
    name = "Test-Fabric-Network-%d"

    subnet             = "10.0.0.0/24"
    provision_start_ip = "10.0.0.2"
    provision_end_ip   = "10.0.0.254"
    gateway            = "10.0.0.1"

    resolvers = [
      "8.8.8.8",
      "8.8.4.4",
    ]

    vlan_id = "${triton_vlan.test.id}"
  }
`, vlanID, vlanID, vlanID)

	both := fmt.Sprintf(`%s
  data "triton_fabric_network" "test" { 
    name    = "Test-Fabric-Network-%d"
    vlan_id = %d
  }
`, resources, vlanID, vlanID)

	return resources, both
}
