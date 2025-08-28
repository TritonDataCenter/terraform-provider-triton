---
page_title: "triton_firewall_rule Resource - triton"
description: |-
    The `triton_firewall_rule` resource represents a rule for the Triton cloud firewall.
---

# triton_firewall_rule (Resource)

The `triton_firewall_rule` resource represents a rule for the Triton cloud firewall.

## Example Usage

### Allow web traffic on ports tcp/80 and tcp/443 to machines with the 'www' tag from any source

```terraform
resource "triton_firewall_rule" "www" {
  description = "Allow web traffic on ports tcp/80 and tcp/443 to machines with the 'www' tag from any source."
  rule        = "FROM any TO tag \"www\" ALLOW tcp (PORT 80 AND PORT 443)"
  enabled     = true
}
```

### Allow ssh traffic on port tcp/22 to all machines from known remote IPs

```terraform
resource "triton_firewall_rule" "22" {
  description = "Allow ssh traffic on port tcp/22 to all machines from known remote IPs."
  rule        = "FROM (ip w.x.y.z OR ip w.x.y.z) TO all vms ALLOW tcp PORT 22"
  enabled     = true
}
```

### Block IMAP traffic on port tcp/143 to all machines

```terraform
resource "triton_firewall_rule" "imap" {
  description = "Block IMAP traffic on port tcp/143 to all machines."
  rule        = "FROM any TO all vms BLOCK tcp PORT 143"
  enabled     = true
}
```

## Argument Reference

The following arguments are supported:

* `rule` - (string, Required) The firewall rule described using the Cloud API rule syntax defined at https://docs.tritondatacenter.com/public-cloud/network/firewall/cloud-firewall-rules-reference. Note: Cloud API will normalize rules based on case-sensitivity, parentheses, ordering of IP addresses, etc. This can result in Terraform updating rules repeatedly if the rule definition differs from the normalized value.

* `enabled` - (boolean, Optional) Default: `false` Whether the rule should be effective.

* `description` - (string, Optional) Description of the firewall rule

## Attribute Reference

The following attributes are exported:

* `id` - (string) - The identifier representing the firewall rule in Triton.

## Import

`triton_firewall` resources can be imported using the firewall rules UUID, for example:

```shell
terraform import triton_firewall_rule.example 2739849e-a2b3-4eb0-bd00-cc1c2ed0e6d5
```
