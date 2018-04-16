package triton

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTritonFabricVLAN_MissingArguments(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTritonFabricVLANMissingArguments,
				ExpectError: regexp.MustCompile("one of `name`, `vlan_id`, or `description` must be assigned"),
			},
		},
	})
}

func TestAccTritonFabricVLAN_BadArguments(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTritonFabricVLANBadArguments,
				ExpectError: regexp.MustCompile(`.* \\"vlan_id\\" value must be between 0 and 4095`),
			},
		},
	})
}

func TestAccTritonFabricVLAN_Basic(t *testing.T) {
	vlanID := acctest.RandIntRange(3, 2048)
	resourcesOnly, config := testAccTritonFabricVLANBasic(vlanID)

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
						"data.triton_fabric_vlan.test",
						"name",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_vlan.test",
						"vlan_id",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_vlan.test",
						"description",
					),
					resource.TestCheckResourceAttr(
						"data.triton_fabric_vlan.test",
						"vlan_id",
						fmt.Sprintf("%d", vlanID),
					),
				),
			},
		},
	})
}

func TestAccTritonFabricVLAN_WildCard(t *testing.T) {
	vlanID := acctest.RandIntRange(3, 2048)
	resourcesOnly, config := testAccTritonFabricVLANWildCard(vlanID)

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
						"data.triton_fabric_vlan.test",
						"name",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_vlan.test",
						"vlan_id",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_vlan.test",
						"description",
					),
					resource.TestCheckResourceAttr(
						"data.triton_fabric_vlan.test",
						"vlan_id",
						fmt.Sprintf("%d", vlanID),
					),
				),
			},
		},
	})
}

func TestAccTritonFabricVLAN_Filters(t *testing.T) {
	vlanID := acctest.RandIntRange(3, 2048)
	resourcesOnly, config := testAccTritonFabricVLANFilters(vlanID)

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
						"data.triton_fabric_vlan.test",
						"name",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_vlan.test",
						"vlan_id",
					),
					resource.TestCheckResourceAttrSet(
						"data.triton_fabric_vlan.test",
						"description",
					),
					resource.TestCheckResourceAttr(
						"data.triton_fabric_vlan.test",
						"vlan_id",
						fmt.Sprintf("%d", vlanID),
					),
					resource.TestCheckResourceAttr(
						"data.triton_fabric_vlan.test",
						"description",
						fmt.Sprintf("Test Fabric VLAN %d", vlanID),
					),
				),
			},
		},
	})
}

func TestAccTritonFabricVLAN_NotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTritonFabricVLANNotFound,
				ExpectError: regexp.MustCompile(`unable to find any Fabric VLANs matching .* try again`),
			},
		},
	})
}

func TestAccTritonFabricVLAN_FiltersNotFound(t *testing.T) {
	vlanID := acctest.RandIntRange(3, 2048)
	resourcesOnly, config := testAccTritonFabricVLANFiltersNotFound(vlanID)

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: resourcesOnly,
			},
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`unable to find any Fabric VLANs matching .* try again`),
			},
		},
	})
}

func TestAccTritonFabricVLAN_MultipleFound(t *testing.T) {
	vlanID := acctest.RandIntRange(3, 2048)
	resourcesOnly, config := testAccTritonFabricVLANMultipleFound(vlanID)

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: resourcesOnly,
			},
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`found multiple Fabric VLANs matching .* try again`),
			},
		},
	})
}

var testAccTritonFabricVLANMissingArguments = `
  data "triton_fabric_vlan" "test" {}
`

var testAccTritonFabricVLANBadArguments = `
  data "triton_fabric_vlan" "test" {
    vlan_id = 12345
  }
`

var testAccTritonFabricVLANNotFound = `
  data "triton_fabric_vlan" "test" {
    name = "Bad-Fabric-VLAN-Name"
  }
`

var testAccTritonFabricVLANBasic = func(vlanID int) (string, string) {
	resources := fmt.Sprintf(`
  resource "triton_vlan" "test" {
    name        = "Test-Fabric-VLAN-%d"
    vlan_id     = %d
    description = "Test Fabric VLAN %d"
  }
`, vlanID, vlanID, vlanID)

	both := fmt.Sprintf(`%s
  data "triton_fabric_vlan" "test" {
    name = "Test-Fabric-VLAN-%d"
  }
`, resources, vlanID)

	return resources, both
}

var testAccTritonFabricVLANWildCard = func(vlanID int) (string, string) {
	resources := fmt.Sprintf(`
  resource "triton_vlan" "test" {
    name        = "Test-Fabric-VLAN-%d"
    vlan_id     = %d
    description = "Test Fabric VLAN %d"
  }
`, vlanID, vlanID, vlanID)

	both := fmt.Sprintf(`%s
  data "triton_fabric_vlan" "test" {
    name = "Tes?-*-VLA?-%d"
  }
`, resources, vlanID)

	return resources, both
}

var testAccTritonFabricVLANFilters = func(vlanID int) (string, string) {
	resources := fmt.Sprintf(`
  resource "triton_vlan" "test_1" {
    name        = "Test-Fabric-VLAN-%d"
    vlan_id     = %d
    description = "Test Fabric VLAN %d"
  }

  resource "triton_vlan" "test_2" {
    name    = "Test-Fabric-VLAN-%d"
    vlan_id = %d
  }
`, vlanID, vlanID, vlanID, vlanID, vlanID+1)

	both := fmt.Sprintf(`%s
  data "triton_fabric_vlan" "test" {
    name        = "Tes?-*-VLA?-*"
    description = "Test * %d"
  }
`, resources, vlanID)

	return resources, both
}

var testAccTritonFabricVLANFiltersNotFound = func(vlanID int) (string, string) {
	resources := fmt.Sprintf(`
  resource "triton_vlan" "test_1" {
    name    = "Test-Fabric-VLAN-%d"
    vlan_id = %d
  }

  resource "triton_vlan" "test_2" {
    name    = "Test-Fabric-VLAN-%d"
    vlan_id = %d
  }
`, vlanID, vlanID, vlanID, vlanID+1)

	both := fmt.Sprintf(`%s
  data "triton_fabric_vlan" "test" {
    name    = "Bad-Fabric-VLAN-Name"
    vlan_id = %d
  }
`, resources, vlanID)

	return resources, both
}

var testAccTritonFabricVLANMultipleFound = func(vlanID int) (string, string) {
	resources := fmt.Sprintf(`
  resource "triton_vlan" "test_1" {
    name    = "Test-Fabric-VLAN-%d"
    vlan_id = %d
  }

  resource "triton_vlan" "test_2" {
    name    = "Test-Fabric-VLAN-%d"
    vlan_id = %d
  }
`, vlanID, vlanID, vlanID, vlanID+1)

	both := fmt.Sprintf(`%s
  data "triton_fabric_vlan" "test" {
    name = "Test-Fabric-VLAN-%d"
  }
`, resources, vlanID)

	return resources, both
}
