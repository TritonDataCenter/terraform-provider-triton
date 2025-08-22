resource "triton_key" "example" {
  name = "Example Key"
  key  = file("keys/id_rsa")
}
