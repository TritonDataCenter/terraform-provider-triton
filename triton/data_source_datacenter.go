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
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	log.Printf("[DEBUG] triton_datacenter: Reading Data Center details.")
	resp, err := client.API().ListDatacentersWithResponse(context.Background(), client.Account())
	if err != nil {
		return fmt.Errorf("error retrieving Data Center details: %s", err)
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error retrieving Data Center details: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	tritonURL := client.URL()
	for name, url := range *resp.JSON200 {
		if url == tritonURL {
			log.Printf("[DEBUG] triton_datacenter: Found matching Data Center: %s -> %s", name, url)
			d.SetId(time.Now().UTC().String())
			d.Set("name", name)
			d.Set("endpoint", url)
			break
		}
	}

	return nil
}
