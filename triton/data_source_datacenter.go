package triton

import (
	"context"
	"log"
	"time"

	"github.com/TritonDataCenter/triton-go/compute"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// dataSourceDataCenter returns schema for the Data Center data source.
func dataSourceDataCenter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDataCenterRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Data Center.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"endpoint": {
				Description: "The endpoint URL of the Data Center.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// dataSourceDataCenterRead retrieves a list of all data centers from Triton
// using the Data Center API. The current Data Center endpoint URL will be
// the same as the one currently configured in the Triton provider.
func dataSourceDataCenterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	c, err := client.Compute()
	if err != nil {
		return errors.Wrap(err, "error creating Compute client")
	}

	log.Printf("[DEBUG] triton_datacenter: Reading Data Center details.")
	dcs, err := c.Datacenters().List(context.Background(), &compute.ListDataCentersInput{})
	if err != nil {
		return errors.Wrap(err, "error retrieving Data Center details")
	}

	tritonURL := client.config.TritonURL
	for _, dc := range dcs {
		if dc.URL == tritonURL {
			log.Printf("[DEBUG] triton_datacenter: Found matching Data Center: %+v", dc)
			d.SetId(time.Now().UTC().String())
			d.Set("name", dc.Name)
			d.Set("endpoint", dc.URL)
			break
		}
	}

	return nil
}
