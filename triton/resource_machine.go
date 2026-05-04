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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// InstanceCNS is a local struct representing CNS configuration.
// The new CloudAPI client has no equivalent — CNS is managed via machine tags
// at the API level — but we keep this for the Terraform schema.
type InstanceCNS struct {
	Disable  bool
	Services []string
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
			"cns": {
				Description: "Container Name Service",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable": {
							Description: "Disable CNS for this instance (after create)",
							Optional:    true,
							Computed:    true,
							Type:        schema.TypeBool,
						},
						"services": {
							Description: "Assign CNS service names to this instance",
							Optional:    true,
							Computed:    true,
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
				Deprecated:  "`locality` was replaced by `affinity` in the underlying Triton API.",
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

			"networks": {
				Description: "Desired network IDs",
				Type:        schema.TypeSet,
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
					return hashcodeString(m["network"].(string))
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
			"deletion_protection_enabled": {
				Description: "Whether to enable deletion protection for this machine",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"delegate_dataset": {
				Description: "Whether to create a delegate dataset for this machine",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"volume": {
				Description: "Volume to attach to the machine",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Computed:    true,
							Description: "The volume attachment mode",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"mountpoint": {
							Description: "Where to attach the volume",
							Required:    true,
							Type:        schema.TypeString,
						},
						"name": {
							Description: "The name of the volume",
							Required:    true,
							Type:        schema.TypeString,
						},
						"type": {
							Computed:    true,
							Description: "The type of volume",
							Optional:    true,
							Type:        schema.TypeString,
						},
					},
				},
			},

			// Metadata and Tags
			"tags": {
				Description: "Machine tags",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"metadata": {
				Description: "Machine metadata",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"root_authorized_keys": {
				Description: "Authorized keys for the root user on this machine",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
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

			// Instance computed parameters
			"state": {
				Description: "Provisioning state of the instance",
				Type:        schema.TypeString,
				Computed:    true,
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
			"primaryip": {
				Description: "Primary (public) IP address for the machine",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"domain_names": {
				Description: "List of domain names from Triton CNS",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
			"compute_node": {
				Description: "UUID of the server on which the instance is located",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func machineStateString(m *cloudapi.Machine) string {
	return string(m.State)
}

func machineTypeString(m *cloudapi.Machine) string {
	v, err := m.Type.AsMachineType0()
	if err != nil {
		v1, err2 := m.Type.AsMachineType1()
		if err2 != nil {
			log.Printf("[WARN] machineTypeString: failed to decode both union branches (type0: %s, type1: %s)", err, err2)
			return ""
		}
		return string(v1)
	}
	return string(v)
}

// resolvePackageValue reconciles the package value returned by CloudAPI
// (always a name) with the value in the Terraform config (which may be a
// UUID).  When the config uses a UUID, we look up the package by name to
// obtain its UUID so the state matches the config and avoids a perpetual
// diff.  When the config uses a name we return apiName unchanged.
func resolvePackageValue(client *Client, apiName, configVal string) (string, error) {
	if !isUUID(configVal) {
		// Config value is a name — no translation needed.
		return apiName, nil
	}

	resp, err := client.API().GetPackageWithResponse(
		context.Background(), client.Account(), apiName)
	if err != nil {
		return apiName, err
	}
	if resp.JSON200 == nil {
		return apiName, fmt.Errorf(
			"error looking up package %q: %s",
			apiName, formatAPIError(resp.StatusCode(), resp.Body))
	}
	return uuidString(resp.JSON200.ID), nil
}

// resolvePackageName is the inverse of resolvePackageValue: given a
// config value that may be a UUID, return the package name so it can
// be compared against the name that CloudAPI returns.
func resolvePackageName(client *Client, configVal string) (string, error) {
	if !isUUID(configVal) {
		// Already a name.
		return configVal, nil
	}

	resp, err := client.API().GetPackageWithResponse(
		context.Background(), client.Account(), configVal)
	if err != nil {
		return configVal, err
	}
	if resp.JSON200 == nil {
		return configVal, fmt.Errorf(
			"error looking up package %q: %s",
			configVal, formatAPIError(resp.StatusCode(), resp.Body))
	}
	return resp.JSON200.Name, nil
}

func resourceMachineCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	var affinity []string
	for _, rule := range d.Get("affinity").([]interface{}) {
		affinity = append(affinity, rule.(string))
	}

	if len(affinity) > 0 {
		client.affinityLock.Lock()
		defer client.affinityLock.Unlock()
	}

	var networks []cloudapi.NetworkObject
	for _, network := range d.Get("networks").(*schema.Set).List() {
		netUUID, err := parseUUID(network.(string))
		if err != nil {
			return fmt.Errorf("invalid network UUID: %s", err)
		}
		// CloudAPI's CreateMachine expects networks as "ipv4_uuid" objects.
		networks = append(networks, cloudapi.NetworkObject{Ipv4UUID: netUUID})
	}

	metadata := map[string]interface{}{}
	for k, v := range d.Get("metadata").(map[string]interface{}) {
		metadata[k] = v
	}
	for argumentName, metadataKey := range metadataArgumentsToKeys {
		if v, ok := d.GetOk(argumentName); ok {
			metadata[metadataKey] = v.(string)
		}
	}

	tags := map[string]interface{}{}
	for k, v := range d.Get("tags").(map[string]interface{}) {
		tags[k] = v
	}

	cns := parseCNSFromSchema(d)
	injectCNSIntoTags(cns, tags)

	imageUUID, err := parseUUID(d.Get("image").(string))
	if err != nil {
		return fmt.Errorf("invalid image UUID: %s", err)
	}

	firewallEnabled := d.Get("firewall_enabled").(bool)
	delegateDataset := d.Get("delegate_dataset").(bool)
	machineName := d.Get("name").(string)

	var volumes []cloudapi.VolumeMount
	if volumesRaw, ok := d.GetOk("volume"); ok {
		volumesList := volumesRaw.(*schema.Set).List()
		for _, v := range volumesList {
			volumeMap, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid volume entry")
			}
			volumeName, ok := volumeMap["name"].(string)
			if !ok {
				return fmt.Errorf("volume entries must specify the volume name")
			}
			vol := cloudapi.VolumeMount{
				Name:       volumeName,
				Mountpoint: volumeMap["mountpoint"].(string),
			}
			if mode, ok := volumeMap["mode"].(string); ok && mode != "" {
				m := cloudapi.MountMode{}
				if err := m.FromMountMode0(cloudapi.MountMode0(mode)); err != nil {
					return fmt.Errorf("invalid volume mount mode: %s", err)
				}
				vol.Mode = &m
			}
			if vtype, ok := volumeMap["type"].(string); ok && vtype != "" {
				vt := cloudapi.VolumeType{}
				if err := vt.FromVolumeType0(cloudapi.VolumeType0(vtype)); err != nil {
					return fmt.Errorf("invalid volume type: %s", err)
				}
				vol.Type = &vt
			}
			volumes = append(volumes, vol)
		}
	}

	// CloudAPI's CreateMachine handler uses restify mapParams and
	// extracts tags/metadata via /^tag\..+/ and /^metadata\..+/
	// regexes on the flattened request parameters.  The OpenAPI
	// struct's nested "tags"/"metadata" objects get flattened to
	// "tags.key" and "metadata.key" which don't match — so we
	// build the request body as a flat map using the legacy format.
	createBody := map[string]interface{}{
		"name":             machineName,
		"package":          d.Get("package").(string),
		"image":            uuidString(imageUUID),
		"firewall_enabled": firewallEnabled,
		"delegate_dataset": delegateDataset,
	}
	if len(networks) > 0 {
		createBody["networks"] = networks
	}
	for k, v := range metadata {
		createBody[fmt.Sprintf("metadata.%s", k)] = v
	}
	for k, v := range tags {
		createBody[fmt.Sprintf("tag.%s", k)] = v
	}
	for _, a := range affinity {
		// CloudAPI accepts affinity as an array.
		if createBody["affinity"] == nil {
			createBody["affinity"] = []string{}
		}
		createBody["affinity"] = append(createBody["affinity"].([]string), a)
	}
	if len(volumes) > 0 {
		createBody["volumes"] = volumes
	}

	if nearRaw, found := d.GetOk("locality.0.close_to"); found {
		nearList := nearRaw.([]interface{})
		localNear := make([]string, len(nearList))
		for i, val := range nearList {
			localNear[i] = val.(string)
		}
		locality := map[string]interface{}{"near": localNear}
		if farRaw, found2 := d.GetOk("locality.0.far_from"); found2 {
			farList := farRaw.([]interface{})
			localFar := make([]string, len(farList))
			for i, val := range farList {
				localFar[i] = val.(string)
			}
			locality["far"] = localFar
		}
		createBody["locality"] = locality
	} else if farRaw, found := d.GetOk("locality.0.far_from"); found {
		farList := farRaw.([]interface{})
		localFar := make([]string, len(farList))
		for i, val := range farList {
			localFar[i] = val.(string)
		}
		createBody["locality"] = map[string]interface{}{"far": localFar}
	}

	bodyJSON, err := json.Marshal(createBody)
	if err != nil {
		return fmt.Errorf("error encoding create machine request: %s", err)
	}
	resp, err := client.API().CreateMachineWithBodyWithResponse(
		context.Background(), client.Account(),
		"application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("error creating machine: %s", err)
	}
	if resp.JSON201 == nil {
		return fmt.Errorf("error creating machine: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	machineUUID := resp.JSON201.ID
	d.SetId(uuidString(machineUUID))

	stateConf := &retry.StateChangeConf{
		Target: []string{machineStateRunning},
		Refresh: func() (interface{}, string, error) {
			r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
			if err != nil {
				return nil, "", err
			}
			if r.JSON200 == nil {
				return nil, "", fmt.Errorf("error polling machine: %s", formatAPIError(r.StatusCode(), r.Body))
			}
			if machineStateString(r.JSON200) == machineStateFailed {
				d.SetId("")
				return nil, "", fmt.Errorf("instance creation failed: %s", r.JSON200.State)
			}
			return r.JSON200, machineStateString(r.JSON200), nil
		},
		Timeout:    machineStateChangeTimeout,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	// Wait for at least one base domain name (e.g.
	// <name>.inst.<account>.<dc>.triton.zone) to propagate before
	// reading state.  On create we only require base names;
	// service-specific CNS names may take longer and will converge
	// on subsequent reads or updates.
	//
	// Only wait when the account has CNS enabled — without CNS,
	// dns_names will never be populated.
	cnsEnabled, err := client.CNSEnabled()
	if err != nil {
		return err
	}
	if cnsEnabled {
		if err := waitForBaseDomainNames(d, client); err != nil {
			return err
		}
	}

	return resourceMachineUpdate(d, meta)
}

func resourceMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)

	machineUUID, err := parseUUID(d.Id())
	if err != nil {
		return false, fmt.Errorf("invalid machine ID: %s", err)
	}

	resp, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
	if err != nil {
		return false, err
	}
	if isNotFound(resp.StatusCode()) {
		return false, nil
	}
	if resp.JSON200 == nil {
		return false, fmt.Errorf("error checking machine: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	return true, nil
}

func resourceMachineRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	machineUUID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid machine ID: %s", err)
	}

	resp, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
	if err != nil {
		return err
	}

	if isNotFound(resp.StatusCode()) {
		log.Printf("Instance %q not found or has been deleted", d.Id())
		d.SetId("")
		return nil
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("error reading machine: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	machine := resp.JSON200

	if machineStateString(machine) == machineStateFailed {
		log.Printf("Instance %q state: `failed` so removing from state", d.Id())
		d.SetId("")
		return nil
	}

	nicsResp, err := client.API().ListNicsWithResponse(context.Background(), client.Account(), machineUUID)
	if err != nil {
		return err
	}
	if nicsResp.JSON200 == nil {
		return fmt.Errorf("error listing NICs: %s", formatAPIError(nicsResp.StatusCode(), nicsResp.Body))
	}

	cns := parseCNSFromMachineTags(machine.Tags)
	cnsRaw := castToSliceRaw(cns)
	d.Set("cns", cnsRaw)

	d.Set("name", machine.Name)
	d.Set("type", machineTypeString(machine))
	d.Set("state", machineStateString(machine))
	d.Set("dataset", uuidString(machine.Image))
	d.Set("image", uuidString(machine.Image))
	d.Set("memory", int(derefUint64(machine.Memory)))
	d.Set("disk", int(machine.Disk))
	d.Set("ips", machine.Ips)

	// Strip CNS tags from user-visible tags
	userTags := map[string]interface{}{}
	for k, v := range machine.Tags {
		if !strings.HasPrefix(k, "triton.cns") {
			userTags[k] = v
		}
	}
	d.Set("tags", userTags)

	d.Set("created", machine.Created.Format(time.RFC3339))
	d.Set("updated", machine.Updated.Format(time.RFC3339))
	pkgVal, err := resolvePackageValue(client, machine.Package, d.Get("package").(string))
	if err != nil {
		log.Printf("[WARN] unable to resolve package UUID: %s", err)
		pkgVal = machine.Package
	}
	d.Set("package", pkgVal)
	d.Set("primaryip", derefString(machine.PrimaryIP))
	d.Set("firewall_enabled", derefBool(machine.FirewallEnabled))
	d.Set("domain_names", derefStringSlice(machine.DNSNames))

	computeNode := ""
	if machine.ComputeNode != nil {
		computeNode = uuidString(*machine.ComputeNode)
	}
	d.Set("compute_node", computeNode)
	d.Set("deletion_protection_enabled", derefBool(machine.DeletionProtection))
	d.Set("delegate_dataset", derefBool(machine.DelegateDataset))

	// Regression guard (#30/#35): populate both "nic" (computed) and
	// "networks" (input) from actual NICs to avoid perpetual diffs with
	// network pools.
	var (
		machineNICs []map[string]interface{}
		networkList []string
	)
	for _, nic := range *nicsResp.JSON200 {
		nicState := ""
		if nic.State != nil {
			nicState = string(*nic.State)
		}
		machineNICs = append(
			machineNICs,
			map[string]interface{}{
				"ip":      nic.IP,
				"mac":     nic.Mac,
				"primary": nic.Primary,
				"netmask": nic.Netmask,
				"gateway": derefString(nic.Gateway),
				"state":   nicState,
				"network": uuidString(nic.Network),
			},
		)
		networkList = append(networkList, uuidString(nic.Network))
	}
	d.Set("nic", machineNICs)
	d.Set("networks", networkList)

	for argumentName, metadataKey := range metadataArgumentsToKeys {
		if machine.Metadata != nil {
			if val, ok := machine.Metadata[metadataKey]; ok {
				d.Set(argumentName, fmt.Sprintf("%v", val))
			} else {
				d.Set(argumentName, "")
			}
			delete(machine.Metadata, metadataKey)
		}
	}

	// Convert metadata values to strings for Terraform
	if machine.Metadata != nil {
		md := map[string]string{}
		for k, v := range machine.Metadata {
			md[k] = fmt.Sprintf("%v", v)
		}
		d.Set("metadata", md)
	}

	primaryIP := derefString(machine.PrimaryIP)
	if primaryIP != "" {
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": primaryIP,
		})
	}

	return nil
}

func resourceMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	machineUUID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid machine ID: %s", err)
	}

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		oldNameInterface, newNameInterface := d.GetChange("name")
		oldName := oldNameInterface.(string)
		newName := newNameInterface.(string)

		if err := client.Typed().RenameMachine(context.Background(), client.Account(), machineUUID,
			cloudapi.RenameMachineRequest{Name: newName}); err != nil {
			return fmt.Errorf("error renaming machine: %s", err)
		}

		stateConf := &retry.StateChangeConf{
			Pending: []string{oldName},
			Target:  []string{newName},
			Refresh: func() (interface{}, string, error) {
				r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
				if err != nil {
					return nil, "", err
				}
				if r.JSON200 == nil {
					return nil, "", fmt.Errorf("error polling machine: %s", formatAPIError(r.StatusCode(), r.Body))
				}
				return r.JSON200, r.JSON200.Name, nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	if (d.HasChange("tags") || d.HasChange("cns")) && !d.IsNewResource() {
		tags := map[string]interface{}{}
		for k, v := range d.Get("tags").(map[string]interface{}) {
			if strings.HasPrefix(k, "triton.cns") {
				continue
			}
			tags[k] = v
		}

		cns := parseCNSFromSchema(d)

		// Compute the expected hash BEFORE injectCNSIntoTags mutates
		// the tags map.  The refresh loop filters CNS out of instTags,
		// so the expected hash must also use user-only tags.
		// Domain name convergence is handled separately by
		// waitForDomainNames after this loop.
		expectedTags, err := hashstructure.Hash([]interface{}{tags, cns}, nil)
		if err != nil {
			return err
		}

		injectCNSIntoTags(cns, tags)

		if len(tags) == 0 {
			resp, err := client.API().DeleteMachineTagsWithResponse(context.Background(), client.Account(), machineUUID)
			if err != nil {
				return fmt.Errorf("error deleting tags: %s", err)
			}
			if resp.StatusCode() >= 400 {
				return fmt.Errorf("error deleting tags: %s", formatAPIError(resp.StatusCode(), resp.Body))
			}
		} else {
			resp, err := client.API().ReplaceMachineTagsWithResponse(context.Background(), client.Account(), machineUUID,
				cloudapi.ReplaceMachineTagsJSONRequestBody(tags))
			if err != nil {
				return fmt.Errorf("error replacing tags: %s", err)
			}
			if resp.StatusCode() >= 400 {
				return fmt.Errorf("error replacing tags: %s", formatAPIError(resp.StatusCode(), resp.Body))
			}
		}

		// expectedTags was computed above before injectCNSIntoTags.
		stateConf := &retry.StateChangeConf{
			Target: []string{strconv.FormatUint(expectedTags, 10)},
			Refresh: func() (interface{}, string, error) {
				r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
				if err != nil {
					return nil, "", err
				}
				if r.JSON200 == nil {
					return nil, "", fmt.Errorf("error polling machine: %s", formatAPIError(r.StatusCode(), r.Body))
				}

				instCNS := parseCNSFromMachineTags(r.JSON200.Tags)
				// Filter out CNS tags for comparison
				instTags := map[string]interface{}{}
				for k, v := range r.JSON200.Tags {
					if !strings.HasPrefix(k, "triton.cns") {
						instTags[k] = v
					}
				}
				hashTags, err := hashstructure.Hash([]interface{}{instTags, instCNS}, nil)
				if err != nil {
					return nil, "", err
				}
				return r.JSON200, strconv.FormatUint(hashTags, 10), nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}

		// Wait for CNS domain names to converge separately from tags.
		// Tags update quickly but DNS propagation may be slower.
		// Skip when the account does not have CNS enabled.
		cnsEnabled, err := client.CNSEnabled()
		if err != nil {
			return err
		}
		if cnsEnabled {
			if err := waitForDomainNames(d, client); err != nil {
				return err
			}
		}
	}

	if d.HasChange("package") && !d.IsNewResource() {
		newPackage := d.Get("package").(string)

		if err := client.Typed().ResizeMachine(context.Background(), client.Account(), machineUUID,
			cloudapi.ResizeMachineRequest{Package: newPackage}); err != nil {
			return fmt.Errorf("error resizing machine: %s", err)
		}

		// The polling target must use the package name because
		// GetMachine always returns the name, not the UUID.
		targetPkgName, err := resolvePackageName(client, newPackage)
		if err != nil {
			return fmt.Errorf("error resolving package name: %s", err)
		}

		stateConf := &retry.StateChangeConf{
			Target: []string{fmt.Sprintf("%s@%s", targetPkgName, "running")},
			Refresh: func() (interface{}, string, error) {
				r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
				if err != nil {
					return nil, "", err
				}
				if r.JSON200 == nil {
					return nil, "", fmt.Errorf("error polling machine: %s", formatAPIError(r.StatusCode(), r.Body))
				}
				return r.JSON200, fmt.Sprintf("%s@%s", r.JSON200.Package, machineStateString(r.JSON200)), nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	if d.HasChange("firewall_enabled") && !d.IsNewResource() {
		enable := d.Get("firewall_enabled").(bool)

		var err error
		if enable {
			err = client.Typed().EnableFirewall(context.Background(), client.Account(), machineUUID, cloudapi.EnableFirewallRequest{})
		} else {
			err = client.Typed().DisableFirewall(context.Background(), client.Account(), machineUUID, cloudapi.DisableFirewallRequest{})
		}
		if err != nil {
			return fmt.Errorf("error updating firewall: %s", err)
		}

		stateConf := &retry.StateChangeConf{
			Target: []string{fmt.Sprintf("%t", enable)},
			Refresh: func() (interface{}, string, error) {
				r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
				if err != nil {
					return nil, "", err
				}
				if r.JSON200 == nil {
					return nil, "", fmt.Errorf("error polling machine: %s", formatAPIError(r.StatusCode(), r.Body))
				}
				return r.JSON200, fmt.Sprintf("%t", derefBool(r.JSON200.FirewallEnabled)), nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	// Regression guard (#104): poll NIC state after add.
	if d.HasChange("networks") && !d.IsNewResource() {
		nicsResp, err := client.API().ListNicsWithResponse(context.Background(), client.Account(), machineUUID)
		if err != nil {
			return err
		}
		if nicsResp.JSON200 == nil {
			return fmt.Errorf("error listing NICs: %s", formatAPIError(nicsResp.StatusCode(), nicsResp.Body))
		}
		nics := *nicsResp.JSON200

		oRaw, nRaw := d.GetChange("networks")
		o := oRaw.(*schema.Set).List()
		n := nRaw.(*schema.Set).List()

		networksToRemove := differenceNetworks(o, n)
		for _, toRemove := range networksToRemove {
			var macId string
			for _, nic := range nics {
				if uuidString(nic.Network) == toRemove {
					macId = nic.Mac
					break
				}
			}

			if macId != "" {
				log.Printf("[DEBUG] Removing NIC with MacId %s", macId)
				resp, err := client.API().RemoveNicWithResponse(context.Background(), client.Account(), machineUUID, macId)
				if err != nil {
					return fmt.Errorf("error removing NIC: %s", err)
				}
				if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
					return fmt.Errorf("error removing NIC: %s", formatAPIError(resp.StatusCode(), resp.Body))
				}
			}
		}

		networksToAdd := differenceNetworks(n, o)
		for _, toAdd := range networksToAdd {
			log.Printf("[DEBUG] Adding NIC with Network %s", toAdd)
			addUUID, err := parseUUID(toAdd)
			if err != nil {
				return fmt.Errorf("invalid network UUID: %s", err)
			}

			addResp, err := client.API().AddNicWithResponse(context.Background(), client.Account(), machineUUID,
				cloudapi.AddNicJSONRequestBody{Network: addUUID})
			if err != nil {
				return fmt.Errorf("error adding NIC: %s", err)
			}
			if addResp.JSON201 == nil {
				return fmt.Errorf("error adding NIC: %s", formatAPIError(addResp.StatusCode(), addResp.Body))
			}

			log.Printf("[DEBUG] NIC added, MAC %s; waiting for machine to return to running", addResp.JSON201.Mac)

			stateConf := &retry.StateChangeConf{
				Target: []string{machineStateRunning},
				Refresh: func() (interface{}, string, error) {
					r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
					if err != nil {
						return nil, "", err
					}
					if r.JSON200 == nil {
						return nil, "", fmt.Errorf("error polling machine after NIC add: %s", formatAPIError(r.StatusCode(), r.Body))
					}
					return r.JSON200, machineStateString(r.JSON200), nil
				},
				Timeout:    machineStateChangeTimeout,
				MinTimeout: 3 * time.Second,
			}
			_, err = stateConf.WaitForState()
			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("deletion_protection_enabled") {
		deletionProtection := d.Get("deletion_protection_enabled").(bool)

		var err error
		if deletionProtection {
			log.Printf("[INFO] Enabling Deletion Protection for %q", d.Id())
			err = client.Typed().EnableDeletionProtection(context.Background(), client.Account(), machineUUID, cloudapi.EnableDeletionProtectionRequest{})
		} else {
			log.Printf("[INFO] Disabling Deletion Protection for %q", d.Id())
			err = client.Typed().DisableDeletionProtection(context.Background(), client.Account(), machineUUID, cloudapi.DisableDeletionProtectionRequest{})
		}
		if err != nil {
			return fmt.Errorf("error updating deletion protection: %s", err)
		}

		stateConf := &retry.StateChangeConf{
			Target: []string{fmt.Sprintf("%t", deletionProtection)},
			Refresh: func() (interface{}, string, error) {
				r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
				if err != nil {
					return nil, "", err
				}
				if r.JSON200 == nil {
					return nil, "", fmt.Errorf("error polling machine: %s", formatAPIError(r.StatusCode(), r.Body))
				}
				return r.JSON200, fmt.Sprintf("%t", derefBool(r.JSON200.DeletionProtection)), nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	metadata := map[string]interface{}{}
	for k, v := range d.Get("metadata").(map[string]interface{}) {
		metadata[k] = v
	}
	if d.HasChange("metadata") && !d.IsNewResource() {
		oldValue, newValue := d.GetChange("metadata")
		newMetadata := newValue.(map[string]interface{})
		for k := range oldValue.(map[string]interface{}) {
			if _, ok := newMetadata[k]; !ok {
				resp, err := client.API().DeleteMachineMetadataWithResponse(context.Background(), client.Account(), machineUUID, k)
				if err != nil {
					return fmt.Errorf("error deleting metadata key %q: %s", k, err)
				}
				if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
					return fmt.Errorf("error deleting metadata key %q: %s", k, formatAPIError(resp.StatusCode(), resp.Body))
				}
			}
		}
	}
	for argumentName, metadataKey := range metadataArgumentsToKeys {
		if val, ok := d.GetOk(argumentName); ok {
			metadata[metadataKey] = val.(string)
		} else {
			if d.HasChange(argumentName) {
				resp, err := client.API().DeleteMachineMetadataWithResponse(context.Background(), client.Account(), machineUUID, metadataKey)
				if err != nil {
					return fmt.Errorf("error deleting metadata key %q: %s", metadataKey, err)
				}
				if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
					return fmt.Errorf("error deleting metadata key %q: %s", metadataKey, formatAPIError(resp.StatusCode(), resp.Body))
				}
			}
		}
	}

	if len(metadata) > 0 {
		resp, err := client.API().AddMachineMetadataWithResponse(context.Background(), client.Account(), machineUUID,
			cloudapi.AddMachineMetadataJSONRequestBody(metadata))
		if err != nil {
			return fmt.Errorf("error updating metadata: %s", err)
		}
		if resp.StatusCode() >= 400 {
			return fmt.Errorf("error updating metadata: %s", formatAPIError(resp.StatusCode(), resp.Body))
		}

		stateConf := &retry.StateChangeConf{
			Target: []string{"converged"},
			Refresh: func() (interface{}, string, error) {
				r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
				if err != nil {
					return nil, "", err
				}
				if r.JSON200 == nil {
					return nil, "", fmt.Errorf("error polling machine: %s", formatAPIError(r.StatusCode(), r.Body))
				}

				for k, v := range metadata {
					vStr := fmt.Sprintf("%v", v)
					if upstream, ok := r.JSON200.Metadata[k]; !ok || fmt.Sprintf("%v", upstream) != vStr {
						return r.JSON200, "converging", nil
					}
				}

				return r.JSON200, "converged", nil
			},
			Timeout:    machineStateChangeTimeout,
			MinTimeout: 3 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	d.Partial(false)

	return resourceMachineRead(d, meta)
}

func resourceMachineDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	machineUUID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid machine ID: %s", err)
	}

	// CloudAPI handles stopping running instances internally on DELETE.
	resp, err := client.API().DeleteMachineWithResponse(context.Background(), client.Account(), machineUUID)
	if err != nil {
		return fmt.Errorf("error deleting machine: %s", err)
	}
	if resp.StatusCode() >= 400 && !isNotFound(resp.StatusCode()) {
		return fmt.Errorf("error deleting machine: %s", formatAPIError(resp.StatusCode(), resp.Body))
	}

	stateConf := &retry.StateChangeConf{
		Target: []string{machineStateDeleted},
		Refresh: func() (interface{}, string, error) {
			r, err := client.API().GetMachineWithResponse(context.Background(), client.Account(), machineUUID)
			if err != nil {
				return nil, "", err
			}
			if isNotFound(r.StatusCode()) {
				return machineStateDeleted, machineStateDeleted, nil
			}
			if r.JSON200 == nil {
				return nil, "", fmt.Errorf("error polling machine: %s", formatAPIError(r.StatusCode(), r.Body))
			}
			return r.JSON200, machineStateString(r.JSON200), nil
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

// parseCNSFromSchema reads CNS configuration from the Terraform schema.
func parseCNSFromSchema(d *schema.ResourceData) InstanceCNS {
	cns := InstanceCNS{}
	if cnsRaw, found := d.GetOk("cns"); found {
		cnsList := cnsRaw.([]interface{})
		if len(cnsList) > 0 {
			cnsMap, ok := cnsList[0].(map[string]interface{})
			if ok {
				if v, ok := cnsMap["disable"]; ok {
					cns.Disable = v.(bool)
				}
				if v, ok := cnsMap["services"]; ok {
					servicesRaw := v.([]interface{})
					cns.Services = make([]string, 0, len(servicesRaw))
					for _, serviceRaw := range servicesRaw {
						cns.Services = append(cns.Services, serviceRaw.(string))
					}
				}
			}
		}
	}
	return cns
}

// parseCNSFromMachineTags extracts CNS configuration from machine tags.
// CloudAPI may return triton.cns.disable as a JSON boolean or a string
// depending on how the tag was originally set, so we coerce via Sprintf.
func parseCNSFromMachineTags(tags map[string]interface{}) InstanceCNS {
	cns := InstanceCNS{}
	if v, ok := tags["triton.cns.disable"]; ok {
		cns.Disable = fmt.Sprintf("%v", v) == "true"
	}
	if v, ok := tags["triton.cns.services"]; ok {
		svc := fmt.Sprintf("%v", v)
		if svc != "" {
			cns.Services = strings.Split(svc, ",")
		}
	}
	return cns
}

// injectCNSIntoTags writes CNS configuration into a tags map for the API.
func injectCNSIntoTags(cns InstanceCNS, tags map[string]interface{}) {
	if cns.Disable {
		// VMAPI requires triton.cns.disable to be a JSON boolean, not a string.
		tags["triton.cns.disable"] = true
	}
	if len(cns.Services) > 0 {
		tags["triton.cns.services"] = strings.Join(cns.Services, ",")
	}
}

// castToTypeList casts an interface slice back into a proper slice of
// strings.
func castToTypeList(sliceRaw interface{}) []string {
	slice := sliceRaw.([]interface{})
	result := make([]string, len(slice))
	for iter, member := range slice {
		result[iter] = fmt.Sprint(member)
	}
	return result
}

// castToSliceRaw casts an InstanceCNS struct to the interface slice that
// Terraform stores them under.
func castToSliceRaw(input InstanceCNS) []interface{} {
	services := make([]interface{}, len(input.Services))
	for i, s := range input.Services {
		services[i] = s
	}
	return []interface{}{
		map[string]interface{}{
			"disable":  input.Disable,
			"services": services,
		},
	}
}

// hasValidDomainNames makes sure domain names have converged.
// Regression guard (#8): CNS domain names don't provision instantaneously.
func hasValidDomainNames(d *schema.ResourceData, inst *cloudapi.Machine) bool {
	domainNames := derefStringSlice(inst.DNSNames)

	if _, hasCNS := d.GetOk("cns"); !hasCNS {
		// No CNS block configured — domain names are informational
		// and we should not block waiting for them.
		return true
	}

	disableRaw := d.Get("cns.0.disable")
	disabled := disableRaw.(bool)
	if disabled {
		if len(domainNames) != 0 {
			return false
		}
	} else {
		// CNS is enabled — the instance should have at least one base
		// domain name (e.g. <name>.inst.<account>.<dc>.triton.zone) even
		// when no explicit services are configured.
		if len(domainNames) == 0 {
			return false
		}
		oldCNS, newCNS := d.GetChange("cns.0.services")
		domains := map[string]bool{}
		for _, domain := range domainNames {
			name := strings.Split(domain, ".")[0]
			domains[name] = true
		}

		checked := map[string]bool{}
		newServices := castToTypeList(newCNS)
		for _, newService := range newServices {
			newServiceTag := strings.Split(newService, ":")[0]
			checked[newServiceTag] = true
			if _, exists := domains[newServiceTag]; !exists {
				return false
			}
		}

		oldServices := castToTypeList(oldCNS)
		for _, oldService := range oldServices {
			oldServiceTag := strings.Split(oldService, ":")[0]
			if _, exists := domains[oldServiceTag]; exists {
				if _, already := checked[oldServiceTag]; !already {
					return false
				}
			}
		}
	}
	return true
}

// waitForBaseDomainNames polls until the machine has at least one DNS
// name assigned.  This is used during create to ensure the base
// instance domain name (e.g. <name>.inst.<account>.<dc>.triton.zone)
// has propagated before returning state.  It does NOT check for
// service-specific CNS names — those are handled by waitForDomainNames
// in the update path.
func waitForBaseDomainNames(d *schema.ResourceData, client *Client) error {
	machineUUID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid machine ID: %s", err)
	}
	stateConf := &retry.StateChangeConf{
		Target: []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			r, err := client.API().GetMachineWithResponse(
				context.Background(), client.Account(), machineUUID)
			if err != nil {
				return nil, "", err
			}
			if r.JSON200 == nil {
				return nil, "", fmt.Errorf(
					"error polling machine for domain names: %s",
					formatAPIError(r.StatusCode(), r.Body))
			}
			names := derefStringSlice(r.JSON200.DNSNames)
			if len(names) > 0 {
				return r.JSON200, "ready", nil
			}
			return r.JSON200, "waiting", nil
		},
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	return err
}

// waitForDomainNames polls until CNS domain names have converged for
// the given machine.  This is separated from tag convergence to avoid
// conflating fast tag updates with slower DNS propagation.
func waitForDomainNames(d *schema.ResourceData, client *Client) error {
	machineUUID, err := parseUUID(d.Id())
	if err != nil {
		return fmt.Errorf("invalid machine ID: %s", err)
	}
	stateConf := &retry.StateChangeConf{
		Target: []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			r, err := client.API().GetMachineWithResponse(
				context.Background(), client.Account(), machineUUID)
			if err != nil {
				return nil, "", err
			}
			if r.JSON200 == nil {
				return nil, "", fmt.Errorf(
					"error polling machine for domain names: %s",
					formatAPIError(r.StatusCode(), r.Body))
			}
			if hasValidDomainNames(d, r.JSON200) {
				return r.JSON200, "ready", nil
			}
			return r.JSON200, "waiting", nil
		},
		Timeout:    machineStateChangeTimeout,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	return err
}

func differenceNetworks(a, b []interface{}) []string {
	mb := map[string]bool{}
	for _, x := range b {
		mb[x.(string)] = true
	}
	ab := []string{}
	for _, x := range a {
		if _, ok := mb[x.(string)]; !ok {
			ab = append(ab, x.(string))
		}
	}
	return ab
}

// https://developer.hashicorp.com/terraform/plugin/sdkv2/guides/v2-upgrade-guide#removal-of-helper-hashcode-package
func hashcodeString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return 0
}
