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
	"strings"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The name of the package.",
					Type:        schema.TypeString,
					Optional:    true,
				},

				"memory": {
					Description: "How much memory will by available (in MiB).",
					Type:        schema.TypeInt,
					Optional:    true,
				},

				"disk": {
					Description: "How much disk space will be available (in MiB).",
					Type:        schema.TypeInt,
					Optional:    true,
				},

				"swap": {
					Description: "How much swap space will be available (in MiB).",
					Type:        schema.TypeInt,
					Optional:    true,
				},

				"lwps": {
					Description: "Maximum number of light-weight processes (threads) allowed.",
					Type:        schema.TypeInt,
					Optional:    true,
				},

				"vcpus": {
					Description: "Number of vCPUs of the package",
					Type:        schema.TypeInt,
					Optional:    true,
				},

				"version": {
					Description: "The version of the package.",
					Type:        schema.TypeString,
					Optional:    true,
				},

				"group": {
					Description: "The group of the package.",
					Type:        schema.TypeString,
					Optional:    true,
				},

				"brand": {
					Description: "The brand of the package (e.g. bhyve, joyent, lx).",
					Type:        schema.TypeString,
					Optional:    true,
				},

				"flexible_disk": {
					Description: "Whether the package uses flexible disk (bhyve only).",
					Type:        schema.TypeBool,
					Optional:    true,
				},
			},
		},
	}
}

func dataSourcePackage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePackageRead,
		Schema: map[string]*schema.Schema{

			"filter": dataSourceFiltersSchema(),

			"name": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"disk": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"swap": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"lwps": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"vcpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"group": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"brand": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"flexible_disk": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourcePackageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	filters := map[string]interface{}{}
	if filterSet, found := d.Get("filter").(*schema.Set); found {
		filterRaw := filterSet.List()[0]
		if filterRaw == nil {
			return fmt.Errorf("please set filters on your package data source")
		}
		filters = filterRaw.(map[string]interface{})
	}

	// Build server-side filter params for all exact-match fields.
	params := &cloudapi.ListPackagesParams{}
	if v := filters["name"].(string); v != "" {
		params.Name = &v
	}
	if v := uint64(filters["memory"].(int)); v > 0 {
		params.Memory = &v
	}
	if v := uint64(filters["disk"].(int)); v > 0 {
		params.Disk = &v
	}
	if v := uint64(filters["swap"].(int)); v > 0 {
		params.Swap = &v
	}
	if v := uint32(filters["lwps"].(int)); v > 0 {
		params.Lwps = &v
	}
	if v := uint32(filters["vcpus"].(int)); v > 0 {
		params.Vcpus = &v
	}
	if v := filters["version"].(string); v != "" {
		params.Version = &v
	}
	if v := filters["group"].(string); v != "" {
		params.Group = &v
	}
	if v := filters["brand"].(string); v != "" {
		params.Brand = &v
	}
	if v := filters["flexible_disk"].(bool); v {
		params.FlexibleDisk = &v
	}

	resp, err := client.API().ListPackagesWithResponse(context.Background(), client.Account(), params)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error listing packages: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	packages := *resp.JSON200

	if len(packages) == 0 {
		return fmt.Errorf("your query returned no results, please change " +
			"your filter criteria and try again")
	}

	// Name uses substring matching, so it stays client-side.
	name := filters["name"].(string)

	var matchIdx = -1
	if name != "" {
		for i, p := range packages {
			if strings.Contains(p.Name, name) {
				matchIdx = i
				break
			}
		}
	}

	if matchIdx < 0 {
		var names []string
		for _, p := range packages {
			if name != "" {
				if strings.Contains(p.Name, name) {
					names = append(names, p.Name)
				}
			} else {
				names = append(names, p.Name)
			}
		}
		return fmt.Errorf(
			"your query returned more than one result (%v),\nplease change "+
				"your filter criteria and try again", strings.Join(names, ", "))
	}

	pkg := packages[matchIdx]

	d.SetId(uuidString(pkg.ID))
	d.Set("name", pkg.Name)
	d.Set("memory", int(pkg.Memory))
	d.Set("disk", int(pkg.Disk))
	d.Set("swap", int(pkg.Swap))

	var lwps int
	if pkg.Lwps != nil {
		lwps = int(*pkg.Lwps)
	}
	d.Set("lwps", lwps)

	var vcpus int
	if pkg.Vcpus != nil {
		vcpus = int(*pkg.Vcpus)
	}
	d.Set("vcpus", vcpus)

	d.Set("version", derefString(pkg.Version))
	d.Set("group", derefString(pkg.Group))
	d.Set("brand", vmBrandString(pkg.Brand))
	d.Set("flexible_disk", pkg.FlexibleDisk != nil && *pkg.FlexibleDisk)

	return nil
}

func vmBrandString(b *cloudapi.VMBrand) string {
	if b == nil {
		return ""
	}
	v, err := b.AsVMBrand0()
	if err != nil {
		v1, err2 := b.AsVMBrand1()
		if err2 != nil {
			v2, err3 := b.AsVMBrand2()
			if err3 != nil {
				log.Printf("[WARN] vmBrandString: failed to decode all union branches (brand0: %s, brand1: %s, brand2: %s)", err, err2, err3)
				return ""
			}
			return string(v2)
		}
		return string(v1)
	}
	return string(v)
}
