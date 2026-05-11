/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Joyent, Inc.
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTritonPackage_basic(t *testing.T) {
	testPackageResultName := testAccConfig(t, "package_query_result")
	testPackageQueryName := testAccConfig(t, "package_query_name")
	testPackageQueryMemory := testAccConfig(t, "package_query_memory")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTritonPackage_basic(testPackageQueryName, testPackageQueryMemory),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonPackageDataSourceID("data.triton_package.base", testPackageResultName),
				),
			},
		},
	})
}

func testAccCheckTritonPackageDataSourceID(name, packageName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find package data source: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("returned package ID should not be empty")
		}
		if rs.Primary.Attributes["name"] != packageName {
			return fmt.Errorf("returned package Name does not match")
		}

		return nil
	}
}

func TestAccTritonPackage_filterByName(t *testing.T) {
	testPackageName := testAccConfig(t, "package_query_result")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "triton_package" "by_name" {
						filter {
							name = "%s"
						}
					}
				`, testPackageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonPackageDataSourceID("data.triton_package.by_name", testPackageName),
					resource.TestCheckResourceAttrSet("data.triton_package.by_name", "memory"),
					resource.TestCheckResourceAttrSet("data.triton_package.by_name", "disk"),
				),
			},
		},
	})
}

func TestAccTritonPackage_filterByBrand(t *testing.T) {
	pkg := testAccFindBrandedPackage(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "triton_package" "by_brand" {
						filter {
							name   = "%s"
							memory = %d
							brand  = "%s"
						}
					}
				`, pkg.Name, pkg.Memory, pkg.Brand),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonPackageDataSourceID("data.triton_package.by_brand", pkg.Name),
					resource.TestCheckResourceAttr("data.triton_package.by_brand", "brand", pkg.Brand),
				),
			},
		},
	})
}

func TestAccTritonPackage_filterWithoutName(t *testing.T) {
	// Find a package that can be uniquely identified by memory alone
	// (i.e. no other package shares its memory value).
	pkgs := testAccDiscoverPackages(t)

	memoryCounts := map[uint64]int{}
	for _, p := range pkgs {
		memoryCounts[p.Memory]++
	}

	var targetName string
	var targetMemory uint64
	for _, p := range pkgs {
		if memoryCounts[p.Memory] == 1 {
			targetName = p.Name
			targetMemory = p.Memory
			break
		}
	}
	if targetName == "" {
		t.Skip("skipping: no package with unique memory value found in this DC")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "triton_package" "no_name" {
						filter {
							memory = %d
						}
					}
				`, targetMemory),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTritonPackageDataSourceID("data.triton_package.no_name", targetName),
					resource.TestCheckResourceAttrSet("data.triton_package.no_name", "disk"),
				),
			},
		},
	})
}

var testAccTritonPackage_basic = func(query string, memory string) string {
	return fmt.Sprintf(`
		data "triton_package" "base" {
			filter {
	   		name = "%s"
	   		memory = %s
			}
		}
		`, query, memory)
}
