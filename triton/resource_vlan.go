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
	"errors"
	"fmt"
	"strconv"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVLAN() *schema.Resource {
	return &schema.Resource{
		Create:   resourceVLANCreate,
		Exists:   resourceVLANExists,
		Read:     resourceVLANRead,
		Update:   resourceVLANUpdate,
		Delete:   resourceVLANDelete,
		Timeouts: fastResourceTimeout,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vlan_id": {
				Description: "Number between 0-4095 indicating VLAN ID",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeInt,
				ValidateFunc: func(val interface{}, field string) (warn []string, err []error) {
					value := val.(int)
					if value < 0 || value > 4095 {
						err = append(err, errors.New("vlan_id must be between 0 and 4095"))
					}
					return
				},
			},
			"name": {
				Description: "Unique name to identify VLAN",
				Required:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description: "Description of the VLAN",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceVLANCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	body := cloudapi.CreateFabricVlanJSONRequestBody{
		VlanID: uint16(d.Get("vlan_id").(int)),
		Name:   d.Get("name").(string),
	}
	if v, ok := d.GetOk("description"); ok {
		body.Description = ptrString(v.(string))
	}
	resp, err := client.API().CreateFabricVlanWithResponse(context.Background(), client.Account(), body)
	if err != nil {
		return fmt.Errorf("error creating VLAN: %s", err)
	}
	if resp.JSON201 == nil {
		return fmt.Errorf("error creating VLAN: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	d.SetId(strconv.Itoa(int(resp.JSON201.VlanID)))
	return resourceVLANRead(d, meta)
}

func resourceVLANExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)

	id, err := resourceVLANIDUint16(d.Id())
	if err != nil {
		return false, err
	}

	resp, err := client.API().GetFabricVlanWithResponse(context.Background(), client.Account(), id)
	if err != nil {
		return false, err
	}
	if isNotFound(resp.StatusCode()) {
		return false, nil
	}
	if resp.JSON200 == nil {
		return false, fmt.Errorf("error checking VLAN: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return true, nil
}

func resourceVLANRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	id, err := resourceVLANIDUint16(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.API().GetFabricVlanWithResponse(context.Background(), client.Account(), id)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error reading VLAN: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	vlan := resp.JSON200
	d.Set("vlan_id", int(vlan.VlanID))
	d.Set("name", vlan.Name)
	d.Set("description", derefString(vlan.Description))

	return nil
}

func resourceVLANUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	id := uint16(d.Get("vlan_id").(int))
	name := d.Get("name").(string)
	body := cloudapi.UpdateFabricVlanJSONRequestBody{
		Name: &name,
	}
	if v, ok := d.GetOk("description"); ok {
		body.Description = ptrString(v.(string))
	}
	resp, err := client.API().UpdateFabricVlanWithResponse(context.Background(), client.Account(), id, body)
	if err != nil {
		return fmt.Errorf("error updating VLAN: %s", err)
	}
	if resp.JSON202 == nil {
		return fmt.Errorf("error updating VLAN: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	d.SetId(strconv.Itoa(int(resp.JSON202.VlanID)))
	return resourceVLANRead(d, meta)
}

func resourceVLANDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	id, err := resourceVLANIDUint16(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.API().DeleteFabricVlanWithResponse(context.Background(), client.Account(), id)
	if err != nil {
		return fmt.Errorf("error deleting VLAN: %s", err)
	}
	if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
		return fmt.Errorf("error deleting VLAN: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return nil
}

func resourceVLANIDUint16(id string) (uint16, error) {
	result, err := strconv.ParseUint(id, 10, 16)
	if err != nil {
		return 0, err
	}

	return uint16(result), nil
}
