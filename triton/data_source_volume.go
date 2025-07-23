package triton

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/TritonDataCenter/triton-go/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceVolume() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVolumeRead,
		Schema: map[string]*schema.Schema{
			"filesystem_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"networks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},

			"type": {
				Description: "Type of volume",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceVolumeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	input := &compute.ListVolumesInput{}
	if name, hasName := d.GetOk("name"); hasName {
		input.Name = name.(string)
	}
	if state, hasState := d.GetOk("state"); hasState {
		input.State = state.(string)
	}
	if size, hasSize := d.GetOk("size"); hasSize {
		input.Size = strconv.Itoa(size.(int))
	}

	volumes, err := c.Volumes().List(context.Background(), input)
	if err != nil {
		return err
	}

	if len(volumes) == 0 {
		return fmt.Errorf("Your query returned no results. Please change " +
			"your search criteria and try again.")
	}

	if len(volumes) > 1 {
		log.Printf("[DEBUG] triton_volume - %d results found", len(volumes))
		return fmt.Errorf("Your query returned more than one result. " +
			"Please try a more specific search criteria.")
	}

	var volume *compute.Volume = volumes[0]

	return tritonVolumeToTerraformVolume(d, volume)
}
