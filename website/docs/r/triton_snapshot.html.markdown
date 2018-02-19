---
layout: "triton"
page_title: "Triton: triton_snapshot"
sidebar_current: "docs-triton-resource-snapshot"
description: |-
    The `triton_snapshot` resource represents a snapshot of a Triton machine.
---

# triton\_snapshot

The `triton_snapshot` resource represents a snapshot of a Triton machine.
Snapshots are not usable with other instances; they are a point-in-time snapshot of the current instance.
Snapshots can also only be taken of instances that are not of brand `kvm`.

## Example Usages

```hcl
data "triton_image" "ubuntu1604" {
  name    = "ubuntu-16.04"
  version = "20170403"
}

resource "triton_machine" "test" {
  image   = "${data.triton_image.ubuntu1604.id}"
  package = "g4-highcpu-128M"
}

resource "triton_snapshot" "test" {
  name       = "my-snapshot"
  machine_id = "${triton_machine.test.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (string, Required)
    The name for the snapshot.

* `machine_id` - (string, Required)
    The ID of the machine of which to take a snapshot.

## Attribute Reference

The following attributes are exported:

* `id` - (string) - The identifier representing the snapshot in Triton.
* `state` - (string) - The current state of the snapshot.
