/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// newPackageReadData builds a ResourceData for the package data source
// with the given filter values.  Only non-zero fields are meaningful;
// zero-valued fields are ignored by dataSourcePackageRead.
func newPackageReadData(t *testing.T, filter map[string]interface{}) *schema.ResourceData {
	t.Helper()

	// Ensure every Optional filter field has a zero value so the type
	// assertions inside dataSourcePackageRead don't panic.
	defaults := map[string]interface{}{
		"name":          "",
		"memory":        0,
		"disk":          0,
		"swap":          0,
		"lwps":          0,
		"vcpus":         0,
		"version":       "",
		"group":         "",
		"brand":         "",
		"flexible_disk": false,
	}
	merged := map[string]interface{}{}
	for k, v := range defaults {
		merged[k] = v
	}
	for k, v := range filter {
		merged[k] = v
	}

	return schema.TestResourceDataRaw(t, dataSourcePackage().Schema, map[string]interface{}{
		"filter": []interface{}{merged},
	})
}

// servePkgs returns an httptest server that responds to any request
// with the JSON-encoded package slice.
func servePkgs(t *testing.T, pkgs []cloudapi.Package) *httptest.Server {
	t.Helper()
	body, err := json.Marshal(pkgs)
	if err != nil {
		t.Fatalf("marshaling packages: %s", err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
}

func TestDataSourcePackageRead_singleResultNoName(t *testing.T) {
	// Server returns exactly one package.  Filter uses memory only
	// (no name).  This should succeed — the single result is
	// unambiguous.
	pkgs := []cloudapi.Package{
		{
			ID:     mustParseUUID(t, "fa5fc249-d1d8-4247-8aba-766fb39c96f4"),
			Name:   "test-4G",
			Memory: 4096,
			Disk:   65536,
			Swap:   8192,
		},
	}

	ts := servePkgs(t, pkgs)
	defer ts.Close()

	client := newTestClient(t, ts)
	d := newPackageReadData(t, map[string]interface{}{
		"memory": 4096,
	})

	err := dataSourcePackageRead(d, client)
	if err != nil {
		t.Fatalf("expected success for single result without name filter, got: %s", err)
	}
	if got := d.Get("name"); got != "test-4G" {
		t.Errorf("name: got %q, want %q", got, "test-4G")
	}
}

func TestDataSourcePackageRead_multipleResultsNoName(t *testing.T) {
	// Server returns two packages.  No name filter.  Should error
	// because the result is ambiguous.
	pkgs := []cloudapi.Package{
		{
			ID:     mustParseUUID(t, "fa5fc249-d1d8-4247-8aba-766fb39c96f4"),
			Name:   "test-4G",
			Memory: 4096,
			Disk:   65536,
			Swap:   8192,
		},
		{
			ID:     mustParseUUID(t, "ba5fc249-d1d8-4247-8aba-766fb39c96f4"),
			Name:   "test-4G-ssd",
			Memory: 4096,
			Disk:   131072,
			Swap:   8192,
		},
	}

	ts := servePkgs(t, pkgs)
	defer ts.Close()

	client := newTestClient(t, ts)
	d := newPackageReadData(t, map[string]interface{}{
		"memory": 4096,
	})

	err := dataSourcePackageRead(d, client)
	if err == nil {
		t.Fatal("expected error for multiple results without name filter, got nil")
	}
	if !strings.Contains(err.Error(), "more than one result") {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestDataSourcePackageRead_nameMatchesOne(t *testing.T) {
	// Two packages returned, name filter matches exactly one.
	pkgs := []cloudapi.Package{
		{
			ID:     mustParseUUID(t, "fa5fc249-d1d8-4247-8aba-766fb39c96f4"),
			Name:   "test-4G",
			Memory: 4096,
			Disk:   65536,
			Swap:   8192,
		},
		{
			ID:     mustParseUUID(t, "ba5fc249-d1d8-4247-8aba-766fb39c96f4"),
			Name:   "other-8G",
			Memory: 8192,
			Disk:   131072,
			Swap:   16384,
		},
	}

	ts := servePkgs(t, pkgs)
	defer ts.Close()

	client := newTestClient(t, ts)
	d := newPackageReadData(t, map[string]interface{}{
		"name": "test",
	})

	err := dataSourcePackageRead(d, client)
	if err != nil {
		t.Fatalf("expected success for single name match, got: %s", err)
	}
	if got := d.Get("name"); got != "test-4G" {
		t.Errorf("name: got %q, want %q", got, "test-4G")
	}
}

func TestDataSourcePackageRead_nameMatchesMultiple(t *testing.T) {
	// Two packages both contain the name substring.  Should error
	// on ambiguity, not silently pick the first.
	pkgs := []cloudapi.Package{
		{
			ID:     mustParseUUID(t, "fa5fc249-d1d8-4247-8aba-766fb39c96f4"),
			Name:   "test-4G",
			Memory: 4096,
			Disk:   65536,
			Swap:   8192,
		},
		{
			ID:     mustParseUUID(t, "ba5fc249-d1d8-4247-8aba-766fb39c96f4"),
			Name:   "test-4G-ssd",
			Memory: 4096,
			Disk:   131072,
			Swap:   8192,
		},
	}

	ts := servePkgs(t, pkgs)
	defer ts.Close()

	client := newTestClient(t, ts)
	d := newPackageReadData(t, map[string]interface{}{
		"name":   "test",
		"memory": 4096,
	})

	err := dataSourcePackageRead(d, client)
	if err == nil {
		t.Fatal("expected error for ambiguous name match, got nil")
	}
	if !strings.Contains(err.Error(), "more than one result") {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestDataSourcePackageRead_nameMatchesNone(t *testing.T) {
	// Server returns packages but none match the name substring.
	pkgs := []cloudapi.Package{
		{
			ID:     mustParseUUID(t, "fa5fc249-d1d8-4247-8aba-766fb39c96f4"),
			Name:   "test-4G",
			Memory: 4096,
			Disk:   65536,
			Swap:   8192,
		},
	}

	ts := servePkgs(t, pkgs)
	defer ts.Close()

	client := newTestClient(t, ts)
	d := newPackageReadData(t, map[string]interface{}{
		"name": "nonexistent",
	})

	err := dataSourcePackageRead(d, client)
	if err == nil {
		t.Fatal("expected error when name matches nothing, got nil")
	}
	if !strings.Contains(err.Error(), "no packages matched") {
		t.Errorf("expected 'no packages matched' error, got: %s", err)
	}
}
