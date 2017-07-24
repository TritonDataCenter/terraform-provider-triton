## 0.1.1 (Unreleased)

FEATURES:

* *New Data Source:* `triton_image` [GH-7]

BUG FIXES:

* `resource/triton_machine`: Instances which fail during provisioning are now detected and tainted rather than timing out [GH-10]
* `resource/triton_machine`: Metadata is now populated correctly [GH-12]

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
