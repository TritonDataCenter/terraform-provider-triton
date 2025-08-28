---
page_title: "triton_machine Resource - triton"
description: |-
    The `triton_machine` resource represents a virtual machine or infrastructure container running in Triton.
---

# triton_machine (Resource)

The `triton_machine` resource represents a virtual machine or infrastructure container running in Triton.

~> **Note:** Starting with Triton 0.2.0, Please note that when you want to specify the networks that you want the machine to be attached to, use the `networks` parameter and not the `nic` parameter.

## Example Usage

### Run a SmartOS base-64 machine.

```terraform
resource "triton_machine" "test-smartos" {
  name = "test-smartos"
  # base-64-lts 24.4.1
  image   = "2f1dc911-6401-4fa4-8e9d-67ea2e39c271"
  package = "g1.nano"

  tags = {
    hello = "world"
    role  = "database"
  }

  cns {
    services = ["web", "frontend"]
  }

  metadata = {
    hello = "again"
  }

  volume {
    name       = "my_volume"
    mountpoint = "/data"
  }
}
```

### Attaching a Machine to Triton public network

```terraform
data "triton_image" "image" {
  name    = "base-64-lts"
  version = "24.4.1"
}

data "triton_network" "public" {
  name = "MNX-Triton-Public"
}

resource "triton_machine" "test" {
  package  = "g1.nano"
  image    = data.triton_image.image.id
  networks = ["${data.triton_network.public.id}"]
}
```

### Run an Ubuntu 24.04 LTS lx-brand machine.

```terraform
resource "triton_machine" "test-ubuntu" {
  name = "test-ubuntu"
  # ubuntu-24.04 20250407 lx-brand
  image                = "8a1b6e3a-00ec-4031-b0a8-8fb0f334c394"
  package              = "g1.small"
  firewall_enabled     = true
  root_authorized_keys = "Example Key"
  user_script          = "#!/bin/bash\necho 'testing user-script' >> /tmp/test.out\nhostname $IMAGENAME"

  tags = {
    purpose = "testing ubuntu lx-brand"
  }
}
```

### Run two SmartOS machine's with placement rules.

```terraform
resource "triton_machine" "test-db" {
  name = "test-db"
  # base-64-lts 24.4.1
  image   = "2f1dc911-6401-4fa4-8e9d-67ea2e39c271"
  package = "g1.medium"

  affinity = ["role!=~web"]

  tags = {
    role = "database"
  }
}

resource "triton_machine" "test-web" {
  name = "test-web"
  # base-64-lts 24.4.1
  image   = "2f1dc911-6401-4fa4-8e9d-67ea2e39c271"
  package = "g1.medium"

  tags = {
    role = "web"
  }
}
```

## Argument Reference

The following arguments are required:

* `package` - (string, Required) The name of the package to use for provisioning.

* `image` - (string, Required) The UUID of the image to provision.

The following arguments are optional:

* `name` - (string, optional) The friendly name for the machine. Triton will generate a name if one is not specified.

* `tags` - (map, optional) A mapping of tags to apply to the machine.

* `cns` - (map of [CNS](#cns-map) attributes, optional) A mapping of [CNS](https://docs.tritondatacenter.com/public-cloud/network/cns) attributes to apply to the machine.

* `metadata` - (map, optional) A mapping of metadata to apply to the machine.

* `networks` - (list[string], optional) The list of networks to associate with the machine. The network ID will be in hex form, e.g `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.

* `affinity` - (list[string] of Affinity rules, optional) A list of valid [Affinity Rules](https://apidocs.tritondatacenter.com/cloudapi/#affinity-rules) to apply to the machine which assist in data center placement. Using this attribute will force resource creation to be serial. NOTE: Affinity rules are best guess and assist in placing instances across a data center. They're used at creation and not referenced after.

* `(Deprecated) locality` - ([Locality](#locality-map) map, optional) A mapping of [Locality](https://apidocs.tritondatacenter.com/cloudapi/#CreateMachine) attributes to apply to the machine that assist in data center placement. NOTE: Locality hints are only used at the time of machine creation and not referenced after. Locality is deprecated as of [CloudAPI v8.3.0](https://apidocs.tritondatacenter.com/cloudapi/#830).

* `firewall_enabled` - (boolean, optional) Default: `false` Whether the cloud firewall should be enabled for this machine.

* `root_authorized_keys` - (string, optional) The public keys authorized for root access via SSH to the machine.

* `user_data` - (string, optional) Data to be copied to the machine on boot. **NOTE:** The content of `user_data` will *not be executed* on boot. The data will only be written to the file on each boot before the content of the script from `user_script` is to be run.

* `user_script` - (string, optional) The user script to run on boot (every boot on SmartMachines). To learn more about both the user script and user data see the [metadata API](https://docs.tritondatacenter.com/private-cloud/instances/using-mdata) documentation and the [TritonDataCenter Metadata Data Dictionary](https://eng.tritondatacenter.com/mdata/datadict.html) specification.

* `administrator_pw` - (string, optional) The initial password for the Administrator user. Only used for Windows virtual machines.

* `cloud_config` - (string, optional) Cloud-init configuration for Linux brand machines, used instead of `user_data`.

* `deletion_protection_enabled` - (bool, optional) Whether an instance is destroyable. Default is `false`.

* `delegate_dataset` - (bool, optional) Whether an instance is created with a delegate dataset. Default is `false`.

* `volume` - ([Volume](#volume-map) map, optional) A volume to attach to the instance. Volume configurations only apply on resource creation. Multiple *volume*'s entries are allowed.

## Attribute Reference

The following attributes are exported:

* `id` - (string) - The identifier representing the machine in Triton.
* `type` - (string) - The type of the machine (`smartmachine` or `virtualmachine`).
* `state` - (string) - The current state of the machine.
* `dataset` - (string) - The dataset URN with which the machine was provisioned.
* `memory` - (int) - The amount of memory the machine has (in Mb).
* `disk` - (int) - The amount of disk the machine has (in Gb).
* `ips` - (list of strings) - IP addresses of the machine.
* `primaryip` - (string) - The primary (public) IP address for the machine.
* `created` - (string) - The time at which the machine was created.
* `updated` - (string) - The time at which the machine was last updated.
* `compute_node` - (string) - UUID of the server on which the instance is located.

* `nic` - A list of the networks that the machine is attached to. Each network is represented by a `nic`, each of which has the following properties:

  * `ip` - The NIC's IPv4 address
  * `mac` - The NIC's MAC address
  * `primary` - Whether this is the machine's primary NIC
  * `netmask` - IPv4 netmask
  * `gateway` - IPv4 Gateway
  * `network` - The ID of the network to which the NIC is attached
  * `state` - The provisioning state of the NIC

### CNS map

The following attributes are used by `cns`:

* `services` - (list of strings) - The list of services that group this instance with others under a shared domain name.
* `disable` - (boolean) - The ability to temporarily disable CNS services domains (optional).

### Locality map

The following attributes are used as `locality` hints:

* `close_to` - (list of strings) - List of container UUIDs that a new instance should be placed alongside, on the same host.
* `far_from` - (list of strings) - List of container UUIDs that a new instance should not be placed onto the same host.

### Volume map

Each *volume* map can entry contain the following attributes:

* `name` - (string) - The name of the volume
* `mountpoint` - (string) - Where the volume will be mounted
* `mode` - (optional, string) - Can be *"rw"* (the default) which means read-write, or *"ro"* for read-only
* `type` - (optional, string) - The type of volume (defaults to *"tritonnfs"*).

## Import

`triton_machine` resources can be imported using the instance UUID, for example:

```shell
terraform import triton_machine.example 4c0bc531-38a4-4919-8065-828a56a3b818
```
