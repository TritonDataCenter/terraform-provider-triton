resource "triton_volume" "my_volume_1" {
}

resource "triton_volume" "my_volume_2" {
}

resource "triton_machine" {
  # base-64-lts 24.4.1
  image   = "2f1dc911-6401-4fa4-8e9d-67ea2e39c271"
  package = "g1.medium"

  volume {
    name       = triton_volume.my_volume_1.name
    mountpoint = "/data1"
  }

  volume {
    name       = triton_volume.my_volume_2.name
    mode       = "ro"
    mountpoint = "/data2"
  }
}
