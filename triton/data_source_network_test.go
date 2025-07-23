package triton

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/TritonDataCenter/triton-go/network"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTritonNetwork_Basic(t *testing.T) {
	publicNetwork := testAccConfig(t, "public_network_name")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonNetworkBasic(publicNetwork),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.triton_network.main", "id"),
					resource.TestCheckResourceAttrSet("data.triton_network.main", "name"),
					resource.TestCheckResourceAttrSet("data.triton_network.main", "public"),
					resource.TestCheckResourceAttrSet("data.triton_network.main", "fabric"),
				),
			},
			{
				Config: testAccTritonNetworkBasic(publicNetwork),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonNetworkDataSourceID("data.triton_network.main", publicNetwork),
				),
			},
		},
	})
}

func TestAccTritonNetwork_NotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTritonNetworkNotFound,
				ExpectError: regexp.MustCompile(`no matching Network with name "Bad-Network-Name" found`),
			},
		},
	})
}

func testAccCheckTritonNetworkDataSourceID(name, networkName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*Client)

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("unable to find Network data source: %s", name)
		}
		if rs.Primary.ID == "" {
			return errors.New("no Network data source ID is set")
		}

		net, err := conn.Network()
		if err != nil {
			return err
		}

		networks, err := net.List(context.Background(), &network.ListInput{})
		if err != nil {
			return err
		}

		var result *network.Network
		for _, network := range networks {
			if network.Id == rs.Primary.ID {
				result = network
				break
			}
		}

		if result.Name != networkName {
			return fmt.Errorf("incorrect Network ID for data source %q: expected %q, got %q",
				name, result.Id, rs.Primary.ID)
		}

		return nil
	}
}

var testAccTritonNetworkBasic = func(name string) string {
	return fmt.Sprintf(`
		data "triton_network" "main" {
  		name = "%s"
		}
	`, name)
}

var testAccTritonNetworkNotFound = `
data "triton_network" "main" {
  name = "Bad-Network-Name"
}
`
