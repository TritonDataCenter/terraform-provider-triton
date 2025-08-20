resource "triton_machine" "test-db" {
  name = "test-db"
  # base-64-lts 24.4.1
  image   = "2f1dc911-6401-4fa4-8e9d-67ea2e39c271"
  package = "g1.medium"

  affinity = ["role!=~web"]

  tags = {
    role = "database"
  }
}

resource "triton_machine" "test-web" {
  name = "test-web"
  # base-64-lts 24.4.1
  image   = "2f1dc911-6401-4fa4-8e9d-67ea2e39c271"
  package = "g1.medium"

  tags = {
    role = "web"
  }
}
