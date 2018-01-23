package triton

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/joyent/triton-go/compute"
	terrors "github.com/joyent/triton-go/errors"
)

func TestAccTritonSnapshot_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonSnapshotConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonSnapshotExists("triton_snapshot.test"),
					func(*terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
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
		c, err := conn.Compute()
		if err != nil {
			return err
		}

		snapshot, err := c.Snapshots().Get(context.Background(), &compute.GetSnapshotInput{
			Name:      rs.Primary.ID,
			MachineID: rs.Primary.Attributes["machine_id"],
		})
		if err != nil {
			return fmt.Errorf("Bad: Check Snapshot Exists: %s", err)
		}

		if snapshot == nil {
			return fmt.Errorf("Bad: Snapshot %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckTritonSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)
	c, err := conn.Compute()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "triton_snapshot" {
			continue
		}

		resp, err := c.Snapshots().Get(context.Background(), &compute.GetSnapshotInput{
			Name:      rs.Primary.ID,
			MachineID: rs.Primary.Attributes["machine_id"],
		})
		if err != nil {
			if terrors.IsResourceNotFound(err) {
				return nil
			}
			return err
		}

		if resp != nil && resp.State != "deleted" {
			return fmt.Errorf("Bad: Snapshot %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTritonSnapshotConfig(rInt int) string {
	return fmt.Sprintf(`
data "triton_image" "ubuntu1604" {
  name = "ubuntu-16.04"
  version = "20170403"
}

resource "triton_machine" "test" {
  image = "${data.triton_image.ubuntu1604.id}"
  package = "g4-highcpu-128M"
}

resource "triton_snapshot" "test" {
  name = "acctest-snap-%d"
  machine_id = "${triton_machine.test.id}"
}`, rInt)
}
