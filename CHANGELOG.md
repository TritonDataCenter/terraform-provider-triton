## 0.1.2 (Unreleased)

FEATURES:

* `resource/triton_machine`: CNS tags can now be specified in a separate stanza [GH-17]
* `resource/triton_machine`: User-defined metadata is now defined in a separate stanza [GH-17]

## 0.1.1 (August 01, 2017)

FEATURES:

* *New Data Source:* `triton_image` ([#7](https://github.com/terraform-providers/terraform-provider-triton/issues/7))
* *New Data Source:* `triton_network` ([#13](https://github.com/terraform-providers/terraform-provider-triton/issues/13))

BUG FIXES:

* `resource/triton_machine`: Instances which fail during provisioning are now detected and tainted rather than timing out ([#10](https://github.com/terraform-providers/terraform-provider-triton/issues/10))
* `resource/triton_machine`: Metadata is now populated correctly ([#12](https://github.com/terraform-providers/terraform-provider-triton/issues/12))

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
