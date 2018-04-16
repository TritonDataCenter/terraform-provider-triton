package triton

import (
	"strings"
	"testing"
)

func TestValidateVLANIdentifier(t *testing.T) {
	cases := []struct {
		value  int
		errors int
	}{
		{
			value:  -1,
			errors: 1,
		},
		{
			value:  0,
			errors: 0,
		},
		{
			value:  4095,
			errors: 0,
		},
		{
			value:  4096,
			errors: 1,
		},
	}

	for _, tc := range cases {
		_, errs := validateVLANIdentifier(tc.value, "vlan_id")
		if len(errs) != tc.errors {
			t.Errorf("expected %d validation errors for value %d, got %d", tc.errors, tc.value, len(errs))
		}
	}

	_, errs := validateVLANIdentifier(12345, "vlan_id")

	e := errs[0]
	if !strings.Contains(e.Error(), `"vlan_id" value must be between 0 and 4095`) {
		t.Errorf("expected error to equal test error, got %s", e)
	}
}
