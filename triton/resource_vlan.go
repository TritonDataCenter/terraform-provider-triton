package triton

import (
	"context"
	"errors"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/network"
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
	n, err := client.Network()
	if err != nil {
		return err
	}

	vlan, err := n.Fabrics().CreateVLAN(context.Background(), &network.CreateVLANInput{
		ID:          d.Get("vlan_id").(int),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(vlan.ID))
	return resourceVLANRead(d, meta)
}

func resourceVLANExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	n, err := client.Network()
	if err != nil {
		return false, err
	}

	id, err := resourceVLANIDInt(d.Id())
	if err != nil {
		return false, err
	}

	return resourceExists(n.Fabrics().GetVLAN(context.Background(), &network.GetVLANInput{
		ID: id,
	}))
}

func resourceVLANRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	n, err := client.Network()
	if err != nil {
		return err
	}

	id, err := resourceVLANIDInt(d.Id())
	if err != nil {
		return err
	}

	vlan, err := n.Fabrics().GetVLAN(context.Background(), &network.GetVLANInput{
		ID: id,
	})
	if err != nil {
		return err
	}

	d.Set("vlan_id", vlan.ID)
	d.Set("name", vlan.Name)
	d.Set("description", vlan.Description)

	return nil
}

func resourceVLANUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	n, err := client.Network()
	if err != nil {
		return err
	}

	vlan, err := n.Fabrics().UpdateVLAN(context.Background(), &network.UpdateVLANInput{
		ID:          d.Get("vlan_id").(int),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(vlan.ID))
	return resourceVLANRead(d, meta)
}

func resourceVLANDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	n, err := client.Network()
	if err != nil {
		return err
	}

	id, err := resourceVLANIDInt(d.Id())
	if err != nil {
		return err
	}

	return n.Fabrics().DeleteVLAN(context.Background(), &network.DeleteVLANInput{
		ID: id,
	})
}

func resourceVLANIDInt(id string) (int, error) {
	result, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return -1, err
	}

	return int(result), nil
}
