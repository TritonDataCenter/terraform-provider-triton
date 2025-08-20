# Declare the data source.
data "triton_network" "private" {
  name = "My-Fabric-Network"
}

# Access unique Network ID using output from the data source. 
output "private_network_id" {
  value = data.triton_network.private.id
}
