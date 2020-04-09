package paymail

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

/*
Alternative checks:

https://dnsviz.net/d/domain.com/dnssec/
https://dnssec-analyzer.verisignlabs.com/domain.com
*/

// DNSCheckResult struct is returned for the DNS check
type DNSCheckResult struct {
	Answer       Answer    `json:"answer"`
	CheckTime    time.Time `json:"time"`
	DNSSEC       bool      `json:"dnssec"`
	Domain       string    `json:"domain,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	NSEC         NSEC      `json:"nsec"`
}

// NSEC struct for NSEC type
type NSEC struct {
	NSEC       *dns.NSEC       `json:"nsec,omitempty"`
	NSEC3      *dns.NSEC3      `json:"nsec_3,omitempty"`
	NSEC3PARAM *dns.NSEC3PARAM `json:"nsec_3_param,omitempty"`
	Type       string          `json:"type,omitempty"`
}

// Answer struct the answer of the DNS question
type Answer struct {
	CalculatedDS      []*DomainDS     `json:"calculate_ds,omitempty"`
	DNSKEYRecordCount int             `json:"dnskey_record_count,omitempty"`
	DNSKEYRecords     []*DomainDNSKEY `json:"dnskey_records,omitempty"`
	DSRecordCount     int             `json:"ds_record_count,omitempty"`
	DSRecords         []*DomainDS     `json:"ds_records,omitempty"`
	Matching          Matching        `json:"matching,omitempty"`
}

// Matching struct for information
type Matching struct {
	DNSKEY []*DomainDNSKEY `json:"dnskey,omitempty"`
	DS     []*DomainDS     `json:"ds,omitempty"`
}

// DomainDS struct
type DomainDS struct {
	Algorithm  uint8  `json:"algorithm,omitempty"`
	Digest     string `json:"digest,omitempty"`
	DigestType uint8  `json:"digest_type,omitempty"`
	KeyTag     uint16 `json:"key_tag,omitempty"`
}

// DomainDNSKEY struct
type DomainDNSKEY struct {
	Algorithm    uint8     `json:"algorithm,omitempty"`
	CalculatedDS *DomainDS `json:"calculate_ds,omitempty"`
	Flags        uint16    `json:"flags,omitempty"`
	Protocol     uint8     `json:"protocol,omitempty"`
	PublicKey    string    `json:"public_key,omitempty"`
}

// Domains that DO NOT work properly
var (

	// todo: find a way to make these work
	// https://network-tools.com/nslookup/ for a heroku app produces 0 results
	domainsWithIssues = []string{
		"herokuapp.com", // CNAME on heroku is a pointer, and thus there is no NS returned
	}
)

// CheckDNSSEC will check the DNSSEC for a given domain
func CheckDNSSEC(domain string, nameServer string) (result *DNSCheckResult) {

	// Start the new result
	result = new(DNSCheckResult)
	result.CheckTime = time.Now()

	var err error

	// Valid domain name (ASCII or IDN)
	if domain, err = idna.ToASCII(domain); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in ToASCII: %s", err.Error())
		return
	}

	// Validate domain
	if domain, err = publicsuffix.EffectiveTLDPlusOne(domain); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in EffectiveTLDPlusOne: %s", err.Error())
		return
	}

	// Set the valid domain now
	result.Domain = domain

	// Check known domain issues
	for _, d := range domainsWithIssues {
		if strings.Contains(result.Domain, d) {
			result.ErrorMessage = fmt.Sprintf("%s cannot be validated due to a known issue with %s", result.Domain, d)
			return
		}
	}

	// Set the TLD
	tld, _ := publicsuffix.PublicSuffix(domain)

	// Set the registry name server
	var registryNameserver string
	if registryNameserver, err = resolveOneNS(tld, nameServer); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in resolveOneNS: %s", err.Error())
		return
	}

	// Set the domain name server
	var domainNameserver string
	if domainNameserver, err = resolveOneNS(domain, nameServer); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in resolveOneNS: %s", err.Error())
		return
	}

	// Domain name servers at registrar Host
	var domainDS []*DomainDS
	if domainDS, err = resolveDomainDS(domain, registryNameserver); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in resolveOneNS: %s", err.Error())
		return
	}

	// Set the records and count
	result.Answer.DSRecords = domainDS
	result.Answer.DSRecordCount = cap(domainDS)

	// Resolve domain DNSKey
	var dnsKey []*DomainDNSKEY
	if dnsKey, err = resolveDomainDNSKEY(domain, domainNameserver); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in resolveDomainDNSKEY: %s", err.Error())
		return
	}

	// Set the DNSKEY records
	result.Answer.DNSKEYRecords = dnsKey
	result.Answer.DNSKEYRecordCount = cap(result.Answer.DNSKEYRecords)

	// Check the digest type
	var digest uint8
	if cap(result.Answer.DSRecords) != 0 {
		digest = result.Answer.DSRecords[0].DigestType
	}

	// Check the DS record
	if result.Answer.DSRecordCount > 0 && result.Answer.DNSKEYRecordCount > 0 {
		var calculatedDS []*DomainDS
		if calculatedDS, err = calculateDSRecord(domain, digest, domainNameserver); err != nil {
			result.ErrorMessage = fmt.Sprintf("failed in calculateDSRecord: %s", err.Error())
			return
		}
		result.Answer.CalculatedDS = calculatedDS
	}

	// Resolve the domain NSEC
	var nsec *dns.NSEC
	if nsec, err = resolveDomainNSEC(domain, nameServer); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in resolveDomainNSEC: %s", err.Error())
		return
	} else if nsec != nil {
		result.NSEC.Type = "nsec"
		result.NSEC.NSEC = nsec
	}

	// Resolve the domain NSEC3
	var nsec3 *dns.NSEC3
	if nsec3, err = resolveDomainNSEC3(domain, nameServer); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in resolveDomainNSEC3: %s", err.Error())
		return
	} else if nsec3 != nil {
		result.NSEC.Type = "nsec3"
		result.NSEC.NSEC3 = nsec3
	}

	// Resolve the domain NSEC3PARAM
	var nsec3param *dns.NSEC3PARAM
	if nsec3param, err = resolveDomainNSEC3PARAM(domain, nameServer); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed in resolveDomainNSEC3PARAM: %s", err.Error())
		return
	} else if nsec3param != nil {
		result.NSEC.Type = "nsec3param"
		result.NSEC.NSEC3PARAM = nsec3param
	}

	// Check the keys and set the DNSSEC flag
	if result.Answer.DSRecordCount > 0 && result.Answer.DNSKEYRecordCount > 0 {
		var filtered []*DomainDS
		var dnsKeys []*DomainDNSKEY
		for _, e := range result.Answer.DSRecords {
			for i, f := range result.Answer.CalculatedDS {
				if f.Digest == e.Digest {
					filtered = append(filtered, f)
					dnsKeys = append(dnsKeys, result.Answer.DNSKEYRecords[i])
				}
			}
		}
		result.Answer.Matching.DS = filtered
		result.Answer.Matching.DNSKEY = dnsKeys
		result.DNSSEC = true
	} else {
		result.DNSSEC = false
	}

	// Done!
	return
}

// resolveOneNS will resolve one name server
func resolveOneNS(domain string, nameServer string) (string, error) {
	var answer []string
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	m.MsgHdr.RecursionDesired = true
	m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameServer+":"+defaultDnsPort)
	if err != nil {
		return "", err
	}
	for _, ain := range in.Answer {
		if a, ok := ain.(*dns.NS); ok {
			answer = append(answer, a.Ns)
		}
	}
	if len(answer) < 1 || answer == nil {
		return "", err
	}
	return answer[0], nil
}

// resolveDomainNSEC will resolve a domain NSEC
func resolveDomainNSEC(domain string, nameServer string) (*dns.NSEC, error) {
	var answer *dns.NSEC
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNSEC)
	m.MsgHdr.RecursionDesired = true
	m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameServer+":"+defaultDnsPort)
	if err != nil {
		return nil, err
	}
	for _, ain := range in.Answer {
		if a, ok := ain.(*dns.NSEC); ok {
			answer = a
			return answer, nil
		}
	}
	return nil, nil
}

// resolveDomainNSEC3 will resolve a domain NSEC3
func resolveDomainNSEC3(domain string, nameServer string) (*dns.NSEC3, error) {
	var answer *dns.NSEC3
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNSEC3)
	m.MsgHdr.RecursionDesired = true
	m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameServer+":"+defaultDnsPort)
	if err != nil {
		return nil, err
	}
	for _, ain := range in.Answer {
		if a, ok := ain.(*dns.NSEC3); ok {
			answer = a
			return answer, nil
		}
	}
	return nil, nil
}

// resolveDomainNSEC3PARAM will resolve a domain NSEC3PARAM
func resolveDomainNSEC3PARAM(domain string, nameServer string) (*dns.NSEC3PARAM, error) {
	var answer *dns.NSEC3PARAM
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNSEC3PARAM)
	m.MsgHdr.RecursionDesired = true
	m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameServer+":"+defaultDnsPort)
	if err != nil {
		return nil, err
	}
	for _, ain := range in.Answer {
		if a, ok := ain.(*dns.NSEC3PARAM); ok {
			answer = a
			return answer, nil
		}
	}
	return nil, nil
}

// resolveDomainDS will resolve a domain DS
func resolveDomainDS(domain string, nameServer string) ([]*DomainDS, error) {
	var ds []*DomainDS
	m := new(dns.Msg)
	m.MsgHdr.RecursionDesired = true
	m.SetQuestion(dns.Fqdn(domain), dns.TypeDS)
	m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameServer+":"+defaultDnsPort)
	if err != nil {
		return ds, err
	}
	for _, ain := range in.Answer {
		if a, ok := ain.(*dns.DS); ok {
			readKey := new(DomainDS)
			readKey.Algorithm = a.Algorithm
			readKey.Digest = a.Digest
			readKey.DigestType = a.DigestType
			readKey.KeyTag = a.KeyTag
			ds = append(ds, readKey)
		}
	}
	return ds, nil
}

// resolveDomainDNSKEY will resolve a domain DNSKEY
func resolveDomainDNSKEY(domain string, nameServer string) ([]*DomainDNSKEY, error) {
	var dnskey []*DomainDNSKEY

	m := new(dns.Msg)
	m.MsgHdr.RecursionDesired = true
	m.SetQuestion(dns.Fqdn(domain), dns.TypeDNSKEY)
	m.SetEdns0(4096, true)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameServer+":"+defaultDnsPort)
	if err != nil {
		return dnskey, err
	}
	for _, ain := range in.Answer {
		if a, ok := ain.(*dns.DNSKEY); ok {
			readKey := new(DomainDNSKEY)
			readKey.Algorithm = a.Algorithm
			readKey.Flags = a.Flags
			readKey.Protocol = a.Protocol
			readKey.PublicKey = a.PublicKey
			dnskey = append(dnskey, readKey)
		}
	}
	return dnskey, err
}

// calculateDSRecord function for generating DS records from the DNSKEY
// Input: domain, digest and name server from the host
// Output: one of more structs with DS information
func calculateDSRecord(domain string, digest uint8, nameServer string) ([]*DomainDS, error) {
	var calculatedDS []*DomainDS

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeDNSKEY)
	m.SetEdns0(4096, true)
	m.MsgHdr.RecursionDesired = true
	c := new(dns.Client)
	in, _, err := c.Exchange(m, nameServer+":"+defaultDnsPort)
	if err != nil {
		return calculatedDS, err
	}
	for _, ain := range in.Answer {
		if a, ok := ain.(*dns.DNSKEY); ok {
			calculatedKey := new(DomainDS)
			calculatedKey.Algorithm = a.ToDS(digest).Algorithm
			calculatedKey.Digest = a.ToDS(digest).Digest
			calculatedKey.DigestType = a.ToDS(digest).DigestType
			calculatedKey.KeyTag = a.ToDS(digest).KeyTag
			calculatedDS = append(calculatedDS, calculatedKey)
		}
	}
	return calculatedDS, nil
}
