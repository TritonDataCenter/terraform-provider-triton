package triton

import (
	"context"
	"fmt"
	"log"

	"github.com/TritonDataCenter/triton-go/network"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// dataSourceNetwork returns schema for the Network data source.
func dataSourceNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetworkRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Network.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"public": {
				Description: "Whether this Network is a public or private [RFC1918](https://tools.ietf.org/html/rfc1918) network.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"fabric": {
				Description: "Whether this Network is created on a [Fabric](https://docs.tritondatacenter.com/public-cloud/network/sdn).",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

// dataSourceNetworkRead retrieves details about all the networks which
// can be used by the given account from the Networks API, then searches
// for a matching network in the list of available networks using network
// name as a filter.
func dataSourceNetworkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	net, err := client.Network()
	if err != nil {
		return errors.Wrap(err, "error creating Network client")
	}

	log.Printf("[DEBUG] triton_network: Reading Network details.")
	networks, err := net.List(context.Background(), &network.ListInput{})
	if err != nil {
		return errors.Wrap(err, "error retrieving Network details")
	}

	networkName := d.Get("name").(string)

	var result *network.Network
	for _, network := range networks {
		if network.Name == networkName {
			log.Printf("[DEBUG] triton_network: Found matching Network: %+v", network)
			result = network
			break
		}
	}
	if result == nil {
		return fmt.Errorf("no matching Network with name %q found", networkName)
	}

	d.SetId(result.Id)
	d.Set("public", result.Public)
	d.Set("fabric", result.Fabric)

	return nil
}
