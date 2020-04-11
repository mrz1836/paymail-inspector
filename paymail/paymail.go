/*
Package paymail encapsulates all the paymail methods and defaults
*/
package paymail

import (
	"context"
	"net"
	"strings"
	"time"
)

// Defaults for paymail functions
const (
	defaultDeadline          = 5     // In seconds
	defaultDnsPort           = "53"  // Default port for DNS / NameServer checks
	defaultNameServerNetwork = "udp" // Default for NS dialer
	defaultTimeout           = 5     // In seconds
	defaultUserAgent         = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36"
	maxSRVRecords            = 1        // Given by paymail specs
	typeBool                 = "bool"   // For bool detection
	typeString               = "string" // For string detection
)

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

// Public defaults for paymail specs
const (
	DefaultBsvAliasVersion = "1.0"      // Default version number for bsvalias
	DefaultPort            = 443        // Default port (from specs)
	DefaultPriority        = 10         // Default priority (from specs)
	DefaultProtocol        = "tcp"      // Default protocol (from specs)
	DefaultServiceName     = "bsvalias" // Default service name (from specs)
	DefaultWeight          = 10         // Default weight (from specs)
	PubKeyLength           = 66         // Required length for a valid PubKey (pki)
)

// ExtractParts will check if it's a domain or address and extract the parts
func ExtractParts(paymailInput string) (domain, address string) {

	// Determine if it's a paymail address vs domain (1 Arg is required)
	if strings.Contains(paymailInput, "@") {

		// Remove any spaces
		address = strings.TrimSpace(paymailInput)

		// Split the parts
		parts := strings.Split(address, "@")

		// Force all domain names to lowercase
		domain = strings.ToLower(parts[1])

		// Combine the address back
		address = parts[0] + "@" + domain

	} else {
		// Force all domain names to lowercase and trim spaces
		domain = strings.TrimSpace(strings.ToLower(paymailInput))
	}
	return
}

// customResolver will return a custom resolver using a given nameServer and network
func customResolver(nameServer, useNetwork string) net.Resolver {
	return net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * defaultTimeout,
			}
			return d.DialContext(ctx, useNetwork, nameServer+":"+defaultDnsPort)
		},
	}
}
