package paymail

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	// BRFCSpecs is a public variable with a list of known BRFC specifications
	BRFCSpecs []*BRFCSpec
)

// LoadSpecifications will load the known specifications into structs from JSON
func LoadSpecifications() (err error) {
	if err = json.Unmarshal([]byte(BRFCKnownSpecifications), &BRFCSpecs); err == nil && len(BRFCSpecs) == 0 {
		err = fmt.Errorf("error loading BRFC specifications, zero results found")
	}
	return
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
