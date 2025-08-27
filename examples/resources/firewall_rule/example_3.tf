resource "triton_firewall_rule" "imap" {
  description = "Block IMAP traffic on port tcp/143 to all machines."
  rule        = "FROM any TO all vms BLOCK tcp PORT 143"
  enabled     = true
}
