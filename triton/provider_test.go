package triton

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"triton": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// getEnv returns the first non-empty value among the given environment
// variable names, or "" if none is set.
func getEnv(names ...string) string {
	for _, n := range names {
		if v := os.Getenv(n); v != "" {
			return v
		}
	}
	return ""
}

func testAccPreCheck(t *testing.T) {
	sdcURL := getEnv("TRITON_URL", "SDC_URL")
	account := getEnv("TRITON_ACCOUNT", "SDC_ACCOUNT")
	keyID := getEnv("TRITON_KEY_ID", "SDC_KEY_ID")

	if sdcURL == "" {
		sdcURL = "https://us-central-1.api.mnx.io"
	}

	if sdcURL == "" || account == "" || keyID == "" {
		t.Fatal("TRITON_ACCOUNT and TRITON_KEY_ID must be set for acceptance tests. To test with the SSH" +
			" private key signer, TRITON_KEY_MATERIAL must also be set.")
	}
}

// testAccPreCheckCNS skips the test if Triton CNS is not enabled on
// the account.  Tests that assert on domain_names require CNS.
func testAccPreCheckCNS(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	config := Config{
		Account: getEnv("TRITON_ACCOUNT", "SDC_ACCOUNT"),
		URL:     getEnv("TRITON_URL", "SDC_URL"),
		KeyID:   getEnv("TRITON_KEY_ID", "SDC_KEY_ID"),
	}
	if config.URL == "" {
		config.URL = "https://us-central-1.api.mnx.io"
	}
	if km := getEnv("TRITON_KEY_MATERIAL", "SDC_KEY_MATERIAL"); km != "" {
		config.KeyMaterial = km
	}
	if err := config.validate(); err != nil {
		t.Fatalf("testAccPreCheckCNS: %s", err)
	}
	client, err := config.newClient()
	if err != nil {
		t.Fatalf("testAccPreCheckCNS: %s", err)
	}
	enabled, err := client.CNSEnabled()
	if err != nil {
		t.Fatalf("testAccPreCheckCNS: %s", err)
	}
	if !enabled {
		t.Skip("skipping: triton_cns_enabled is false on this account")
	}
}

func testAccConfig(t *testing.T, key string) string {
	if key == "URL" {
		return getEnv("TRITON_URL", "SDC_URL")
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

	case "package_query_brand":
		return "joyent"

	default:
		t.Fatalf("Unknown acceptance test config key '%s'", key)
		return ""
	}
}
