---
layout: "triton"
page_title: "Triton: triton_network"
sidebar_current: "docs-triton-datasource-network"
description: |-
    The `triton_network` data source queries the Triton Network API for network IDs.
---

# triton\_network

The `triton_network` data source queries the Triton Network API for a network ID
based on it's name.

## Example Usages

Find the ID of the Joyent-SDC-Private network.

```hcl
data "triton_network" "private" {
    name = "Joyent-SDC-Private"
}

output "private_network_id" {
    value = "${data.triton_network.private.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (string)
    The name of the network

