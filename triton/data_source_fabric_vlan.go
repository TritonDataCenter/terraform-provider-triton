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
	"errors"
	"log"
	"time"

	"context"
	"fmt"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// filterVLANFunc is a function that is called to filter a Fabric VLAN from
// a slice of Fabric VLANs based on a predicate.
type filterVLANFunc func(*cloudapi.FabricVlan) bool

// dataSourceFabricVLAN returns schema for the Fabric VLAN data source.
func dataSourceFabricVLAN() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFabricVLANRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vlan_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateVLANIdentifier,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

// dataSourceFabricVLANRead retrieves details about all the Fabric VLANs
// which are available in the current Data Center from the Fabrics API,
// then searches for a matching Fabric VLAN using either name, VLAN ID
// or description as filter, or a combination of thereof.
func dataSourceFabricVLANRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	vlanName, vlanNameOk := d.GetOk("name")
	vlanID, vlanIDOk := d.GetOk("vlan_id")
	vlanDesc, vlanDescOk := d.GetOk("description")

	if !vlanNameOk && !vlanIDOk && !vlanDescOk {
		return errors.New("one of `name`, `vlan_id`, or `description` must be assigned")
	}

	log.Printf("[DEBUG] triton_fabric_vlan: Reading Fabric VLAN details.")
	resp, err := client.API().ListFabricVlansWithResponse(context.Background(), client.Account())
	if err != nil {
		return fmt.Errorf("error retrieving Fabric VLAN details: %s", err)
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error retrieving Fabric VLAN details: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	vlans := *resp.JSON200
	matches := vlans

	if vlanIDOk {
		matches = filterVLANs(matches, func(v *cloudapi.FabricVlan) bool {
			return int(v.VlanID) == vlanID.(int)
		})
	}
	if vlanNameOk {
		matches = filterVLANs(matches, func(v *cloudapi.FabricVlan) bool {
			return wildcardMatch(vlanName.(string), v.Name)
		})
	}
	if vlanDescOk {
		matches = filterVLANs(matches, func(v *cloudapi.FabricVlan) bool {
			return wildcardMatch(vlanDesc.(string), derefString(v.Description))
		})
	}

	if len(matches) == 0 {
		return errors.New("unable to find any Fabric VLANs matching the " +
			"current search criteria, please change your search criteria " +
			"and try again")
	}

	if len(matches) > 1 {
		log.Printf("[DEBUG] triton_fabric_vlan: Found multiple matching Fabric VLANs: %+v", matches)
		return errors.New("found multiple Fabric VLANs matching the " +
			"current search criteria, please change your search criteria " +
			"and try again")
	}

	vlan := matches[0]

	log.Printf("[DEBUG] triton_fabric_vlan: Found matching Fabric VLAN: %+v", vlan)
	d.SetId(time.Now().UTC().String())

	d.Set("name", vlan.Name)
	d.Set("vlan_id", int(vlan.VlanID))
	d.Set("description", derefString(vlan.Description))

	return nil
}

// filterVLANs iterates over a slice of Fabric VLANs, and returns a slice that
// contains all of the Fabric VLANs the predicate returns a value of true for.
func filterVLANs(vlans []cloudapi.FabricVlan, f filterVLANFunc) (results []cloudapi.FabricVlan) {
	for i := range vlans {
		if f(&vlans[i]) {
			results = append(results, vlans[i])
		}
	}
	return
}
