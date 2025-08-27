data "triton_package" "nano" {
  filter {
    name   = "nano"
    memory = 512
  }
}

output "package_id" {
  value = data.triton_package.nano.id
}
