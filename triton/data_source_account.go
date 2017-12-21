package triton

import (
	"context"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/account"
)

func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAccountRead,
		Schema: map[string]*schema.Schema{
			"cns_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Account()
	if err != nil {
		return err
	}

	acc, err := c.Get(context.Background(), &account.GetInput{})
	if err != nil {
		return err
	}

	d.SetId(acc.ID)

	d.Set("cns_enabled", acc.TritonCNSEnabled)

	return nil
}
