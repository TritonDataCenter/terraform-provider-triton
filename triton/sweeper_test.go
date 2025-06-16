package triton

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	triton "github.com/TritonDataCenter/triton-go"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedConfigForRegion(region string) (interface{}, error) {
	if triton.GetEnv("ACCOUNT") == "" {
		return nil, fmt.Errorf("empty TRITON_ACCOUNT")
	}

	if triton.GetEnv("KEY_ID") == "" {
		return nil, fmt.Errorf("empty TRITON_KEY_ID")
	}

	config := Config{
		Account:               triton.GetEnv("ACCOUNT"),
		URL:                   triton.GetEnv("URL"),
		KeyID:                 triton.GetEnv("KEY_ID"),
		InsecureSkipTLSVerify: false,
	}

	if triton.GetEnv("KEY_MATERIAL") != "" {
		config.KeyMaterial = triton.GetEnv("KEY_MATERIAL")
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
