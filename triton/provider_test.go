package triton

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	triton "github.com/joyent/triton-go"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"triton": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	sdcURL := triton.GetEnv("URL")
	account := triton.GetEnv("ACCOUNT")
	keyID := triton.GetEnv("KEY_ID")

	if sdcURL == "" {
		sdcURL = "https://us-west-1.api.joyentcloud.com"
	}

	if sdcURL == "" || account == "" || keyID == "" {
		t.Fatal("TRITON_ACCOUNT and TRITON_KEY_ID must be set for acceptance tests. To test with the SSH" +
			" private key signer, TRITON_KEY_MATERIAL must also be set.")
	}
}

func testAccConfig(t *testing.T, key string) string {
	if key == "URL" {
		return triton.GetEnv("URL")
	}

	var env_value = os.Getenv(fmt.Sprintf("testacc_%s", key))
	if env_value != "" {
		return env_value
	}

	switch key {
	case "dc_name":
		return "us-sw-1"

	case "test_package_name":
		return "g4-highcpu-128M"

	case "test_network_name":
		return "Joyent-SDC-Public"

	case "public_network_name":
		return "Joyent-SDC-Public"

	case "package_query_name":
		return "highcpu"

	case "package_query_memory":
		return "128"

	case "package_query_result":
		return "g4-hughcpu-128M"

	default:
		t.Fatalf("Unknown acceptance test config key '%s'", key)
		return ""
	}
}
