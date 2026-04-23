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

// dataSourceFabricNetwork returns schema for the Fabric Network data source.
func dataSourceFabricNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFabricNetworkRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"fabric": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provision_start_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provision_end_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resolvers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"routes": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"internet_nat": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"vlan_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateVLANIdentifier,
			},
		},
	}
}

// dataSourceFabricNetworkRead retrieves details about all the Fabric Networks
// from a specific VLAN in the current Data Center from the Fabrics API, then
// searches for a matching Fabric Network in the list of available networks
// using network name as a filter.
func dataSourceFabricNetworkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	fabricName := d.Get("name").(string)
	vlanID := uint16(d.Get("vlan_id").(int))

	log.Printf("[DEBUG] triton_fabric_network: Reading Fabric Network details on VLAN %d", vlanID)
	resp, err := client.API().ListFabricNetworksWithResponse(context.Background(), client.Account(), vlanID)
	if err != nil {
		return fmt.Errorf("error retrieving Fabric Network details: %s", err)
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error retrieving Fabric Network details: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	var result *cloudapi.Network
	for i := range *resp.JSON200 {
		fabric := &(*resp.JSON200)[i]
		if derefBool(fabric.Fabric) && fabric.Name == fabricName {
			log.Printf("[DEBUG] triton_fabric_network: Found matching Fabric Network: %+v", fabric)
			result = fabric
			break
		}
	}
	if result == nil {
		return fmt.Errorf("unable to find any Fabric Network with name %q "+
			"on the VLAN %d, please change your search criteria "+
			"and try again", fabricName, vlanID)
	}

	d.SetId(uuidString(result.ID))
	d.Set("name", result.Name)
	d.Set("public", result.Public)
	d.Set("fabric", derefBool(result.Fabric))
	d.Set("description", derefString(result.Description))
	d.Set("subnet", derefString(result.Subnet))
	d.Set("provision_start_ip", derefString(result.ProvisionStartIP))
	d.Set("provision_end_ip", derefString(result.ProvisionEndIP))
	d.Set("gateway", derefString(result.Gateway))
	d.Set("resolvers", derefStringSlice(result.Resolvers))
	d.Set("routes", result.Routes)
	d.Set("internet_nat", derefBool(result.InternetNat))
	d.Set("vlan_id", int(vlanID))

	return nil
}
