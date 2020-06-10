/*
Package paymail encapsulates all the paymail methods and defaults
*/
package paymail

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mrz1836/go-sanitize"
	"github.com/mrz1836/go-validate"
)

// Defaults for paymail functions
const (
	defaultDnsPort           = "53"         // Default port for DNS / NameServer checks
	defaultDnsTimeout        = 5            // In seconds
	defaultGetTimeout        = 15           // In seconds
	defaultNameServerNetwork = "udp"        // Default for NS dialer
	defaultPostTimeout       = 15           // In seconds
	defaultSSLDeadline       = 10           // In seconds
	defaultSSLTimeout        = 10           // In seconds
	defaultUserAgent         = "go:paymail" // Default user agent
	maxSRVRecords            = 1            // Given by paymail specs
)

// UserAgent is for customizing the user agent
var UserAgent = defaultUserAgent

// Public defaults for paymail specs
/*
	http://bsvalias.org/02-01-host-discovery.html

	Service	  bsvalias
	Proto	  tcp
	Name	  <domain>.<tld>.
	TTL	      3600 (see notes)
	Class	  IN
	Priority  10
	Weight	  10
	Port	  443
	Target	  <endpoint-discovery-host>

	Max SRV Records:  1
*/
const (
	DefaultBsvAliasVersion = "1.0"      // Default version number for bsvalias
	DefaultPort            = 443        // Default port (from specs)
	DefaultPriority        = 10         // Default priority (from specs)
	DefaultProtocol        = "tcp"      // Default protocol (from specs)
	DefaultServiceName     = "bsvalias" // Default service name (from specs)
	DefaultWeight          = 10         // Default weight (from specs)
	PubKeyLength           = 66         // Required length for a valid PubKey (pki)
)

// StandardResponse is the standard fields returned on all responses
type StandardResponse struct {
	StatusCode int             `json:"status_code"` // Status code returned on the request
	Tracing    resty.TraceInfo `json:"tracing"`     // Trace information if enabled on the request
}

// JsonError is the standard error response from a paymail server
type JsonError struct {
	Message string `json:"message"` // Shows the error message returned by the server
}

// ExtractParts will check if it's a domain or address and extract the parts
func ExtractParts(paymailInput string) (domain, address string) {

	// Determine if it's a paymail address vs domain (1 Arg is required)
	if strings.Contains(paymailInput, "@") {

		// Sanitize the paymail address
		address = sanitize.Email(paymailInput, false)

		// Split the parts
		parts := strings.Split(address, "@")

		// Sanitize the domain name
		domain, _ = sanitize.Domain(parts[1], false, true)

	} else {
		// Sanitize the domain name
		domain, _ = sanitize.Domain(paymailInput, false, true)
	}
	return
}

// customResolver will return a custom resolver using a given nameServer and network
func customResolver(nameServer, useNetwork string) net.Resolver {
	return net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * defaultDnsTimeout,
			}
			return d.DialContext(ctx, useNetwork, nameServer+":"+defaultDnsPort)
		},
	}
}

// ValidatePaymail will do a basic validation on the paymail format
func ValidatePaymail(paymailAddress string) error {

	// Validate the format for the paymail address (paymail addresses follow conventional email requirements)
	if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
		return fmt.Errorf("paymail address failed format validation: %s", err.Error())

	} else if !ok {
		return fmt.Errorf("paymail address failed format validation: unknown reason")
	}

	return nil
}

// ValidateDomain will do a basic validation on the domain format
func ValidateDomain(domain string) error {

	// Check for a real domain (require at least one period)
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("domain name is invalid: %s", domain)

	} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
		return fmt.Errorf("domain name failed DNS check: %s", domain)
	}

	return nil
}

// ValidatePaymailAndDomain will do a basic validation on the paymail and domain format
func ValidatePaymailAndDomain(paymailAddress, domain string) error {
	if err := ValidatePaymail(paymailAddress); err != nil {
		return err
	}
	if err := ValidateDomain(domain); err != nil {
		return err
	}

	return nil
}

// ParseIfHandcashHandle will parse a paymail input to check if
// it is a HandCash handle (eg. $user). If it is, then it will
// return the paymail address equivalent (eg. user@handcash.io),
// else it will just return the input unchanged.
func ParseIfHandcashHandle(paymailInput string) string {
	var validID = regexp.MustCompile(`^\$[a-zA-Z0-9\-_.]{4,}$`)

	if validID.MatchString(paymailInput) {
		return paymailInput[1:] + "@handcash.io"
	} else {
		return paymailInput
	}
}
