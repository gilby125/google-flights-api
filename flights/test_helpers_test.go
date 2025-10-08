package flights

import (
	"os"
	"testing"
)

var runIntegrationTests = os.Getenv("ENABLE_INTEGRATION_TESTS") == "1"

func skipUnlessIntegration(t *testing.T) {
	if !runIntegrationTests {
		t.Skip("set ENABLE_INTEGRATION_TESTS=1 to run live Google Flights integration tests")
	}
}
