Triton Terraform Provider
=========================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.8 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-triton`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-triton
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-triton
$ make build
```

Initialize your Terraform project by passing in the directory that contains your custom built provider binary, `terraform-provider-triton`. This is typically `$GOPATH/bin`.

```sh
$ terraform version
Terraform v0.10.0

$ terraform init --plugin-dir=$GOPATH/bin
```

Using the provider
------------------

If you haven't already done so, [create a Triton account](https://docs.joyent.com/public-cloud/getting-started) and read the getting started guide to complete the account setup and get your environment configured.

The provider automatically picks up [Triton environment variables](https://docs.joyent.com/public-cloud/api-access/cloudapi#environment-variables). You can also [configure it manually](https://www.terraform.io/docs/providers/triton/index.html#argument-reference).

There are a wide range of resources available when configuring the Triton Terraform Provider. Here's one example of how to configure a LX branded zone running Ubuntu.

```hcl
provider "triton" {}

data "triton_image" "lx-ubuntu" {
    name = "ubuntu-16.04"
    version = "20170403"
}

resource "triton_machine" "test-cns" {
    name    = "test-cns"
    package = "g4-highcpu-256M"
    image   = "${data.triton_image.lx-ubuntu.id}"

    cns {
        services = ["frontend", "app"]
    }
}
```

Visit Terraform's website for official [Triton Provider documentation](https://www.terraform.io/docs/providers/triton/index.html).

Developing the Provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-$PROVIDER_NAME
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

In order to _update_ existing vendored `triton-go` libraries, run `make vendor-triton`.

*NOTE*: Contributors will still need to add new sub-packages when new resources are added that depend on them.

```sh
$ make vendor-triton
```
