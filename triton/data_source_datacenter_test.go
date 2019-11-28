package triton

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTritonDataCenter(t *testing.T) {
	url := testAccConfig(t, "URL")
	config := testAccTritonDataCenter(url)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.triton_datacenter.current", "name", testAccConfig(t, "dc_name")),
					resource.TestCheckResourceAttr("data.triton_datacenter.current", "endpoint", url),
				),
			},
		},
	})
}

var testAccTritonDataCenter = func(url string) string {
	return fmt.Sprintf(`
		provider "triton" {
		  url = "%s"
		}

		data "triton_datacenter" "current" {}
	`, url)
}
