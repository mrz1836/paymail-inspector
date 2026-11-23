package cli

import (
	"regexp"
	"strings"

	"github.com/mrz1836/go-sanitize"
)

// Example values used in help text and docs; all of them are live. RelayX
// does not publish the verify capability, so verify examples use Simply Cash.
const (
	exampleDomain        = "relayx.io"
	exampleAlias         = "mrz"
	exampleAddress       = exampleAlias + "@" + exampleDomain
	exampleVerifyAddress = exampleAlias + "@simply.cash"
	examplePubKey        = "022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec"
)

// Where the paymail specifications live. The original bsvalias.org domain
// lapsed and now serves unrelated content, so it is never linked; the same
// pages are published from the bitcoin-sv-specs/brfc-paymail repository.
const (
	specMirror      = "https://bitcoin-sv-specs.github.io/brfc-paymail/"
	paymailStandard = "https://tsc.bsvblockchain.org/standards/paymail/"
)

// legacySpecHost matches links to the lapsed bsvalias.org domain.
var legacySpecHost = regexp.MustCompile(`^https?://(www\.)?bsvalias\.org/`)

// specURL rewrites a link to the lapsed bsvalias.org domain (as found in
// go-paymail's bundled BRFC list) to the same page on the spec mirror.
func specURL(u string) string {
	return legacySpecHost.ReplaceAllString(u, specMirror)
}

// provider is a public paymail provider that issues handles.
type provider struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	Link   string `json:"link"`
}

// defaultProviders lists the public paymail providers that whois searches.
// Every entry was verified live (SRV record and/or /.well-known/bsvalias)
// when it was added; providers that shut down are removed rather than left
// to time out.
func defaultProviders() []provider {
	return []provider{
		{Domain: "handcash.io", Name: "HandCash", Link: "https://handcash.io"},
		{Domain: "relayx.io", Name: "RelayX", Link: "https://relayx.com"},
		{Domain: "simply.cash", Name: "Simply Cash", Link: "https://simply.cash"},
		{Domain: "paymail.us", Name: "Paymail.us", Link: "https://paymail.us"},
		{Domain: "1sat.app", Name: "1Sat Ordinals", Link: "https://1sat.app"},
	}
}

// providerFor returns the known provider for domain, or a generic entry
// pointing at the domain itself.
func (a *app) providerFor(domain string) provider {
	domain = strings.ToLower(domain)
	if clean, err := sanitize.Domain(domain, false, true); err == nil && clean != "" {
		domain = clean
	}

	for _, p := range a.providers {
		if p.Domain == domain {
			return p
		}
	}

	return provider{Domain: domain, Name: domain, Link: "https://" + domain}
}
