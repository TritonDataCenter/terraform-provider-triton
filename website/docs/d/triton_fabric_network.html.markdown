---
layout: "triton"
page_title: "Triton: triton_fabric_network"
sidebar_current: "docs-triton-datasource-fabric-network"
description: |-
  The `triton_fabric_network` data source queries Triton for Fabric Network
  information (e.g., subnet CIDR, gateway, static routes, etc.) based on the
  name of the Fabric Network and ID of the VLAN on which the network has been
  created.
---

# triton_fabric_network

The `triton_fabric_network` data source queries Triton for [Fabric Network][4]
information (e.g., subnet CIDR, gateway, state routes, etc.) based on the
name of the Fabric Network and ID of the VLAN on which the network has been
created.

## Example Usages

Find the subnet CIDR of a Fabric Network:

```hcl
# Declare the data source to retrieve Fabric VLAN details.
data "triton_fabric_vlan" "private" {
  name = "Private-VLAN-Production"
}

# Declare the data source to retrieve Fabric Network details.
data "triton_fabric_network" "private" {
  name     = "Private-Network-Production"
  vland_id = "${data.triton_fabric_vlan.private.vlan_id}"
}

# Access subnet CIDR using output from the data source.
output "private_network_cidr" {
  value = "${data.triton_fabric_network.private.subnet}"
}
```

## Argument Reference

~> **NOTE:** You can use the [triton_fabric_vlan][1] data source to
retrieve details about a [Fabric VLAN][2] for reference.

The following arguments are supported:

* `name` - (string)
    **Required.** The name of the Fabric Network.

* `vlan_id` - (integer)
    **Required.** The unique identifier (VLAN ID) of the Fabric VLAN.

## Attribute Reference

* `name` - (string)
    The name of the Fabric Network.

* `public` - (boolean)
    Whether this Fabric Network is a public or private [RFC1918][3] network.

* `fabric` - (boolean)
    Whether this network is created on a [Fabric][4]. This is always
    **true** for a Fabric Network.

* `description` - (string)
    The description of the Fabric Network, if any.

* `subnet` - (string)
    A [CIDR][5] block used for the Fabric Network.

* `provision_start_ip` - (string)
    The first IP address on this network that may be assigned.

* `provision_end_ip` - (string)
    The last IP address on this network that may be assigned.

* `gateway` - (string)
    An IP address of the gateway on this network, if any.

* `resolvers` - (list)
    A list of IP addresses of DNS resolvers on this network.

* `routes` - (map)
    A map of static routes (using the [CIDR][5] notation) and corresponding gateways on this network, if any.

* `internet_nat` - (boolean)
    Whether the gateway on this network is also provisioned with the
    Internet NAT zone.

* `vlan_id` - (integer)
    The unique identifier (VLAN ID) of the Fabric VLAN.

[1]: /docs/providers/triton/d/triton_fabric_vlan.html
[2]: https://docs.joyent.com/public-cloud/network/sdn#vlans
[3]: https://tools.ietf.org/html/rfc1918
[4]: https://docs.joyent.com/public-cloud/network/sdn
[5]: https://tools.ietf.org/html/rfc4632
