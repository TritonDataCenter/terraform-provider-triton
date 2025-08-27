package triton

import (
	"context"
	"fmt"
	"log"

	"github.com/TritonDataCenter/triton-go/compute"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceImageRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the image.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"os": {
				Description: "The underlying operating system for the image.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"version": {
				Description: "The version for the image.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"public": {
				Description: "Whether to return public as well as private images",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},

			"state": {
				Description: "The state of the image. By default, only `active` images are shown. Must be one of: `active`, `unactivated`, `disabled`, `creating`, `failed` or `all`, though the default is sufficient in almost every case.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"owner": {
				Description: "The UUID of the account which owns the image.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"type": {
				Description: "The image type. Must be one of: `zone-dataset`, `lx-dataset`, `zvol`, `docker` or `other`.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"most_recent": {
				Description: "If more than one result is returned, use the most recent Image.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
		},
	}
}

func mostRecentImages(images []*compute.Image) *compute.Image {
	return sortImages(images)[0]
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

	var image *compute.Image
	if len(images) == 0 {
		return fmt.Errorf("Your query returned no results. Please change " +
			"your search criteria and try again.")
	}

	if len(images) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] triton_image - multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			image = mostRecentImages(images)
		} else {
			return fmt.Errorf("Your query returned more than one result. " +
				"Please try a more specific search criteria.")
		}
	} else {
		image = images[0]
	}

	d.SetId(image.ID)
	return nil
}
