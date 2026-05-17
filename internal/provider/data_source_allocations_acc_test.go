package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceAllocations_basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "ipam_allocations" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ipam_allocations.all", "items.#"),
				),
			},
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "ipam_allocations" "filtered" {
  tag_key   = %q
  tag_value = %q
}
`, "env", "dev"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ipam_allocations.filtered", "items.#"),
				),
			},
		},
	})
}
