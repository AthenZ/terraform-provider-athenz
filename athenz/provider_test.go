package athenz

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

func funcProvider() (*schema.Provider, error) {
	return testAccProvider, nil
}

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]func() (*schema.Provider, error){
		"athenz": funcProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ATHENZ_ZMS_URL"); v == "" {
		t.Fatal("ATHENZ_ZMS_URL must be set for acceptance tests")
	}
}
