data "triton_image" "ubuntu2404" {
  name    = "ubuntu-24.04"
  version = "20250407"
}

resource "triton_machine" "test" {
  image   = "${data.triton_image.ubuntu2404.id}"
  package = "g1.nano"
}

resource "triton_snapshot" "test" {
  name       = "my-snapshot"
  machine_id = "${triton_machine.test.id}"
}
