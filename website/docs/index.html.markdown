---
layout: "triton"
page_title: "Provider: Joyent Triton"
sidebar_current: "docs-triton-index"
description: |-
  Used to provision infrastructure in Joyent's Triton public or on-premise clouds.
---

# Joyent Triton Provider

The Triton provider is used to interact with resources in Joyent's Triton cloud. It is compatible with both public- and on-premise installations of Triton. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "triton" {
  account = "AccountName"
  key_id  = "25:d4:a9:fe:ef:e6:c0:bf:b4:4b:4b:d4:a8:8f:01:0f"

  # If using a private installation of Triton, specify the URL, otherwise
  # set the URL according to the region you wish to provision.
  url = "https://us-west-1.api.joyentcloud.com"

  # If you want to use a triton account other than the main account, then
  # you can specify the username as follows
  user = "myusername"
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

* `account` - (Required) This is the name of the Triton account. It can also be
provided via the `SDC_ACCOUNT` or `TRITON_ACCOUNT` environment variables.
* `user` - (Optional) This is the username to interact with the triton API. It
can be provided via the `SDC_USER` or `TRITON_USER` environment variables.
* `key_material` - (Optional) This is the private key of an SSH key associated
with the Triton account to be used. If this is not set, the private key corresponding
to the fingerprint in `key_id` must be available via an SSH Agent. It can be provided
via the `SDC_KEY_MATERIAL` or `TRITON_KEY_MATERIAL` environment variables.
* `key_id` - (Required) This is the fingerprint of the public key matching the key
specified in `key_path`. It can be obtained via the command `ssh-keygen -l -E md5 -f /path/to/key`.
It can be provided via the `SDC_KEY_ID` or `TRITON_KEY_ID` environment variables.
* `url` - (Optional) This is the URL to the Triton API endpoint. It is required
if using a private installation of Triton. The default is to use the Joyent public
cloud us-west-1 endpoint. Valid public cloud endpoints include: `us-east-1`, `us-east-2`,
`us-east-3`, `us-sw-1`, `us-west-1`, `eu-ams-1`. It can be provided
via the `SDC_URL` or `TRITON_URL` environment variables.
* `insecure_skip_tls_verify` (Optional - defaults to false) This allows skipping
TLS verification of the Triton endpoint. It is useful when connecting to a temporary
Triton installation such as Cloud-On-A-Laptop which does not generally use a certificate
signed by a trusted root CA.
