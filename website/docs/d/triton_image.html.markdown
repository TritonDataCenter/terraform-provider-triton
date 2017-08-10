---
layout: "triton"
page_title: "Triton: triton_image"
sidebar_current: "docs-triton-datasource-image"
description: |-
    The `triton_image` data source queries the Triton Image API for image IDs.
---

# triton\_image

The `triton_image` data source queries the Triton Image API for an image ID based
on a variety of different parameters.

## Example Usages

Find the ID of a Base 64 LTS image.

```hcl
data "triton_image" "base" {
	name = "base-64-lts"
	version = "16.4.1"
}

output "image_id" {
    value = "${data.triton_image.base.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (string)
    The name of the image

* `os` - (string)
    The underlying operating system for the image

* `version` - (string)
    The version for the image

* `public` - (boolean)
    Whether to return public as well as private images

* `state` - (string)
    The state of the image. By default, only `active` images are shown. Must be one of:
    `active`, `unactivated`, `disabled`, `creating`, `failed` or `all`, though the
    default is sufficient in almost every case.

* `owner` - (string)
    The UUID of the account which owns the image

* `type` - (string)
    The image type. Must be one of: `zone-dataset`, `lx-dataset`, `zvol`, `docker` or
    `other`.

* `most_recent` - (bool) If more than one result is returned, use the most recent Image.
