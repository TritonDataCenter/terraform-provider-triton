---
layout: "triton"
page_title: "Triton: triton_account"
sidebar_current: "docs-triton-datasource-account"
description: |-
    The `triton_account` data source queries the Triton Account API for account information.
---

# triton_account

The `triton_account` data source queries the Triton Account API for account information.

## Example Usages

```hcl
data "triton_account" "main" {}

output "account_id" {
    value = "${data.triton_account.main.id}"
}
```

## Argument Reference

The data source uses the name of the account currently configured to interact with the Triton API.

## Attribute Reference

The following attributes are supported:

* `cns_enabled` - (bool) Whether CNS is enabled for the account.
