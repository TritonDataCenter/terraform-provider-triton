package triton

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/terraform"
)

func resourceFabricMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found Fabric State v0; migrating to v1")
		return migrateFabricStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateFabricStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before Migration: %#v", is.Attributes)

	if is.Attributes["internet_nat"] != "true" {
		is.Attributes["internet_nat"] = "false"
	}

	log.Printf("[DEBUG] Attributes after State Migration: %#v", is.Attributes)

	return is, nil
}
