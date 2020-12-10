package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-sanitize"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/spf13/cobra"
	"github.com/tonicpow/go-paymail"
	"github.com/ttacon/chalk"
)

// brfcCmd represents the brfc command (Bitcoin SV Request-For-Comments)
// http://bsvalias.org/01-brfc-specifications.html
var brfcCmd = &cobra.Command{
	Use:        "brfc",
	Short:      "List all specs, search by keyword, or generate a new BRFC ID",
	Aliases:    []string{"spec", "b"},
	SuggestFor: []string{"specs", "specifications"},
	Example: applicationName + ` brfc list
` + applicationName + ` brfc search nChain
` + applicationName + ` brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1`,
	Long: chalk.Green.NewStyle().WithTextStyle(chalk.Bold).Style(`
___.           _____       
\_ |__________/ ____\____  
 | __ \_  __ \   __\/ ___\ 
 | \_\ \  | \/|  | \  \___ 
 |___  /__|   |__|  \___  >
     \/                 \/`) + `
` + chalk.Yellow.Color(`
Use the [list] argument to show all known BRFC protocols.

Use the [generate] argument with required flags to generate a new BRFC ID.

Use the [search] argument to show any matching BRFCs by either ID, Title or Author.

BRFC (Bitcoin SV Request-For-Comments) Specifications describe functionality across the ecosystem. 
"bsvalias" protocols and paymail implementations are described across a series of BRFC documents.

Whilst this is not the authoritative definition of the BRFC process, a summary is included here 
as the BRFC process is the nominated mechanism through which extensions to the paymail system 
are defined and discovered during Service Discovery.

Read more at: `+chalk.Cyan.Color("http://bsvalias.org/01-brfc-specifications.html")),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("brfc requires either [list] or [generate] or [search]")
		}
		if args[0] != "list" && args[0] != "generate" && args[0] != "search" {
			return chalker.Error("brfc requires either [list] or [generate] or [search]")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Load the BRFC specifications via new client
		client, err := paymail.NewClient(nil, nil, nil)
		if err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Error loading BRFC specifications: %s", err.Error()))
			return
		}

		// Search command
		if args[0] == "search" {

			// Did we find some specifications?
			if len(client.Options.BRFCSpecs) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("No existing brfc specs found in: %s", "BRFCSpecs"))
				return
			}

			// No second argument?
			if len(args) == 1 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Search requires a second argument: %s", "search term"))
				return
			}

			// Basic sanitation
			searchTerm := strings.TrimSpace(sanitize.SingleLine(args[1]))

			// Loop the list
			found := 0
			for _, brfc := range client.Options.BRFCSpecs {
				if simpleSearch(brfc.ID, searchTerm) || simpleSearch(brfc.Title, searchTerm) || simpleSearch(brfc.Author, searchTerm) {
					showBrfc(brfc)
					found = found + 1
				}
			}

			// Show success message
			if found > 0 {
				chalker.Log(chalker.SUCCESS, fmt.Sprintf("Total BRFC specification(s) found: %d searching: %s", found, searchTerm))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("No BRFC specifications found searching for: %s", searchTerm))
			}
			return
		}

		// List command
		if args[0] == "list" {

			// Did we find some specifications?
			if len(client.Options.BRFCSpecs) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("No existing brfc specs found in: %s", "BRFCSpecs"))
				return
			}

			// Loop the list
			for _, brfc := range client.Options.BRFCSpecs {

				// Skip an invalid specs in the JSON (there should NOT be any invalid specs)
				if len(brfc.Title) == 0 || len(brfc.Version) == 0 || len(brfc.ID) == 0 {
					continue
				}

				displayHeader(chalker.BOLD, brfc.Title+" v"+brfc.Version)

				valid := chalk.Green.Color("(Valid)")
				// Validate the BRFC ID
				if !skipBrfcValidation {
					if ok, id, err := brfc.Validate(); err != nil {
						chalker.Log(chalker.ERROR, fmt.Sprintf("Error validating brfc %s: %s", brfc.ID, err.Error()))
					} else if !ok {
						valid = chalk.Yellow.Color("(Invalid ID, generator says: " + id + ")")
						// valid = chalk.Magenta.Color("(Invalid ID)")
					}
				}

				chalker.Log(chalker.DEFAULT, fmt.Sprintf("ID        : %s %s", chalk.Cyan.Color(brfc.ID), valid))

				if len(brfc.Author) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("Author(s) : %s", chalk.Cyan.Color(brfc.Author)))
				}
				if len(brfc.Alias) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("Alias     : %s", chalk.Cyan.Color(brfc.Alias)))
				}
				if len(brfc.URL) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("URL       : %s", chalk.Cyan.Color(brfc.URL)))
				}
			}

			// Show success message
			chalker.Log(chalker.SUCCESS, fmt.Sprintf("Total BRFC specifications found: %d", len(client.Options.BRFCSpecs)))

			return
		}

		// Generate command
		if args[0] == "generate" {

			// Validate title, author, version
			if len(brfcTitle) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Missing required flag: %s", "--title"))
				return
			}

			// Create the new BRFC
			brfc := &paymail.BRFCSpec{
				Author:  brfcAuthor,
				Title:   brfcTitle,
				Version: brfcVersion,
			}

			// Generate the ID
			var err error
			if err = brfc.Generate(); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error generating BRFC ID: %s", err.Error()))
				return
			}

			// Check that it doesn't exist
			for _, existingBrfc := range client.Options.BRFCSpecs {
				if existingBrfc.ID == brfc.ID {
					chalker.Log(chalker.ERROR, fmt.Sprintf("BRFC already exists: %s", brfc.ID))
					return
				}
			}

			displayHeader(chalker.BOLD, "Generating BRFC ID...")

			// Show the generated ID
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Generated ID: %s", chalk.Cyan.Color(brfc.ID)))
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("Title       : %s", chalk.Cyan.Color(brfc.Title)))

			// Show optional fields
			if len(brfc.Author) > 0 {
				chalker.Log(chalker.DEFAULT, fmt.Sprintf("Author      : %s", chalk.Cyan.Color(brfc.Author)))
			}
			if len(brfc.Version) > 0 {
				chalker.Log(chalker.DEFAULT, fmt.Sprintf("Version     : %s", chalk.Cyan.Color(brfc.Version)))
			}

			// Done
			return
		}
	},
}

func simpleSearch(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

// showBrfc will show a given brfc
func showBrfc(brfc *paymail.BRFCSpec) {

	// Header
	displayHeader(chalker.BOLD, brfc.Title+" v"+brfc.Version)

	// Validate the BRFC ID
	valid := chalk.Green.Color("(Valid)")
	if !skipBrfcValidation {
		if ok, id, err := brfc.Validate(); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Error validating brfc %s: %s", brfc.ID, err.Error()))
		} else if !ok {
			valid = chalk.Yellow.Color("(Invalid ID, generator says: " + id + ")")
		}
	}

	chalker.Log(chalker.DEFAULT, fmt.Sprintf("ID        : %s %s", chalk.Cyan.Color(brfc.ID), valid))

	if len(brfc.Author) > 0 {
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Author(s) : %s", chalk.Cyan.Color(brfc.Author)))
	}
	if len(brfc.Alias) > 0 {
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Alias     : %s", chalk.Cyan.Color(brfc.Alias)))
	}
	if len(brfc.URL) > 0 {
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("URL       : %s", chalk.Cyan.Color(brfc.URL)))
	}
}

func init() {
	rootCmd.AddCommand(brfcCmd)

	// Set the title of the brfc
	brfcCmd.Flags().StringVar(&brfcTitle, "title", "", "Title of the new BRFC specification")

	// Set the author of the brfc
	brfcCmd.Flags().StringVar(&brfcAuthor, "author", "", "Author(s) new BRFC specification")

	// Set the version of the brfc
	brfcCmd.Flags().StringVar(&brfcVersion, "version", "", "Version of the new BRFC specification")

	// Skip validating the BRFC ids
	brfcCmd.Flags().BoolVar(&skipBrfcValidation, "skip-validation", false, "Skip validating the existing BRFC IDs")
}
