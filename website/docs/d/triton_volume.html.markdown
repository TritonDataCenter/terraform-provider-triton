---
layout: "triton"
page_title: "Triton: Volume Data Source"
sidebar_current: "docs-triton-datasource-volume"
description: |-
    The `triton_volume` data source queries the Triton API for an existing volume.
---

# triton\_volume

The `triton_volume` data source represents a storage volume instance running in Triton.

## Example Usages

Find the ID of a volume with a given name:

```hcl
data "triton_volume" "myvol" {
  name    = "my-volume-name"
}

output "volume_id" {
  value = "${data.triton_volume.myvol.id}"
}
```

## Argument Reference

These arguments can be supplied when querying for an existing volume:

* `name` - (string)
    The name of the volume.

* `size` - (integer)
    The size of the volume.

* `state` - (string)
    The state of the volume (one of *creating*, *ready*, *deleting*, *deleted*
    or *failed*).

## Attribute Reference

The following attributes are exported on a volume data source:

* `id` - (string) - The identifier representing the volume in Triton.
* `filesystem_path` - (string) - The NFS path that the volume can be referenced
  through.
* `networks` - (list of strings) - The ID of the networks which the volume is
  attached to, and thus over which it can be accessed.
* `state` - (string) - The current state of the volume. Can be one of
  *creating*, *ready*, *deleting*, *deleted* or *failed*.
* `tags` - (map) - A mapping of tags the volume is using.
* `type` - (string) - The type of the volume.
