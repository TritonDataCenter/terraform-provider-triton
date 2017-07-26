package triton

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/network"
)

func dataSourceNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetworkRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func dataSourceNetworkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	net, err := client.Network()
	if err != nil {
		return err
	}

	networks, err := net.List(context.Background(), &network.ListInput{})
	if err != nil {
		return err
	}
	if len(networks) == 0 {
		return fmt.Errorf("Your query returned no results. Please change " +
			"your search criteria and try again.")
	}

	var networkName string
	if name, hasName := d.GetOk("name"); hasName {
		networkName = name.(string)
	}

	var network *network.Network
	for _, found := range networks {
		if found.Name == networkName {
			network = found
		}
	}
	if network == nil {
		return fmt.Errorf("No Networks found by name %q", networkName)
	}

	d.SetId(network.Id)
	return nil
}
