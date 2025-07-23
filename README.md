Triton Terraform Provider
=========================

Provider documentation: https://registry.terraform.io/providers/TritonDataCenter/triton/latest/docs

Learning about terraform
------------
- Website: https://developer.hashicorp.com/terraform
- Forums: [HashiCorp Discuss](https://discuss.hashicorp.com/c/terraform-core)
- Documentation: [https://developer.hashicorp.com/terraform/docs](https://developer.hashicorp.com/terraform/docs)
- Tutorials: [HashiCorp's Learn Platform](https://developer.hashicorp.com/terraform/tutorials)

<img alt="Terraform" src="https://www.datocms-assets.com/2885/1731373310-terraform_white.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 1.0
-	[Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)

Building The Provider
---------------------

Clone repository:

```sh
$ git clone git@github.com:terraform-providers/terraform-provider-triton
```

Enter the provider directory and build the provider

```sh
$ cd ./terraform-provider-triton
$ make build
```

Initialize your Terraform project by passing in the directory that contains your custom built provider binary, `terraform-provider-triton`. This is typically `$GOPATH/bin`.

```sh
$ terraform version
Terraform v1.12.0

$ terraform init --plugin-dir=$GOPATH/bin
```

Using the provider
------------------

If you haven't already done so, [create a Triton account](https://docs.tritondatacenter.com/public-cloud/getting-started) and read the getting started guide to complete the account setup and get your environment configured.

### Setup ###

The provider takes [many configuration arguments](https://www.terraform.io/docs/providers/triton/index.html#argument-reference) for setting up your Triton account within Terraform. The following example shows you how to explicitly configure the provider using your account information.

```hcl
provider "triton" {
  account = "AccountName"
  key_id  = "25:d4:a9:fe:ef:e6:c0:bf:b4:4b:4b:d4:a8:8f:01:0f"

  # If using a private installation of Triton, specify the URL, otherwise
  # set the URL to the CloudAPI endpoint of the region you wish to provision.
  url = "https://us-central-1.api.mnx.io"
}
```

The following arguments are supported.

- `account` - (Required) This is the name of the Triton account. It can also be provided via the SDC_ACCOUNT environment variable.
- `key_material` - (Optional) This is the private key of an SSH key associated with the Triton account to be used. If this is not set, the private key corresponding to the fingerprint in key_id must be available via an SSH Agent.
- `key_id` - (Required) This is the fingerprint of the public key matching the key specified in key_path. It can be obtained via the command ssh-keygen -l -E md5 -f /path/to/key
- `url` - (Optional) This is the URL to the Triton API endpoint. It is required if using a private installation of Triton. The default is to use the MNX.io public cloud us-west-1 endpoint. Valid public cloud endpoints include: us-east-1, us-east-2, us-east-3, us-sw-1, us-west-1, eu-ams-1
- `insecure_skip_tls_verify` (Optional - defaults to false) This allows skipping TLS verification of the Triton endpoint. It is useful when connecting to a temporary or development Triton installation.

Another option is to pass in account information through Triton's commonly used [environment variables](https://docs.tritondatacenter.com/public-cloud/api-access/cloudapi#environment-variables). The provider takes the following environment variables...

- `TRITON_ACCOUNT` or `SDC_ACCOUNT` with your Triton account name.
- `TRITON_KEY_MATERIAL` or `SDC_KEY_MATERIAL` with the contents of your private key attached to your Triton account.
- `TRITON_KEY_ID` or `SDC_KEY_ID` with a key id used to reference your Triton account's SSH key.
- `TRITON_URL` or `SDC_URL` with the URL to your CloudAPI endpoint, handy if using Terraform with a private Triton installation.
- `TRITON_SKIP_TLS_VERIFY` to skip TLS verification when connecting to `TRITON_URL`.

Finally, the provider will automatically pick up your Triton SSH key if you do not set `key_material` but are [using `ssh-agent`](https://docs.tritondatacenter.com/public-cloud/getting-started/ssh-keys).

### Resources and Data Providers ###

There are a wide range of [Triton resources and data providers](https://www.terraform.io/docs/providers/triton/index.html) available when building with the Triton Terraform Provider.

- [`triton_image`](https://www.terraform.io/docs/providers/triton/d/triton_image.html)
- [`triton_network`](https://www.terraform.io/docs/providers/triton/d/triton_network.html)
- [`triton_key`](https://www.terraform.io/docs/providers/triton/r/triton_key.html)
- [`triton_firewall_rule`](https://www.terraform.io/docs/providers/triton/r/triton_firewall_rule.html)
- [`triton_vlan`](https://www.terraform.io/docs/providers/triton/r/triton_vlan.html)
- [`triton_fabric`](https://www.terraform.io/docs/providers/triton/r/triton_fabric.html)
- [`triton_machine`](https://www.terraform.io/docs/providers/triton/r/triton_machine.html)
- [`triton_volume`](https://www.terraform.io/docs/providers/triton/r/triton_volume.html)

### Example ###

The following example shows you how to configure a LX branded zone running Ubuntu.

```hcl
# use env vars and SSH agent to configure the provider
provider "triton" {}

data "triton_image" "lx-ubuntu" {
    name = "ubuntu-24.04"
    version = "20250407"
}

resource "triton_machine" "test-cns" {
    name    = "test-cns"
    package = "g1.nano"
    image   = "${data.triton_image.lx-ubuntu.id}"

    cns {
        services = ["frontend", "app"]
    }
}
```

Visit Terraform's website for official [Triton Provider documentation](https://www.terraform.io/docs/providers/triton/index.html).

Developing the Provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.21+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-triton
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

The acceptance tests will pick up basic triton account configuration from your environment as the provider itself does. You can set this using the `node-triton` CLI tool in your shell:

```sh
$ eval "$(triton env)"
$ make testacc
...
```

To run only a sub-group of tests, for example all tests starting with `TestAccTritonFabricVLAN*`, you can use:

```sh
TESTARGS='-run=TestAccTritonFabricVLAN' make testacc
```

The provider's [log level](https://developer.hashicorp.com/terraform/plugin/log/managing) can be controlled by inserting `TF_LOG=DEBUG` before `go test` in `GNUmakefile`.

The acceptance tests are configured, by default, to run against the MNX.io Public Cloud. You can adjust the parameters used during test provisions by setting the configuration keys below to be appropriate to your environment. The tests expect the deployment to support fabric overlay networks - if these arent't set-up then the tests will fail. They also expect a version of the `base-64-lts@24.4.1` image to be available.

The default values for these parameters are specified in `triton/provider_test.go` and can be overridden by setting environment variables with the name prefixed with `testacc_`

* `dc_name`
  The datacenter name reported by CloudAPI

* `test_package_name`
  The name of a simple package that is used in the bulk of test provisions - this can be the lowest resource package available on your platform.

* `test_network_name`
  The name of a network that is available to the test user that the vast majority of test machines will be provisioned into.

* `public_network_name`
  The name of a _public_ network that is available to the test user

* `package_query_name`, `package_query_memory`, `package_query_result`
  These three values are used to test the package lookup - the `package_query_name` and `package_query_memory` should resolve to a single package with the name specified in `package_query_result`

You can also use the following environment variables to assist in debugging:

* `TRITON_TRACE_HTTP` when set this will print the HTTP requests and responses
  to stderr.

# Publishing

Follow the instructions here for publishing to the Terraform registry:
https://developer.hashicorp.com/terraform/registry/providers/publishing
