# Declare the data source.
data "triton_fabric_vlan" "public" {
  name = "Public-VLAN-Production"
}

# Access unique VLAN ID using output from the data source.
output "public_vlan_id" {
  value = data.triton_fabric_vlan.public.vlan_id
}
