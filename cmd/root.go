/*
Package cmd is all the available commands for the CLI application
*/
package cmd

import (
	"fmt"

	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/database"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	DisableAutoGenTag: true,
	Use:               applicationName,
	Short:             "Inspect, validate domains or resolve paymail addresses",
	Example:           applicationName + " -h",
	Long: chalk.Green.NewStyle().WithTextStyle(chalk.Bold).Style(`
__________                             .__.__    .___                                     __                
\______   \_____  ___.__. _____ _____  |__|  |   |   | ____   ____________   ____   _____/  |_  ___________ 
 |     ___/\__  \<   |  |/     \\__  \ |  |  |   |   |/    \ /  ___/\____ \_/ __ \_/ ___\   __\/  _ \_  __ \
 |    |     / __ \\___  |  Y Y  \/ __ \|  |  |__ |   |   |  \\___ \ |  |_> >  ___/\  \___|  | (  <_> )  | \/
 |____|    (____  / ____|__|_|  (____  /__|____/ |___|___|  /____  >|   __/ \___  >\___  >__|  \____/|__|   
                \/\/          \/     \/                   \/     \/ |__|        \/     \/     `+Version) + `
` + chalk.Yellow.Color("Author: MrZ © 2021 github.com/mrz1836/"+applicationFullName) + `

This CLI app is used for interacting with paymail service providers.

Help contribute via Github!
`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	// Create a database connection (Don't require DB for now)
	if err := database.Connect(applicationName); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("Error connecting to database: %s", err.Error()))
	} else {
		// Set this flag for caching detection
		databaseEnabled = true

		// Defer the database disconnection
		defer func() {
			dbErr := database.GarbageCollection()
			if dbErr != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error in database GarbageCollection: %s", dbErr.Error()))
			}

			if dbErr = database.Disconnect(); dbErr != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error in database Disconnect: %s", dbErr.Error()))
			}
		}()
	}

	// Run root command
	er(rootCmd.Execute())

	// Generate documentation from all commands
	if generateDocs {
		generateDocumentation()
	}

	// Flush cache?
	if flushCache && databaseEnabled {
		if dbErr := database.Flush(); dbErr != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Error in database Flush: %s", dbErr.Error()))
		} else {
			chalker.Log(chalker.SUCCESS, "Successfully flushed the local database cache")
		}
	}
}
