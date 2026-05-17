package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	envAccEnabled = "IPAM_ACC"
	envBaseURL    = "IPAM_BASE_URL"
	envUsername   = "IPAM_USERNAME"
	envPassword   = "IPAM_PASSWORD"
)

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv(envAccEnabled) != "1" {
		t.Skip("Acceptance tests are disabled. Set IPAM_ACC=1 to enable.")
	}
	for _, key := range []string{envBaseURL, envUsername, envPassword} {
		if os.Getenv(key) == "" {
			t.Fatalf("%s must be set for acceptance tests", key)
		}
	}
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"ipam": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccProviderConfig() string {
	return `
provider "ipam" {
  base_url = "` + os.Getenv(envBaseURL) + `"
  username = "` + os.Getenv(envUsername) + `"
  password = "` + os.Getenv(envPassword) + `"
}
`
}

func checkDestroyNoop(_ *terraform.State) error {
	return nil
}
