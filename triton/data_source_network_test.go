package triton

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/joyent/triton-go/network"
)

func TestAccTritonNetwork_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonNetwork_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonNetworkDataSourceID("data.triton_network.base", "Joyent-SDC-Public"),
				),
			},
		},
	})
}

func testAccCheckTritonNetworkDataSourceID(name, networkName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find Network data source: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Network data source ID not set")
		}

		conn := testAccProvider.Meta().(*Client)
		net, err := conn.Network()
		if err != nil {
			return err
		}

		networks, err := net.List(context.Background(), &network.ListInput{})
		if err != nil {
			return err
		}
		var network *network.Network
		for _, found := range networks {
			if found.Id == rs.Primary.ID {
				network = found
			}
		}
		if network.Name != networkName {
			return fmt.Errorf("Bad ID for data source %q: expected %q, got %q",
				name, network.Id, rs.Primary.ID)
		}
		return nil
	}
}

var testAccTritonNetwork_basic = `
data "triton_network" "base" {
	name = "SDC-Public"
}
`
