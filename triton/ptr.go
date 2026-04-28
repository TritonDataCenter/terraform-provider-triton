/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2025 Edgecast Cloud LLC.
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool       { return &b }

func derefString(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}

func derefBool(p *bool) bool {
	if p != nil {
		return *p
	}
	return false
}

func derefUint64(p *uint64) uint64 {
	if p != nil {
		return *p
	}
	return 0
}

func derefStringSlice(p *[]string) []string {
	if p != nil {
		return *p
	}
	return nil
}

func parseUUID(s string) (openapi_types.UUID, error) {
	return uuid.Parse(s)
}

func isUUID(s string) bool {
	_, err := parseUUID(s)
	return err == nil
}

func uuidString(u openapi_types.UUID) string {
	return u.String()
}

func uuidSliceToStrings(uuids []openapi_types.UUID) []string {
	out := make([]string, len(uuids))
	for i, u := range uuids {
		out[i] = uuidString(u)
	}
	return out
}

func uuidPtrSliceToStrings(uuids *[]openapi_types.UUID) []string {
	if uuids == nil {
		return nil
	}
	return uuidSliceToStrings(*uuids)
}
