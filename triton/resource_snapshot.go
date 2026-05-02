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
	"strings"
	"time"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	snapshotCreateTimeout = 30 * time.Minute
)

func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceSnapshotCreate,
		Read:   resourceSnapshotRead,
		Delete: resourceSnapshotDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				machineId, snapshotName, err := resourceSnapshotParseIds(d.Id())

				if err != nil {
					return nil, err
				}

				d.Set("machine_id", machineId)
				d.SetId(snapshotName)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name for the snapshot.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"machine_id": {
				Description: "The ID of the machine of which to take a snapshot.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"state": {
				Description: "The current state of the snapshot.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func snapshotStateString(s cloudapi.SnapshotState) string {
	v, err := s.AsSnapshotState0()
	if err != nil {
		// Fall back to the raw union for unknown states.
		v1, err2 := s.AsSnapshotState1()
		if err2 != nil {
			log.Printf("[WARN] snapshotStateString: failed to decode both union branches (state0: %s, state1: %s)", err, err2)
			return "unknown"
		}
		return string(v1)
	}
	return string(v)
}

func resourceSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	machineID, err := parseUUID(d.Get("machine_id").(string))
	if err != nil {
		return fmt.Errorf("invalid machine_id: %s", err)
	}

	snapshotName := d.Get("name").(string)
	resp, err := client.API().CreateMachineSnapshotWithResponse(context.Background(), client.Account(), machineID,
		cloudapi.CreateMachineSnapshotJSONRequestBody{
			Name: &snapshotName,
		})
	if err != nil {
		return fmt.Errorf("error creating snapshot: %s", err)
	}
	if resp.JSON201 == nil {
		return fmt.Errorf("error creating snapshot: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	d.SetId(resp.JSON201.Name)

	// Poll via ListMachineSnapshots instead of GetMachineSnapshot
	// because the GET endpoint can return stale data (e.g. "deleted"
	// state from a previous snapshot) while the list endpoint is
	// authoritative.
	stateConf := &retry.StateChangeConf{
		Pending: []string{"queued", "creating"},
		Target:  []string{"created"},
		Refresh: func() (interface{}, string, error) {
			r, err := client.API().ListMachineSnapshotsWithResponse(context.Background(), client.Account(), machineID)
			if err != nil {
				return nil, "", err
			}
			if r.JSON200 == nil {
				return nil, "", fmt.Errorf("error polling snapshot: %s", formatAPIError(r.StatusCode(), r.Body))
			}
			for _, snap := range *r.JSON200 {
				if snap.Name == d.Id() {
					state := snapshotStateString(snap.State)
					if state == "failed" {
						return nil, "", fmt.Errorf("snapshot entered terminal state %q during creation", state)
					}
					return &snap, state, nil
				}
			}
			// Snapshot not yet visible in the list; treat as queued.
			return nil, "queued", nil
		},
		Timeout:    snapshotCreateTimeout,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceSnapshotRead(d, meta)
}

func resourceSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	machineID, err := parseUUID(d.Get("machine_id").(string))
	if err != nil {
		return fmt.Errorf("invalid machine_id: %s", err)
	}

	resp, err := client.API().GetMachineSnapshotWithResponse(context.Background(), client.Account(), machineID, d.Id())
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error reading snapshot: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	snapshot := resp.JSON200
	d.Set("name", snapshot.Name)
	d.Set("state", snapshotStateString(snapshot.State))

	return nil
}

func resourceSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	machineID, err := parseUUID(d.Get("machine_id").(string))
	if err != nil {
		return fmt.Errorf("invalid machine_id: %s", err)
	}

	resp, err := client.API().DeleteMachineSnapshotWithResponse(context.Background(), client.Account(), machineID, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting snapshot: %s", err)
	}
	if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
		return fmt.Errorf("error deleting snapshot: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return nil
}

func resourceSnapshotParseIds(id string) (string, string, error) {
	parts := strings.SplitN(id, ".", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected machineId.snapshotName", id)
	}

	return parts[0], parts[1], nil
}
