// +build windows darwin

package main

func monitor_switch(cmds chan BlenderCommand) {
	// monitor gpio pin and turn off lights when gpio is off
	// this is the windows/osx implementation though so just do nothing.
}