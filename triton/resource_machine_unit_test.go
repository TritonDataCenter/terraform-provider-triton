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
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
		_, _ = w.Write(pkgJSON)
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
		_, _ = w.Write(pkgJSON)
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

// cnsEqual compares two InstanceCNS values field by field.
func cnsEqual(t *testing.T, got, want InstanceCNS) {
	t.Helper()
	if got.Disable != want.Disable {
		t.Errorf("Disable: got %v, want %v", got.Disable, want.Disable)
	}
	if !reflect.DeepEqual(got.Services, want.Services) {
		t.Errorf("Services: got %v, want %v", got.Services, want.Services)
	}
}

// accountJSON builds a JSON-encoded cloudapi.Account for httptest mocking.
func accountJSON(t *testing.T, cnsEnabled *bool) []byte {
	t.Helper()
	acc := cloudapi.Account{
		ID:               mustParseUUID(t, "00000000-0000-0000-0000-000000000001"),
		Login:            "testaccount",
		Email:            "test@example.com",
		Created:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		TritonCnsEnabled: cnsEnabled,
	}
	b, err := json.Marshal(acc)
	if err != nil {
		t.Fatalf("marshaling account: %s", err)
	}
	return b
}

func TestParseCNSFromMachineTags(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]interface{}
		want InstanceCNS
	}{
		{
			name: "nil tags",
			tags: nil,
			want: InstanceCNS{},
		},
		{
			name: "empty tags",
			tags: map[string]interface{}{},
			want: InstanceCNS{},
		},
		{
			name: "disable true as string",
			tags: map[string]interface{}{"triton.cns.disable": "true"},
			want: InstanceCNS{Disable: true},
		},
		{
			name: "disable true as bool",
			tags: map[string]interface{}{"triton.cns.disable": true},
			want: InstanceCNS{Disable: true},
		},
		{
			name: "disable false as string",
			tags: map[string]interface{}{"triton.cns.disable": "false"},
			want: InstanceCNS{Disable: false},
		},
		{
			name: "single service",
			tags: map[string]interface{}{"triton.cns.services": "web"},
			want: InstanceCNS{Services: []string{"web"}},
		},
		{
			name: "multiple services",
			tags: map[string]interface{}{"triton.cns.services": "web,api,db"},
			want: InstanceCNS{Services: []string{"web", "api", "db"}},
		},
		{
			name: "empty services string",
			tags: map[string]interface{}{"triton.cns.services": ""},
			want: InstanceCNS{},
		},
		{
			name: "disable and services",
			tags: map[string]interface{}{
				"triton.cns.disable":  "true",
				"triton.cns.services": "web,api",
			},
			want: InstanceCNS{Disable: true, Services: []string{"web", "api"}},
		},
		{
			name: "mixed with user tags",
			tags: map[string]interface{}{
				"triton.cns.services": "web",
				"env":                 "prod",
				"team":                "platform",
			},
			want: InstanceCNS{Services: []string{"web"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseCNSFromMachineTags(tc.tags)
			cnsEqual(t, got, tc.want)
		})
	}
}

func TestInjectCNSIntoTags(t *testing.T) {
	tests := []struct {
		name     string
		cns      InstanceCNS
		existing map[string]interface{}
		want     map[string]interface{}
	}{
		{
			name:     "zero value",
			cns:      InstanceCNS{},
			existing: map[string]interface{}{},
			want:     map[string]interface{}{},
		},
		{
			name:     "disable only",
			cns:      InstanceCNS{Disable: true},
			existing: map[string]interface{}{},
			want:     map[string]interface{}{"triton.cns.disable": true},
		},
		{
			name:     "disable false not injected",
			cns:      InstanceCNS{Disable: false},
			existing: map[string]interface{}{},
			want:     map[string]interface{}{},
		},
		{
			name:     "services only",
			cns:      InstanceCNS{Services: []string{"web", "api"}},
			existing: map[string]interface{}{},
			want:     map[string]interface{}{"triton.cns.services": "web,api"},
		},
		{
			name:     "single service",
			cns:      InstanceCNS{Services: []string{"web"}},
			existing: map[string]interface{}{},
			want:     map[string]interface{}{"triton.cns.services": "web"},
		},
		{
			name:     "disable and services",
			cns:      InstanceCNS{Disable: true, Services: []string{"db"}},
			existing: map[string]interface{}{},
			want: map[string]interface{}{
				"triton.cns.disable":  true,
				"triton.cns.services": "db",
			},
		},
		{
			name:     "preserves existing tags",
			cns:      InstanceCNS{Services: []string{"web"}},
			existing: map[string]interface{}{"env": "prod"},
			want: map[string]interface{}{
				"env":                 "prod",
				"triton.cns.services": "web",
			},
		},
		{
			name:     "empty services slice",
			cns:      InstanceCNS{Services: []string{}},
			existing: map[string]interface{}{},
			want:     map[string]interface{}{},
		},
		{
			name:     "nil services slice",
			cns:      InstanceCNS{Services: nil},
			existing: map[string]interface{}{},
			want:     map[string]interface{}{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := make(map[string]interface{})
			for k, v := range tc.existing {
				tags[k] = v
			}
			injectCNSIntoTags(tc.cns, tags)
			if !reflect.DeepEqual(tags, tc.want) {
				t.Errorf("got %v, want %v", tags, tc.want)
			}
			// Verify triton.cns.disable is a bool, not a string.
			if v, ok := tags["triton.cns.disable"]; ok {
				if _, isBool := v.(bool); !isBool {
					t.Errorf("triton.cns.disable is %T, want bool", v)
				}
			}
		})
	}
}

func TestCastToSliceRaw(t *testing.T) {
	tests := []struct {
		name string
		cns  InstanceCNS
		want []interface{}
	}{
		{
			name: "zero value",
			cns:  InstanceCNS{},
			want: []interface{}{
				map[string]interface{}{
					"disable":  false,
					"services": []interface{}{},
				},
			},
		},
		{
			name: "disable true",
			cns:  InstanceCNS{Disable: true},
			want: []interface{}{
				map[string]interface{}{
					"disable":  true,
					"services": []interface{}{},
				},
			},
		},
		{
			name: "with services",
			cns:  InstanceCNS{Services: []string{"web", "api"}},
			want: []interface{}{
				map[string]interface{}{
					"disable":  false,
					"services": []interface{}{"web", "api"},
				},
			},
		},
		{
			name: "disable and services",
			cns:  InstanceCNS{Disable: true, Services: []string{"db"}},
			want: []interface{}{
				map[string]interface{}{
					"disable":  true,
					"services": []interface{}{"db"},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := castToSliceRaw(tc.cns)
			if len(got) != 1 {
				t.Fatalf("expected 1-element slice, got %d", len(got))
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCNSEnabled(t *testing.T) {
	t.Run("CNS enabled", func(t *testing.T) {
		body := accountJSON(t, ptrBool(true))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer ts.Close()

		client := newTestClient(t, ts)
		got, err := client.CNSEnabled()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !got {
			t.Error("expected true, got false")
		}
	})

	t.Run("CNS disabled (false)", func(t *testing.T) {
		body := accountJSON(t, ptrBool(false))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer ts.Close()

		client := newTestClient(t, ts)
		got, err := client.CNSEnabled()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got {
			t.Error("expected false, got true")
		}
	})

	t.Run("CNS field missing (nil)", func(t *testing.T) {
		body := accountJSON(t, nil)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer ts.Close()

		client := newTestClient(t, ts)
		got, err := client.CNSEnabled()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got {
			t.Error("expected false when field is nil, got true")
		}
	})

	t.Run("API error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"InternalError","message":"boom"}`))
		}))
		defer ts.Close()

		client := newTestClient(t, ts)
		got, err := client.CNSEnabled()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "error checking CNS status") {
			t.Errorf("unexpected error message: %s", err)
		}
		if got {
			t.Error("expected false on error, got true")
		}
	})

	t.Run("result is cached", func(t *testing.T) {
		var hits int64
		body := accountJSON(t, ptrBool(true))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&hits, 1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer ts.Close()

		client := newTestClient(t, ts)
		for i := 0; i < 3; i++ {
			got, err := client.CNSEnabled()
			if err != nil {
				t.Fatalf("call %d: unexpected error: %s", i, err)
			}
			if !got {
				t.Errorf("call %d: expected true", i)
			}
		}
		if n := atomic.LoadInt64(&hits); n != 1 {
			t.Errorf("expected 1 API call, got %d", n)
		}
	})

	t.Run("error is cached", func(t *testing.T) {
		var hits int64
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&hits, 1)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"InternalError","message":"boom"}`))
		}))
		defer ts.Close()

		client := newTestClient(t, ts)
		for i := 0; i < 2; i++ {
			_, err := client.CNSEnabled()
			if err == nil {
				t.Fatalf("call %d: expected error", i)
			}
		}
		if n := atomic.LoadInt64(&hits); n != 1 {
			t.Errorf("expected 1 API call, got %d", n)
		}
	})
}

func TestParseCNSRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		cns  InstanceCNS
	}{
		{
			name: "disable and services",
			cns:  InstanceCNS{Disable: true, Services: []string{"web", "api"}},
		},
		{
			name: "services only",
			cns:  InstanceCNS{Services: []string{"db"}},
		},
		{
			name: "disable only",
			cns:  InstanceCNS{Disable: true},
		},
		{
			name: "zero value",
			cns:  InstanceCNS{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := map[string]interface{}{}
			injectCNSIntoTags(tc.cns, tags)
			got := parseCNSFromMachineTags(tags)
			cnsEqual(t, got, tc.cns)
		})
	}
}
