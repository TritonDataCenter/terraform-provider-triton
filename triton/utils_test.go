package triton

import "testing"

func TestWildcardMatch(t *testing.T) {
	cases := []struct {
		value    string
		pattern  string
		expected bool
	}{
		{
			"",
			"",
			true,
		},
		{
			"",
			"*",
			true,
		},
		{
			"",
			"?",
			false,
		},
		{
			"",
			"triton",
			false,
		},
		{
			"triton",
			"",
			false,
		},
		{
			"triton",
			"*",
			true,
		},
		{
			"triton",
			"?",
			false,
		},
		{
			"triton",
			"*n",
			true,
		},
		{
			"triton",
			"?*",
			true,
		},
		{
			"triton",
			"t*",
			true,
		},
		{
			"triton",
			"t*n",
			true,
		},
		{
			"triton",
			"?*n",
			true,
		},
		{
			"triton",
			"triton",
			true,
		},
		{
			"triton",
			"??????",
			true,
		},
		{
			"triton",
			"trito?",
			true,
		},
		{
			"triton",
			"?riton",
			true,
		},
		{
			"triton",
			"t?it?n",
			true,
		},
		{
			"triton",
			"*triton",
			true,
		},
		{
			"triton",
			"triton*",
			true,
		},
		{
			"triton",
			"?triton",
			false,
		},
		{
			"triton",
			"triton?",
			false,
		},
		{
			"txrxixtxoxnx",
			"t*r*i*t*o*nx",
			true,
		},
		{
			"txrxixtxoxnn",
			"t*r*i*t*o*n*",
			true,
		},
		{
			"trxrrxritonx",
			"t*r?t*o*n*x*",
			true,
		},
		{
			"trxrrxritonn",
			"t*r?t*o*n*x",
			false,
		},
	}

	for _, tc := range cases {
		actual := wildcardMatch(tc.pattern, tc.value)
		if actual != tc.expected {
			t.Errorf("expected %q to match %q, got %t", tc.pattern, tc.value, actual)
		}
	}
}
