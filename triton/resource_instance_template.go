package triton

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/services"
)

const templateStateChangeTimeout = 2 * time.Minute

func resourceInstanceTemplate() *schema.Resource {
	return &schema.Resource{
		Create:   resourceTemplateCreate,
		Exists:   resourceTemplateExists,
		Read:     resourceTemplateRead,
		Delete:   resourceTemplateDelete,
		Timeouts: slowResourceTimeout,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"template_name": {
				Description:  "Friendly name for the instance template",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: resourceTemplateValidateName,
			},
			"image": {
				Description: "UUID of the image",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"package": {
				Description: "Package name used for provisioning",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"firewall_enabled": {
				Description: "Whether to enable the firewall for group instances",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"tags": {
				Description: "Tags for group instances",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"networks": {
				Description: "Network IDs for group instances",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Description: "Metadata for group instances",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"userdata": {
				Description: "Data copied to instance on boot",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
		},
	}
}

func resourceTemplateValidateName(value interface{}, name string) ([]string, []error) {
	warnings := []string{}
	errors := []error{}

	r := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\_\.\-]*$`)
	if !r.Match([]byte(value.(string))) {
		errors = append(errors, fmt.Errorf(`"%s" is not a valid %s`, value.(string), name))
	}

	return warnings, errors
}

func resourceTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	svc, err := client.Services()
	if err != nil {
		return err
	}

	var networks []string
	for _, network := range d.Get("networks").([]interface{}) {
		networks = append(networks, network.(string))
	}

	metadata := map[string]string{}
	for k, v := range d.Get("metadata").(map[string]interface{}) {
		metadata[k] = v.(string)
	}

	tags := map[string]string{}
	for k, v := range d.Get("tags").(map[string]interface{}) {
		tags[k] = v.(string)
	}

	templateName := d.Get("template_name").(string)

	ctx := context.Background()
	tmpl, err := svc.Templates().Create(ctx, &services.CreateTemplateInput{
		TemplateName:    templateName,
		Package:         d.Get("package").(string),
		ImageID:         d.Get("image").(string),
		FirewallEnabled: d.Get("firewall_enabled").(bool),
		Networks:        networks,
		Userdata:        d.Get("userdata").(string),
		Metadata:        metadata,
		Tags:            tags,
	})
	if err != nil {
		return err
	}

	d.SetId(tmpl.ID)

	// refresh state after provisioning
	return resourceTemplateRead(d, meta)
}

func resourceTemplateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Services()
	if err != nil {
		return err
	}

	ctx := context.Background()
	tmpl, err := c.Templates().Get(ctx, &services.GetTemplateInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	d.Set("template_name", tmpl.TemplateName)
	d.Set("package", tmpl.Package)
	d.Set("image", tmpl.ImageID)
	d.Set("firewall_enabled", tmpl.FirewallEnabled)
	d.Set("networks", tmpl.Networks)
	d.Set("userdata", tmpl.Userdata)
	d.Set("metadata", tmpl.Metadata)
	d.Set("tags", tmpl.Tags)

	return nil
}

func resourceTemplateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	c, err := client.Services()
	if err != nil {
		return false, err
	}

	ctx := context.Background()
	tmpl, err := c.Templates().Get(ctx, &services.GetTemplateInput{
		ID: d.Id(),
	})
	if err != nil {
		return false, err
	}
	if tmpl != nil {
		return true, nil
	}

	return false, fmt.Errorf("failed to find instance template by name %v", d.Id())
}

func resourceTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	svc, err := client.Services()
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = svc.Templates().Delete(ctx, &services.DeleteTemplateInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	return nil
}
