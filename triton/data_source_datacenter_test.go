package triton

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTritonDataCenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonDataCenter,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.triton_datacenter.current", "name", "us-sw-1"),
					resource.TestCheckResourceAttr("data.triton_datacenter.current", "endpoint", "https://us-sw-1.api.joyentcloud.com"),
				),
			},
		},
	})
}

func TestAccTritonDataCenterOldURL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonDataCenterOldURL,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.triton_datacenter.current", "name", "us-east-1"),
					resource.TestCheckResourceAttr("data.triton_datacenter.current", "endpoint", "https://us-east-1.api.joyentcloud.com"),
				),
			},
		},
	})
}

var testAccTritonDataCenter = `
provider "triton" {
  url = "https://us-sw-1.api.joyent.com"
}

data "triton_datacenter" "current" {}
`

var testAccTritonDataCenterOldURL = `
provider "triton" {
  url = "https://us-east-1.api.joyentcloud.com"
}

data "triton_datacenter" "current" {}
`
