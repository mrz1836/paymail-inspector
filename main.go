/*
Paymail Inspector

Author: MrZ © 2020 github.com/mrz1836/paymail-inspector

This CLI tool can help you inspect, validate or resolve a paymail domain/address.

Help contribute via Github!
*/
package main

import (
	"github.com/mrz1836/paymail-inspector/cmd"
)

// main will load the all the commands and kick-start the application
func main() {
	cmd.Execute()
}
