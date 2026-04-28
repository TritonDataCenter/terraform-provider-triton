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
	"sync"
	"testing"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// newTestClient creates a Client backed by the given httptest server.
func newTestClient(t *testing.T, ts *httptest.Server) *Client {
	t.Helper()
	api, err := cloudapi.NewClientWithResponses(ts.URL)
	if err != nil {
		t.Fatalf("creating test client: %s", err)
	}
	return &Client{
		api:          api,
		account:      "testaccount",
		url:          ts.URL,
		affinityLock: &sync.RWMutex{},
	}
}

func TestResolvePackageValue(t *testing.T) {
	const (
		pkgName = "sample-4G"
		pkgUUID = "fa5fc249-d1d8-4247-8aba-766fb39c96f4"
	)

	pkgJSON, _ := json.Marshal(cloudapi.Package{
		ID:     mustParseUUID(t, pkgUUID),
		Name:   pkgName,
		Memory: 4096,
		Disk:   65536,
		Swap:   8192,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Expect: GET /testaccount/packages/sample-4G
		w.Header().Set("Content-Type", "application/json")
		w.Write(pkgJSON)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)

	t.Run("config is UUID, API returns name", func(t *testing.T) {
		got, err := resolvePackageValue(client, pkgName, pkgUUID)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got != pkgUUID {
			t.Errorf("expected %q, got %q", pkgUUID, got)
		}
	})

	t.Run("config is name, API returns name", func(t *testing.T) {
		got, err := resolvePackageValue(client, pkgName, pkgName)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got != pkgName {
			t.Errorf("expected %q, got %q", pkgName, got)
		}
	})
}

func TestResolvePackageName(t *testing.T) {
	const (
		pkgName = "sample-4G"
		pkgUUID = "fa5fc249-d1d8-4247-8aba-766fb39c96f4"
	)

	pkgJSON, _ := json.Marshal(cloudapi.Package{
		ID:     mustParseUUID(t, pkgUUID),
		Name:   pkgName,
		Memory: 4096,
		Disk:   65536,
		Swap:   8192,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(pkgJSON)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)

	t.Run("config is UUID, resolves to name", func(t *testing.T) {
		got, err := resolvePackageName(client, pkgUUID)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got != pkgName {
			t.Errorf("expected %q, got %q", pkgName, got)
		}
	})

	t.Run("config is name, returns unchanged", func(t *testing.T) {
		got, err := resolvePackageName(client, pkgName)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got != pkgName {
			t.Errorf("expected %q, got %q", pkgName, got)
		}
	})
}

func mustParseUUID(t *testing.T, s string) openapi_types.UUID {
	t.Helper()
	u, err := parseUUID(s)
	if err != nil {
		t.Fatalf("parsing UUID %q: %s", s, err)
	}
	return u
}
