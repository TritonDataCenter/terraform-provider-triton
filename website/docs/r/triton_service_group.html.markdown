---
layout: "triton"
page_title: "Triton: triton_service_group"
sidebar_current: "docs-triton-resource-service-group"
description: |-
    The `triton_service_group` resource represents a Triton Service Group.
---

# triton_service_group

The `triton_service_group` resource represents a Triton Service Group.

~> **NOTE:**  Triton Service Groups are in Preview and only supported in specific regions at this time. They will become Generally Available in the near future.


## Example Usages

```hcl
resource "triton_service_group" "web" {
  group_name = "web_group"
  template   = "${triton_instance_template.base.id}"
  capacity   = 3
}
```

## Argument Reference

The following arguments are supported:

* `group_name` - (string, Required) Friendly name for the service group.

* `template` - (string, Required)  Identifier of an instance template.

* `capacity` - (int, Optional) Number of instances to launch and monitor.

## Attribute Reference

The following attributes are exported:

* `id` - (string) - The identifier representing the Triton Service Group.
