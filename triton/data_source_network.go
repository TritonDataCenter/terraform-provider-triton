package triton

import (
	"context"
	"fmt"
	"strings"

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

	var netName string
	if name, hasName := d.GetOk("name"); hasName {
		netName = name.(string)
	}

	var network *network.Network
	for _, found := range networks {
		if strings.Contains(found.Name, netName) {
			network = found
		}
	}

	d.SetId(network.Id)
	return nil
}
