---
page_title: "triton_datacenter Data Source - triton"
description: |-
  The `triton_datacenter` data source queries Triton for Data Center information.
---

# triton_datacenter (Data Source)

The `triton_datacenter` data source queries Triton for Data Center information.

~> **NOTE:** This data source uses the endpoint `url` of the Data Center currently configured in the [Triton provider](/providers/tritondatacenter/triton/latest/docs).

## Example Usage

Find current Data Center endpoint URL:

```terraform
# Declare the data source.
data "triton_datacenter" "current" {}

# Access current endpoint URL using output from the data source.
output "endpoint" {
  value = data.triton_datacenter.current.endpoint
}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

The following attributes are supported:

* `name` - (string) The name of the Data Center.

* `endpoint` - (string) The endpoint URL of the Data Center.
