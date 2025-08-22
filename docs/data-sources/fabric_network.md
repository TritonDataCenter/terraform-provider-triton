---
page_title: "triton_fabric_network Data Source - triton"
description: |-
  The `triton_fabric_network` data source queries Triton for Fabric Network
  information (e.g., subnet CIDR, gateway, static routes, etc.) based on the
  name of the Fabric Network and ID of the VLAN on which the network has been
  created.
---

# triton_fabric_network (Data Source)

The `triton_fabric_network` data source queries Triton for [Fabric Network](https://docs.tritondatacenter.com/public-cloud/network/sdn) information (e.g., subnet CIDR, gateway, state routes, etc.) based on the name of the Fabric Network and ID of the VLAN on which the network has been created.

## Example Usage

Find the subnet CIDR of a Fabric Network:

```terraform
# Declare the data source to retrieve Fabric VLAN details.
data "triton_fabric_vlan" "private" {
  name = "Private-VLAN-Production"
}

# Declare the data source to retrieve Fabric Network details.
data "triton_fabric_network" "private" {
  name     = "Private-Network-Production"
  vland_id = data.triton_fabric_vlan.private.vlan_id
}

# Access subnet CIDR using output from the data source.
output "private_network_cidr" {
  value = data.triton_fabric_network.private.subnet
}
```

## Argument Reference

~> **NOTE:** You can use the [triton_fabric_vlan](/docs/providers/triton/d/triton_fabric_vlan.html) data source to retrieve details about a [Fabric VLAN](https://docs.tritondatacenter.com/public-cloud/network/sdn#vlans) for reference.

The following arguments are supported:

* `name` - (string) **Required.** The name of the Fabric Network.

* `vlan_id` - (integer) **Required.** The unique identifier (VLAN ID) of the Fabric VLAN.

## Attribute Reference

* `name` - (string) The name of the Fabric Network.

* `public` - (boolean) Whether this Fabric Network is a public or private [RFC1918](https://tools.ietf.org/html/rfc1918) network.

* `fabric` - (boolean) Whether this network is created on a [Fabric](https://docs.tritondatacenter.com/public-cloud/network/sdn). This is always **true** for a Fabric Network.

* `description` - (string) The description of the Fabric Network, if any.

* `subnet` - (string) A [CIDR](https://tools.ietf.org/html/rfc4632) block used for the Fabric Network.

* `provision_start_ip` - (string) The first IP address on this network that may be assigned.

* `provision_end_ip` - (string) The last IP address on this network that may be assigned.

* `gateway` - (string) An IP address of the gateway on this network, if any.

* `resolvers` - (list) A list of IP addresses of DNS resolvers on this network.

* `routes` - (map) A map of static routes (using the [CIDR](https://tools.ietf.org/html/rfc4632) notation) and corresponding gateways on this network, if any.

* `internet_nat` - (boolean) Whether the gateway on this network is also provisioned with the Internet NAT zone.

* `vlan_id` - (integer) The unique identifier (VLAN ID) of the Fabric VLAN.
