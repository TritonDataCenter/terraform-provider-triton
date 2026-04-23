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
	"strconv"
	"strings"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFabric() *schema.Resource {
	return &schema.Resource{
		Create: resourceFabricCreate,
		Exists: resourceFabricExists,
		Read:   resourceFabricRead,
		Delete: resourceFabricDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				// d.Id() is the last argument passed to the `terraform import RESOURCE_TYPE.RESOURCE_NAME RESOURCE_ID` command
				// We need to parse both the fabric vlan ID and the fabric UUID to import it
				vlanIdString, fabricId, err := resourceFabricParseIds(d.Id())

				if err != nil {
					return nil, err
				}

				vlanIdInt, err := strconv.Atoi(vlanIdString)

				if err != nil {
					return nil, err
				}

				d.Set("vlan_id", vlanIdInt)
				d.SetId(fabricId)

				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaVersion: 1,
		MigrateState:  resourceFabricMigrateState,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Network name",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"public": {
				Description: "Whether or not this is an RFC1918 network",
				Computed:    true,
				Type:        schema.TypeBool,
			},
			"fabric": {
				Description: "Whether or not this network is on a fabric",
				Computed:    true,
				Type:        schema.TypeBool,
			},
			"description": {
				Description: "Description of network",
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"subnet": {
				Description: "CIDR formatted string describing network address space",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"provision_start_ip": {
				Description: "First IP on the network that can be assigned",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"provision_end_ip": {
				Description: "Last assignable IP on the network",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"gateway": {
				Description: "Gateway IP",
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"resolvers": {
				Description: "List of IP addresses for DNS resolvers",
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"routes": {
				Description: "Map of CIDR block to Gateway IP address",
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeMap,
			},
			"internet_nat": {
				Description: "Whether or not a NAT zone is provisioned at the Gateway IP address",
				Default:     true,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeBool,
			},
			"vlan_id": {
				Description: "VLAN on which the network exists",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeInt,
			},
		},
	}
}

func resourceFabricCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	vlanID := uint16(d.Get("vlan_id").(int))
	internetNat := d.Get("internet_nat").(bool)
	provisionStartIP := d.Get("provision_start_ip").(string)
	provisionEndIP := d.Get("provision_end_ip").(string)
	subnet := d.Get("subnet").(string)

	body := cloudapi.CreateFabricNetworkJSONRequestBody{
		Name:             d.Get("name").(string),
		Subnet:           subnet,
		ProvisionStartIP: provisionStartIP,
		ProvisionEndIP:   provisionEndIP,
		InternetNat:      &internetNat,
	}

	// Regression guard (#28): Only set optional fields when the user provides
	// a value. The pointer-based API omits nil fields, avoiding zero-value
	// submission that the API rejects.
	if v, ok := d.GetOk("description"); ok {
		body.Description = ptrString(v.(string))
	}
	if v, ok := d.GetOk("gateway"); ok {
		body.Gateway = ptrString(v.(string))
	}
	if v, ok := d.GetOk("resolvers"); ok {
		var resolvers []string
		for _, resolver := range v.([]interface{}) {
			resolvers = append(resolvers, resolver.(string))
		}
		body.Resolvers = &resolvers
	}
	if v, ok := d.GetOk("routes"); ok {
		routes := map[string]string{}
		for cidr, ip := range v.(map[string]interface{}) {
			ipStr, ok := ip.(string)
			if !ok {
				return fmt.Errorf(`cannot use "%v" as an IP address`, ip)
			}
			routes[cidr] = ipStr
		}
		body.Routes = routes
	}

	resp, err := client.API().CreateFabricNetworkWithResponse(context.Background(), client.Account(), vlanID, body)
	if err != nil {
		return fmt.Errorf("error creating fabric network: %s", err)
	}
	if resp.JSON201 == nil {
		return fmt.Errorf("error creating fabric network: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	d.SetId(uuidString(resp.JSON201.ID))

	return resourceFabricRead(d, meta)
}

func resourceFabricExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)

	vlanID := uint16(d.Get("vlan_id").(int))
	fabricID, err := parseUUID(d.Id())
	if err != nil {
		return false, fmt.Errorf("invalid fabric network ID: %s", err)
	}

	resp, err := client.API().GetFabricNetworkWithResponse(context.Background(), client.Account(), vlanID, fabricID)
	if err != nil {
		return false, err
	}
	if isNotFound(resp.StatusCode()) {
		return false, nil
	}
	if resp.JSON200 == nil {
		return false, fmt.Errorf("error checking fabric network: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return true, nil
}

func resourceFabricRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	vlanID := uint16(d.Get("vlan_id").(int))
	fabricID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid fabric network ID: %s", err)
	}

	resp, err := client.API().GetFabricNetworkWithResponse(context.Background(), client.Account(), vlanID, fabricID)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error reading fabric network: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	fabric := resp.JSON200
	d.SetId(uuidString(fabric.ID))
	d.Set("name", fabric.Name)
	d.Set("public", fabric.Public)
	d.Set("fabric", derefBool(fabric.Fabric))
	d.Set("description", derefString(fabric.Description))
	d.Set("subnet", derefString(fabric.Subnet))
	d.Set("provision_start_ip", derefString(fabric.ProvisionStartIP))
	d.Set("provision_end_ip", derefString(fabric.ProvisionEndIP))
	d.Set("gateway", derefString(fabric.Gateway))
	d.Set("resolvers", derefStringSlice(fabric.Resolvers))
	d.Set("routes", fabric.Routes)
	d.Set("internet_nat", derefBool(fabric.InternetNat))
	d.Set("vlan_id", d.Get("vlan_id").(int))

	return nil
}

func resourceFabricDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	vlanID := uint16(d.Get("vlan_id").(int))
	fabricID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid fabric network ID: %s", err)
	}

	// Retry on conflict errors (e.g. when instances are still using the fabric).
	_, err2 := retryOnError(func(err error) bool {
		return strings.Contains(err.Error(), "InvalidArgument") ||
			strings.Contains(err.Error(), "409")
	}, func() (interface{}, error) {
		resp, err := client.API().DeleteFabricNetworkWithResponse(context.Background(), client.Account(), vlanID, fabricID)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
			return nil, fmt.Errorf("error deleting fabric network: %s", formatAPIError(resp.StatusCode(), resp.Body))
		}
		return nil, nil
	})

	return err2
}

func resourceFabricParseIds(id string) (string, string, error) {
	parts := strings.SplitN(id, ".", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected vlanId.fabricId", id)
	}

	return parts[0], parts[1], nil
}
