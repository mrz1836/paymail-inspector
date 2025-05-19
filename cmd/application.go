package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/integrations/baemail"
	"github.com/mrz1836/paymail-inspector/integrations/bitpic"
	"github.com/mrz1836/paymail-inspector/integrations/powping"
	"github.com/mrz1836/paymail-inspector/integrations/roundesk"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
	"github.com/tonicpow/go-paymail"
)

const versionPrefix = ": v"

// Core application loader (runs before every cmd)
func init() {

	// Set up the application resources
	setupAppResources()

	// Load the configuration
	cobra.OnInitialize(initConfig)

	// Set the user agent for the application's external integrations
	baemail.UserAgent = applicationFullName + versionPrefix + Version
	bitpic.UserAgent = applicationFullName + versionPrefix + Version
	powping.UserAgent = applicationFullName + versionPrefix + Version
	roundesk.UserAgent = applicationFullName + versionPrefix + Version

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
		log.Println("Error:", err.Error())
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set
func initConfig() {

	// Custom configuration file and location
	if configFile != "" {

		chalker.Log(chalker.INFO, fmt.Sprintf("Loading custom configuration file: %s...", configFile))

		// Use config file from the flag
		viper.SetConfigFile(configFile)
	} else {

		// Make a dummy file if it doesn't exist
		file, err := os.OpenFile(filepath.Join(applicationDirectory, configFileDefault+".yaml"), os.O_RDONLY|os.O_CREATE, 0600)
		er(err)
		_ = file.Close() // Error is not needed here, just close and continue

		// Search config in home directory with name "." (without extension)
		viper.AddConfigPath(applicationDirectory)
		viper.SetConfigName(configFileDefault)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("Error reading config file: %s", err.Error()))
	}

	// chalker.Log(chalker.INFO, fmt.Sprintf("...loaded config file: %s", viper.ConfigFileUsed()))
}

// generateDocumentation will generate all documentation about each command
func generateDocumentation() {

	// Replace the colorful logs in terminal (displays in Cobra docs) (color numbers generated)
	replacer := strings.NewReplacer("[32m", "```", "[33m", "```\n", "[39m", "", "[22m", "", "[36m", "", "[1m", "", "[40m", "", "[49m", "", "\u001B", "", "[0m", "")
	rootCmd.Long = replacer.Replace(rootCmd.Long)

	// Loop all command, adjust the Long description, re-add command
	for _, command := range rootCmd.Commands() {
		rootCmd.RemoveCommand(command)
		command.Long = replacer.Replace(command.Long)
		rootCmd.AddCommand(command)
	}

	// Generate the Markdown docs
	if err := doc.GenMarkdownTree(rootCmd, docsLocation); err != nil {
		chalker.Log(chalker.ERROR, fmt.Sprintf("Error generating docs: %s", err.Error()))
		return
	}

	// Success
	chalker.Log(chalker.SUCCESS, fmt.Sprintf("Successfully generated documentation for %d commands", len(rootCmd.Commands())))
}

// setupAppResources will set up the local application directories
func setupAppResources() {

	// Find home directory
	home, err := homedir.Dir()
	er(err)

	// Set the path
	applicationDirectory = filepath.Join(home, applicationName)

	// Detect if we have a program folder (windows)
	_, err = os.Stat(applicationDirectory)
	if err != nil {
		// If it does not exist, make one!
		if os.IsNotExist(err) {
			er(os.MkdirAll(applicationDirectory, os.ModePerm)) //nolint:gosec // G301 - running locally
		}
	}
}

// RandomHex returns a random hex string and error
func RandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
