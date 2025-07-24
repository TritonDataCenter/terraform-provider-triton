resource "triton_firewall_rule" "22" {
  description = "Allow ssh traffic on port tcp/22 to all machines from known remote IPs."
  rule        = "FROM (ip w.x.y.z OR ip w.x.y.z) TO all vms ALLOW tcp PORT 22"
  enabled     = true
}
