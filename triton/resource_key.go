package triton

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/joyent/triton-go/account"
)

func resourceKey() *schema.Resource {
	return &schema.Resource{
		Create:   resourceKeyCreate,
		Exists:   resourceKeyExists,
		Read:     resourceKeyRead,
		Delete:   resourceKeyDelete,
		Timeouts: fastResourceTimeout,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the key (generated from the key comment if not set)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"key": {
				Description: "Content of public key from disk in OpenSSH format",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	a, err := client.Account()
	if err != nil {
		return err
	}

	if keyName := d.Get("name").(string); keyName == "" {
		parts := strings.SplitN(d.Get("key").(string), " ", 3)
		if len(parts) == 3 {
			d.Set("name", parts[2])
		} else {
			return errors.New("No key name specified, and key material has no comment")
		}
	}

	_, err = a.Keys().Create(context.Background(), &account.CreateKeyInput{
		Name: d.Get("name").(string),
		Key:  d.Get("key").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))

	return resourceKeyRead(d, meta)
}

func resourceKeyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	a, err := client.Account()
	if err != nil {
		return false, err
	}

	_, err = a.Keys().Get(context.Background(), &account.GetKeyInput{
		KeyName: d.Id(),
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func resourceKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	a, err := client.Account()
	if err != nil {
		return err
	}

	key, err := a.Keys().Get(context.Background(), &account.GetKeyInput{
		KeyName: d.Id(),
	})
	if err != nil {
		return err
	}

	d.Set("name", key.Name)
	d.Set("key", key.Key)

	return nil
}

func resourceKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	a, err := client.Account()
	if err != nil {
		return err
	}

	return a.Keys().Delete(context.Background(), &account.DeleteKeyInput{
		KeyName: d.Id(),
	})
}
