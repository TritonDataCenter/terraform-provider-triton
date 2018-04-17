---
layout: "triton"
page_title: "Triton: triton_instance_template"
sidebar_current: "docs-triton-resource-instance-template"
description: |-
    The `triton_instance_template` resource represents a Triton Service Group instance template.
---

# triton_instance_template

The `triton_instance_template` resource represents a Triton Service Group instance template.

~> **NOTE:**  Triton Service Groups are in Preview and only supported in specific regions at this time. They will become Generally Available in the near future.

## Example Usages

```hcl
data "triton_image" "base" {
  name    = "base-64-lts"
  version = "16.4.1"
}

data "triton_network" "private" {
  name = "Joyent-SDC-Private"
}

resource "triton_instance_template" "base" {
  template_name    = "Base template"
  image            = "${data.triton_image.base.id}"
  package          = "g4-highcpu-128M"
  
  firewall_enabled = false
  
  networks         = ["${data.triton_network.private.id}"]
  
  tags {
    hello = "world"
    role  = "database"
  }
}
```

## Argument Reference

The following arguments are supported:

* `template_name` - (string, Required) Friendly name for the instance template.

* `image` - (string, Required)  UUID of the image.

* `package` - (string, Required) Package name used for provisioning.

* `firewall_enabled` - (boolean, Optional) Whether to enable the firewall for group instances. Default is `false`.

* `tags` - (map, Optional) Tags for group instances. 

* `networks` - (list, Optional) Network IDs for group instances.

* `metadata` - (map, Optional) Metadata for group instances.

* `userdata` - (string, Optional) Data copied to instance on boot.

## Attribute Reference

The following attributes are exported:

* `id` - (string) - The identifier representing the Triton Service Group instance template.
