package triton

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTritonAccount(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonAccountBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.triton_account.main", "id"),
					resource.TestCheckResourceAttrSet("data.triton_account.main", "login"),
					resource.TestCheckResourceAttrSet("data.triton_account.main", "email"),
					resource.TestCheckResourceAttrSet("data.triton_account.main", "cns_enabled"),
				),
			},
		},
	})
}

var testAccTritonAccountBasic = `
data "triton_account" "main" {}
`
