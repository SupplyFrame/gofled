// +build linux,arm

package main


import (
	"time"
	"fmt"
	"os"
	"runtime"
	"github.com/kellydunn/go-opc"
)


var opcClient *opc.Client

func Renderer(numLEDs int, blender *Blender) {
	// setup a client
	opcClient = opc.NewClient()

	err := opcClient.Connect("tcp", "localhost:7890")

	if err != nil {
		fmt.Println("Error while connecting to OPC server at localhost:7890 : ", err.Error())
		os.Exit(1)
	}

	fmt.Println("Setting up ", ledWidth*ledHeight, " leds, in ", ledWidth, "x", ledHeight)


	start := time.Now()
	frameCount := 0

	for {
		blender.Redraw()
		// construct a message
		m := opc.NewMessage(0)
		m.SetLength(uint16(ledHeight*ledWidth)*3)
		buffer := blender.GetBuffer()
		
		// fill with data in serpentine fashion
		for y := 0; y < ledHeight; y++ {
			for x := 0; x < ledWidth; x++ {
				ledPos := 0
				i := y*ledWidth + x
				dataPos := ((ledHeight -1 - y)*ledWidth + x) * 3
				if y%2==0 {
					ledPos = i
				} else {
					ledPos = i
					dataPos = ((ledHeight - 1 - y)*ledWidth + (ledWidth-1) - x) * 3
				}
				brightness := blender.brightness
				
				// scale brightness to our preferred range
				// 0.20 is lowest, 0.75 is highest
				//brightness = 0.2 + brightness*0.55

				m.SetPixelColor(ledPos,
					byte(float64(buffer[dataPos])*brightness),
					byte(float64(buffer[dataPos+1])*brightness), 
					byte(float64(buffer[dataPos+2])*brightness))
			}
		}
		
		opcClient.Send(m)

		frameCount++
		if frameCount >= 60 {
			elapsed := time.Since(start)
			start = time.Now()
			fps := 1.0 / (elapsed.Seconds() / float64(frameCount))
			fmt.Println("FPS : ", fps)
			frameCount = 0
		}
		runtime.Gosched()
		time.Sleep(16*time.Millisecond)
	}
}
