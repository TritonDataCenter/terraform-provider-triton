provider "triton" {
  account = "AccountName"
  key_id  = "25:d4:a9:fe:ef:e6:c0:bf:b4:4b:4b:d4:a8:8f:01:0f"

  # If using a private installation of Triton, specify the URL, otherwise
  # set the URL according to the region you wish to provision.
  url = "https://us-central-1.api.mnx.io"

  # If you want to use a triton sub user of the main account, then
  # you can specify the username as follows
  #user = "myusername"

  # If using a test Triton installation (self-signed certifcate), use:
  #insecure_skip_tls_verify = true
}
