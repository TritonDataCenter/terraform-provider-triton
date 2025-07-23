package triton

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	triton "github.com/TritonDataCenter/triton-go"
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
		sdcURL = "https://us-central-1.api.mnx.io"
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
		return "us-central-1"

	case "test_package_name":
		return "g1.nano"

	case "test_network_name":
		return "My-Fabric-Network"

	case "public_network_name":
		return "MNX-Triton-Public"

	case "package_query_name":
		return "nano"

	case "package_query_memory":
		return "512"

	case "package_query_result":
		return "g1.nano"

	default:
		t.Fatalf("Unknown acceptance test config key '%s'", key)
		return ""
	}
}
