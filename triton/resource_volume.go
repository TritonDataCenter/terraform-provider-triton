package triton

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/TritonDataCenter/triton-go/compute"
	"github.com/TritonDataCenter/triton-go/errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
	c, err := client.Compute()
	if err != nil {
		return err
	}

	var networks []string
	for _, network := range d.Get("networks").([]interface{}) {
		networks = append(networks, network.(string))
	}

	tags := map[string]string{}
	for k, v := range d.Get("tags").(map[string]interface{}) {
		tags[k] = v.(string)
	}

	createInput := &compute.CreateVolumeInput{
		Networks: networks,
		Tags:     tags,
	}

	if value, ok := d.GetOk("name"); ok {
		createInput.Name = value.(string)
	}

	if value, ok := d.GetOk("type"); ok {
		createInput.Type = value.(string)
	}

	if value, ok := d.GetOk("size"); ok {
		createInput.Size = int64(value.(int))
	}

	volume, err := c.Volumes().Create(context.Background(), createInput)
	if err != nil {
		return err
	}

	d.SetId(volume.ID)
	stateConf := &resource.StateChangeConf{
		Target: []string{volumeStateReady},
		Refresh: func() (interface{}, string, error) {
			volume, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
				ID: d.Id(),
			})
			if err != nil {
				return nil, "", err
			}
			if volume.State == volumeStateFailed {
				d.SetId("")
				return nil, "", fmt.Errorf("volume creation failed: %s", volume.State)
			}

			return volume, volume.State, nil
		},
		Timeout:    *slowResourceTimeout.Create,
		MinTimeout: defaultPollInterval,
	}
	v, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	return tritonVolumeToTerraformVolume(d, v.(*compute.Volume))
}

func resourceVolumeExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return false, err
	}

	return resourceExists(c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
		ID: d.Id(),
	}))
}

func tritonVolumeToTerraformVolume(d *schema.ResourceData, volume *compute.Volume) error {
	d.SetId(volume.ID)

	d.Set("filesystem_path", volume.FileSystemPath)
	d.Set("name", volume.Name)
	d.Set("networks", volume.Networks)
	d.Set("size", volume.Size)
	d.Set("state", volume.State)
	d.Set("tags", volume.Tags)
	d.Set("type", volume.Type)

	return nil
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
		if errors.IsSpecificStatusCode(err, http.StatusNotFound) || errors.IsSpecificStatusCode(err, http.StatusGone) {
			log.Printf("Volume %q not found or has been deleted", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if volume.State == volumeStateFailed {
		log.Printf("Volume %q state: `failed` so removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return tritonVolumeToTerraformVolume(d, volume)
}

func resourceVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		oldNameInterface, newNameInterface := d.GetChange("name")
		oldName := oldNameInterface.(string)
		newName := newNameInterface.(string)

		err := c.Volumes().Update(context.Background(), &compute.UpdateVolumeInput{
			ID:   d.Id(),
			Name: newName,
		})
		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{oldName},
			Target:  []string{newName},
			Refresh: func() (interface{}, string, error) {
				volume, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
					ID: d.Id(),
				})
				if err != nil {
					return nil, "", err
				}

				return volume, volume.Name, nil
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
	c, err := client.Compute()
	if err != nil {
		return err
	}

	err = c.Volumes().Delete(context.Background(), &compute.DeleteVolumeInput{
		ID: d.Id(),
	})
	if err != nil {
		// Allow it to be already deleted.
		if errors.IsSpecificStatusCode(err, http.StatusNotFound) || errors.IsSpecificStatusCode(err, http.StatusGone) {
			return nil
		}
		return err
	}

	stateConf := &resource.StateChangeConf{
		Target: []string{volumeStateDeleted},
		Refresh: func() (interface{}, string, error) {
			inst, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
				ID: d.Id(),
			})
			if err != nil {
				if errors.IsSpecificStatusCode(err, http.StatusNotFound) || errors.IsSpecificStatusCode(err, http.StatusGone) {
					return inst, "deleted", nil
				}
				return nil, "", err
			}

			return inst, inst.State, nil
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
