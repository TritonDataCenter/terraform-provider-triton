package triton

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/compute"
)

func dataSourceImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceImageRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"os": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"public": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"state": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"owner": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func dataSourceImageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	input := &compute.ListImagesInput{}
	if name, hasName := d.GetOk("name"); hasName {
		input.Name = name.(string)
	}
	if os, hasOS := d.GetOk("os"); hasOS {
		input.OS = os.(string)
	}
	if version, hasVersion := d.GetOk("version"); hasVersion {
		input.Version = version.(string)
	}
	if public, hasPublic := d.GetOk("public"); hasPublic {
		input.Public = public.(bool)
	}
	if state, hasState := d.GetOk("state"); hasState {
		input.State = state.(string)
	}
	if owner, hasOwner := d.GetOk("owner"); hasOwner {
		input.Owner = owner.(string)
	}
	if imageType, hasImageType := d.GetOk("type"); hasImageType {
		input.Type = imageType.(string)
	}

	images, err := c.Images().List(context.Background(), input)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		return fmt.Errorf("Your query returned no results. Please change " +
			"your search criteria and try again.")
	}

	if len(images) > 1 {
		return fmt.Errorf("Your query returned more than one result. " +
			"Please try a more specific search criteria.")
	}

	d.SetId(images[0].ID)
	return nil
}
