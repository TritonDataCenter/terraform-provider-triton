---
layout: "triton"
page_title: "Triton: triton_network"
sidebar_current: "docs-triton-datasource-network"
description: |-
  The `triton_network` data source queries Triton for Network information
  (e.g., Network ID, etc.) based on the name of the Network.
---

# triton_network

The `triton_network` data source queries Triton for Network information
(e.g., Network ID, etc.) based on the name of the Network.

## Example Usages

Find the Network ID of the `Joyent-SDC-Private` network.

```hcl
# Declare the data source.
data "triton_network" "private" {
  name = "Joyent-SDC-Private"
}

# Access unique Network ID using output from the data source. 
output "private_network_id" {
  value = "${data.triton_network.private.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (string)
    **Required.** The name of the Network.

## Attribute Reference

The following attributes are supported:

* `id` - (string)
    The unique identifier of the Network.

* `public` - (boolean)
    Whether this Network is a public or private [RFC1918][1] network.
    
* `fabric` - (boolean)
    Whether this Network is created on a [Fabric][2].

[1]: https://tools.ietf.org/html/rfc1918
[2]: https://docs.joyent.com/public-cloud/network/sdn
