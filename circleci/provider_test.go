package circleci

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testOrg string = os.Getenv("CIRCLECI_TEST_ORGANIZATION")
var testrepo string = os.Getenv("CIRCLECI_TEST_REPO")

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"circleci": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("CIRCLECI_API_TOKEN"); v == "" {
		t.Fatal("CIRCLECI_API_TOKEN must be set for acceptance tests")
	}

	if v := os.Getenv("CIRCLECI_TEST_ORGANIZATION"); v == "" {
		t.Fatal("CIRCLECI_TEST_ORGANIZATION must be set for acceptance tests")
	}

	if v := os.Getenv("CIRCLECI_TEST_REPO"); v == "" {
		t.Fatal("CIRCLECI_TEST_REPO must be set for acceptance tests")
	}
}
