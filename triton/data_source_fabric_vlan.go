package triton

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/joyent/triton-go/network"
	"github.com/pkg/errors"
)

// filterVLANFunc is a function that is called to filter a Fabric VLAN from
// a slice of Fabric VLANs based on a predicate.
type filterVLANFunc func(*network.FabricVLAN) bool

// dataSourceFabricVLAN returns schema for the Fabric VLAN data source.
func dataSourceFabricVLAN() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFabricVLANRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vlan_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateVLANIdentifier,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

// dataSourceFabricVLANRead retrieves details about all the Fabric VLANs
// which are available in the current Data Center from the Fabrics API,
// then searches for a matching Fabric VLAN using either name, VLAN ID
// or description as filter, or a combination of thereof.
func dataSourceFabricVLANRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	vlanName, vlanNameOk := d.GetOk("name")
	vlanID, vlanIDOk := d.GetOk("vlan_id")
	vlanDesc, vlanDescOk := d.GetOk("description")

	if !vlanNameOk && !vlanIDOk && !vlanDescOk {
		return errors.New("one of `name`, `vlan_id`, or `description` must be assigned")
	}

	net, err := client.Network()
	if err != nil {
		return errors.Wrap(err, "error creating Network client")
	}

	log.Printf("[DEBUG] triton_fabric_vlan: Reading Fabric VLAN details.")
	vlans, err := net.Fabrics().ListVLANs(context.Background(), &network.ListVLANsInput{})
	if err != nil {
		return errors.Wrap(err, "error retrieving Fabric VLAN details")
	}

	matches := vlans

	// There can be many Fabric VLANs sharing the same name and description
	// within the data center (neither name nor description are unique), and
	// the only way to uniquely identify a single Fabric VLAN would be either
	// its VLAN ID (which is always a match) and either a very specific name
	// or description, or a combination of thereof. We allow the end-user to
	// use multiple attributes as filters together to granularly narrow down
	// results so that only a single Fabric VLAN would be found. All of the
	// filters create an implicit AND relationship between one another, and
	// in a case of the name and description attributes, a simple wildcard
	// match can be used.
	if vlanIDOk {
		matches = filterVLANs(matches, func(v *network.FabricVLAN) bool {
			return v.ID == vlanID.(int)
		})
	}
	if vlanNameOk {
		matches = filterVLANs(matches, func(v *network.FabricVLAN) bool {
			return wildcardMatch(vlanName.(string), v.Name)
		})
	}
	if vlanDescOk {
		matches = filterVLANs(matches, func(v *network.FabricVLAN) bool {
			return wildcardMatch(vlanDesc.(string), v.Description)
		})
	}

	var vlan *network.FabricVLAN
	if len(matches) == 0 {
		return errors.New("unable to find any Fabric VLANs matching the " +
			"current search criteria, please change your search criteria " +
			"and try again")
	}

	if len(matches) > 1 {
		log.Printf("[DEBUG] triton_fabric_vlan: Found multiple matching Fabric VLANs: %+v", matches)
		return errors.New("found multiple Fabric VLANs matching the " +
			"current search criteria, please change your search criteria " +
			"and try again")
	}

	vlan = matches[0]

	log.Printf("[DEBUG] triton_fabric_vlan: Found matching Fabric VLAN: %+v", vlan)
	d.SetId(time.Now().UTC().String())

	d.Set("name", vlan.Name)
	d.Set("vlan_id", vlan.ID)
	d.Set("description", vlan.Description)

	return nil
}

// filterVLANs iterates over a slice of Fabric VLANs, and returns a slice that
// contains all of the Fabric VLANs the predicate returns a value of true for.
func filterVLANs(vlans []*network.FabricVLAN, f filterVLANFunc) (results []*network.FabricVLAN) {
	for _, vlan := range vlans {
		if f(vlan) {
			results = append(results, vlan)
		}
	}
	return
}
