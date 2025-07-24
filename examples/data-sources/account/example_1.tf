# Declare the data source.
data "triton_account" "main" {}

# Access unique Account ID using output from the data source.
output "account_id" {
  value = "${data.triton_account.main.id}"
}
