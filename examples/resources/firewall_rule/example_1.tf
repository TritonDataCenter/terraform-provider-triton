resource "triton_firewall_rule" "www" {
  description = "Allow web traffic on ports tcp/80 and tcp/443 to machines with the 'www' tag from any source."
  rule        = "FROM any TO tag \"www\" ALLOW tcp (PORT 80 AND PORT 443)"
  enabled     = true
}
