data "triton_image" "base" {
  name    = "base-64-lts"
  version = "24.4.1"
}

output "image_id" {
  value = data.triton_image.base.id
}
