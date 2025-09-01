package triton

import (
	"fmt"
	"strings"
	"testing"

	triton "github.com/TritonDataCenter/triton-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	if triton.GetEnv("URL") == "" {
		return nil, fmt.Errorf("empty TRITON_URL")
	}

	if !strings.Contains(triton.GetEnv("URL"), region) {
		return nil, fmt.Errorf("SWEEP region " + region + " does not match TRITON_URL " + triton.GetEnv("URL") + ", aborting")
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
