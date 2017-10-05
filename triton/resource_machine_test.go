package triton

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/joyent/triton-go/compute"
)

func init() {
	resource.AddTestSweepers("triton_machine", &resource.Sweeper{
		Name: "triton_machine",
		F:    testSweepMachines,
	})

}

func testSweepMachines(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*Client)
	a, err := client.Compute()
	if err != nil {
		return err
	}

	instances, err := a.Instances().List(context.Background(), &compute.ListInstancesInput{})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Found %d instances to sweep", len(instances))

	for _, v := range instances {
		if strings.HasPrefix(v.Name, "acctest-") {
			log.Printf("Destroying instance %s", v.Name)

			if err := a.Instances().Delete(context.Background(), &compute.DeleteInstanceInput{
				ID: v.ID,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccTritonMachine_basic(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	config := fmt.Sprintf(testAccTritonMachine_basic, machineName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					func(*terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
		},
	})
}

func TestAccTritonMachine_affinity(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	config := fmt.Sprintf(testAccTritonMachine_affinity, machineName, machineName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test-1"),
					testCheckTritonMachineExists("triton_machine.test-2"),
					func(*terraform.State) error {
						time.Sleep(10 * time.Second)
						return nil
					},
				),
			},
		},
	})
}

func TestAccTritonMachine_dns(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	dns_output := fmt.Sprintf(testAccTritonMachine_dns, machineName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: dns_output,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					func(state *terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
					resource.TestMatchOutput("domain_names", regexp.MustCompile(".*acctest-.*")),
				),
			},
		},
	})
}

func TestAccTritonMachine_nic(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	config := testAccTritonMachine_singleNIC(machineName, acctest.RandIntRange(1024, 2048), acctest.RandIntRange(0, 256))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					func(*terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
					resource.TestCheckResourceAttr("triton_machine.test", "networks.#", "1"),
				),
			},
		},
	})
}

func TestAccTritonMachine_addNIC(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	vlanNumber := acctest.RandIntRange(1024, 2048)
	subnetNumber := acctest.RandIntRange(0, 256)

	singleNICConfig := testAccTritonMachine_singleNIC(machineName, vlanNumber, subnetNumber)
	dualNICConfig := testAccTritonMachine_dualNIC(machineName, vlanNumber, subnetNumber)
	publicNetworkConfigAndDualNIC := testAccTritonMachine_multipleNIC(machineName, vlanNumber, subnetNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: singleNICConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr("triton_machine.test", "networks.#", "1"),
					func(*terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
			{
				Config: dualNICConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr("triton_machine.test", "networks.#", "2"),
					func(*terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
			{
				Config: publicNetworkConfigAndDualNIC,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr("triton_machine.test", "networks.#", "3"),
					func(*terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
		},
	})
}

func testCheckTritonMachineExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		conn := testAccProvider.Meta().(*Client)
		c, err := conn.Compute()
		if err != nil {
			return err
		}

		instance, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
			ID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("Bad: Check Machine Exists: %s", err)
		}

		if instance == nil {
			return fmt.Errorf("Bad: Machine %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckTritonMachineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Client)
	c, err := conn.Compute()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "triton_machine" {
			continue
		}

		resp, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
			ID: rs.Primary.ID,
		})
		if err != nil {
			if compute.IsResourceNotFound(err) {
				return nil
			}
			return err
		}

		if resp != nil && resp.State != machineStateDeleted {
			return fmt.Errorf("Bad: Machine %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func TestAccTritonMachine_firewall(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	disabled_config := fmt.Sprintf(testAccTritonMachine_firewall_0, machineName)
	enabled_config := fmt.Sprintf(testAccTritonMachine_firewall_1, machineName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: enabled_config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "firewall_enabled", "true"),
				),
			},
			{
				Config: disabled_config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "firewall_enabled", "false"),
				),
			},
			{
				Config: enabled_config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "firewall_enabled", "true"),
				),
			},
		},
	})
}

func TestAccTritonMachine_metadata(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	basic := fmt.Sprintf(testAccTritonMachine_metadata_1, machineName)
	add_metadata := fmt.Sprintf(testAccTritonMachine_metadata_1, machineName)
	add_metadata_2 := fmt.Sprintf(testAccTritonMachine_metadata_2, machineName)
	add_metadata_3 := fmt.Sprintf(testAccTritonMachine_metadata_3, machineName)
	add_metadata_4 := fmt.Sprintf(testAccTritonMachine_metadata_4, machineName)
	add_metadata_5 := fmt.Sprintf(testAccTritonMachine_metadata_5, machineName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: basic,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
				),
			},
			{
				Config: add_metadata,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"user_data", "hello"),
				),
			},
			{
				Config: add_metadata_2,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"tags.test", "hello!"),
				),
			},
			{
				Config: add_metadata_3,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"tags.test", "hello!"),
				),
			},
			{
				Config: add_metadata_4,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"metadata.custom_meta", "hello-again"),
				),
			},
			{
				Config: add_metadata_5,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"metadata.custom_meta", "hello-two"),
				),
			},
		},
	})
}

func TestAccTritonMachine_cns(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	// add cns service frontend
	cns_fixture_1 := fmt.Sprintf(testAccTritonMachine_cns_1, machineName)
	// add cns service frontend and web
	cns_fixture_2 := fmt.Sprintf(testAccTritonMachine_cns_2, machineName)
	// add cns disable
	cns_fixture_3 := fmt.Sprintf(testAccTritonMachine_cns_3, machineName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: cns_fixture_1,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "cns.0.services.0", "frontend"),
				),
			},
			{
				Config: cns_fixture_2,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "cns.0.services.1", "web"),
				),
			},
			{
				Config: cns_fixture_3,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "cns.0.disable", "true"),
				),
			},
		},
	})
}

func TestAccTritonMachine_locality(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())
	locality_fixture_1 := fmt.Sprintf(testAccTritonMachine_locality_1, machineName, machineName, machineName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: locality_fixture_1,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test3"),
					resource.TestCheckResourceAttrSet(
						"triton_machine.test3", "locality.0.far_from.0"),
					resource.TestCheckResourceAttrSet(
						"triton_machine.test3", "locality.0.close_to.0"),
				),
			},
		},
	})
}

var testAccTritonMachine_basic = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-general-4G"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  tags = {
	test = "hello!"
  }
}
`

var testAccTritonMachine_affinity = `
resource "triton_machine" "test-1" {
  name = "%s-1"
  package = "g4-general-4G"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  tags = {
	service = "one"
  }
}

resource "triton_machine" "test-2" {
  name = "%s-2"
  package = "g4-general-4G"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  affinity = ["service!=one"]

  tags = {
	service = "two"
  }
}
`

var testAccTritonMachine_firewall_0 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-general-4G"
  image = "c20b4b7c-e1a6-11e5-9a4d-ef590901732e"

	firewall_enabled = 0
}
`
var testAccTritonMachine_firewall_1 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-general-4G"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

	firewall_enabled = 1
}
`

var testAccTritonMachine_metadata_1 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-general-4G"
  image = "c20b4b7c-e1a6-11e5-9a4d-ef590901732e"

  user_data = "hello"

  tags {
	test = "hello!"
	}
}
`
var testAccTritonMachine_metadata_2 = `
variable "tags" {
  default = {
	test = "hello!"
  }
}
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  user_data = "hello"

  tags = "${var.tags}"
}
`
var testAccTritonMachine_metadata_3 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  user_data = "hello"

  tags {
	test = "hello!"
  }
}
`
var testAccTritonMachine_metadata_4 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  user_data = "hello"

  tags {
	test = "hello!"
  }

  metadata {
	custom_meta = "hello-again"
  }
}
`
var testAccTritonMachine_metadata_5 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  user_data = "hello"

  tags {
	test = "hello!"
  }

  metadata {
	custom_meta = "hello-two"
  }
}
`
var testAccTritonMachine_cns_1 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  cns {
	services = ["frontend"]
  }
}
`
var testAccTritonMachine_cns_2 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  cns {
	services = ["frontend", "web"]
  }
}
`
var testAccTritonMachine_cns_3 = `
resource "triton_machine" "test" {
  name = "%s"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  cns {
	disable = true
	services = ["frontend", "web"]
  }
}
`

var testAccTritonMachine_locality_1 = `
resource "triton_machine" "test1" {
  name = "%s-1"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"
}

resource "triton_machine" "test2" {
  name = "%s-2"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"
}

resource "triton_machine" "test3" {
  name = "%s-3"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

  locality {
	far_from = ["${triton_machine.test1.id}"]
	close_to = ["${triton_machine.test2.id}"]
  }
}
`

var testAccTritonMachine_singleNIC = func(name string, vlanNumber int, subnetNumber int) string {
	return fmt.Sprintf(`resource "triton_vlan" "test" {
	  vlan_id = %d
	  name = "%s-vlan"
	  description = "test vlan"
}

resource "triton_fabric" "test" {
	name = "%s-network"
	description = "test network"
	vlan_id = "${triton_vlan.test.vlan_id}"

	subnet = "10.%d.0.0/24"
	gateway = "10.%d.0.1"
	provision_start_ip = "10.%d.0.10"
	provision_end_ip = "10.%d.0.250"

	resolvers = ["8.8.8.8", "8.8.4.4"]
}

resource "triton_machine" "test" {
	name = "%s-instance"
	package = "g4-highcpu-128M"
	image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

	tags = {
		test = "Test"
	}

	networks = ["${triton_fabric.test.id}"]
}`, vlanNumber, name, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name)
}

var testAccTritonMachine_multipleNIC = func(name string, vlanNumber, subnetNumber int) string {
	return fmt.Sprintf(`resource "triton_vlan" "test" {
	  vlan_id = %d
	  name = "%s-vlan"
	  description = "test vlan"
}

data "triton_network" "public" {
    name = "Joyent-SDC-Public"
}

resource "triton_fabric" "test" {
	name = "%s-network"
	description = "test network"
	vlan_id = "${triton_vlan.test.vlan_id}"

	subnet = "10.%d.0.0/24"
	gateway = "10.%d.0.1"
	provision_start_ip = "10.%d.0.10"
	provision_end_ip = "10.%d.0.250"

	resolvers = ["8.8.8.8", "8.8.4.4"]
}

resource "triton_fabric" "test_add" {
	name = "%s-network-2"
	description = "test network 2"
	vlan_id = "${triton_vlan.test.vlan_id}"

	subnet = "172.23.%d.0/24"
	gateway = "172.23.%d.1"
	provision_start_ip = "172.23.%d.10"
	provision_end_ip = "172.23.%d.250"

	resolvers = ["8.8.8.8", "8.8.4.4"]
}

resource "triton_machine" "test" {
	name = "%s-instance"
	package = "g4-highcpu-128M"
	image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

	tags = {
		test = "Test"
	}

	networks = ["${triton_fabric.test.id}", "${triton_fabric.test_add.id}", "${data.triton_network.public.id}"]
}`, vlanNumber, name, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name)
}

var testAccTritonMachine_dualNIC = func(name string, vlanNumber, subnetNumber int) string {
	return fmt.Sprintf(`resource "triton_vlan" "test" {
	  vlan_id = %d
	  name = "%s-vlan"
	  description = "test vlan"
}

resource "triton_fabric" "test" {
	name = "%s-network"
	description = "test network"
	vlan_id = "${triton_vlan.test.vlan_id}"

	subnet = "10.%d.0.0/24"
	gateway = "10.%d.0.1"
	provision_start_ip = "10.%d.0.10"
	provision_end_ip = "10.%d.0.250"

	resolvers = ["8.8.8.8", "8.8.4.4"]
}

resource "triton_fabric" "test_add" {
	name = "%s-network-2"
	description = "test network 2"
	vlan_id = "${triton_vlan.test.vlan_id}"

	subnet = "172.23.%d.0/24"
	gateway = "172.23.%d.1"
	provision_start_ip = "172.23.%d.10"
	provision_end_ip = "172.23.%d.250"

	resolvers = ["8.8.8.8", "8.8.4.4"]
}

resource "triton_machine" "test" {
	name = "%s-instance"
	package = "g4-highcpu-128M"
	image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"

	tags = {
		test = "Test"
	}

	networks = ["${triton_fabric.test.id}", "${triton_fabric.test_add.id}"]
}`, vlanNumber, name, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name)
}

var testAccTritonMachine_dns = `
provider "triton" {
}

resource "triton_machine" "test" {
  name = "%s"
  package = "g4-highcpu-128M"
  image = "fb5fe970-e6e4-11e6-9820-4b51be190db9"
}

output "domain_names" {
  value = "${join(", ", triton_machine.test.domain_names)}"
}
`
