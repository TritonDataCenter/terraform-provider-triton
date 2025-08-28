---
page_title: "triton_key Resource - triton"
description: |-
    The `triton_key` resource represents a SSH public key for a Triton account.
---

# triton_key (Resource)

The `triton_key` resource represents a [SSH public key for a Triton account](https://docs.tritondatacenter.com/public-cloud/getting-started/ssh-keys), used for authentication.

## Example Usage

Add a public key to the Triton account, using the [file](https://developer.hashicorp.com/terraform/language/functions/file) function

```terraform
resource "triton_key" "example-file" {
  name = "Example Key"
  key  = file("keys/id_rsa.pub")
}
```

Add a public key to the Triton account

```terraform
resource "triton_key" "example-string" {
  name = "Example Key"
  key  = "ssh-rsa AAAAB... user@hostname"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (string, Change forces new resource) The name of the key. If this is left empty, the name is inferred from the comment in the SSH public key material.

* `key` - (string, Required, Change forces new resource) The SSH public key material. In order to read this from a file, use the [file](https://developer.hashicorp.com/terraform/language/functions/file) function.

## Import

`triton_key` resources can be imported using the SSH public key `name`, for example:

```shell
terraform import triton_key.example "Example Key"
```
