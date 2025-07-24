data "triton_volume" "myvol" {
  name    = "my-volume-name"
}

output "volume_id" {
  value = "${data.triton_volume.myvol.id}"
}
