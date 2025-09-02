package triton

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTritonDataVolume_basic(t *testing.T) {
	volumeName := fmt.Sprintf("acctest-volume-%d", acctest.RandInt())
	config := fmt.Sprintf(`
		resource "triton_volume" "test_volume" {
			name = "%s"
			tags = {
				Name = "Database Volume"
			}
		}

		data "triton_volume" "my_volume" {
			name = "${triton_volume.test_volume.name}"
			size = "${triton_volume.test_volume.size}"
		}
	`, volumeName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.triton_volume.my_volume", "id"),
					resource.TestCheckResourceAttrSet("data.triton_volume.my_volume", "size"),
					resource.TestCheckResourceAttr("data.triton_volume.my_volume", "name", volumeName),
					resource.TestCheckResourceAttr("data.triton_volume.my_volume", "state", volumeStateReady),
					resource.TestCheckResourceAttr("data.triton_volume.my_volume", "tags.%", "1"),
					resource.TestCheckResourceAttr("data.triton_volume.my_volume", "tags.Name", "Database Volume"),
					resource.TestCheckResourceAttr("data.triton_volume.my_volume", "type", "tritonnfs"),
				),
			},
		},
	})
}

func TestAccTritonDataVolume_noResults(t *testing.T) {
	config := `
		data "triton_volume" "myvol" {
			name = "missing-volume"
	}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`your query returned no results`),
			},
		},
	})
}

func TestAccTritonDataVolume_BadSize(t *testing.T) {
	config := `
		data "triton_volume" "myvol" {
			size = "one million"
	}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`Inappropriate value for attribute "size": a number is required.`),
			},
		},
	})
}
