package paymail

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// BRFCSpec is a full BRFC specification document
// See more: http://bsvalias.org/01-brfc-specifications.html
type BRFCSpec struct {
	Alias      string `json:"alias,omitempty"`      // Alias is used in the list of capabilities
	Author     string `json:"author"`               // Free-form, could include a name, alias, paymail address, GitHub/social media handle, etc.
	ID         string `json:"id"`                   // Public BRFC ID
	Supersedes string `json:"supersedes,omitempty"` // A BRFC ID (or list of IDs) that this document supersedes
	Title      string `json:"title"`                // Title of the brfc
	URL        string `json:"url,omitempty"`        // Public URL to view the specification
	Valid      bool   `json:"valid"`                // Validated the ID -> (title,author,version)
	Version    string `json:"version"`              // No set format; could be a sequence number, publication date, or any other scheme
}

var (
	// ListOfBRFCSpecs is a public variable with a list of known BRFC specifications
	ListOfBRFCSpecs []*BRFCSpec
)

func init() {

	// Service Discovery
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Author:  "andy (nChain), Ryan X. Charles (Money Button)",
		ID:      "b2aa66e26b43",
		Title:   "bsvalias Service Discovery",
		URL:     "http://bsvalias.org/02-service-discovery.html",
		Version: "1",
	})

	// Public Key Infrastructure
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Alias:   "pki",
		Author:  "andy (nChain), Ryan X. Charles (Money Button)",
		ID:      "0c4339ef99c2",
		Title:   "bsvalias Public Key Infrastructure",
		URL:     "http://bsvalias.org/03-public-key-infrastructure.html",
		Version: "1",
	})

	// Basic Address Resolution
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Alias:   "paymentDestination",
		Author:  "andy (nChain), Ryan X. Charles (Money Button)",
		ID:      "759684b1a19a",
		Title:   "bsvalias Payment Addressing (Basic Address Resolution)",
		URL:     "http://bsvalias.org/04-01-basic-address-resolution.html",
		Version: "1",
	})

	// Sender Validation
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "6745385c3fc0",
		Title:   "bsvalias Payment Addressing (Payer Validation)",
		URL:     "http://bsvalias.org/04-02-sender-validation.html",
		Version: "1",
	})

	// Receiver Approvals
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "3d7c2ca83a46",
		Title:   "bsvalias Payment Addressing (Payee Approvals)",
		URL:     "http://bsvalias.org/04-03-receiver-approvals.html",
		Version: "1",
	})

	// PayTo Protocol Prefix
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "7bd25e5a1fc6",
		Title:   "bsvalias Payment Addressing (PayTo Protocol Prefix)",
		URL:     "http://bsvalias.org/04-04-payto-protocol-prefix.html",
		Version: "1",
	})

	// Verify Public Key Owner
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Author:  "andy (nChain), Ryan X. Charles (Money Button), Miguel Duarte (Money Button)",
		ID:      "a9f510c16bde",
		Title:   "bsvalias public key verify (Verify Public Key Owner)",
		URL:     "http://bsvalias.org/05-verify-public-key-owner.html",
		Version: "1",
	})

	// P2P Transactions
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Author:  "Ryan X. Charles (Money Button), Miguel Duarte (Money Button), Rafa Jimenez Seibane (Handcash), Ivan Mlinarić  (Handcash)",
		ID:      "5f1323cddf31",
		Title:   "P2P Transactions",
		URL:     "https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html",
		Version: "1",
	})

	// P2P Payment Destination
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		Author:  "Ryan X. Charles (Money Button), Miguel Duarte (Money Button), Rafa Jimenez Seibane (Handcash), Ivan Mlinarić  (Handcash)",
		ID:      "2a40af698840",
		Title:   "P2P Payment Destination",
		URL:     "https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html",
		Version: "1",
	})

	// Public Profile // todo: still unknown where the source is
	ListOfBRFCSpecs = append(ListOfBRFCSpecs, &BRFCSpec{
		ID:    "f12f968c92d6",
		Title: "Public Profile",
	})

}

// Generate will generate a new BRFC ID from the given specification
// See more: http://bsvalias.org/01-02-brfc-id-assignment.html
func (b *BRFCSpec) Generate() (id string, err error) {

	// Validate the title, author or version
	if len(b.Title) == 0 {
		err = fmt.Errorf("invalid brfc title, length: 0")
		return
	}

	// Start a new SHA256 hash
	h := sha256.New()

	// Append all values (trim leading & trailing whitespace)
	h.Write([]byte(strings.TrimSpace(b.Title) + strings.TrimSpace(b.Author) + strings.TrimSpace(b.Version)))

	// Start the double SHA256
	h2 := sha256.New()

	// Write the first SHA256 result
	h2.Write(h.Sum(nil))

	// Create the final double SHA256
	doubleHash := h2.Sum(nil)

	// fmt.Printf("doubleHash: %x\n", doubleHash)

	// Reverse the order
	for i, j := 0, len(doubleHash)-1; i < j; i, j = i+1, j-1 {
		doubleHash[i], doubleHash[j] = doubleHash[j], doubleHash[i]
	}

	// fmt.Printf("doubleHash reversed: %x\n", doubleHash)

	// Hex encode the value
	hexDoubleHash := make([]byte, hex.EncodedLen(len(doubleHash)))
	hex.Encode(hexDoubleHash, doubleHash)

	// fmt.Printf("hex.Encode: %x\n", hexDoubleHash)

	// Extract the ID and set
	if len(hexDoubleHash) >= 12 {
		id = string(hexDoubleHash[:12])
	} else {
		err = fmt.Errorf("failed to generate a valid id, length was %d", len(hexDoubleHash))
	}

	return
}

// Validate will check if the BRFC is valid or not (and set b.Valid)
// Returns the ID that was generated to compare
func (b *BRFCSpec) Validate() (valid bool, id string, err error) {

	// Run the generate method to return an ID
	if id, err = b.Generate(); err != nil {
		return
	}

	// Are we the same ID as given?
	if b.ID == id {
		valid = true
		b.Valid = valid
	}

	return
}
