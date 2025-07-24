# Declare the data source to retrieve Fabric VLAN details.
data "triton_fabric_vlan" "private" {
  name = "Private-VLAN-Production"
}

# Declare the data source to retrieve Fabric Network details.
data "triton_fabric_network" "private" {
  name     = "Private-Network-Production"
  vland_id = "${data.triton_fabric_vlan.private.vlan_id}"
}

# Access subnet CIDR using output from the data source.
output "private_network_cidr" {
  value = "${data.triton_fabric_network.private.subnet}"
}
