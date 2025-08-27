resource "triton_machine" "test-ubuntu" {
  name = "test-ubuntu"
  # ubuntu-24.04 20250407 lx-brand
  image                = "8a1b6e3a-00ec-4031-b0a8-8fb0f334c394"
  package              = "g1.small"
  firewall_enabled     = true
  root_authorized_keys = "Example Key"
  user_script          = "#!/bin/bash\necho 'testing user-script' >> /tmp/test.out\nhostname $IMAGENAME"

  tags = {
    purpose = "testing ubuntu lx-brand"
  }
}
