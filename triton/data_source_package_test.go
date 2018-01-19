package triton

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const (
	testPackageName = "g4-highcpu-128M"
)

func TestAccTritonPackage_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonPackage_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonPackageDataSourceID("data.triton_package.base", testPackageName),
				),
			},
		},
	})
}

func testAccCheckTritonPackageDataSourceID(name, packageName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find package data source: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("returned package ID should not be empty")
		}
		if rs.Primary.Attributes["name"] != packageName {
			return fmt.Errorf("returned package Name does not match")
		}

		return nil
	}
}

var testAccTritonPackage_basic = `
data "triton_package" "base" {
	filter {
	   name = "highcpu"
	   memory = 128
	}
}
`
