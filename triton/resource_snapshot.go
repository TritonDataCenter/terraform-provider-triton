package triton

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TritonDataCenter/triton-go/compute"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				// d.Id() is the last argument passed to the `terraform import RESOURCE_TYPE.RESOURCE_NAME RESOURCE_ID` command
				// We need to parse both the instance UUID and the snapshot UUID to import it
				machineId, snapshotId, err := resourceSnapshotParseIds(d.Id())

				if err != nil {
					return nil, err
				}

				d.Set("machine_id", machineId)
				d.SetId(snapshotId)

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

func resourceSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	createInput := &compute.CreateSnapshotInput{
		MachineID: d.Get("machine_id").(string),
		Name:      d.Get("name").(string),
	}

	snapshot, err := c.Snapshots().Create(context.Background(), createInput)
	if err != nil {
		return err
	}

	d.SetId(snapshot.Name)

	stateConf := &resource.StateChangeConf{
		Target: []string{"created"},
		Refresh: func() (interface{}, string, error) {
			snapshot, err := c.Snapshots().Get(context.Background(), &compute.GetSnapshotInput{
				MachineID: d.Get("machine_id").(string),
				Name:      d.Id(),
			})
			if err != nil {
				return nil, "", err
			}
			return snapshot, snapshot.State, nil
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
	c, err := client.Compute()
	if err != nil {
		return err
	}

	snapshot, err := c.Snapshots().Get(context.Background(), &compute.GetSnapshotInput{
		MachineID: d.Get("machine_id").(string),
		Name:      d.Id(),
	})
	if err != nil {
		return err
	}

	d.Set("name", snapshot.Name)
	d.Set("state", snapshot.State)

	return nil
}

func resourceSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	return c.Snapshots().Delete(context.Background(), &compute.DeleteSnapshotInput{
		Name:      d.Id(),
		MachineID: d.Get("machine_id").(string),
	})
}

func resourceSnapshotParseIds(id string) (string, string, error) {
	parts := strings.SplitN(id, ".", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected machineId.snapshotId", id)
	}

	return parts[0], parts[1], nil
}
