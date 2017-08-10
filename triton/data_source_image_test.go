package triton

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const base64LTS = "1f32508c-e6e9-11e6-bc05-8fea9e979940"

func TestAccTritonImage_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonImage_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonImageDataSourceID("data.triton_image.base", base64LTS),
				),
			},
		},
	})
}

func TestAccTritonImage_mostRecent(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonImage_mostRecent,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonImageDataSourceID("data.triton_image.base", base64LTS),
				),
			},
		},
	})
}

func testAccCheckTritonImageDataSourceID(name, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find Image data source: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Image data source ID not set")
		}

		if rs.Primary.ID != id {
			return fmt.Errorf("Bad ID for data source %q: expected %q, got %q",
				name, id, rs.Primary.ID)
		}
		return nil
	}
}

var testAccTritonImage_basic = `
data "triton_image" "base" {
	name = "base-64-lts"
	version = "16.4.1"
}
`

var testAccTritonImage_mostRecent = `
data "triton_image" "base" {
	name = "base-64-lts"
	most_recent = true
}
`
