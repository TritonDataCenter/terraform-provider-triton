package triton

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/compute"
	"github.com/mitchellh/hashstructure"
)

const (
	machineStateDeleted      = "deleted"
	machineStateFailed       = "failed"
	machineStateProvisioning = "provisioning"
	machineStateRunning      = "running"
	machineStateStopped      = "stopped"
	machineStateStopping     = "stopping"

	machineStateChangeTimeout = 10 * time.Minute
)

// semantics: "argument_name": "metadata-key"
var metadataArgumentsToKeys = map[string]string{
	"administrator_pw":     "administrator-pw",
	"cloud_config":         "cloud-init:user-data",
	"root_authorized_keys": "root_authorized_keys",
	"user_data":            "user-data",
	"user_script":          "user-script",
}

func resourceMachine() *schema.Resource {
	return &schema.Resource{
		Create:   resourceMachineCreate,
		Exists:   resourceMachineExists,
		Read:     resourceMachineRead,
		Update:   resourceMachineUpdate,
		Delete:   resourceMachineDelete,
		Timeouts: slowResourceTimeout,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Friendly name for machine",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: resourceMachineValidateName,
			},
			"type": {
				Description: "Machine type (smartmachine or virtualmachine)",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dataset": {
				Description: "Dataset URN with which the machine was provisioned",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"memory": {
				Description: "Amount of memory allocated to the machine (in Mb)",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"disk": {
				Description: "Amount of disk allocated to the machine (in Gb)",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"ips": {
				Description: "IP addresses assigned to the machine",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Description: "Machine tags",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"cns": {
				Description: "Container Name Service",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable": {
							Description: "Disable CNS for this instance (after create)",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"services": {
							Description: "Assign CNS service names to this instance",
							Optional:    true,
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"affinity": {
				Description: "Label based affinity rules for assisting instance placement",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"locality": {
				Description: "UUID based locality hints for assisting placement behavior",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"close_to": {
							Description: "UUIDs of other instances to attempt to provision alongside",
							Optional:    true,
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"far_from": {
							Description: "UUIDs of other instances to attempt not to provision alongside",
							Optional:    true,
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"metadata": {
				Description: "Machine metadata",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"created": {
				Description: "When the machine was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated": {
				Description: "When the machine was updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"package": {
				Description: "The package for use for provisioning",
				Type:        schema.TypeString,
				Required:    true,
			},
			"image": {
				Description: "UUID of the image",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"primaryip": {
				Description: "Primary (public) IP address for the machine",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"networks": {
				Description: "Desired network IDs",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"nic": {
				Description: "Network interface",
				Type:        schema.TypeSet,
				Computed:    true,
				Optional:    true,
				Set: func(v interface{}) int {
					m := v.(map[string]interface{})
					return hashcode.String(m["network"].(string))
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Description: "NIC's IPv4 address",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"mac": {
							Description: "NIC's MAC address",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"primary": {
							Description: "Whether this is the machine's primary NIC",
							Computed:    true,
							Type:        schema.TypeBool,
						},
						"netmask": {
							Description: "IPv4 netmask",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"gateway": {
							Description: "IPv4 gateway",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"network": {
							Description: "ID of the network to which the NIC is attached",
							Required:    true,
							Type:        schema.TypeString,
						},
						"state": {
							Description: "Provisioning state of the NIC",
							Computed:    true,
							Type:        schema.TypeString,
						},
					},
				},
			},
			"firewall_enabled": {
				Description: "Whether to enable the firewall for this machine",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"domain_names": {
				Description: "List of domain names from Triton CNS",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// computed resources from metadata
			"root_authorized_keys": {
				Description: "Authorized keys for the root user on this machine",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			// resources derived from metadata
			"user_script": {
				Description: "User script to run on boot (every boot on SmartMachines)",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"cloud_config": {
				Description: "copied to machine on boot",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"user_data": {
				Description: "Data copied to machine on boot",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"administrator_pw": {
				Description: "Administrator's initial password (Windows only)",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceMachineCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	var affinity []string
	for _, rule := range d.Get("affinity").([]interface{}) {
		affinity = append(affinity, rule.(string))
	}

	if len(affinity) > 0 {
		client.affinityLock.Lock()
		defer client.affinityLock.Unlock()
	}

	var networks []string
	for _, network := range d.Get("networks").([]interface{}) {
		networks = append(networks, network.(string))
	}

	metadata := map[string]string{}
	for k, v := range d.Get("metadata").(map[string]interface{}) {
		metadata[k] = v.(string)
	}
	for argumentName, metadataKey := range metadataArgumentsToKeys {
		if v, ok := d.GetOk(argumentName); ok {
			metadata[metadataKey] = v.(string)
		}
	}

	tags := map[string]string{}
	for k, v := range d.Get("tags").(map[string]interface{}) {
		tags[k] = v.(string)
	}

	cns := compute.InstanceCNS{}
	if cnsRaw, found := d.GetOk("cns"); found {
		cnsList := cnsRaw.([]interface{})
		cnsMap, ok := cnsList[0].(map[string]interface{})
		if len(cnsList) > 0 && ok {
			for k, v := range cnsMap {
				switch k {
				case "disable":
					// NOTE: we can't provision an instance with CNS disabled
					d.Set("cns.0.disable", false)
				case "services":
					servicesRaw := v.([]interface{})
					cns.Services = make([]string, 0, len(servicesRaw))
					for _, serviceRaw := range servicesRaw {
						cns.Services = append(cns.Services, serviceRaw.(string))
					}
				default:
					return fmt.Errorf("unsupported CNS attribute %q", k)
				}
			}
		}
	}

	createInput := &compute.CreateInstanceInput{
		Name:            d.Get("name").(string),
		Package:         d.Get("package").(string),
		Image:           d.Get("image").(string),
		Networks:        networks,
		Metadata:        metadata,
		Affinity:        affinity,
		Tags:            tags,
		CNS:             cns,
		FirewallEnabled: d.Get("firewall_enabled").(bool),
	}

	if nearRaw, found := d.GetOk("locality.0.close_to"); found {
		nearList := nearRaw.([]interface{})
		localNear := make([]string, len(nearList))
		for i, val := range nearList {
			valStr := val.(string)
			if valStr != "" {
				localNear[i] = valStr
			}
		}
		createInput.LocalityNear = localNear
	}

	if farRaw, found := d.GetOk("locality.0.far_from"); found {
		farList := farRaw.([]interface{})
		localFar := make([]string, len(farList))
		for i, val := range farList {
			valStr := val.(string)
			if valStr != "" {
				localFar[i] = valStr
			}
		}
		createInput.LocalityFar = localFar
	}

	machine, err := c.Instances().Create(context.Background(), createInput)
	if err != nil {
		return err
	}

	d.SetId(machine.ID)
	stateConf := &resource.StateChangeConf{
		Target: []string{machineStateRunning},
		Refresh: func() (interface{}, string, error) {
			inst, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
				ID: d.Id(),
			})
			if err != nil {
				return nil, "", err
			}
			if inst.State == machineStateFailed {
				d.SetId("")
				return nil, "", fmt.Errorf("instance creation failed: %s", inst.State)
			}

			if hasInitDomainNames(d, inst) {
				return inst, inst.State, nil
			}
			return inst, inst.State, nil
		},
		Timeout:    machineStateChangeTimeout,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	// refresh state after it provisions
	return resourceMachineRead(d, meta)
}

func resourceMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return false, err
	}

	return resourceExists(c.Instances().Get(context.Background(), &compute.GetInstanceInput{
		ID: d.Id(),
	}))
}

func resourceMachineRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	machine, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	nics, err := c.Instances().ListNICs(context.Background(), &compute.ListNICsInput{
		InstanceID: d.Id(),
	})
	if err != nil {
		return err
	}

	d.Set("name", machine.Name)
	d.Set("type", machine.Type)
	d.Set("state", machine.State)
	d.Set("dataset", machine.Image)
	d.Set("image", machine.Image)
	d.Set("memory", machine.Memory)
	d.Set("disk", machine.Disk)
	d.Set("ips", machine.IPs)
	d.Set("cns", machine.CNS)
	d.Set("tags", machine.Tags)
	d.Set("created", machine.Created)
	d.Set("updated", machine.Updated)
	d.Set("package", machine.Package)
	d.Set("image", machine.Image)
	d.Set("primaryip", machine.PrimaryIP)
	d.Set("firewall_enabled", machine.FirewallEnabled)
	d.Set("domain_names", machine.DomainNames)

	// create and update NICs
	var (
		machineNICs []map[string]interface{}
		networks    []string
	)
	for _, nic := range nics {
		machineNICs = append(
			machineNICs,
			map[string]interface{}{
				"ip":      nic.IP,
				"mac":     nic.MAC,
				"primary": nic.Primary,
				"netmask": nic.Netmask,
				"gateway": nic.Gateway,
				"state":   nic.State,
				"network": nic.Network,
			},
		)
		networks = append(networks, nic.Network)
	}
	d.Set("nic", machineNICs)

	for argumentName, metadataKey := range metadataArgumentsToKeys {
		d.Set(argumentName, machine.Metadata[metadataKey])
		delete(machine.Metadata, metadataKey)
	}
	d.Set("metadata", machine.Metadata)

	return nil
}

func resourceMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	d.Partial(true)

	if d.HasChange("name") {
		oldNameInterface, newNameInterface := d.GetChange("name")
		oldName := oldNameInterface.(string)
		newName := newNameInterface.(string)

		err := c.Instances().Rename(context.Background(), &compute.RenameInstanceInput{
			ID:   d.Id(),
			Name: newName,
		})
		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{oldName},
			Target:  []string{newName},
			Refresh: func() (interface{}, string, error) {
				inst, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
					ID: d.Id(),
				})
				if err != nil {
					return nil, "", err
				}

				return inst, inst.Name, nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}

		d.SetPartial("name")
	}

	if d.HasChange("tags") || d.HasChange("cns") {
		tags := map[string]string{}
		for k, v := range d.Get("tags").(map[string]interface{}) {
			if strings.HasPrefix(k, "triton.cns") {
				delete(tags, k)
			} else {
				tags[k] = v.(string)
			}
		}

		cns := compute.InstanceCNS{}
		if cnsRaw, found := d.GetOk("cns"); found {
			cnsList := cnsRaw.([]interface{})
			cnsMap, ok := cnsList[0].(map[string]interface{})
			if len(cnsList) > 0 && ok {
				for k, v := range cnsMap {
					switch k {
					case "disable":
						b := v.(bool)
						if b {
							cns.Disable = b
						}
					case "services":
						servicesRaw := v.([]interface{})
						cns.Services = make([]string, 0, len(servicesRaw))
						for _, serviceRaw := range servicesRaw {
							cns.Services = append(cns.Services, serviceRaw.(string))
						}
					default:
						return fmt.Errorf("unsupported CNS attribute %q", k)
					}
				}
			}
		}

		var err error
		if len(tags) == 0 && len(cns.Services) == 0 {
			err = c.Instances().DeleteTags(context.Background(), &compute.DeleteTagsInput{
				ID: d.Id(),
			})
		} else {
			err = c.Instances().ReplaceTags(context.Background(), &compute.ReplaceTagsInput{
				ID:   d.Id(),
				Tags: tags,
				CNS:  cns,
			})
		}
		if err != nil {
			return err
		}

		expectedTags, err := hashstructure.Hash([]interface{}{tags, cns, true}, nil)
		if err != nil {
			return err
		}
		stateConf := &resource.StateChangeConf{
			Target: []string{strconv.FormatUint(expectedTags, 10)},
			Refresh: func() (interface{}, string, error) {
				inst, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
					ID: d.Id(),
				})
				if err != nil {
					return nil, "", err
				}

				domainCheck := hasValidDomainNames(d, inst)
				hashTags, err := hashstructure.Hash([]interface{}{inst.Tags, inst.CNS, domainCheck}, nil)
				if err != nil {
					return nil, "", err
				}
				return inst, strconv.FormatUint(hashTags, 10), nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}

		if d.HasChange("tags") {
			d.SetPartial("tags")
		}
		if d.HasChange("cns") {
			d.SetPartial("cns")
		}
	}

	if d.HasChange("package") {
		newPackage := d.Get("package").(string)

		err := c.Instances().Resize(context.Background(), &compute.ResizeInstanceInput{
			ID:      d.Id(),
			Package: newPackage,
		})
		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Target: []string{fmt.Sprintf("%s@%s", newPackage, "running")},
			Refresh: func() (interface{}, string, error) {
				inst, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
					ID: d.Id(),
				})
				if err != nil {
					return nil, "", err
				}

				return inst, fmt.Sprintf("%s@%s", inst.Package, inst.State), nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}

		d.SetPartial("package")
	}

	if d.HasChange("firewall_enabled") {
		enable := d.Get("firewall_enabled").(bool)

		var err error
		if enable {
			err = c.Instances().EnableFirewall(context.Background(), &compute.EnableFirewallInput{
				ID: d.Id(),
			})
		} else {
			err = c.Instances().DisableFirewall(context.Background(), &compute.DisableFirewallInput{
				ID: d.Id(),
			})
		}
		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Target: []string{fmt.Sprintf("%t", enable)},
			Refresh: func() (interface{}, string, error) {
				inst, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
					ID: d.Id(),
				})
				if err != nil {
					return nil, "", err
				}

				return inst, fmt.Sprintf("%t", inst.FirewallEnabled), nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}

		d.SetPartial("firewall_enabled")
	}

	if d.HasChange("networks") {

		nics, err := c.Instances().ListNICs(context.Background(), &compute.ListNICsInput{
			InstanceID: d.Id(),
		})
		if err != nil {
			return err
		}

		oRaw, nRaw := d.GetChange("networks")
		o := oRaw.([]interface{})
		n := nRaw.([]interface{})

		for _, new := range n {
			var nicId string

			for _, old := range o {
				exists := false
				nicId = old.(string)
				if old.(string) == new.(string) {
					exists = true
				}
				if !exists {
					var macId string
					for _, nic := range nics {
						if nic.Network == nicId {
							macId = nic.MAC
						}
					}

					log.Printf("[DEBUG] Removing NIC with MacId %s", macId)
					_, err := retryOnError(compute.IsResourceFound, func() (interface{}, error) {
						err := c.Instances().RemoveNIC(context.Background(), &compute.RemoveNICInput{
							InstanceID: d.Id(),
							MAC:        macId,
						})
						return nil, err
					})
					if err != nil {
						return err
					}
				}
			}
		}

		for _, old := range o {

			for _, new := range n {
				exists := false
				if old.(string) == new.(string) {
					exists = true
				}
				if !exists {

					log.Printf("[DEBUG] Adding NIC with Network %s", new.(string))
					_, err := retryOnError(compute.IsResourceFound, func() (interface{}, error) {
						_, err := c.Instances().AddNIC(context.Background(), &compute.AddNICInput{
							InstanceID: d.Id(),
							Network:    new.(string),
						})
						return nil, err
					})
					if err != nil {
						return err
					}
				}
			}
		}

		d.SetPartial("networks")
	}

	metadata := map[string]string{}
	for k, v := range d.Get("metadata").(map[string]interface{}) {
		metadata[k] = v.(string)
	}
	if d.HasChange("metadata") {
		oldValue, newValue := d.GetChange("metadata")
		newMetadata := newValue.(map[string]interface{})
		for k, _ := range oldValue.(map[string]interface{}) {
			if _, ok := newMetadata[k]; !ok {
				if err := c.Instances().DeleteMetadata(context.Background(), &compute.DeleteMetadataInput{
					ID:  d.Id(),
					Key: k,
				}); err != nil {
					return err
				}
			}
		}
	}
	for argumentName, metadataKey := range metadataArgumentsToKeys {
		if val, ok := d.GetOk(argumentName); ok {
			metadata[metadataKey] = val.(string)
		} else {
			if d.HasChange(argumentName) {
				if err := c.Instances().DeleteMetadata(context.Background(), &compute.DeleteMetadataInput{
					ID:  d.Id(),
					Key: metadataKey,
				}); err != nil {
					return err
				}
			}
		}
	}

	if len(metadata) > 0 {
		if _, err := c.Instances().UpdateMetadata(context.Background(), &compute.UpdateMetadataInput{
			ID:       d.Id(),
			Metadata: metadata,
		}); err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Target: []string{"converged"},
			Refresh: func() (interface{}, string, error) {
				inst, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
					ID: d.Id(),
				})
				if err != nil {
					return nil, "", err
				}

				for k, v := range metadata {
					if upstream, ok := inst.Metadata[k]; !ok || v != upstream {
						return inst, "converging", nil
					}
				}

				return inst, "converged", nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err := stateConf.WaitForState()
		if err != nil {
			return err
		}

		for argumentName := range metadataArgumentsToKeys {
			if d.HasChange(argumentName) {
				d.SetPartial(argumentName)
			}
		}
	}

	d.Partial(false)

	return resourceMachineRead(d, meta)
}

func resourceMachineDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	c, err := client.Compute()
	if err != nil {
		return err
	}

	err = c.Instances().Delete(context.Background(), &compute.DeleteInstanceInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Target: []string{machineStateDeleted},
		Refresh: func() (interface{}, string, error) {
			inst, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
				ID: d.Id(),
			})
			if err != nil {
				if compute.IsResourceNotFound(err) {
					return inst, "deleted", nil
				}
				return nil, "", err
			}

			return inst, inst.State, nil
		},
		Timeout:    machineStateChangeTimeout,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return nil
}

func resourceMachineValidateName(value interface{}, name string) (warnings []string, errors []error) {
	warnings = []string{}
	errors = []error{}

	r := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\_\.\-]*$`)
	if !r.Match([]byte(value.(string))) {
		errors = append(errors, fmt.Errorf(`"%s" is not a valid %s`, value.(string), name))
	}

	return warnings, errors
}

// castToTypeList casts an interface slice back into a proper slice of
// strings. This handles pulling services out of various nested interface
// collections that Terraform stores them under.
func castToTypeList(sliceRaw interface{}) []string {
	slice := sliceRaw.([]interface{})
	result := make([]string, len(slice))
	for iter, member := range slice {
		result[iter] = fmt.Sprint(member)
	}
	return result
}

// hasValidDomainNames makes sure domain names have converged for various
// reasons. This could be because CNS services have been added, changed, or
// disabled. We could also have nothing to do with CNS and we only need to
// validate our normal instance domains.
//
// This helps store the proper converged domain names in our
// state file that match our instance name and CNS services tag.
func hasValidDomainNames(d *schema.ResourceData, inst *compute.Instance) bool {
	// If we no longer have CNS than we don't want empty domain names.
	if _, hasCNS := d.GetOk("cns"); !hasCNS {
		if len(inst.DomainNames) == 0 {
			return false
		}
	}

	// If CNS has been disabled than we need domain names.
	disableRaw := d.Get("cns.0.disable")
	disabled := disableRaw.(bool)
	if disabled {
		if len(inst.DomainNames) != 0 {
			return false
		}
	} else {
		oldCNS, newCNS := d.GetChange("cns.0.services")
		// Index domains so we O(1) our checks
		domains := map[string]bool{}
		for _, domain := range inst.DomainNames {
			name := strings.Split(domain, ".")[0]
			domains[name] = true
		}

		// check domains for new services that are missing
		checked := map[string]bool{}
		newServices := castToTypeList(newCNS)
		for _, newService := range newServices {
			checked[newService] = true
			if _, exists := domains[newService]; !exists {
				return false
			}
		}

		oldServices := castToTypeList(oldCNS)
		// check domains for any services that have expired
		for _, oldService := range oldServices {
			if _, exists := domains[oldService]; exists {
				if _, already := checked[oldService]; !already {
					return false
				}
			}
		}
	}
	return true
}

// hasInitDomainNames makes sure domain names have propagated properly after
// provisioning a new instance. See hasValidDomainNames.
//
// This helps store the proper converged domain names in our
// state file that match our instance name and CNS services tag.
func hasInitDomainNames(d *schema.ResourceData, inst *compute.Instance) bool {
	// If we don't have CNS than we also don't want empty domain names
	if _, hasCNS := d.GetOk("cns"); !hasCNS {
		if len(inst.DomainNames) > 0 {
			return true
		}
	}
	servicesRaw, hasServices := d.GetOk("cns.0.services")
	newServices := castToTypeList(servicesRaw)
	if hasServices {
		// Index domains so we O(1) our checks
		domains := map[string]bool{}
		for _, domain := range inst.DomainNames {
			name := strings.Split(domain, ".")[0]
			domains[name] = true
		}
		// check domains for new services that are missing
		for _, newService := range newServices {
			if _, exists := domains[newService]; !exists {
				return false
			}
		}
	} else {
		if len(inst.DomainNames) == 0 {
			return false
		}
	}
	return true
}
