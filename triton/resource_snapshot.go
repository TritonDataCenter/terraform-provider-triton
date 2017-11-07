package triton

import (
	"context"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/compute"
)

const (
	snapshotCreateTimeout = 30 * time.Minute
)

func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceSnapshotCreate,
		Read:   resourceSnapshotRead,
		Delete: resourceSnapshotDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"machine_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"state": {
				Type:     schema.TypeString,
				Computed: true,
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
