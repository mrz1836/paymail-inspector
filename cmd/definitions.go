package cmd

import (
	"github.com/mrz1836/go-sanitize"
	"github.com/mrz1836/paymail-inspector/integrations/baemail"
	"github.com/mrz1836/paymail-inspector/integrations/bitpic"
	"github.com/mrz1836/paymail-inspector/integrations/powping"
	"github.com/mrz1836/paymail-inspector/integrations/roundesk"
	"github.com/tonicpow/go-paymail"
)

// Version is set manually (also make:build overwrites this value from Github's latest tag)
var Version = "v0.3.13"

// Default flag values for various commands
var (
	amount             uint64 // cmd: resolve
	brfcAuthor         string // cmd: brfc
	brfcTitle          string // cmd: brfc
	brfcVersion        string // cmd: brfc
	configFile         string // cmd: root
	disableCache       bool   // cmd: root
	flushCache         bool   // cmd: root
	generateDocs       bool   // cmd: root
	nameServer         string // cmd: validate
	port               uint16 // cmd: validate
	priority           uint16 // cmd: validate
	protocol           string // cmd: validate
	purpose            string // cmd: resolve
	satoshis           uint64 // cmd: resolve
	serviceName        string // cmd: validate
	signature          string // cmd: resolve
	skipBaemail        bool   // cmd: resolve
	skipBitpic         bool   // cmd: resolve
	skipBrfcValidation bool   // cmd: brfc
	skipDNSCheck       bool   // cmd: validate
	skipPki            bool   // cmd: resolve
	skipPowPing        bool   // cmd: resolve
	skipPublicProfile  bool   // cmd: resolve
	skipRoundesk       bool   // cmd: resolve
	skipSrvCheck       bool   // cmd: validate
	skipSSLCheck       bool   // cmd: validate
	skipTracing        bool   // cmd: root
	weight             uint16 // cmd: validate
)

// Application global variables
var (
	applicationDirectory string // Folder path for the application resources
	databaseEnabled      bool   // Flag is set if DB loads successfully
)

// Defaults for the application
const (
	applicationFullName = "paymail-inspector" // Full name of the application (long version)
	applicationName     = "paymail"           // Application name (binary) (short version
	configFileDefault   = "config"            // Config file name
	defaultDomainName   = "moneybutton.com"   // Used in examples
	defaultNameServer   = "8.8.8.8"           // Default DNS NameServer
	docsLocation        = "docs/commands"     // Default location for command documentation
	flagBsvAlias        = "bsvalias"          // Flag for a known, common key
	flagSenderHandle    = "sender-handle"
	flagSenderName      = "sender-name"
)

// Provider is the paymail provider information
type Provider struct {
	Domain string `json:"domain"`
	Link   string `json:"link"`
}

// providers is a list of providers that user's can obtain a paymail
var providers = []*Provider{
	{"moneybutton.com", "https://tncpw.co/4c58a26f"},
	{"handcash.io", "https://tncpw.co/742b1f09"},
	{"relayx.io", "https://tncpw.co/4897634e"},
	{"centbee.com", "https://tncpw.co/4350c72f"},
	{"simply.cash", "https://tncpw.co/1ce8f70f"},
	{"dotwallet.com", "https://tncpw.co/5745c80e"},
	{"mypaymail.co", "https://tncpw.co/ee243a15"},
	{"volt.id", "https://tncpw.co/e9ff2b0c"},
}

// getProvider will return a provider given the domain name
func getProvider(domain string) *Provider {
	domain, _ = sanitize.Domain(domain, false, true)
	for _, provider := range providers {
		if domain == provider.Domain {
			return provider
		}
	}
	return nil
}

// PaymailDetails is all the info about one paymail address
type PaymailDetails struct {
	Baemail       *baemail.Response      `json:"baemail"`
	Bitpic        string                 `json:"bitpic"`
	Bitpics       *bitpic.SearchResponse `json:"bitpics"`
	Dimely        string                 `json:"dimely"`
	Handle        string                 `json:"handle"`
	PKI           *paymail.PKI           `json:"pki"`
	PowPing       *powping.Response      `json:"powping"`
	Provider      *Provider              `json:"provider"`
	PublicProfile *paymail.PublicProfile `json:"public_profile"`
	Resolution    *paymail.Resolution    `json:"resolution"`
	Roundesk      *roundesk.Response     `json:"roundesk"`
}
