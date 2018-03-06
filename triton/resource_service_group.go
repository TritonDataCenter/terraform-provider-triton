package triton

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/services"
)

const groupStateChangeTimeout = 2 * time.Minute

func resourceServiceGroup() *schema.Resource {
	return &schema.Resource{
		Create:   resourceGroupCreate,
		Exists:   resourceGroupExists,
		Read:     resourceGroupRead,
		Update:   resourceGroupUpdate,
		Delete:   resourceGroupDelete,
		Timeouts: slowResourceTimeout,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group_name": {
				Description:  "Friendly name for the service group",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: resourceGroupValidateName,
			},
			"template": {
				Description: "Identifier of an instance template",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"capacity": {
				Description: "Number of instances to launch and monitor",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"health_check_interval": {
				Description: "Interval in which to test for healthy instances",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceGroupValidateName(value interface{}, name string) ([]string, []error) {
	warnings := []string{}
	errors := []error{}

	r := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\_\.\-]*$`)
	if !r.Match([]byte(value.(string))) {
		errors = append(errors, fmt.Errorf(`"%s" is not a valid %s`, value.(string), name))
	}

	return warnings, errors
}

func resourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	svc, err := client.Services()
	if err != nil {
		return err
	}

	groupName := d.Get("group_name").(string)

	templateID, err := strconv.ParseInt(d.Get("template").(string), 10, 64)
	if err != nil {
		return err
	}

	err = svc.Groups().Create(context.Background(), &services.CreateGroupInput{
		GroupName:           groupName,
		TemplateID:          templateID,
		Capacity:            d.Get("capacity").(int),
		HealthCheckInterval: d.Get("health_check_interval").(int),
	})
	if err != nil {
		return err
	}

	d.SetId(groupName)

	// refresh state after provisioning
	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Services()
	if err != nil {
		return err
	}

	ctx := context.Background()
	grp, err := c.Groups().Get(ctx, &services.GetGroupInput{
		Name: d.Id(),
	})
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", grp.ID))

	d.Set("group_name", grp.GroupName)
	d.Set("template", grp.TemplateID)
	d.Set("capacity", grp.Capacity)
	d.Set("health_check_interval", grp.HealthCheckInterval)

	return nil
}

func resourceGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	c, err := client.Services()
	if err != nil {
		return false, err
	}

	ctx := context.Background()
	grp, err := c.Groups().Get(ctx, &services.GetGroupInput{
		Name: d.Id(),
	})
	if err != nil {
		return false, err
	}
	if grp != nil {
		return true, nil
	}

	return false, fmt.Errorf("failed to find v% service group", d.Id())
}

func resourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	svc, err := client.Services()
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = svc.Groups().Delete(ctx, &services.DeleteGroupInput{
		Name: d.Id(),
	})
	if err != nil {
		return err
	}

	return nil
}
