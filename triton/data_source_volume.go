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

	resp, err := client.API().ListVolumesWithResponse(context.Background(), client.Account())
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error listing volumes: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	allVolumes := *resp.JSON200

	// Client-side filtering
	filterName, hasName := d.GetOk("name")
	filterState, hasState := d.GetOk("state")
	filterSize, hasSize := d.GetOk("size")

	var filtered []cloudapi.Volume
	for _, v := range allVolumes {
		if hasName && v.Name != filterName.(string) {
			continue
		}
		if hasState && string(v.State) != filterState.(string) {
			continue
		}
		if hasSize && int(v.Size) != filterSize.(int) {
			continue
		}
		filtered = append(filtered, v)
	}

	if len(filtered) == 0 {
		return fmt.Errorf("your query returned no results, please change " +
			"your search criteria and try again")
	}

	if len(filtered) > 1 {
		log.Printf("[DEBUG] triton_volume - %d results found", len(filtered))
		return fmt.Errorf("your query returned more than one result, " +
			"please try a more specific search criteria")
	}

	volume := filtered[0]

	return cloudapiVolumeToTerraform(d, &volume)
}
