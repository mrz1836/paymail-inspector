package cmd

import (
	"fmt"

	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

// brfcCmd represents the brfc command (Bitcoin SV Request-For-Comments)
// http://bsvalias.org/01-brfc-specifications.html
var brfcCmd = &cobra.Command{
	Use:   "brfc",
	Short: "List all known BRFC specs or Generate a new BRFC number",
	Example: configDefault + ` brfc list
` + configDefault + ` brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1`,
	Long: chalk.Green.Color(`
___.           _____       
\_ |__________/ ____\____  
 | __ \_  __ \   __\/ ___\ 
 | \_\ \  | \/|  | \  \___ 
 |___  /__|   |__|  \___  >
     \/                 \/`) + `
` + chalk.Yellow.Color(`
Use the [list] argument to show all known BRFC protocols.

Use the [generate] argument with required flags to generate a new BRFC ID.

BRFC (Bitcoin SV Request-For-Comments) Specifications describe functionality across the ecosystem. 
"bsvalias" protocols and paymail implementations are described across a series of BRFC documents.

Whilst this is not the authoritative definition of the BRFC process, a summary is included here 
as the BRFC process is the nominated mechanism through which extensions to the paymail system 
are defined and discovered during Service Discovery.

Read more at: `+chalk.Cyan.Color("http://bsvalias.org/01-brfc-specifications.html")),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("brfc requires either [list] or [generate]")
		} else if len(args) > 1 {
			return chalker.Error("brfc only supports one argument")
		}
		if args[0] != "list" && args[0] != "generate" {
			return chalker.Error("brfc requires either [list] or [generate]")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Load the BRFC specifications
		if len(paymail.BRFCSpecs) == 0 {
			if err := paymail.LoadSpecifications(); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error loading BRFC specifications: %s", err.Error()))
				return
			}
		}

		// List command
		if args[0] == "list" {

			// Did we find some specifications?
			if len(paymail.BRFCSpecs) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("no existing brfc specs found in: %s", "BRFCSpecs"))
				return
			}

			// Show success message
			chalker.Log(chalker.SUCCESS, fmt.Sprintf("total brfc specs found: %d", len(paymail.BRFCSpecs)))

			// Loop the list
			for _, brfc := range paymail.BRFCSpecs {
				chalker.Log(chalker.DEFAULT, "-----------------------------------------------------------------")
				if len(brfc.ID) == 0 {
					chalker.Log(chalker.WARN, fmt.Sprintf("invalid brfc detected, missing attribute: %s", "id"))
					continue
				}
				if len(brfc.ID) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("id: %s", chalk.Cyan.Color(brfc.ID)))
				}
				if len(brfc.Title) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("title: %s", chalk.Cyan.Color(brfc.Title)))
				}
				if len(brfc.Author) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("author: %s", chalk.Cyan.Color(brfc.Author)))
				}
				if len(brfc.Version) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("version: %s", chalk.Cyan.Color(brfc.Version)))
				}
				if len(brfc.Alias) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("alias: %s", chalk.Cyan.Color(brfc.Alias)))
				}
				if len(brfc.URL) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("url: %s", chalk.Cyan.Color(brfc.URL)))
				}

				// Validate the BRFC ID
				if !skipBrfcValidation {
					if ok, id, err := brfc.Validate(); err != nil {
						chalker.Log(chalker.ERROR, fmt.Sprintf("error validating brfc %s: %s", brfc.ID, err.Error()))
					} else if ok {
						chalker.Log(chalker.DEFAULT, fmt.Sprintf("validation: %s", chalk.Green.Color("success")))
					} else {
						chalker.Log(chalker.DEFAULT, fmt.Sprintf("validation: %s", chalk.Magenta.Color("failed, generated id: "+id)))
					}
				}
			}
			return
		}

		// Generate command
		if args[0] == "generate" {

			// Validate title, author, version
			if len(brfcTitle) == 0 {
				chalker.Log(chalker.ERROR, fmt.Sprintf("missing required flag: %s", "--title"))
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
			if brfc.ID, err = brfc.Generate(); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error generating id: %s", err.Error()))
				return
			}

			// Check that it doesn't exist
			if len(paymail.BRFCSpecs) > 0 {
				for _, existingBrfc := range paymail.BRFCSpecs {
					if existingBrfc.ID == brfc.ID {
						chalker.Log(chalker.ERROR, fmt.Sprintf("brfc already exists: %s", brfc.ID))
						return
					}
				}
			}

			// Show the generated ID
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("generated id: %s", chalk.Cyan.Color(brfc.ID)))
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("title: %s", chalk.Cyan.Color(brfc.Title)))

			// Show optional fields
			if len(brfc.Author) > 0 {
				chalker.Log(chalker.DEFAULT, fmt.Sprintf("author: %s", chalk.Cyan.Color(brfc.Author)))
			}
			if len(brfc.Version) > 0 {
				chalker.Log(chalker.DEFAULT, fmt.Sprintf("version: %s", chalk.Cyan.Color(brfc.Version)))
			}

			// Done
			return
		}
	},
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
