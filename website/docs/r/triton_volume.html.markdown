---
layout: "triton"
page_title: "Triton: triton_volume"
sidebar_current: "docs-triton-resource-volume"
description: |-
    The `triton_volume` resource represents a volume for a Triton account.
---

# triton_volume

The `triton_volume` resource represents a volume for a Triton account.

## Example

```hcl
data "triton_network" "my_fabric" {
  name = "My-Fabric-Network"
}
resource "triton_volume" "test" {
  name = "volume-1"
  networks = ["${data.triton_network.my_fabric.id}"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (string, Required)
   The desired name for the volume.

* `size` - (int, Optional)
   The desired minimum storage capacity for that volume in mebibytes. Default value is `10240` mebibytes (10 gibibytes).

* `type` - (string, Optional)
   The type of volume. Currently only `tritonnfs` is supported.

* `networks` - (list, required)
   A list of UUIDs representing networks on which the volume is reachable. These networks must be fabric networks owned by the user sending the request.


## Attribute Reference

* `owner` - the UUID of the volume's owner.
* `filesystem_path` - the path that can be used by a NFS client to mount the NFS remote filesystem in the host's filesystem.