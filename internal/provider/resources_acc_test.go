package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func uniq(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestAccUserResource_basic(t *testing.T) {
	testAccPreCheck(t)
	tName := uniq("tf-acc-tenancy-user")
	uName := uniq("tf-acc-user")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "ipam_tenancy" "t" {
  name           = %q
  description    = "tenancy for user acc"
  admin_username = %q
  admin_password = "AccPass123"
}

resource "ipam_user" "u" {
  username  = %q
  password  = "AccPass123"
  role      = "TenantUser"
  tenancy_id = ipam_tenancy.t.id
}
`, tName, uniq("tf-acc-tadmin"), uName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ipam_user.u", "username", uName),
				),
			},
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "ipam_tenancy" "t" {
  name           = %q
  description    = "tenancy for user acc"
  admin_username = %q
  admin_password = "AccPass123"
}

resource "ipam_user" "u" {
  username  = %q
  password  = "AccPass123"
  role      = "TenantUser"
  tenancy_id = ipam_tenancy.t.id
}
`, tName, uniq("tf-acc-tadmin"), uName+"-upd"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ipam_user.u", "username", uName+"-upd"),
				),
			},
			{ResourceName: "ipam_user.u", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"password"}},
		},
	})
}

func TestAccPrivateSubnetResource_basic(t *testing.T) {
	testAccPreCheck(t)
	tName := uniq("tf-acc-tenancy-ps")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyNoop,
		Steps: []resource.TestStep{
			{Config: testAccProviderConfig() + fmt.Sprintf(`
resource "ipam_tenancy" "t" {
  name           = %q
  description    = "tenancy for subnet"
  admin_username = %q
  admin_password = "AccPass123"
}

resource "ipam_private_subnet" "s" {
  tenancy_id = ipam_tenancy.t.id
  cidr = "10.90.0.0/24"
  name = "subnet-a"
  description = "private subnet"
}
`, tName, uniq("tf-acc-tadmin")),
				Check: resource.TestCheckResourceAttr("ipam_private_subnet.s", "name", "subnet-a")},
			{ResourceName: "ipam_private_subnet.s", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccSharedSubnetResource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyNoop,
		Steps: []resource.TestStep{
			{Config: testAccProviderConfig() + `
resource "ipam_shared_subnet" "s" {
  cidr = "172.30.0.0/24"
  name = "shared-a"
  description = "shared subnet"
}
`, Check: resource.TestCheckResourceAttr("ipam_shared_subnet.s", "name", "shared-a")},
			{ResourceName: "ipam_shared_subnet.s", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccSharedSubnetAccessResource_basic(t *testing.T) {
	testAccPreCheck(t)
	tName := uniq("tf-acc-tenancy-access")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyNoop,
		Steps: []resource.TestStep{
			{Config: testAccProviderConfig() + fmt.Sprintf(`
resource "ipam_tenancy" "t" {
  name           = %q
  description    = "tenancy for access"
  admin_username = %q
  admin_password = "AccPass123"
}

resource "ipam_shared_subnet" "s" {
  cidr = "172.31.0.0/24"
  name = "shared-access"
  description = "shared subnet"
}

resource "ipam_shared_subnet_access" "a" {
  subnet_id = ipam_shared_subnet.s.id
  tenancy_id = ipam_tenancy.t.id
}
`, tName, uniq("tf-acc-tadmin")),
				Check: resource.TestCheckResourceAttrSet("ipam_shared_subnet_access.a", "id")},
			{ResourceName: "ipam_shared_subnet_access.a", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccExclusionResource_basic(t *testing.T) {
	testAccPreCheck(t)
	tName := uniq("tf-acc-tenancy-excl")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyNoop,
		Steps: []resource.TestStep{
			{Config: testAccProviderConfig() + fmt.Sprintf(`
resource "ipam_tenancy" "t" {
  name           = %q
  description    = "tenancy for exclusion"
  admin_username = %q
  admin_password = "AccPass123"
}

resource "ipam_private_subnet" "s" {
  tenancy_id = ipam_tenancy.t.id
  cidr = "10.91.0.0/24"
  name = "subnet-ex"
  description = "private subnet"
}

resource "ipam_exclusion" "e" {
  subnet_id = ipam_private_subnet.s.id
  start = "10.91.0.1"
  end = "10.91.0.1"
  description = "gateway"
}
`, tName, uniq("tf-acc-tadmin")),
				Check: resource.TestCheckResourceAttr("ipam_exclusion.e", "description", "gateway")},
			{ResourceName: "ipam_exclusion.e", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccAllocationAndTagsResources_basic(t *testing.T) {
	testAccPreCheck(t)
	tName := uniq("tf-acc-tenancy-alloc")
	tAdmin := uniq("tf-acc-tenant-admin")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
provider "ipam" {
  alias = "tenant"
  base_url = "%s"
  username = "%s"
  password = "AccPass123"
}

resource "ipam_tenancy" "t" {
  name           = %q
  description    = "tenancy for allocation"
  admin_username = %q
  admin_password = "AccPass123"
}

resource "ipam_private_subnet" "s" {
  tenancy_id = ipam_tenancy.t.id
  cidr = "10.92.0.0/24"
  name = "subnet-alloc"
  description = "private subnet"
}

resource "ipam_allocation" "a" {
  provider = ipam.tenant
  subnet_id = ipam_private_subnet.s.id
  description = "host-1"
}

resource "ipam_allocation_tags" "tags" {
  provider = ipam.tenant
  allocation_id = ipam_allocation.a.id
  tags = {
    env = "dev"
    owner = "platform"
  }
}
`, os.Getenv(envBaseURL), tAdmin, tName, tAdmin),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ipam_allocation.a", "ip_address"),
					resource.TestCheckResourceAttr("ipam_allocation_tags.tags", "tags.env", "dev"),
				),
			},
			{ResourceName: "ipam_allocation.a", ImportState: true, ImportStateVerify: true},
			{ResourceName: "ipam_allocation_tags.tags", ImportState: true, ImportStateVerify: true},
		},
	})
}
