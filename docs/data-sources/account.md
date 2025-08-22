---
page_title: "triton_account Data Source - triton"
description: |-
  The `triton_account` data source queries Triton for Account information.
---

# triton_account (Data Source)

The `triton_account` data source queries Triton for Account information.

~> **NOTE:** This data source uses the name of the `account` currently configured in the [Triton provider](/providers/tritondatacenter/triton/latest/docs).

## Example Usage

Find current Account unique identifier:

```terraform
# Declare the data source.
data "triton_account" "main" {}

# Access unique Account ID using output from the data source.
output "account_id" {
  value = data.triton_account.main.id
}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

The following attributes are supported:

* `id` - (string) The unique identifier representing the Account in Triton.

* `login` - (string) The login name associated with the Account.

* `email` - (string) An e-mail address that is current set in the Account.

* `cns_enabled` - (boolean) Whether the Container Name Service (CNS) is enabled for the Account.
