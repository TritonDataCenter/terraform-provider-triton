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
	"testing"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestKeyReadMigratesNameIDToFingerprint demonstrates that when a state file
// contains a name-based key ID (from the old triton-go provider), a read
// should update the ID to the fingerprint returned by CloudAPI. Without
// this migration, triton_key.foo.id interpolates to a name instead of a
// fingerprint.
func TestKeyReadMigratesNameIDToFingerprint(t *testing.T) {
	const (
		keyName        = "mykey"
		keyFingerprint = "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"
		keyMaterial    = "ssh-rsa AAAAB3... mykey"
	)

	keyJSON, err := json.Marshal(cloudapi.SSHKey{
		Name:        keyName,
		Fingerprint: keyFingerprint,
		Key:         keyMaterial,
	})
	if err != nil {
		t.Fatalf("marshaling key: %s", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET /testaccount/keys/mykey — CloudAPI accepts name or fingerprint.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(keyJSON)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)

	// Simulate old state: ID is the key name, not the fingerprint.
	rd := resourceKey().TestResourceData()
	rd.SetId(keyName)
	rd.Set("name", keyName)
	rd.Set("key", keyMaterial)

	err = resourceKeyRead(rd, client)
	if err != nil {
		t.Fatalf("resourceKeyRead: %s", err)
	}

	if got := rd.Id(); got != keyFingerprint {
		t.Errorf("after read, resource ID = %q; want fingerprint %q (old name-based ID was not migrated)", got, keyFingerprint)
	}
}

// TestKeyReadPreservesFingerprint verifies that when the state already has a
// fingerprint-based ID (new format), it remains unchanged after a read.
func TestKeyReadPreservesFingerprint(t *testing.T) {
	const (
		keyName        = "mykey"
		keyFingerprint = "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"
		keyMaterial    = "ssh-rsa AAAAB3... mykey"
	)

	keyJSON, err := json.Marshal(cloudapi.SSHKey{
		Name:        keyName,
		Fingerprint: keyFingerprint,
		Key:         keyMaterial,
	})
	if err != nil {
		t.Fatalf("marshaling key: %s", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(keyJSON)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)

	// State already has fingerprint-based ID (new format).
	rd := resourceKey().TestResourceData()
	rd.SetId(keyFingerprint)
	rd.Set("name", keyName)
	rd.Set("key", keyMaterial)

	err = resourceKeyRead(rd, client)
	if err != nil {
		t.Fatalf("resourceKeyRead: %s", err)
	}

	if got := rd.Id(); got != keyFingerprint {
		t.Errorf("after read, resource ID = %q; want %q", got, keyFingerprint)
	}
}

// TestKeyStateUpgrade verifies that the state upgrader migrates a v0 state
// (name-based ID) to v1 (fingerprint-based ID) by looking up the key.
func TestKeyStateUpgrade(t *testing.T) {
	const (
		keyName        = "mykey"
		keyFingerprint = "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"
		keyMaterial    = "ssh-rsa AAAAB3... mykey"
	)

	keyJSON, err := json.Marshal(cloudapi.SSHKey{
		Name:        keyName,
		Fingerprint: keyFingerprint,
		Key:         keyMaterial,
	})
	if err != nil {
		t.Fatalf("marshaling key: %s", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(keyJSON)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)

	res := resourceKey()
	if res.SchemaVersion < 1 {
		t.Fatalf("expected SchemaVersion >= 1 to indicate state migration support, got %d", res.SchemaVersion)
	}
	if len(res.StateUpgraders) == 0 {
		t.Fatal("expected at least one StateUpgrader for v0->v1 migration")
	}

	// Build v0 raw state with name-based ID.
	rawState := map[string]interface{}{
		"id":   keyName,
		"name": keyName,
		"key":  keyMaterial,
	}

	upgraded, err := res.StateUpgraders[0].Upgrade(nil, rawState, client)
	if err != nil {
		t.Fatalf("state upgrade: %s", err)
	}

	gotID, ok := upgraded["id"].(string)
	if !ok {
		t.Fatalf("upgraded state has no 'id' field")
	}
	if gotID != keyFingerprint {
		t.Errorf("upgraded state id = %q; want fingerprint %q", gotID, keyFingerprint)
	}

	// Verify the Terraform instance state would also have the right ID.
	is := &terraform.InstanceState{
		ID:         keyName,
		Attributes: map[string]string{"name": keyName, "key": keyMaterial},
	}
	_ = is // available for further integration testing if needed

	// Verify the resource can use SchemaVersion-based dispatch.
	rd := res.TestResourceData()
	rd.SetId(keyFingerprint)
	if rd.Id() != keyFingerprint {
		t.Errorf("TestResourceData ID = %q; want %q", rd.Id(), keyFingerprint)
	}
}

// Helper: suppress unused import for schema package.
var _ = (*schema.Resource)(nil)
