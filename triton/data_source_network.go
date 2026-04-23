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

	log.Printf("[DEBUG] triton_network: Reading Network details.")
	resp, err := client.API().ListNetworksWithResponse(context.Background(), client.Account())
	if err != nil {
		return fmt.Errorf("error retrieving Network details: %s", err)
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error retrieving Network details: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	networkName := d.Get("name").(string)

	var result *cloudapi.Network
	for i := range *resp.JSON200 {
		n := &(*resp.JSON200)[i]
		if n.Name == networkName {
			log.Printf("[DEBUG] triton_network: Found matching Network: %+v", n)
			result = n
			break
		}
	}
	if result == nil {
		return fmt.Errorf("no matching Network with name %q found", networkName)
	}

	d.SetId(uuidString(result.ID))
	d.Set("public", result.Public)
	d.Set("fabric", derefBool(result.Fabric))

	return nil
}
