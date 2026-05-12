# Network Pool UUID Drift in `triton_machine`

## Problem

When a `triton_machine` resource specifies a network pool UUID in its
`networks` set, `terraform plan` shows perpetual drift after every apply.

### Example

```hcl
data "triton_network" "public" {
  name = "MNX-Triton-Public"
}

resource "triton_machine" "example" {
  name    = "my-instance"
  package = "g1.nano"
  image   = "..."
  networks = [data.triton_network.public.id]
}
```

After apply, every subsequent plan shows:

```
~ networks = [
    - "0efd931d-ae63-40ca-bfa9-f028573bbb0a",
    + "5ff1fe03-075b-4e4c-b85b-73de0c452f77",
  ]
```

## Root Cause

Triton network pools are virtual containers of one or more real networks.
When you provision a machine on a pool (e.g. `5ff1fe03`), NAPI selects a
member network (e.g. `0efd931d`) and attaches the NIC to that member.  The
NIC's `network` field is the member UUID, not the pool UUID.

In `resourceMachineRead` (resource_machine.go), the provider reads back the
NIC list and writes their actual network UUIDs into the `networks` state
attribute:

```go
for _, nic := range *nicsResp.JSON200 {
    networkList = append(networkList, uuidString(nic.Network))
}
d.Set("networks", networkList)
```

This writes `0efd931d` (the member) into state.  On the next plan, Terraform
compares state (`0efd931d`) against the config (`5ff1fe03` from the data
source) and sees a diff.

The `networks` schema field is `Optional` but not `Computed`, so Terraform
always diffs it against the config value.

## Why the Write-Back Exists

The comment at the write-back site says:

> Regression guard (#30/#35): populate both "nic" (computed) and "networks"
> (input) from actual NICs to avoid perpetual diffs with network pools.

The intent was that by writing back the actual NIC UUIDs, subsequent reads
would return the same values and avoid drift.  This works when the config
also uses the actual network UUID (i.e. non-pool networks), but fails when
the config references a pool UUID because the config value never changes.

## Who Is Affected

Any user whose `networks` list includes a network pool UUID.  Networks that
are not pools (e.g. fabric networks) are unaffected because the NIC's network
UUID matches the config UUID exactly.

You can identify pools by their lack of `subnet`, `gateway`, and `vlan_id`
properties:

```
$ triton network get MNX-Triton-Public
{
    "id": "5ff1fe03-075b-4e4c-b85b-73de0c452f77",
    "name": "MNX-Triton-Public",
    "public": true
}
```

(No subnet, gateway, or VLAN — this is a pool.)

## Possible Fixes

### Option A: Resolve NIC UUIDs Back to Pool UUIDs

In `resourceMachineRead`, for each NIC network UUID, check if the config
specifies a pool UUID that contains that network.  If so, write the pool
UUID to state instead of the member UUID.

**Pros:** Config stays authoritative; no schema change needed.
**Cons:** Requires additional API calls to CloudAPI's `ListNetworkMembers`
or equivalent to determine pool membership.  The CloudAPI endpoint for this
may need to be verified/added.

### Option B: Add `Computed: true` to `networks`

Change the schema to `Optional: true, Computed: true`.  This tells Terraform
that the provider may set a value that differs from the config, suppressing
the diff.

**Pros:** Simple one-line change.
**Cons:** With `Computed`, if the user removes a network from config,
Terraform may not detect the removal (the provider-set value takes
precedence).  This subtly changes the semantics of the field and may
surprise users.

### Option C: Stop Writing `networks` Back in Read

Remove the `d.Set("networks", networkList)` call.  The `nic` computed block
already provides the actual NIC details.  Let the config value be
authoritative for `networks`.

**Pros:** Simple.  Config always matches state.
**Cons:** `terraform import` would not populate `networks`, requiring users
to add it to their config manually.  Also, if a NIC is removed externally,
Terraform would not detect the drift.

### Option D: DiffSuppressFunc

Add a `DiffSuppressFunc` on the `networks` field that calls the API to
check if the old value is a member of the pool identified by the new value
(or vice versa).

**Pros:** Most precise — only suppresses pool-related diffs.
**Cons:** API call in the diff function; complex to implement correctly for
sets with multiple elements.

## Workaround

Until this is fixed, users can avoid the drift by specifying the member
network UUID directly instead of the pool UUID.  Use `triton network list`
to find the actual network UUID that the pool resolves to.
