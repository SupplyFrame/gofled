// +build windows darwin

package main

import (
	"fmt"
	"time"
)
func monitor_switch(cmds chan BlenderCommand) {
	// monitor gpio pin and turn off lights when gpio is off
	// this is the windows/osx implementation though so just do nothing.

	fmt.Println("Monitoring switch")

	defer func() {
		fmt.Println("Closing switch monitor")
	}()

	for {
		select {
			case <- quit: // test quit channel for a value, when closed this will return from this function immediately
				return
			default:
				// continue as normal
		}

		// test switch (windows/osx it always completes)
		time.Sleep(1*time.Second)
	}
}