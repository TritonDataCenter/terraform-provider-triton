---
layout: "triton"
page_title: "Triton: triton_fabric_vlan"
sidebar_current: "docs-triton-datasource-fabric-vlan"
description: |-
  The `triton_fabric_vlan` data source queries Triton for Fabric VLAN information
  (e.g., VLAN ID, etc.) based either on the name, VLAN ID or description of the
  Fabric VLAN.
---

# triton_fabric_vlan

The `triton_fabric_vlan` data source queries Triton for [Fabric VLAN][1]
information (e.g., VLAN ID, etc.) based either on the name, VLAN ID or
description of the Fabric VLAN.

## Example Usages

Find the VLAN ID using the name of the Fabric VLAN as a search filter:

```hcl
# Declare the data source.
data "triton_fabric_vlan" "public" {
  name = "Public-VLAN-Production"
}

# Access unique VLAN ID using output from the data source.
output "public_vlan_id" {
  value = "${data.triton_fabric_vlan.public.vlan_id}"
}
```

Find the VLAN ID using name (with a wildcard match) and description of
the Fabric VLAN as a search filters:

```hcl
# Declare the data source, and use a combination of two arguments
# to form a search filter. Use a wildcard match for the name.
data "triton_fabric_vlan" "private_database_vlan" {
  name        = "Private-VLAN-*"
  description = "A secure VLAN for production database servers"
}

# Access unique VLAN ID using output from the data source.
output "private_database_vlan_id" {
  value = "${data.triton_fabric_vlan.private_database_vlan.vlan_id}"
}
```
## Argument Reference

~> **NOTE:** The arguments of this data source act as filters when searching for
a matching Fabric VLAN and can be combined together, but at lease one of `name`,
`vlan_id` or `description` must be assigned.

The following arguments are supported:

* `name` - (string)
    Optional. The name of the Fabric VLAN.

* `vlan_id` - (integer)
    Optional. The unique identifier (VLAN ID) of the Fabric VLAN.

* `description` - (string)
    Optional. The description of the Fabric VLAN.

~> **NOTE:** Both the `name` and `description` arguments support a simple wildcard
pattern matching using two common wildcards, such as **`*`** (asterisk) and **`?`**.
There is no support for either ranges or character classes. More details about
wildcard patterm matching can be found [here][2].

## Attribute Reference

The following attributes are exported:

* `name` - (string)
    The name of the Fabric VLAN, if any.

* `vlan_id` - (integer)
    The unique identifier (VLAN ID) of the Fabric VLAN.

* `description` - (string)
    The description of the Fabric VLAN, if any.

[1]: https://docs.joyent.com/public-cloud/network/sdn#vlans
[2]: https://en.wikipedia.org/wiki/Glob_(programming)
