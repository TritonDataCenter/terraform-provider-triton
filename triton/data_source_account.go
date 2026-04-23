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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	log.Printf("[DEBUG] triton_account: Reading Account details.")
	resp, err := client.API().GetAccountWithResponse(context.Background(), client.Account())
	if err != nil {
		return fmt.Errorf("error retrieving Account details: %s", err)
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error retrieving Account details: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	acc := resp.JSON200
	log.Printf("[DEBUG] triton_account: Found matching Account: %+v", acc)
	d.SetId(uuidString(acc.ID))

	d.Set("login", acc.Login)
	d.Set("email", acc.Email)
	d.Set("cns_enabled", derefBool(acc.TritonCnsEnabled))

	return nil
}
