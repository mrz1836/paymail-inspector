/*
Package cmd is all the available commands for the CLI application
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// The config file if used
var (
	configFile      string
	bsvAliasVersion string
)

// Defaults for the application
const (
	configDefault     = "paymail-inspector"  // Config file and application name
	defaultDomainName = "moneybutton.com"    // Used in examples
	defaultNameServer = "8.8.8.8"            // Default DNS NameServer
	logPrefix         = "paymail-inspector:" // Prefix for the logs in the CLI application output
)

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	Use:     configDefault,
	Short:   "Inspect, validate or resolve paymail domains and addresses",
	Long:    `This CLI tool can help you inspect, validate or resolve a paymail domain/address`,
	Version: "0.0.4",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	// Load the configuration
	cobra.OnInitialize(initConfig)

	// Add config option
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/."+configDefault+".yaml)")

	// Add a bsvalias version to target
	rootCmd.PersistentFlags().StringVar(&bsvAliasVersion, "bsvalias", paymail.DefaultBsvAliasVersion, "The bsvalias version (default is "+paymail.DefaultBsvAliasVersion+")")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" {

		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {

		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name "."+configDefault (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("." + configDefault)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
