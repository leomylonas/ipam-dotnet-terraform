package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTenancyResource_basic(t *testing.T) {
	testAccPreCheck(t)

	name := "tf-acc-tenancy-basic"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "ipam_tenancy" "test" {
  name           = %q
  description    = "Terraform acceptance test tenancy"
  admin_username = "tf-acc-admin-basic"
  admin_password = "Accpass1234"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ipam_tenancy.test", "name", name),
					resource.TestCheckResourceAttr("ipam_tenancy.test", "description", "Terraform acceptance test tenancy"),
				),
			},
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "ipam_tenancy" "test" {
  name           = %q
  description    = "Updated description"
  admin_username = "tf-acc-admin-basic"
  admin_password = "Accpass1234"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ipam_tenancy.test", "description", "Updated description"),
				),
			},
			{
				ResourceName:      "ipam_tenancy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"admin_password",
					"admin_username",
				},
			},
		},
	})
}
