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
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTritonSnapshot_basic(t *testing.T) {
	snapshotName := fmt.Sprintf("acctest-snap-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonSnapshotConfig(t, snapshotName),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonSnapshotExists("triton_snapshot.test"),
					resource.TestCheckResourceAttr("triton_snapshot.test", "name", snapshotName),
					func(*terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
			{
				ResourceName:      "triton_snapshot.test",
				ImportState:       true,
				ImportStateIdFunc: testAccTritonSnapshotImportStateIdFunc("triton_snapshot.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckTritonSnapshotExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		conn := testAccProvider.Meta().(*Client)

		machineID, err := parseUUID(rs.Primary.Attributes["machine_id"])
		if err != nil {
			return fmt.Errorf("Bad: invalid machine_id: %s", err)
		}

		resp, err := conn.API().GetMachineSnapshotWithResponse(context.Background(), conn.Account(), machineID, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Bad: Check Snapshot Exists: %s", err)
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("Bad: Snapshot %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckTritonSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "triton_snapshot" {
			continue
		}

		machineID, err := parseUUID(rs.Primary.Attributes["machine_id"])
		if err != nil {
			return fmt.Errorf("invalid machine_id: %s", err)
		}

		resp, err := conn.API().ListMachineSnapshotsWithResponse(context.Background(), conn.Account(), machineID)
		if err != nil {
			return err
		}
		if resp.JSON200 != nil {
			for _, snap := range *resp.JSON200 {
				if snap.Name == rs.Primary.ID {
					return fmt.Errorf("Bad: Snapshot %q still exists", rs.Primary.ID)
				}
			}
		}
	}

	return nil
}

func testAccTritonSnapshotConfig(t *testing.T, snapshotName string) string {
	var packageName = testAccConfig(t, "test_package_name")

	return testAccTritonMachine_base(t, fmt.Sprintf(`
		resource "triton_machine" "test" {
		  image = "${data.triton_image.base.id}"
		  networks = [data.triton_network.test.id]

		  package = "%s"
		}

		resource "triton_snapshot" "test" {
		  name = "%s"
		  machine_id = "${triton_machine.test.id}"
		}
	`, packageName, snapshotName))
}

func TestAccTritonSnapshot_noStateDrift(t *testing.T) {
	snapshotName := fmt.Sprintf("acctest-snap-%d", acctest.RandInt())
	config := testAccTritonSnapshotConfig(t, snapshotName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonSnapshotExists("triton_snapshot.test"),
					resource.TestCheckResourceAttr(
						"triton_snapshot.test", "state", "created"),
					func(*terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func testAccTritonSnapshotImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s.%s", rs.Primary.Attributes["machine_id"], rs.Primary.ID), nil
	}
}
