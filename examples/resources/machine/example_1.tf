resource "triton_machine" "test-smartos" {
  name = "test-smartos"
  # base-64-lts 24.4.1
  image   = "2f1dc911-6401-4fa4-8e9d-67ea2e39c271"
  package = "g1.nano"

  tags = {
    hello = "world"
    role  = "database"
  }

  cns {
    services = ["web", "frontend"]
  }

  metadata = {
    hello = "again"
  }

  volume {
    name       = "my_volume"
    mountpoint = "/data"
  }
}
