package triton

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/errors"
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
	log.Printf("[DEBUG] Found %d instances", len(instances))

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
	config := testAccTritonMachine_basic(t, machineName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttrSet("triton_machine.test", "compute_node"),
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
	config := testAccTritonMachine_affinity(t, machineName)

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
	dns_output := testAccTritonMachine_dns(t, machineName)

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
	config := testAccTritonMachine_singleNIC(t, machineName, acctest.RandIntRange(1024, 2048), acctest.RandIntRange(0, 256))

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

	singleNICConfig := testAccTritonMachine_singleNIC(t, machineName, vlanNumber, subnetNumber)
	dualNICConfig := testAccTritonMachine_dualNIC(t, machineName, vlanNumber, subnetNumber)
	publicNetworkConfigAndDualNIC := testAccTritonMachine_multipleNIC(t, machineName, vlanNumber, subnetNumber)

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
			if errors.IsSpecificStatusCode(err, http.StatusNotFound) || errors.IsSpecificStatusCode(err, http.StatusGone) {
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
	disabled_config := testAccTritonMachine_firewall(t, machineName, "firewall_enabled = false")
	enabled_config := testAccTritonMachine_firewall(t, machineName, "firewall_enabled = true")

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
	// testAccTriton_metadata(t, <name>, <outer prepend>, <machine append>)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonMachine_metadata(t, machineName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
				),
			},
			{
				Config: testAccTritonMachine_metadata(t, machineName, "", `
					user_data = "hello"
				`),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"user_data", "hello"),
				),
			},
			{
				Config: testAccTritonMachine_metadata(t, machineName, `
					variable "tags" {
  					default = {
							test = "hello!"
		  			}
					}
				`, `
					user_data = "hello"
					tags = "${var.tags}"
				`),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"tags.test", "hello!"),
				),
			},
			{
				Config: testAccTritonMachine_metadata(t, machineName, "",
					`
		  		user_data = "hello"
		  		tags = {
						test = "hello!"
		  		}
				`),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"tags.test", "hello!"),
				),
			},
			{
				Config: testAccTritonMachine_metadata(t, machineName, "", `
				  user_data = "hello"

				  tags = {
						test = "hello!"
		  		}

		  		metadata = {
						custom_meta = "hello-again"
		  		}
		  	`),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test",
						"metadata.custom_meta", "hello-again"),
				),
			},
			{
				Config: testAccTritonMachine_metadata(t, machineName, "", `
					user_data = "hello"

		  		tags = {
						test = "hello!"
		  		}

		  		metadata = {
						custom_meta = "hello-two"
		  		}
		  	`),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{

			// add cns service frontend
			{
				Config: testAccTritonMachine_cns(t, machineName, `
					cns {
						services = ["frontend"]
					}
				`),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "cns.0.services.0", "frontend"),
				),
			},

			// add cns service frontend and web
			{
				Config: testAccTritonMachine_cns(t, machineName, `
					cns {
						services = ["frontend", "web"]
					}
				`),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "cns.0.services.1", "web"),
				),
			},

			// add cns disable
			{
				Config: testAccTritonMachine_cns(t, machineName, `
					cns {
						disable = true
						services = ["frontend", "web"]
					}
				`),
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
	locality_fixture_1 := testAccTritonMachine_locality_1(t, machineName)

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

func TestAccTritonMachine_deletionProtection(t *testing.T) {
	machineName := fmt.Sprintf("acctest-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckTritonMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonMachine_deletionProtection(t, machineName, ""),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "deletion_protection_enabled", "false"),
				),
			},
			{
				Config: testAccTritonMachine_deletionProtection(t, machineName, "deletion_protection_enabled = true"),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "deletion_protection_enabled", "true"),
				),
			},
			{
				Config: testAccTritonMachine_deletionProtection(t, machineName, "deletion_protection_enabled = false"),
				Check: resource.ComposeTestCheckFunc(
					testCheckTritonMachineExists("triton_machine.test"),
					resource.TestCheckResourceAttr(
						"triton_machine.test", "deletion_protection_enabled", "false"),
				),
			},
		},
	})
}

var testAccTritonMachine_base = func(t *testing.T, append string) string {
	var networkName = testAccConfig(t, "test_network_name")

	return fmt.Sprintf(`
		data "triton_network" "test" {
			name = "%s"
		}
		data "triton_image" "base" {
			name = "base-64-lts"
			version = "16.4.1"
			most_recent = true
		}

		%s
	`, networkName, append)
}

var testAccTritonMachine_singleMachine = func(t *testing.T, machineName string, machineAppend string) string {
	var packageName = testAccConfig(t, "test_package_name")

	return testAccTritonMachine_base(t, fmt.Sprintf(`
		resource "triton_machine" "test" {
		  name = "%s"
		  package = "%s"
		  image = "${data.triton_image.base.id}"

		  networks = [data.triton_network.test.id]

		  %s
		}
	`, machineName, packageName, machineAppend))
}

// all of these tests are just single machines with some additional config string injected, so share the same fixture
var testAccTritonMachine_deletionProtection = testAccTritonMachine_singleMachine
var testAccTritonMachine_firewall = testAccTritonMachine_singleMachine
var testAccTritonMachine_cns = testAccTritonMachine_singleMachine

// a "Basic" is just a singleMachine with no additional config
var testAccTritonMachine_basic = func(t *testing.T, machineName string) string {
	return testAccTritonMachine_singleMachine(t, machineName, "")
}

// metadata config let's us add a string to the _outer_ config, as well as inside the machine
var testAccTritonMachine_metadata = func(t *testing.T, machineName string, outerConfig string, machineAppend string) string {
	var machineConfig = testAccTritonMachine_singleMachine(t, machineName, machineAppend)
	return fmt.Sprintf("%s\n%s", outerConfig, machineConfig)
}

var testAccTritonMachine_affinity = func(t *testing.T, machinePrefix string) string {
	var packageName = testAccConfig(t, "test_package_name")

	return testAccTritonMachine_base(t, fmt.Sprintf(`
		resource "triton_machine" "test-1" {
		  name = "%s-1"
		  package = "%s"
		  image = "${data.triton_image.base.id}"
			
			networks = [data.triton_network.test.id]
		  
		  tags = {
			service = "one"
		  }
		}

		resource "triton_machine" "test-2" {
		  name = "%s-2"
		  package = "%s"
		  image = "${data.triton_image.base.id}"

		  affinity = ["service!=one"]

			networks = [data.triton_network.test.id]

		  tags = {
			service = "two"
		  }
		}
	`, machinePrefix, packageName, machinePrefix, packageName))
}

var testAccTritonMachine_locality_1 = func(t *testing.T, machinePrefix string) string {
	var packageName = testAccConfig(t, "test_package_name")

	return testAccTritonMachine_base(t, fmt.Sprintf(`
		resource "triton_machine" "test1" {
		  name = "%s-1"
		  package = "%s"
		  image = "${data.triton_image.base.id}"

		  networks = [data.triton_network.test.id]
		}

		resource "triton_machine" "test2" {
		  name = "%s-2"
		  package = "%s"
		  image = "${data.triton_image.base.id}"

		  networks = [data.triton_network.test.id]
		}

		resource "triton_machine" "test3" {
		  name = "%s-3"
		  package = "%s"
		  image = "${data.triton_image.base.id}"

			networks = [data.triton_network.test.id]

		  locality {
			far_from = ["${triton_machine.test1.id}"]
			close_to = ["${triton_machine.test2.id}"]
		  }
		}
	`, machinePrefix, packageName, machinePrefix, packageName, machinePrefix, packageName))
}

var testAccTritonMachine_singleNIC = func(t *testing.T, name string, vlanNumber int, subnetNumber int) string {
	var packageName = testAccConfig(t, "test_package_name")

	return testAccTritonMachine_base(t, fmt.Sprintf(`
		resource "triton_vlan" "test" {
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
			package = "%s"
			image = "${data.triton_image.base.id}"

			tags = {
				test = "Test"
			}

			networks = ["${data.triton_network.test.id}"]
		}
	`, vlanNumber, name, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name, packageName))
}

var testAccTritonMachine_multipleNIC = func(t *testing.T, name string, vlanNumber, subnetNumber int) string {
	var packageName = testAccConfig(t, "test_package_name")

	return testAccTritonMachine_base(t, fmt.Sprintf(`
		resource "triton_vlan" "test" {
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
			package = "%s"
			image = "${data.triton_image.base.id}"

			tags = {
				test = "Test"
			}

			networks = ["${data.triton_network.test.id}", "${triton_fabric.test.id}", "${triton_fabric.test_add.id}"]
		}
	`, vlanNumber, name, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name, packageName))
}

var testAccTritonMachine_dualNIC = func(t *testing.T, name string, vlanNumber, subnetNumber int) string {
	var packageName = testAccConfig(t, "test_package_name")

	return testAccTritonMachine_base(t, fmt.Sprintf(`
		resource "triton_vlan" "test" {
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
			package = "%s"
			image = "${data.triton_image.base.id}"

			tags = {
				test = "Test"
			}

			networks = ["${data.triton_network.test.id}", "${triton_fabric.test.id}"]
		}
	`, vlanNumber, name, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name, subnetNumber, subnetNumber, subnetNumber, subnetNumber, name, packageName))
}

var testAccTritonMachine_dns = func(t *testing.T, name string) string {
	var packageName = testAccConfig(t, "test_package_name")

	return testAccTritonMachine_base(t, fmt.Sprintf(`
		resource "triton_machine" "test" {
		  name = "%s"
		  package = "%s"
		  image = "${data.triton_image.base.id}"

		  networks = ["${data.triton_network.test.id}"]
		}

		output "domain_names" {
		  value = "${join(", ", triton_machine.test.domain_names)}"
		}
	`, name, packageName))
}
