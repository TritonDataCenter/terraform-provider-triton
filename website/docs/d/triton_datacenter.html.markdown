---
layout: "triton"
page_title: "Triton: triton_datacenter"
sidebar_current: "docs-triton-datasource-datacenter"
description: |-
    The `triton_datacenter` data source queries the Triton Account API for datacenter information.
---

# triton_datacenter

The `triton_datacenter` data source queries the Triton Account API for datacenter information.

## Example Usages

```hcl
data "triton_datacenter" "current" {}

output "endpoint" {
    value = "${data.triton_datacenter.current.endpoint}"
}
```

## Argument Reference

The data source uses the endpoint currently configured to interact with the Triton API.

## Attribute Reference

The following attributes are supported:

* `name` - (string) The name of the datacenter.
* `endpoint` - (string) The endpoint url of the datacenter
