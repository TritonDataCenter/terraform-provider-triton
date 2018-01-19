package triton

import (
	"testing"

	"regexp"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccTritonImage_multipleResults(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTritonImage_multipleResults,
				ExpectError: regexp.MustCompile(`Your query returned more than one result`),
			},
		},
	})
}

func TestAccTritonImage_noResults(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTritonImage_noResults,
				ExpectError: regexp.MustCompile(`Your query returned no results`),
			},
		},
	})
}

func TestAccTritonImage_mostRecent(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonImage_mostRecent,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.triton_image.base", "id"),
				),
			},
		},
	})
}

func TestAccTritonImage_nameVersionAndMostRecent(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonImage_nameVersionAndMostRecent,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.triton_image.base", "id"),
				),
			},
		},
	})
}

var testAccTritonImage_noResults = `
data "triton_image" "base" {
	name = "missing-image"
}
`

var testAccTritonImage_multipleResults = `
data "triton_image" "base" {
	name = "base-64-lts"
}
`

var testAccTritonImage_mostRecent = `
data "triton_image" "base" {
	name = "base-64-lts"
	most_recent = true
}
`

var testAccTritonImage_nameVersionAndMostRecent = `
data "triton_image" "base" {
	name = "base-64-lts"
	version = "16.4.1"
	most_recent = true
}
`
