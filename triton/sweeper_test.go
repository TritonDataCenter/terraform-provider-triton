package triton

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedConfigForRegion(region string) (interface{}, error) {
	if os.Getenv("TRITON_ACCOUNT") == "" {
		return nil, fmt.Errorf("empty TRITON_ACCOUNT")
	}

	if os.Getenv("TRITON_KEY_ID") == "" {
		return nil, fmt.Errorf("empty TRITON_KEY_ID")
	}

	regionUrl := fmt.Sprintf("https://%s.api.joyentcloud.com", region)

	config := Config{
		Account: os.Getenv("TRITON_ACCOUNT"),
		URL:     regionUrl,
		KeyID:   os.Getenv("TRITON_KEY_ID"),
		InsecureSkipTLSVerify: false,
	}

	if os.Getenv("TRITON_KEY_MATERIAL") != "" {
		config.KeyMaterial = os.Getenv("TRITON_KEY_MATERIAL")
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
