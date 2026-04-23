package triton

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedConfigForRegion(region string) (interface{}, error) {
	if getEnv("TRITON_ACCOUNT", "SDC_ACCOUNT") == "" {
		return nil, fmt.Errorf("empty TRITON_ACCOUNT")
	}

	if getEnv("TRITON_KEY_ID", "SDC_KEY_ID") == "" {
		return nil, fmt.Errorf("empty TRITON_KEY_ID")
	}

	if getEnv("TRITON_URL", "SDC_URL") == "" {
		return nil, fmt.Errorf("empty TRITON_URL")
	}

	tritonURL := getEnv("TRITON_URL", "SDC_URL")
	if !strings.Contains(tritonURL, region) {
		return nil, fmt.Errorf("SWEEP region %s does not match TRITON_URL %s, aborting", region, tritonURL)
	}

	config := Config{
		Account:               getEnv("TRITON_ACCOUNT", "SDC_ACCOUNT"),
		URL:                   tritonURL,
		KeyID:                 getEnv("TRITON_KEY_ID", "SDC_KEY_ID"),
		InsecureSkipTLSVerify: false,
	}

	if km := getEnv("TRITON_KEY_MATERIAL", "SDC_KEY_MATERIAL"); km != "" {
		config.KeyMaterial = km
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	client, err := config.newClient()
	if err != nil {
		return nil, err
	}

	return client, nil
}
