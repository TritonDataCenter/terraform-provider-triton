---
page_title: "triton_network Data Source - triton"
description: |-
  The `triton_network` data source queries Triton for Network information
  (e.g., Network ID, etc.) based on the name of the Network.
---

# triton_network (Data Source)

The `triton_network` data source queries Triton for Network information (e.g., Network ID, etc.) based on the name of the Network.

## Example Usage

Find the Network ID of the `My-Fabric-Network` network.

```terraform
# Declare the data source.
data "triton_network" "private" {
  name = "My-Fabric-Network"
}

# Access unique Network ID using output from the data source. 
output "private_network_id" {
  value = data.triton_network.private.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (string) **Required.** The name of the Network.

## Attribute Reference

The following attributes are supported:

* `id` - (string) The unique identifier of the Network.

* `public` - (boolean) Whether this Network is a public or private [RFC1918](https://tools.ietf.org/html/rfc1918) network.

* `fabric` - (boolean) Whether this Network is created on a [Fabric](https://docs.tritondatacenter.com/public-cloud/network/sdn).
