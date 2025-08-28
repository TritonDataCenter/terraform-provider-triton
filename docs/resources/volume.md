---
page_title: "triton_volume Resource - triton"
description: |-
    The `triton_volume` resource represents a storage volume instance running in Triton.
---

# triton_volume (Resource)

The `triton_volume` resource represents a storage volume instance running in Triton.

## Example Usage

### Creating a volume

```terraform
resource "triton_volume" "my-volume" {
  name = "my-volume"

  tags {
    hello = "world"
    role  = "database"
  }
}
```

### Creating a volume on a specific network with a specific size

```terraform
data "triton_network" "my_fabric" {
  name = "My-Fabric-Network"
}

resource "triton_volume" "my_volume" {
  networks = ["${data.triton_network.my_fabric.id}"]
  size     = 10240
}
```

### Creating two volumes and one machine that uses them both

```terraform
resource "triton_volume" "my_volume_1" {
}

resource "triton_volume" "my_volume_2" {
}

resource "triton_machine" {
  # base-64-lts 24.4.1
  image   = "2f1dc911-6401-4fa4-8e9d-67ea2e39c271"
  package = "g1.medium"

  volume {
    name       = triton_volume.my_volume_1.name
    mountpoint = "/data1"
  }

  volume {
    name       = triton_volume.my_volume_2.name
    mode       = "ro"
    mountpoint = "/data2"
  }
}
```

## Argument Reference

These arguments can be supplied when creating a volume:

* `name` - (string, optional) The friendly name for the volume. Triton will generate a name if one is not specified.

* `size` - (integer, optional) The size of the volume.

* `networks` - (list, optional) The list of networks for which the volume will be accessible on. The network ID will be in hex form, e.g `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.

* `tags` - (map, optional) A mapping of tags to apply to the volume.

* `type` - (string, optional) The type of volume Triton should create (defaults to *tritonnfs*).

## Attribute Reference

The following attributes are exported on a volume resource:

* `id` - (string) - The identifier representing the volume in Triton.
* `filesystem_path` - (string) - The NFS path that the volume can be referenced through.
* `networks` - (list of strings) - The ID of the networks which the volume is attached to, and thus over which it can be accessed.
* `state` - (string) - The current state of the volume. Can be one of *creating*, *ready*, *deleting*, *deleted* or *failed*.
* `tags` - (map) - A mapping of tags the volume is using.
* `type` - (string) - The type of the volume.

## Import

`triton_volume` resources can be imported using the volume UUID, for example:

```shell
terraform import triton_volume.example 4c0bc531-38a4-4919-8065-828a56a3b818
```
