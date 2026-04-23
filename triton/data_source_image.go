/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2021 Joyent, Inc.
 * Copyright 2022 MNX Cloud, Inc.
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"context"
	"fmt"
	"log"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
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

func mostRecentImages(images []cloudapi.Image) cloudapi.Image {
	return sortImages(images)[0]
}

func dataSourceImageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	params := &cloudapi.ListImagesParams{}
	if name, hasName := d.GetOk("name"); hasName {
		params.Name = ptrString(name.(string))
	}
	if os, hasOS := d.GetOk("os"); hasOS {
		params.Os = ptrString(os.(string))
	}
	if version, hasVersion := d.GetOk("version"); hasVersion {
		params.Version = ptrString(version.(string))
	}
	if public, hasPublic := d.GetOk("public"); hasPublic {
		params.Public = ptrBool(public.(bool))
	}
	if state, hasState := d.GetOk("state"); hasState {
		s := cloudapi.ImageState{}
		s.FromImageState0(cloudapi.ImageState0(state.(string)))
		params.State = &s
	}
	if owner, hasOwner := d.GetOk("owner"); hasOwner {
		ownerUUID, err := parseUUID(owner.(string))
		if err != nil {
			return fmt.Errorf("invalid owner UUID: %s", err)
		}
		params.Owner = &ownerUUID
	}
	if imageType, hasImageType := d.GetOk("type"); hasImageType {
		t := cloudapi.ImageType{}
		t.FromImageType0(cloudapi.ImageType0(imageType.(string)))
		params.Type = &t
	}

	resp, err := client.API().ListImagesWithResponse(context.Background(), client.Account(), params)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error listing images: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	images := *resp.JSON200

	if len(images) == 0 {
		return fmt.Errorf("your query returned no results, please change " +
			"your search criteria and try again")
	}

	var image cloudapi.Image
	if len(images) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] triton_image - multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			image = mostRecentImages(images)
		} else {
			return fmt.Errorf("your query returned more than one result, " +
				"please try a more specific search criteria")
		}
	} else {
		image = images[0]
	}

	d.SetId(uuidString(image.ID))
	return nil
}
