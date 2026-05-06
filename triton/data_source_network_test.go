/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Joyent, Inc.
 * Copyright 2025 MNX Cloud, Inc.
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

		resp, err := conn.API().ListNetworksWithResponse(context.Background(), conn.Account())
		if err != nil {
			return err
		}
		if resp.JSON200 == nil {
			return fmt.Errorf("error listing networks: unexpected status %d", resp.StatusCode())
		}

		var resultID, resultName string
		for i := range *resp.JSON200 {
			n := &(*resp.JSON200)[i]
			if uuidString(n.ID) == rs.Primary.ID {
				resultID = uuidString(n.ID)
				resultName = n.Name
				break
			}
		}

		if resultName != networkName {
			return fmt.Errorf("incorrect Network ID for data source %q: expected %q, got %q",
				name, resultID, rs.Primary.ID)
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
