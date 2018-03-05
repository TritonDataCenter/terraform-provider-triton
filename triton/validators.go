package triton

import "fmt"

// validateVLANIdentifier validates that the integer value is a valid VLAN ID,
// which for the Fabric VLAN must be in the range between 0 and 4095 inclusive.
func validateVLANIdentifier(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 0 || value > 4095 {
		errors = append(errors, fmt.Errorf("%q value must be between 0 and 4095", k))
	}
	return
}
