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

	// Build server-side filter params.
	params := &cloudapi.ListVolumesParams{}
	if v, ok := d.GetOk("name"); ok {
		s := v.(string)
		params.Name = &s
	}
	if v, ok := d.GetOk("state"); ok {
		s := v.(string)
		params.State = &s
	}
	if v, ok := d.GetOk("size"); ok {
		sz := uint64(v.(int))
		params.Size = &sz
	}

	resp, err := client.API().ListVolumesWithResponse(context.Background(), client.Account(), params)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error listing volumes: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	volumes := *resp.JSON200

	if len(volumes) == 0 {
		return fmt.Errorf("your query returned no results, please change " +
			"your search criteria and try again")
	}

	if len(volumes) > 1 {
		log.Printf("[DEBUG] triton_volume - %d results found", len(volumes))
		return fmt.Errorf("your query returned more than one result, " +
			"please try a more specific search criteria")
	}

	volume := volumes[0]

	return cloudapiVolumeToTerraform(d, &volume)
}
