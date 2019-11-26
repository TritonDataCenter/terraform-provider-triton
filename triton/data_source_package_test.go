package triton

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTritonPackage_basic(t *testing.T) {
	testPackageResultName := testAccConfig(t, "package_query_result")
	testPackageQueryName := testAccConfig(t, "package_query_name")
	testPackageQueryMemory := testAccConfig(t, "package_query_memory")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonPackage_basic(testPackageQueryName, testPackageQueryMemory),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonPackageDataSourceID("data.triton_package.base", testPackageResultName),
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

var testAccTritonPackage_basic = func(query string, memory string) (string) {
	return fmt.Sprintf(`
		data "triton_package" "base" {
			filter {
	   		name = "%s"
	   		memory = %s
			}
		}
		`, query, memory)
}
