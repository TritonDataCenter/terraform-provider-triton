resource "triton_fabric" "dmz" {
  vlan_id            = 100
  name               = "dmz"
  description        = "DMZ Network"
  subnet             = "10.60.1.0/24"
  provision_start_ip = "10.60.1.10"
  provision_end_ip   = "10.60.1.240"
  gateway            = "10.60.1.1"
  resolvers          = ["8.8.8.8", "8.8.4.4"]

  internet_nat = true
}
