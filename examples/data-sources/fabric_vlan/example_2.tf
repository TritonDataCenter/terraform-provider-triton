# Declare the data source, and use a combination of two arguments
# to form a search filter. Use a wildcard match for the name.
data "triton_fabric_vlan" "private_database_vlan" {
  name        = "Private-VLAN-*"
  description = "A secure VLAN for production database servers"
}

# Access unique VLAN ID using output from the data source.
output "private_database_vlan_id" {
  value = data.triton_fabric_vlan.private_database_vlan.vlan_id
}
