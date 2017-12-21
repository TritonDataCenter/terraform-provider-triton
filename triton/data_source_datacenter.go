package triton

import (
	"context"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/compute"
)

func dataSourceDataCenter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDataCenterRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDataCenterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	dcs, err := c.Datacenters().List(context.Background(), &compute.ListDataCentersInput{})
	if err != nil {
		return err
	}

	for _, dc := range dcs {
		if dc.URL == client.config.TritonURL {
			d.SetId(time.Now().UTC().String())
			d.Set("name", dc.Name)
			d.Set("endpoint", dc.URL)
		}
	}

	return nil
}
