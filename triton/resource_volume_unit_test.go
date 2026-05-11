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
	"time"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TestResourceVolumeSchema_immutableFieldsForceNew verifies that fields
// which CloudAPI's UpdateVolume endpoint does not support are marked
// ForceNew in the Terraform schema.  Without ForceNew, Terraform calls
// Update on changes to these fields, but resourceVolumeUpdate silently
// ignores them — producing a perpetual plan diff.
func TestResourceVolumeSchema_immutableFieldsForceNew(t *testing.T) {
	immutableFields := []string{"size", "networks", "tags", "type"}

	res := resourceVolume()
	for _, field := range immutableFields {
		s, ok := res.Schema[field]
		if !ok {
			t.Errorf("field %q not found in schema", field)
			continue
		}
		if !s.ForceNew {
			t.Errorf("field %q: ForceNew should be true "+
				"(CloudAPI does not support updating this field after creation), "+
				"got false", field)
		}
	}
}

// TestResourceVolumeUpdate_refreshesState verifies that
// resourceVolumeUpdate reads the volume back from the API so
// Terraform state reflects the server-side values after an update.
func TestResourceVolumeUpdate_refreshesState(t *testing.T) {
	volID := mustParseUUID(t, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	fsPath := "/mnt/test-vol"

	volType := cloudapi.VolumeType{}
	if err := volType.FromVolumeType0("tritonnfs"); err != nil {
		t.Fatalf("creating VolumeType: %s", err)
	}

	vol := cloudapi.Volume{
		ID:             volID,
		Name:           "test-vol",
		Size:           10240,
		State:          "ready",
		Type:           volType,
		Created:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		OwnerUUID:      mustParseUUID(t, "11111111-2222-3333-4444-555555555555"),
		FilesystemPath: &fsPath,
		Networks:       &[]openapi_types.UUID{mustParseUUID(t, "cccccccc-dddd-eeee-ffff-000000000000")},
	}

	body, err := json.Marshal(vol)
	if err != nil {
		t.Fatalf("marshaling volume: %s", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)

	d := schema.TestResourceDataRaw(t, resourceVolume().Schema, map[string]interface{}{
		"name": "test-vol",
		"type": "tritonnfs",
	})
	d.SetId(uuidString(volID))

	if err := resourceVolumeUpdate(d, client); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// After update completes, state should be refreshed with
	// server-side values (e.g. filesystem_path, size, owner).
	if got := d.Get("filesystem_path"); got != fsPath {
		t.Errorf("filesystem_path not refreshed after update: got %q, want %q", got, fsPath)
	}
	if got := d.Get("size"); got != 10240 {
		t.Errorf("size not refreshed after update: got %v, want 10240", got)
	}
}
