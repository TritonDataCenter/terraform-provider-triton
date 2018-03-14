package triton

import (
	"context"

	"fmt"
	"time"

	"strings"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/compute"
)

const (
	volumeStateReady   = "ready"
	volumeStateFailed  = "failed"
	volumeStateDeleted = "deleted"
)

func resourceVolume() *schema.Resource {
	return &schema.Resource{
		Create:   resourceVolumeCreate,
		Read:     resourceVolumeRead,
		Update:   resourceVolumeUpdate,
		Delete:   resourceVolumeDelete,
		Timeouts: fastResourceTimeout,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Required: true,
				Type:     schema.TypeString,
			},
			"networks": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"size": {
				Optional: true,
				Default:  10240,
				ForceNew: true,
				Type:     schema.TypeInt,
			},
			"type": {
				Optional: true,
				Default:  "tritonnfs",
				ForceNew: true,
				Type:     schema.TypeString,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filesystem_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	var networks []string
	for _, network := range d.Get("networks").(*schema.Set).List() {
		networks = append(networks, network.(string))
	}

	vol, err := c.Volumes().Create(context.Background(), &compute.CreateVolumeInput{
		Name:     d.Get("name").(string),
		Size:     int64(d.Get("size").(int)),
		Type:     d.Get("type").(string),
		Networks: networks,
	})
	if err != nil {
		return err
	}

	d.SetId(vol.ID)

	stateConf := &resource.StateChangeConf{
		Target: []string{volumeStateReady},
		Refresh: func() (interface{}, string, error) {
			inst, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
				ID: d.Id(),
			})
			if err != nil {
				return nil, "", err
			}
			if inst.State == volumeStateFailed {
				d.SetId("")
				return nil, "", fmt.Errorf("volume creation failed: %s", inst.State)
			}

			return inst, inst.State, nil
		},
		Timeout:    1 * time.Minute,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceVolumeRead(d, meta)
}

func resourceVolumeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	volume, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	d.Set("name", volume.Name)
	d.Set("type", volume.Type)
	d.Set("size", volume.Size)
	d.Set("owner", volume.Owner)
	d.Set("filesystem_path", volume.FileSystemPath)
	d.Set("networks", volume.Networks)

	return nil
}

func resourceVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	err = c.Volumes().Update(context.Background(), &compute.UpdateVolumeInput{
		ID:   d.Id(),
		Name: d.Get("name").(string),
	})
	if err != nil {
		return err
	}

	return resourceVolumeRead(d, meta)
}

func resourceVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	err = c.Volumes().Delete(context.Background(), &compute.DeleteVolumeInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Target: []string{volumeStateDeleted},
		Refresh: func() (interface{}, string, error) {
			inst, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
				ID: d.Id(),
			})
			if err != nil {
				if strings.Contains(err.Error(), "VolumeNotFound") {
					return inst, "deleted", nil
				}
				return nil, "", err
			}

			return inst, inst.State, nil
		},
		Timeout:    3 * time.Minute,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return nil
}
