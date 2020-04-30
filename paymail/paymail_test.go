package paymail

import (
	"strings"
	"testing"
)

// TestExtractParts will test the ExtractParts method
func TestExtractParts(t *testing.T) {

	testAlias := "user"
	testDomain := "domain.com"
	testString := testAlias + "@" + testDomain

	domain, address := ExtractParts(testString)
	if domain != testDomain {
		t.Fatalf("expected domain name: %s but got: %s", testDomain, domain)
	}
	if address != testString {
		t.Fatalf("expected address: %s but got: %s", testString, address)
	}

	// Test removing spaces and normalizing (lowercase)
	testAlias = " User"
	testDomain = "Domain.com "
	testString = testAlias + "@" + testDomain

	domain, address = ExtractParts(testString)
	if domain != strings.TrimSpace(strings.ToLower(testDomain)) {
		t.Fatalf("expected domain name: %s but got: %s", strings.TrimSpace(strings.ToLower(testDomain)), domain)
	}
	if address != "user@domain.com" {
		t.Fatalf("expected alias: %s but got: %s", "user@domain.com", address)
	}
}
