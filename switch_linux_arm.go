// +build linux,arm

package main

import (
	"time"
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

func monitor_switch(cmds chan BlenderCommand) {
	// monitor gpio pin and turn off lights when gpio is off
	embd.InitGPIO()

	pin, err := embd.NewDigitalPin("GPIO_5")
	if err!=nil {
		fmt.Println("Error opening GPIO_5: ", err.Error())
		return
	}
	pin.SetDirection(embd.In)
	pin.PullUp()

	lightsOn := true

	fmt.Println("Monitoring switch")

	defer func() {
		fmt.Println("Closing switch monitor")
		pin.Close()
		embd.CloseGPIO()
	}()

	for {
		select {
			case <- quit: // test quit channel for a value, when closed this will return from this function immediately
				return
			default:
				// continue as normal
		}

		// if pin goes low send off state to system via channel
		v, err := pin.Read()

		if err!=nil {
			fmt.Println("Error reading GPIO_5: ", err.Error())
			time.Sleep(5*time.Second)
			continue
		}

		if v==0 && lightsOn != false {
			lightsOn = false
			// send message
			cmds <- BlenderCommand{ Type: "lights-off" }
		} else if v==1 && lightsOn != true {
			lightsOn = true
			// send message
			cmds <- BlenderCommand{ Type: "lights-on" }
		}

		time.Sleep(1*time.Second)
	}
}
