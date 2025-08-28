resource "triton_key" "example-file" {
  name = "Example Key"
  key  = file("keys/id_rsa.pub")
}
