/*
Package cmd is all the available commands for the CLI application
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/database"
	"github.com/mrz1836/paymail-inspector/integrations/bitpic"
	"github.com/mrz1836/paymail-inspector/integrations/roundesk"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

// Default flag values for various commands
var (
	amount             uint64 // cmd: resolve
	brfcAuthor         string // cmd: brfc
	brfcTitle          string // cmd: brfc
	brfcVersion        string // cmd: brfc
	configFile         string // cmd: root
	databaseEnabled    bool   // cmd: root
	disableCache       bool   // cmd: root
	flushCache         bool   // cmd: root
	generateDocs       bool   // cmd: root
	nameServer         string // cmd: validate
	port               int    // cmd: validate
	priority           int    // cmd: validate
	protocol           string // cmd: validate
	purpose            string // cmd: resolve
	satoshis           uint64 // cmd: resolve
	serviceName        string // cmd: validate
	signature          string // cmd: resolve
	skipBitpic         bool   // cmd: resolve
	skipBrfcValidation bool   // cmd: brfc
	skipDnsCheck       bool   // cmd: validate
	skipPki            bool   // cmd: resolve
	skipPublicProfile  bool   // cmd: resolve
	skipRoundesk       bool   // cmd: resolve
	skipSrvCheck       bool   // cmd: validate
	skipSSLCheck       bool   // cmd: validate
	skipTracing        bool   // cmd: root
	weight             int    // cmd: validate
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

// Version is set manually (also make:build overwrites this value from Github's latest tag)
var Version = "v0.1.4"

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	DisableAutoGenTag: true,
	Use:               applicationName,
	Short:             "Inspect, validate domains or resolve paymail addresses",
	Example:           applicationName + " -h",
	Long: chalk.Green.Color(`
__________                             .__.__    .___                                     __                
\______   \_____  ___.__. _____ _____  |__|  |   |   | ____   ____________   ____   _____/  |_  ___________ 
 |     ___/\__  \<   |  |/     \\__  \ |  |  |   |   |/    \ /  ___/\____ \_/ __ \_/ ___\   __\/  _ \_  __ \
 |    |     / __ \\___  |  Y Y  \/ __ \|  |  |__ |   |   |  \\___ \ |  |_> >  ___/\  \___|  | (  <_> )  | \/
 |____|    (____  / ____|__|_|  (____  /__|____/ |___|___|  /____  >|   __/ \___  >\___  >__|  \____/|__|   
                \/\/          \/     \/                   \/     \/ |__|        \/     \/     `+Version) + `
` + chalk.Yellow.Color("Author: MrZ © 2020 github.com/mrz1836/"+applicationFullName) + `

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

	// Generate docs from all commands
	if generateDocs {

		// Replace the colorful logs in terminal (displays in Cobra docs) (color numbers generated)
		replacer := strings.NewReplacer("[32m", "```", "[33m", "```\n", "[39m", "", "[36m", "", "\u001B", "")
		rootCmd.Long = replacer.Replace(rootCmd.Long)

		// Loop all command, adjust the Long description, re-add command
		for _, command := range rootCmd.Commands() {
			rootCmd.RemoveCommand(command)
			command.Long = replacer.Replace(command.Long)
			rootCmd.AddCommand(command)
		}

		// Generate the markdown docs
		if err := doc.GenMarkdownTree(rootCmd, docsLocation); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Error generating docs: %s", err.Error()))
			return
		}
		chalker.Log(chalker.SUCCESS, fmt.Sprintf("Successfully generated documentation for %d commands", len(rootCmd.Commands())))
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

func init() {

	// Load the configuration
	cobra.OnInitialize(initConfig)

	// Set the user agent for the application's external integrations
	bitpic.UserAgent = applicationFullName + ": v" + Version
	paymail.UserAgent = applicationFullName + ": v" + Version
	roundesk.UserAgent = applicationFullName + ": v" + Version

	// Add config option
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Custom config file (default is $HOME/"+applicationName+"/"+configFileDefault+".yaml)")

	// Add document generation for all commands
	rootCmd.PersistentFlags().BoolVar(&generateDocs, "docs", false, "Generate docs from all commands (./"+docsLocation+")")

	// Add a toggle for request tracing
	rootCmd.PersistentFlags().BoolVarP(&skipTracing, "skip-tracing", "t", false, "Turn off request tracing information")

	// Add a toggle for disabling request caching
	rootCmd.PersistentFlags().BoolVar(&disableCache, "no-cache", false, "Turn off caching for this specific command")

	// Add a toggle for flushing all the local database cache
	rootCmd.PersistentFlags().BoolVar(&flushCache, "flush-cache", false, "Flushes ALL cache, empties local database")

	// Add a bsvalias version to target
	rootCmd.PersistentFlags().String(flagBsvAlias, paymail.DefaultBsvAliasVersion, fmt.Sprintf("The %s version", flagBsvAlias))
	er(viper.BindPFlag(flagBsvAlias, rootCmd.PersistentFlags().Lookup(flagBsvAlias)))
}

// er is a basic helper method to catch errors loading the application
func er(err error) {
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" {

		chalker.Log(chalker.INFO, fmt.Sprintf("Loading custom configuration file: %s...", configFile))

		// Use config file from the flag
		viper.SetConfigFile(configFile)
	} else {

		// Find home directory
		home, err := homedir.Dir()
		er(err)

		// Set the path
		path := filepath.Join(home, applicationName)

		// Make a dummy file if it doesn't exist
		var file *os.File
		file, err = os.OpenFile(filepath.Join(path, configFileDefault+".yaml"), os.O_RDONLY|os.O_CREATE, 0644)
		er(err)
		_ = file.Close() // Error is not needed here, just close and continue

		// Search config in home directory with name "." (without extension)
		viper.AddConfigPath(path)
		viper.SetConfigName(configFileDefault)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("Error reading config file: %s", err.Error()))
	}

	// chalker.Log(chalker.INFO, fmt.Sprintf("...loaded config file: %s", viper.ConfigFileUsed()))
}
