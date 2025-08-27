---
page_title: "triton_vlan Resource - triton"
description: |-
    The `triton_vlan` resource represents an VLAN for a Triton account.
---

# triton_vlan (Resource)

The `triton_vlan` resource represents an Triton VLAN. A VLAN provides a low level way to segregate and subdivide the network. Traffic on one VLAN cannot, *on its own*, reach another VLAN.

## Example Usage

### Create a VLAN

```terraform
resource "triton_vlan" "dmz" {
  vlan_id     = 100
  name        = "dmz"
  description = "DMZ VLAN"
}
```

## Argument Reference

The following arguments are supported:

* `vlan_id` - (int, Required, Change forces new resource) Number between 0-4095 indicating VLAN ID

* `name` - (string, Required) Unique name to identify VLAN

* `description` - (string, Optional) Description of the VLAN

## Import

`triton_vlan` resources can be imported using the VLAN ID, for example:

```shell
terraform import triton_vlan.example 100
```
