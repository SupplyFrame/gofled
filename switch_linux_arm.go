// +build linux,arm

package main

import (
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

func monitor_switch(cmds chan BlenderCommand) {
	// monitor gpio pin and turn off lights when gpio is off
	embd.InitGPIO()
	defer embd.CloseGPIO()

	pin, _ := embd.NewDigitalPin("GPIO_04")
	pin.SetDirection(embd.In)
	pin.PullUp()

	lightsOn := true

	for {
		// if pin goes low send off state to system via channel
		v, err := pin.Read()
		if v==0 && lightsOn != false {
			lightsOn = false
			// send message
			src.commands <- BlenderCommand{ Type: "lights-off" }
		} else if lightsOn != true {
			lightsOn = true
			// send message
			src.commands <- BlenderCommand{ Type: "lights-on" }
		}
		// if it goes high send on state to system via channel
	}
}