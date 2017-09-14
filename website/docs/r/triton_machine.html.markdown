---
layout: "triton"
page_title: "Triton: triton_machine"
sidebar_current: "docs-triton-resource-machine"
description: |-
    The `triton_machine` resource represents a virtual machine or infrastructure container running in Triton.
---

# triton\_machine

The `triton_machine` resource represents a virtual machine or infrastructure container running in Triton.

~> **Note:** Starting with Triton 0.2.0, Please note that when you want to specify the networks that you want the machine to be attached to, use the `networks` parameter
and not the `nic` parameter.

## Example Usages

### Run a SmartOS base-64 machine.

```hcl
resource "triton_machine" "test-smartos" {
  name    = "test-smartos"
  package = "g3-standard-0.25-smartos"
  image   = "842e6fa6-6e9b-11e5-8402-1b490459e334"

  tags {
    hello = "world"
    role = "database"
  }

  cns {
    services = ["web", "frontend"]
  }

  metadata {
    hello = "again"
  }

}
```

### Attaching a Machine to Joyent public network

```hcl
data "triton_image" "image" {
    name = "base-64-lts"
    version = "16.4.1"
}

data "triton_network" "public" {
    name = "Joyent-SDC-Public"
}

resource "triton_machine" "test" {
    package = "g4-highcpu-128M"
    image   = "${data.triton_image.image.id}"
    networks = ["${data.triton_network.public.id}"]
   }
```

### Run an Ubuntu 14.04 LTS machine.

```hcl
resource "triton_machine" "test-ubuntu" {
  name                 = "test-ubuntu"
  package              = "g4-general-4G"
  image                = "1996a1d6-c0d9-11e6-8b80-4772e39dc920"
  firewall_enabled     = true
  root_authorized_keys = "Example Key"
  user_script          = "#!/bin/bash\necho 'testing user-script' >> /tmp/test.out\nhostname $IMAGENAME"

  tags = {
    purpose = "testing ubuntu"
  } ## tags
} ## resource
```

### Run two SmartOS machine's with placement rules.

```hcl
resource "triton_machine" "test-db" {
  name    = "test-db"
  package = "g4-highcpu-8G"
  image   = "842e6fa6-6e9b-11e5-8402-1b490459e334"

  affinity = ["role!=~web"]

  tags {
    role = "database"
  }
}

resource "triton_machine" "test-web" {
  name    = "test-web"
  package = "g4-highcpu-8G"
  image   = "842e6fa6-6e9b-11e5-8402-1b490459e334"

  tags {
    role = "web"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (string)
    The friendly name for the machine. Triton will generate a name if one is not specified.

* `tags` - (map)
    A mapping of tags to apply to the machine.

* `cns` - (map of CNS attributes, Optional)
    A mapping of [CNS](https://docs.joyent.com/public-cloud/network/cns) attributes to apply to the machine.

* `metadata` - (map, optional)
    A mapping of metadata to apply to the machine.

* `package` - (string, Required)
    The name of the package to use for provisioning.

* `image` - (string, Required)
    The UUID of the image to provision.

* `networks` - (list, optional)
    The list of networks to associate with the machine. The network ID will be in hex form, e.g `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.

* `affinity` - (list of Affinity rules, Optional)
    A list of valid [Affinity Rules](https://apidocs.joyent.com/cloudapi/#affinity-rules) to apply to the machine which assist in data center placement. Using this attribute will force resource creation to be serial. NOTE: Affinity rules are best guess and assist in placing instances across a data center. They're used at creation and not referenced after.

* `locality` - (map of Locality hints, Optional)
    A mapping of [Locality](https://apidocs.joyent.com/cloudapi/#CreateMachine) attributes to apply to the machine that assist in data center placement. NOTE: Locality hints are only used at the time of machine creation and not referenced after.

* `firewall_enabled` - (boolean)  Default: `false`
    Whether the cloud firewall should be enabled for this machine.

* `root_authorized_keys` - (string)
    The public keys authorized for root access via SSH to the machine.

* `user_data` - (string)
    Data to be copied to the machine on boot.

* `user_script` - (string)
    The user script to run on boot (every boot on SmartMachines).

* `administrator_pw` - (string)
    The initial password for the Administrator user. Only used for Windows virtual machines.

* `cloud_config` - (string)
    Cloud-init configuration for Linux brand machines, used instead of `user_data`.

## Attribute Reference

The following attributes are exported:

* `id` - (string) - The identifier representing the firewall rule in Triton.
* `type` - (string) - The type of the machine (`smartmachine` or `virtualmachine`).
* `state` - (string) - The current state of the machine.
* `dataset` - (string) - The dataset URN with which the machine was provisioned.
* `memory` - (int) - The amount of memory the machine has (in Mb).
* `disk` - (int) - The amount of disk the machine has (in Gb).
* `ips` - (list of strings) - IP addresses of the machine.
* `primaryip` - (string) - The primary (public) IP address for the machine.
* `created` - (string) - The time at which the machine was created.
* `updated` - (string) - The time at which the machine was last updated.

* `nic` - A list of the networks that the machine is attached to. Each network is represented by a `nic`, each of which has the following properties:

* `ip` - The NIC's IPv4 address
* `mac` - The NIC's MAC address
* `primary` - Whether this is the machine's primary NIC
* `netmask` - IPv4 netmask
* `gateway` - IPv4 Gateway
* `network` - The ID of the network to which the NIC is attached
* `state` - The provisioning state of the NIC

The following attributes are used by `cns`:

* `services` - (list of strings) - The list of services that group this instance with others under a shared domain name.
* `disable` - (boolean) - The ability to temporarily disable CNS services domains (optional).

The following attributes are used as `locality` hints:

* `close_to` - (list of strings) - List of container UUIDs that a new instance should be placed alongside, on the same host.
* `far_from` - (list of strings) - List of container UUIDs that a new instance should not be placed onto the same host.
