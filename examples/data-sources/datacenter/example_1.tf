# Declare the data source.
data "triton_datacenter" "current" {}

# Access current endpoint URL using output from the data source.
output "endpoint" {
  value = data.triton_datacenter.current.endpoint
}
