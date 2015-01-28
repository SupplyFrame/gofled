// +build windows darwin

package main

import (
	"time"
	"fmt"
)

func Renderer(numLEDs int, blender *Blender) {

	start := time.Now()
	frameCount := 0
	// at whatever rate we can achieve, render the current frame in src
	for {
		blender.Redraw()
		// TODO: Modify this to send the current frame out over SPI using embd
		//fmt.Println("Frame = ", ActiveSource.current)
		// TODO: instead of sleep we might want to just call the scheduler so we yield to the other processes
		frameCount++

		if frameCount >= 60 {
			elapsed := time.Since(start)
			start = time.Now()
			// compute average
			fps := 1.0 / (elapsed.Seconds() / float64(frameCount))
			fmt.Println("FPS : ", fps)
			frameCount = 0
		}
		time.Sleep(16*time.Millisecond)
	}
}