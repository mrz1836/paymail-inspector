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

// TestValidatePaymail will test the ValidatePaymail method
func TestValidatePaymail(t *testing.T) {

	testPaymail := "User@domain.com"

	err := ValidatePaymail(testPaymail)

	if err != nil {
		t.Fatalf("expected 'nil' but got: %s", err.Error())
	}
}

// TestValidateDomain will test the ValidateDomain method
func TestValidateDomain(t *testing.T) {

	testDomain := "domain.com"

	err := ValidateDomain(testDomain)

	if err != nil {
		t.Fatalf("expected 'nil' but got: %s", err.Error())
	}
}

// TestValidatePaymailAndDomain will test the ValidatePaymailAndDomain method
func TestValidatePaymailAndDomain(t *testing.T) {

	testAlias := "user"
	testDomain := "domain.com"
	testAddress := testAlias + "@" + testDomain

	err := ValidatePaymailAndDomain(testAddress, testDomain)

	if err != nil {
		t.Fatalf("expected 'nil' but got: %s", err.Error())
	}
}

// TestParseIfHandcashHandle will test the ParseIfHandcashHandle method
func TestParseIfHandcashHandle(t *testing.T) {

	expected := "user@handcash.io"

	t.Run("handle passed", func(t *testing.T) {
		testInput := "$user"

		res := ParseIfHandcashHandle(testInput)

		if res != expected {
			t.Fatalf("expected: %s, got: %s", expected, res)
		}
	})

	t.Run("paymail address passed", func(t *testing.T) {
		testInput := "user@handcash.io"

		res := ParseIfHandcashHandle(testInput)

		if res != expected {
			t.Fatalf("expected: %s, got: %s", expected, res)
		}
	})
}
