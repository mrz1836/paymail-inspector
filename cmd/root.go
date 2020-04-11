/*
Package cmd is all the available commands for the CLI application
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// The config file if used
var (
	configFile string
)

// Defaults for the application
const (
	configDefault     = "paymail-inspector" // Config file and application name
	defaultDomainName = "moneybutton.com"   // Used in examples
	defaultNameServer = "8.8.8.8"           // Default DNS NameServer
)

// These are keys for known flags that are used in the configuration
const (
	flagBsvAlias = "bsvalias"
)

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	Use:   configDefault,
	Short: "Inspect, validate or resolve paymail domains and addresses",
	Long: `
__________                             .__.__    .___                                     __                
\______   \_____  ___.__. _____ _____  |__|  |   |   | ____   ____________   ____   _____/  |_  ___________ 
 |     ___/\__  \<   |  |/     \\__  \ |  |  |   |   |/    \ /  ___/\____ \_/ __ \_/ ___\   __\/  _ \_  __ \
 |    |     / __ \\___  |  Y Y  \/ __ \|  |  |__ |   |   |  \\___ \ |  |_> >  ___/\  \___|  | (  <_> )  | \/
 |____|    (____  / ____|__|_|  (____  /__|____/ |___|___|  /____  >|   __/ \___  >\___  >__|  \____/|__|   
                \/\/          \/     \/                   \/     \/ |__|        \/     \/
Author: MrZ © 2020 github.com/mrz1836/paymail-validator

This CLI tool can help you inspect, validate or resolve a paymail domain/address.

Feel free to contribute via Github
`,
	Version: "0.0.10",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	er(rootCmd.Execute())
}

func init() {

	// Set chalker application prefix
	chalker.SetPrefix("paymail-inspector:")

	// Load the configuration
	cobra.OnInitialize(initConfig)

	// Add config option
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/."+configDefault+".yaml)")

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

		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {

		// Find home directory.
		home, err := homedir.Dir()
		er(err)

		// Search config in home directory with name "."+configDefault (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("." + configDefault)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		chalker.Log(chalker.INFO, fmt.Sprintf("loaded config file: %s", viper.ConfigFileUsed()))
	}
}
