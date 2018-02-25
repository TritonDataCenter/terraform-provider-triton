---
layout: "triton"
page_title: "Triton: triton_account"
sidebar_current: "docs-triton-datasource-account"
description: |-
  The `triton_account` data source queries Triton for Account information.
---

# triton_account

The `triton_account` data source queries Triton for Account information.

~> **NOTE:** This data source uses the name of the Account currently
configured in the [Trition provider][1].

## Example Usages

Find current Account unique identifier:

```hcl
# Declare the data source.
data "triton_account" "main" {}

# Access unique Account ID using output from the data source.
output "account_id" {
  value = "${data.triton_account.main.id}"
}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

The following attributes are supported:

* `id` - (string)
    The unique identifier representing the Account in Triton.

* `login` - (string)
    The login name associated with the Account.

* `email` - (string)
    An e-mail address that is current set in the Account.

* `cns_enabled` - (boolean)
    Whether the Container Name Service (CNS) is enabled for the Account.

[1]: /docs/providers/triton/index.html
