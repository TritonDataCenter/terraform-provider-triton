resource "triton_volume" "my-volume" {
  name    = "my-volume"

  tags {
    hello = "world"
    role  = "database"
  }
}
