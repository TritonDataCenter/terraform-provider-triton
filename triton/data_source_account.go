package triton

import (
	"context"
	"log"

	"github.com/TritonDataCenter/triton-go/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// dataSourceAccount returns schema for the Account data source.
func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAccountRead,
		Schema: map[string]*schema.Schema{
			"login": {
				Description: "The login name associated with the Account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "An e-mail address that is current set in the Account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cns_enabled": {
				Description: "Whether the Container Name Service (CNS) is enabled for the Account.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

// dataSourceAccountRead retrieves details about current Account from Triton
// using the Account API. The current Account name will be the same as the
// one currently configured in the Triton provider.
func dataSourceAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	c, err := client.Account()
	if err != nil {
		return errors.Wrap(err, "error creating Account client")
	}

	log.Printf("[DEBUG] triton_account: Reading Account details.")
	acc, err := c.Get(context.Background(), &account.GetInput{})
	if err != nil {
		return errors.Wrap(err, "error retrieving Account details")
	}

	log.Printf("[DEBUG] triton_account: Found matching Account: %+v", acc)
	d.SetId(acc.ID)

	d.Set("login", acc.Login)
	d.Set("email", acc.Email)
	d.Set("cns_enabled", acc.TritonCNSEnabled)

	return nil
}
