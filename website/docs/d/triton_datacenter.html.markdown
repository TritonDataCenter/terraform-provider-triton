---
layout: "triton"
page_title: "Triton: triton_datacenter"
sidebar_current: "docs-triton-datasource-datacenter"
description: |-
  The `triton_datacenter` data source queries Triton for Data Center information.
---

# triton_datacenter

The `triton_datacenter` data source queries Triton for Data Center information.

~> **NOTE:** This data source uses the endpoint URL of the Data Center currently
configured in the [Trition provider][1].

## Example Usages

Find current Data Center endpoint URL:

```hcl
# Declare the data source.
data "triton_datacenter" "current" {}

# Access current endpoint URL using output from the data source.
output "endpoint" {
  value = "${data.triton_datacenter.current.endpoint}"
}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

~> **NOTE:** When using the [Triton Public Cloud](https://www.joyent.com/triton),
the `endpoint` attribute might include an old, but still fully supported, domain
name "joyentcloud.com" (e.g. https://us-east-1.api.joyentcloud.com), even when
the new domain name "joyent.com" has been used to configure the cloud endpoint
URL in the [Trition provider][1].

The following attributes are supported:

* `name` - (string)
    The name of the Data Center.

* `endpoint` - (string)
    The endpoint URL of the Data Center.

[1]: /docs/providers/triton/index.html
