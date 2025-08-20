---
page_title: "Provider: Triton"
sidebar_current: "docs-triton-index"
description: |-
  Used to provision infrastructure in TritonDataCenter public or on-premise clouds.
---

# Triton Provider

The Triton provider is used to interact with resources in [TritonDataCenter clouds](https://www.tritondatacenter.com/). It is compatible with both public and on-premise installations of TritonDataCenter. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```terraform
provider "triton" {
  account = "AccountName"
  key_id  = "25:d4:a9:fe:ef:e6:c0:bf:b4:4b:4b:d4:a8:8f:01:0f"

  # If using a private installation of Triton, specify the URL, otherwise
  # set the URL according to the region you wish to provision.
  url = "https://us-central-1.api.mnx.io"

  # If you want to use a triton sub user of the main account, then
  # you can specify the username as follows
  #user = "myusername"

  # If using a test Triton installation (self-signed certifcate), use:
  #insecure_skip_tls_verify = true
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

### Required

- `account` (String) This is the name of the Triton account. It can also be provided via the `SDC_ACCOUNT` or `TRITON_ACCOUNT` environment variables.
- `key_id` (String) This is the fingerprint of the public key matching the key specified in `key_path`. It can be obtained via the command `ssh-keygen -l -E md5 -f /path/to/key`. It can be provided via the `SDC_KEY_ID` or `TRITON_KEY_ID` environment variables.

### Optional

- `insecure_skip_tls_verify` (Boolean) Defaults to `false`. This allows skipping TLS verification of the Triton endpoint. It is useful when connecting to a temporary Triton installation such as Cloud-On-A-Laptop which does not generally use a certificate signed by a trusted root CA.
- `key_material` (String) This is the private key of an SSH key associated with the Triton account to be used. If this is not set, the private key corresponding to the fingerprint in `key_id` must be available via an SSH Agent. It can be provided via the `SDC_KEY_MATERIAL` or `TRITON_KEY_MATERIAL` environment variables.
- `user` (String) This is the username of a sub user to interact with the Triton API. It can be provided via the `SDC_USER` or `TRITON_USER` environment variables.
- `url` (String) This is the URL to the Triton API endpoint. It is required if using a private installation of Triton. The default is to use the MNX.io public cloud `us-central-1` endpoint. It can be provided via the `SDC_URL` or `TRITON_URL` environment variables.

## Source Code

The source for the Terraform Triton provider is available through GitHub: https://github.com/TritonDataCenter/terraform-provider-triton/
