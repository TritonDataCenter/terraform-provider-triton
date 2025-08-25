data "triton_network" "my_fabric" {
  name = "My-Fabric-Network"
}

resource "triton_volume" "my_volume" {
  networks = ["${data.triton_network.my_fabric.id}"]
  size     = 10240
}
