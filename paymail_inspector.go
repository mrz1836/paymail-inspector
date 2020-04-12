/*
__________                             .__.__    .___                                     __
\______   \_____  ___.__. _____ _____  |__|  |   |   | ____   ____________   ____   _____/  |_  ___________
 |     ___/\__  \<   |  |/     \\__  \ |  |  |   |   |/    \ /  ___/\____ \_/ __ \_/ ___\   __\/  _ \_  __ \
 |    |     / __ \\___  |  Y Y  \/ __ \|  |  |__ |   |   |  \\___ \ |  |_> >  ___/\  \___|  | (  <_> )  | \/
 |____|    (____  / ____|__|_|  (____  /__|____/ |___|___|  /____  >|   __/ \___  >\___  >__|  \____/|__|
                \/\/          \/     \/                   \/     \/ |__|        \/     \/
Author: MrZ © 2020 github.com/mrz1836/paymail-inspector

This CLI tool can help you inspect, validate or resolve a paymail domain/address.

Help contribute via Github!
*/
package main

import "github.com/mrz1836/paymail-inspector/cmd"

// main will load the all the commands and kick-start the application
func main() {
	cmd.Execute()
}
