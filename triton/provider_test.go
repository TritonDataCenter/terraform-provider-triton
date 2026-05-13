/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Joyent, Inc.
 * Copyright 2025 MNX Cloud, Inc.
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
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

// testAccNewClient creates an authenticated CloudAPI client from
// environment variables.  It is intended for test helpers that need to
// query the API outside of a Terraform test step (e.g. pre-checks,
// resource discovery).
func testAccNewClient(t *testing.T) *Client {
	t.Helper()
	config := testAccBuildConfig()
	if err := config.validate(); err != nil {
		t.Fatalf("testAccNewClient: %s", err)
	}
	client, err := config.newClient()
	if err != nil {
		t.Fatalf("testAccNewClient: %s", err)
	}
	return client
}

// testAccBuildConfig constructs a Config from the standard environment
// variables used by the provider.
func testAccBuildConfig() Config {
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
	if os.Getenv("TRITON_SKIP_TLS_VERIFY") != "" {
		config.InsecureSkipTLSVerify = true
	}
	return config
}

// testAccPreCheckCNS skips the test if Triton CNS is not enabled on
// the account.  Tests that assert on domain_names require CNS.
func testAccPreCheckCNS(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	client := testAccNewClient(t)
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

	default:
		t.Fatalf("Unknown acceptance test config key '%s'", key)
		return ""
	}
}

// ---------------------------------------------------------------------------
// Package discovery helpers
// ---------------------------------------------------------------------------

var (
	discoverPkgOnce sync.Once
	discoveredPkgs  []cloudapi.Package
	discoverPkgErr  error
)

// testAccDiscoverPackages fetches all packages from the target DC once
// per test run and caches the result.
func testAccDiscoverPackages(t *testing.T) []cloudapi.Package {
	t.Helper()
	discoverPkgOnce.Do(func() {
		config := testAccBuildConfig()
		if err := config.validate(); err != nil {
			discoverPkgErr = fmt.Errorf("package discovery: %s", err)
			return
		}
		client, err := config.newClient()
		if err != nil {
			discoverPkgErr = fmt.Errorf("package discovery: %s", err)
			return
		}
		resp, err := client.API().ListPackagesWithResponse(
			context.Background(), client.Account(),
			&cloudapi.ListPackagesParams{},
		)
		if err != nil {
			discoverPkgErr = fmt.Errorf("package discovery: %s", err)
			return
		}
		if resp.JSON200 == nil {
			discoverPkgErr = fmt.Errorf("package discovery: %s",
				formatAPIError(resp.StatusCode(), resp.Body))
			return
		}
		discoveredPkgs = *resp.JSON200
	})
	if discoverPkgErr != nil {
		t.Skipf("skipping: %s", discoverPkgErr)
	}
	return discoveredPkgs
}

type discoveredPkg struct {
	Name   string
	Memory uint64
	Brand  string
}

// testAccFindBrandedPackage returns the first package that has a
// non-empty brand.  Skips the calling test if no such package exists.
func testAccFindBrandedPackage(t *testing.T) discoveredPkg {
	t.Helper()
	for _, p := range testAccDiscoverPackages(t) {
		if b := vmBrandString(p.Brand); b != "" {
			return discoveredPkg{Name: p.Name, Memory: p.Memory, Brand: b}
		}
	}
	t.Skip("skipping: no branded packages available in this DC")
	return discoveredPkg{}
}
