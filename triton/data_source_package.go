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
	"strings"

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

	resp, err := client.API().ListPackagesWithResponse(context.Background(), client.Account())
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error listing packages: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	allPackages := *resp.JSON200

	// Client-side filtering (the new CloudAPI client has no server-side
	// filter parameters for ListPackages).
	filterMemory := uint64(filters["memory"].(int))
	filterDisk := uint64(filters["disk"].(int))
	filterSwap := uint64(filters["swap"].(int))
	filterLwps := uint32(filters["lwps"].(int))
	filterVcpus := uint32(filters["vcpus"].(int))
	filterVersion := filters["version"].(string)
	filterGroup := filters["group"].(string)

	var filtered []int
	for i, p := range allPackages {
		if filterMemory > 0 && p.Memory != filterMemory {
			continue
		}
		if filterDisk > 0 && p.Disk != filterDisk {
			continue
		}
		if filterSwap > 0 && p.Swap != filterSwap {
			continue
		}
		if filterLwps > 0 && (p.Lwps == nil || *p.Lwps != filterLwps) {
			continue
		}
		if filterVcpus > 0 && (p.Vcpus == nil || *p.Vcpus != filterVcpus) {
			continue
		}
		if filterVersion != "" && (p.Version == nil || *p.Version != filterVersion) {
			continue
		}
		if filterGroup != "" && (p.Group == nil || *p.Group != filterGroup) {
			continue
		}
		filtered = append(filtered, i)
	}

	if len(filtered) == 0 {
		return fmt.Errorf("your query returned no results, please change " +
			"your filter criteria and try again")
	}

	iname, hasName := filters["name"]
	name := iname.(string)

	var matchIdx = -1
	if hasName && name != "" {
		for _, idx := range filtered {
			if strings.Contains(allPackages[idx].Name, name) {
				matchIdx = idx
				break
			}
		}
	}

	if matchIdx < 0 {
		names := make([]string, 0)
		for _, idx := range filtered {
			if hasName && name != "" {
				if strings.Contains(allPackages[idx].Name, name) {
					names = append(names, allPackages[idx].Name)
				}
			} else {
				names = append(names, allPackages[idx].Name)
			}
		}
		return fmt.Errorf(
			"your query returned more than one result (%v),\nplease change "+
				"your filter criteria and try again", strings.Join(names, ", "))
	}

	pkg := allPackages[matchIdx]

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

	return nil
}
