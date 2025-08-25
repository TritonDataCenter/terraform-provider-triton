resource "triton_vlan" "dmz" {
  vlan_id     = 100
  name        = "dmz"
  description = "DMZ VLAN"
}
