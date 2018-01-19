package triton

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/compute"
)

func dataSourceFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Optional: true,
				},

				"memory": {
					Type:     schema.TypeInt,
					Optional: true,
				},

				"disk": {
					Type:     schema.TypeInt,
					Optional: true,
				},

				"swap": {
					Type:     schema.TypeInt,
					Optional: true,
				},

				"lwps": {
					Type:     schema.TypeInt,
					Optional: true,
				},

				"vcpus": {
					Type:     schema.TypeInt,
					Optional: true,
				},

				"version": {
					Type:     schema.TypeString,
					Optional: true,
				},

				"group": {
					Type:     schema.TypeString,
					Optional: true,
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
	c, err := client.Compute()
	if err != nil {
		return err
	}

	filters := map[string]interface{}{}
	if filterSet, found := d.Get("filter").(*schema.Set); found {
		filterRaw := filterSet.List()[0]
		if filterRaw == nil {
			return fmt.Errorf("Please set filters on your package data source.")
		}
		filters = filterRaw.(map[string]interface{})
	}

	input := &compute.ListPackagesInput{}

	if memory := int64(filters["memory"].(int)); memory > 0 {
		input.Memory = memory
	}
	if disk := int64(filters["disk"].(int)); disk > 0 {
		input.Disk = disk
	}
	if swap := int64(filters["swap"].(int)); swap > 0 {
		input.Swap = swap
	}
	if lwps := int64(filters["lwps"].(int)); lwps > 0 {
		input.LWPs = lwps
	}
	if vcpus := int64(filters["vcpus"].(int)); vcpus > 0 {
		input.VCPUs = vcpus
	}
	if version := filters["version"].(string); version != "" {
		input.Version = version
	}
	if group := filters["group"].(string); group != "" {
		input.Group = group
	}

	packages, err := c.Packages().List(context.Background(), input)
	if err != nil {
		return err
	}
	if len(packages) == 0 {
		return fmt.Errorf("Your query returned no results. Please change " +
			"your filter criteria and try again.")
	}

	iname, hasName := filters["name"]
	name := iname.(string)

	var pkg *compute.Package
	if hasName {
		for _, p := range packages {
			if strings.Contains(p.Name, name) {
				pkg = p
				break
			}
		}
	}

	if pkg == nil {
		names := make([]string, 0)
		for _, pkg := range packages {
			if hasName {
				if strings.Contains(pkg.Name, name) {
					names = append(names, pkg.Name)
				}
			} else {
				names = append(names, pkg.Name)
			}
		}
		return fmt.Errorf(
			"Your query returned more than one result (%v).\nPlease change "+
				"your filter criteria and try again.", strings.Join(names, ", "))
	}

	d.SetId(pkg.ID)
	d.Set("name", pkg.Name)
	d.Set("memory", pkg.Memory)
	d.Set("disk", pkg.Disk)
	d.Set("swap", pkg.Swap)
	d.Set("lwps", pkg.LWPs)
	d.Set("vcpus", pkg.VCPUs)
	d.Set("version", pkg.Version)
	d.Set("group", pkg.Group)

	return nil
}
