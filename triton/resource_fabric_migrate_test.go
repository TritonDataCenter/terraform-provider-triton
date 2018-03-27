package triton

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestFabricMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		ID           string
		Attributes   map[string]string
		Expected     string
		Meta         interface{}
	}{
		"v0_1_not_set_should_be_false": {
			StateVersion: 0,
			ID:           "tf-testing-file",
			Attributes:   map[string]string{},
			Expected:     "false",
		},
		"v0_1_set_to_true_do_nothing": {
			StateVersion: 0,
			ID:           "tf-testing-file",
			Attributes: map[string]string{
				"internet_nat": "true",
			},
			Expected: "true",
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := resourceFabricMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		if is.Attributes["internet_nat"] != tc.Expected {
			t.Fatalf("Bad internet_nat migration: %s\n\n expected: %s", is.Attributes["internet_nat"], tc.Expected)
		}
	}
}
