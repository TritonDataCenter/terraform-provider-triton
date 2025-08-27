resource "triton_service_group" "web" {
  group_name = "web_group"
  template   = triton_instance_template.base.id
  capacity   = 3
}
