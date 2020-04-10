package paymail

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// GetSRVRecord will get the SRV record for a given domain name
// Specs: http://bsvalias.org/02-01-host-discovery.html
func GetSRVRecord(service, protocol, domainName, nameServer string) (srv *net.SRV, err error) {

	// Invalid parameters?
	if len(service) == 0 {
		err = fmt.Errorf("invalid parameter: service")
		return
	} else if len(protocol) == 0 {
		err = fmt.Errorf("invalid parameter: protocol")
		return
	} else if len(domainName) == 0 || len(domainName) > 255 {
		err = fmt.Errorf("invalid parameter: name")
		return
	}

	// Force the case
	protocol = strings.TrimSpace(strings.ToLower(protocol))

	// Setup the custom resolver
	r := customResolver(nameServer, defaultNameServerNetwork)

	// The final cname to check for
	cnameCheck := fmt.Sprintf("_%s._%s.%s.", service, protocol, domainName)

	// Lookup the SRV record
	var cname string
	var records []*net.SRV
	if cname, records, err = r.LookupSRV(context.Background(), service, protocol, domainName); err != nil {
		return
	}

	// No SRV record found
	if len(records) == 0 {
		err = fmt.Errorf("zero SRV records found using: %s", cnameCheck)
		return
	}

	// More than X records (spec calls for 1 record only)
	if len(records) > maxSRVRecords {
		err = fmt.Errorf("only %d SRV record(s) should exist, found %d records", maxSRVRecords, len(records))
		return
	}

	//  Basic CNAME check (sanity check!)
	if cname != cnameCheck {
		err = fmt.Errorf("cname was invalid or not found using: %s looking for: %s", cnameCheck, cname)
		return
	}

	// Only return the first record (in case multiple are returned)
	srv = records[0]

	return
}

// ValidateSRVRecord will check for a valid SRV record for paymail
// Specs: http://bsvalias.org/02-01-host-discovery.html
func ValidateSRVRecord(srv *net.SRV, nameServer string, port, priority, weight int) (err error) {

	// Check the params first
	if srv == nil {
		err = fmt.Errorf("invalid parameter: srv is missing or nil")
		return
	} else if port <= 0 {
		err = fmt.Errorf("invalid parameter: port")
		return
	} else if priority <= 0 {
		err = fmt.Errorf("invalid parameter: priority")
		return
	} else if weight <= 0 {
		err = fmt.Errorf("invalid parameter: weight")
		return
	}

	// Check the basics of the SRV record
	if len(srv.Target) == 0 {
		err = fmt.Errorf("target is invalid or empty")
		return
	} else if srv.Port != uint16(port) {
		err = fmt.Errorf("port %d does not match %d", srv.Port, port)
		return
	} else if srv.Priority != uint16(priority) {
		err = fmt.Errorf("priority %d does not match %d", srv.Priority, priority)
		return
	} else if srv.Weight != uint16(weight) {
		err = fmt.Errorf("weight %d does not match %d", srv.Weight, weight)
		return
	}

	// Setup the custom resolver
	r := customResolver(nameServer, defaultNameServerNetwork)

	// Do we have a hostname that resolves?
	var addresses []string
	if addresses, err = r.LookupHost(context.Background(), srv.Target); err != nil {
		return
	} else if len(addresses) == 0 {
		err = fmt.Errorf("target %s could not resolve a host", srv.Target)
	}

	return
}
