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
	"regexp"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	volumeStateCreating = "creating"
	volumeStateDeleted  = "deleted"
	volumeStateDeleting = "deleting"
	volumeStateFailed   = "failed"
	volumeStateReady    = "ready"
)

func resourceVolume() *schema.Resource {
	return &schema.Resource{
		Create:   resourceVolumeCreate,
		Exists:   resourceVolumeExists,
		Read:     resourceVolumeRead,
		Update:   resourceVolumeUpdate,
		Delete:   resourceVolumeDelete,
		Timeouts: slowResourceTimeout,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Friendly name for volume",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: resourceVolumeValidateName,
			},
			"networks": {
				Description: "Desired network IDs",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"size": {
				Description: "The size of the volume (Mb)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"tags": {
				Description: "Volume tags",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"type": {
				Description: "Type of volume",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "tritonnfs",
			},

			// Volume computed parameters
			"filesystem_path": {
				Description: "NFS mounting path",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"owner": {
				Description: "Who owns the volume",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the volume",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	body := cloudapi.CreateVolumeJSONRequestBody{}

	if v, ok := d.GetOk("size"); ok {
		sz := uint64(v.(int))
		body.Size = &sz
	}

	if value, ok := d.GetOk("name"); ok {
		body.Name = ptrString(value.(string))
	}

	if value, ok := d.GetOk("type"); ok {
		vt := cloudapi.VolumeType{}
		vt.FromVolumeType0(cloudapi.VolumeType0(value.(string)))
		body.Type = &vt
	}

	if value, ok := d.GetOk("networks"); ok {
		var networks []openapi_types.UUID
		for _, n := range value.([]interface{}) {
			id, err := parseUUID(n.(string))
			if err != nil {
				return fmt.Errorf("invalid network UUID: %s", err)
			}
			networks = append(networks, id)
		}
		body.Networks = &networks
	}

	if value, ok := d.GetOk("tags"); ok {
		tags := cloudapi.Tags{}
		for k, v := range value.(map[string]interface{}) {
			tags[k] = v
		}
		body.Tags = &tags
	}

	resp, err := client.API().CreateVolumeWithResponse(context.Background(), client.Account(), body)
	if err != nil {
		return fmt.Errorf("error creating volume: %s", err)
	}
	if resp.JSON201 == nil {
		return fmt.Errorf("error creating volume: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	d.SetId(uuidString(resp.JSON201.ID))

	stateConf := &retry.StateChangeConf{
		Target: []string{volumeStateReady},
		Refresh: func() (interface{}, string, error) {
			volID, _ := parseUUID(d.Id())
			r, err := client.API().GetVolumeWithResponse(context.Background(), client.Account(), volID)
			if err != nil {
				return nil, "", err
			}
			if r.JSON200 == nil {
				return nil, "", fmt.Errorf("error polling volume: %s", formatAPIError(r.StatusCode(), r.Body))
			}
			if string(r.JSON200.State) == volumeStateFailed {
				d.SetId("")
				return nil, "", fmt.Errorf("volume creation failed: %s", r.JSON200.State)
			}
			return r.JSON200, string(r.JSON200.State), nil
		},
		Timeout:    *slowResourceTimeout.Create,
		MinTimeout: defaultPollInterval,
	}
	v, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	vol := v.(*cloudapi.Volume)
	return cloudapiVolumeToTerraform(d, vol)
}

func resourceVolumeExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)

	volID, err := parseUUID(d.Id())
	if err != nil {
		return false, fmt.Errorf("invalid volume ID: %s", err)
	}

	resp, err := client.API().GetVolumeWithResponse(context.Background(), client.Account(), volID)
	if err != nil {
		return false, err
	}
	if isNotFound(resp.StatusCode()) {
		return false, nil
	}
	if resp.JSON200 == nil {
		return false, fmt.Errorf("error checking volume: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return true, nil
}

func cloudapiVolumeToTerraform(d *schema.ResourceData, volume *cloudapi.Volume) error {
	d.SetId(uuidString(volume.ID))

	d.Set("filesystem_path", derefString(volume.FilesystemPath))
	d.Set("name", volume.Name)
	d.Set("networks", uuidPtrSliceToStrings(volume.Networks))
	d.Set("size", int(volume.Size))
	d.Set("state", string(volume.State))
	d.Set("type", volumeTypeString(&volume.Type))

	if volume.Tags != nil {
		d.Set("tags", *volume.Tags)
	}

	return nil
}

func volumeTypeString(vt *cloudapi.VolumeType) string {
	if vt == nil {
		return ""
	}
	v, err := vt.AsVolumeType0()
	if err != nil {
		v1, err2 := vt.AsVolumeType1()
		if err2 != nil {
			return ""
		}
		return string(v1)
	}
	return string(v)
}

func resourceVolumeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	volID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid volume ID: %s", err)
	}

	resp, err := client.API().GetVolumeWithResponse(context.Background(), client.Account(), volID)
	if err != nil {
		return err
	}

	if isNotFound(resp.StatusCode()) {
		log.Printf("Volume %q not found or has been deleted", d.Id())
		d.SetId("")
		return nil
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error reading volume: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	if string(resp.JSON200.State) == volumeStateFailed {
		log.Printf("Volume %q state: `failed` so removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return cloudapiVolumeToTerraform(d, resp.JSON200)
}

func resourceVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		volID, err := parseUUID(d.Id())
		if err != nil {
			return fmt.Errorf("invalid volume ID: %s", err)
		}

		oldNameInterface, newNameInterface := d.GetChange("name")
		oldName := oldNameInterface.(string)
		newName := newNameInterface.(string)

		if err := client.Typed().UpdateVolume(context.Background(), client.Account(), volID,
			cloudapi.UpdateVolumeRequest{Name: &newName}); err != nil {
			return fmt.Errorf("error updating volume: %s", err)
		}

		stateConf := &retry.StateChangeConf{
			Pending: []string{oldName},
			Target:  []string{newName},
			Refresh: func() (interface{}, string, error) {
				r, err := client.API().GetVolumeWithResponse(context.Background(), client.Account(), volID)
				if err != nil {
					return nil, "", err
				}
				if r.JSON200 == nil {
					return nil, "", fmt.Errorf("error polling volume: %s", formatAPIError(r.StatusCode(), r.Body))
				}
				return r.JSON200, r.JSON200.Name, nil
			},
			Timeout:    *slowResourceTimeout.Update,
			MinTimeout: defaultPollInterval,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	d.Partial(false)

	return nil
}

func resourceVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	volID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid volume ID: %s", err)
	}

	resp, err := client.API().DeleteVolumeWithResponse(context.Background(), client.Account(), volID)
	if err != nil {
		return fmt.Errorf("error deleting volume: %s", err)
	}
	if isNotFound(resp.StatusCode()) {
		return nil
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("error deleting volume: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	stateConf := &retry.StateChangeConf{
		Target: []string{volumeStateDeleted},
		Refresh: func() (interface{}, string, error) {
			r, err := client.API().GetVolumeWithResponse(context.Background(), client.Account(), volID)
			if err != nil {
				return nil, "", err
			}
			if isNotFound(r.StatusCode()) {
				return volumeStateDeleted, volumeStateDeleted, nil
			}
			if r.JSON200 == nil {
				return nil, "", fmt.Errorf("error polling volume: %s", formatAPIError(r.StatusCode(), r.Body))
			}
			return r.JSON200, string(r.JSON200.State), nil
		},
		Timeout:    *slowResourceTimeout.Delete,
		MinTimeout: defaultPollInterval,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return nil
}

func resourceVolumeValidateName(value interface{}, name string) (warnings []string, errors []error) {
	warnings = []string{}
	errors = []error{}

	r := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\_\.\-]+$`)
	if !r.Match([]byte(value.(string))) {
		errors = append(errors, fmt.Errorf(`"%s" is not a valid %s`, value.(string), name))
	}

	return warnings, errors
}
