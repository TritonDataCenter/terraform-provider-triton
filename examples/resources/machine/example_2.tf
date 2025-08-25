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
