/*
Package cmd is all the available commands for the CLI application
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/mrz1836/paymail-inspector/bitpic"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/mrz1836/paymail-inspector/roundesk"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

// Default flag values for various commands
var (
	amount             uint64
	brfcAuthor         string
	brfcTitle          string
	brfcVersion        string
	configFile         string
	generateDocs       bool
	nameServer         string
	port               int
	priority           int
	protocol           string
	purpose            string
	satoshis           uint64
	serviceName        string
	signature          string
	skipBitpic         bool
	skipBrfcValidation bool
	skipDnsCheck       bool
	skipPki            bool
	skipPublicProfile  bool
	skipRoundesk       bool
	skipSrvCheck       bool
	skipSSLCheck       bool
	skipTracing        bool
	weight             int
)

// Defaults for the application
const (
	applicationName   = "paymail"           // Application name (binary)
	configFileDefault = "paymail-inspector" // Config file and application name
	defaultDomainName = "moneybutton.com"   // Used in examples
	defaultNameServer = "8.8.8.8"           // Default DNS NameServer
	docsLocation      = "docs/commands"     // Default location for command documentation
	flagBsvAlias      = "bsvalias"          // Flag for a known, common key
	flagSenderHandle  = "sender-handle"
	flagSenderName    = "sender-name"
)

// Version is set manually (also make:build overwrites this value from Github's latest tag)
var Version = "v0.1.2"

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
` + chalk.Yellow.Color("Author: MrZ © 2020 github.com/mrz1836/"+configFileDefault) + `

This CLI app is used for interacting with paymail service providers.

Help contribute via Github!
`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

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
}

func init() {

	// Load the configuration
	cobra.OnInitialize(initConfig)

	// Set the user agent for the application's external integrations
	bitpic.UserAgent = configFileDefault + ": v" + Version
	paymail.UserAgent = configFileDefault + ": v" + Version
	roundesk.UserAgent = configFileDefault + ": v" + Version

	// Add config option
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file (default is $HOME/."+configFileDefault+".yaml)")

	// Add document generation for all commands
	rootCmd.PersistentFlags().BoolVar(&generateDocs, "docs", false, "Generate docs from all commands (./"+docsLocation+")")

	// Add a toggle for request tracing
	rootCmd.PersistentFlags().BoolVarP(&skipTracing, "skip-tracing", "t", false, "Turn off request tracing information")

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

		// Use config file from the flag
		viper.SetConfigFile(configFile)
	} else {

		// Find home directory
		home, err := homedir.Dir()
		er(err)

		// Search config in home directory with name "." (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigName("." + configFileDefault)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		chalker.Log(chalker.INFO, fmt.Sprintf("...loaded config file: %s", viper.ConfigFileUsed()))
	}
}
