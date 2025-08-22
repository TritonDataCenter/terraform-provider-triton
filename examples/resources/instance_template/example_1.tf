data "triton_image" "base" {
  name    = "base-64-lts"
  version = "24.4.1"
}

data "triton_network" "private" {
  name = "My-Fabric-Network"
}

resource "triton_instance_template" "base" {
  template_name = "Base template"
  image         = data.triton_image.base.id
  package       = "g1.nano"

  firewall_enabled = false

  networks = ["${data.triton_network.private.id}"]

  tags {
    hello = "world"
    role  = "database"
  }
}
