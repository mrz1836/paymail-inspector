package paymail

import "testing"

/*
Test cases from: http://bsvalias.org/01-02-brfc-id-assignment.html
*/

// TestBRFCSpec_Generate will test the Generate() method
func TestBRFCSpec_Generate(t *testing.T) {

	// Test Case #1
	brfc := &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "57dd1f54fc67",
		Title:   "BRFC Specifications",
		Version: "1",
	}

	// Generate
	id, err := brfc.Generate()
	if err != nil {
		t.Fatalf("error occurred: %s", err.Error())
	}

	// If ID does NOT match
	if id != brfc.ID {
		t.Fatalf("generate failed, id expected: %s, got: %s", brfc.ID, id)
	}

	// Test Case #2
	brfc = &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "74524c4d6274",
		Title:   "bsvalias Payment Addressing (PayTo Protocol Prefix)",
		Version: "1",
	}

	// Generate
	id, err = brfc.Generate()
	if err != nil {
		t.Fatalf("error occurred: %s", err.Error())
	}

	// If ID does NOT match
	if id != brfc.ID {
		t.Fatalf("generate failed, id expected: %s, got: %s", brfc.ID, id)
	}

	// Test Case #3
	brfc = &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "0036f9b8860f",
		Title:   "bsvalias Integration with Simplified Payment Protocol",
		Version: "1",
	}

	// Generate
	id, err = brfc.Generate()
	if err != nil {
		t.Fatalf("error occurred: %s", err.Error())
	}

	// If ID does NOT match
	if id != brfc.ID {
		t.Fatalf("generate failed, id expected: %s, got: %s", brfc.ID, id)
	}
}

// TestBRFCSpec_Validate will test the Validate() method
func TestBRFCSpec_Validate(t *testing.T) {

	// Test Case #1
	brfc := &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "57dd1f54fc67",
		Title:   "BRFC Specifications",
		Version: "1",
	}

	ok, _, err := brfc.Validate()
	if err != nil {
		t.Fatalf("error occurred: %s", err.Error())
	} else if !ok {
		t.Fatalf("validation failed: %s", brfc.ID)
	}

	// Test Case #2
	brfc = &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "74524c4d6274",
		Title:   "bsvalias Payment Addressing (PayTo Protocol Prefix)",
		Version: "1",
	}

	ok, _, err = brfc.Validate()
	if err != nil {
		t.Fatalf("error occurred: %s", err.Error())
	} else if !ok {
		t.Fatalf("validation failed: %s", brfc.ID)
	}

	// Test Case #3
	brfc = &BRFCSpec{
		Author:  "andy (nChain)",
		ID:      "0036f9b8860f",
		Title:   "bsvalias Integration with Simplified Payment Protocol",
		Version: "1",
	}

	ok, _, err = brfc.Validate()
	if err != nil {
		t.Fatalf("error occurred: %s", err.Error())
	} else if !ok {
		t.Fatalf("validation failed: %s", brfc.ID)
	}
}
