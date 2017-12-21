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
				Config: testAccTritonAccount_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.triton_account.main", "id"),
					resource.TestCheckResourceAttrSet("data.triton_account.main", "cns_enabled"),
				),
			},
		},
	})
}

var testAccTritonAccount_basic = `
data "triton_account" "main" {}
`
