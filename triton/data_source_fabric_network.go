package triton

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/network"
	"github.com/pkg/errors"
)

// dataSourceFabricNetwork returns schema for the Fabric Network data source.
func dataSourceFabricNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFabricNetworkRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"fabric": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provision_start_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provision_end_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resolvers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"routes": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"internet_nat": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"vlan_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateVLANIdentifier,
			},
		},
	}
}

// dataSourceFabricNetworkRead retrieves details about all the Fabric Networks
// from a specific VLAN in the current Data Center from the Fabrics API, then
// searches for a matching Fabric Network in the list of available networks
// using network name as a filter.
func dataSourceFabricNetworkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	fabricName := d.Get("name").(string)
	vlanID := d.Get("vlan_id").(int)

	net, err := client.Network()
	if err != nil {
		return errors.Wrap(err, "error creating Network client")
	}

	log.Printf("[DEBUG] triton_fabric_network: Reading Fabric Network details on VLAN %d", vlanID)
	fabrics, err := net.Fabrics().List(context.Background(), &network.ListFabricsInput{
		FabricVLANID: vlanID,
	})
	if err != nil {
		return errors.Wrap(err, "error retrieving Fabric Network details")
	}

	var result *network.Network
	for _, fabric := range fabrics {
		if fabric.Fabric && fabric.Name == fabricName {
			log.Printf("[DEBUG] triton_fabric_network: Found matching Fabric Network: %+v", fabric)
			result = fabric
			break
		}
	}
	if result == nil {
		return fmt.Errorf("unable to find any Fabric Network with name %q "+
			"on the VLAN %d, please change your search criteria "+
			"and try again", fabricName, vlanID)
	}

	d.SetId(result.Id)
	d.Set("name", result.Name)
	d.Set("public", result.Public)
	d.Set("fabric", result.Fabric)
	d.Set("description", result.Description)
	d.Set("subnet", result.Subnet)
	d.Set("provision_start_ip", result.ProvisioningStartIP)
	d.Set("provision_end_ip", result.ProvisioningEndIP)
	d.Set("gateway", result.Gateway)
	d.Set("resolvers", result.Resolvers)
	d.Set("routes", result.Routes)
	d.Set("internet_nat", result.InternetNAT)
	d.Set("vlan_id", vlanID) // The VLAN ID is not part of the `network.Network` type.

	return nil
}
