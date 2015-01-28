// +build linux,arm

package main

import (
	"time"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

func off(numLEDs int, startFrame []byte, endFrame []byte, spiBus embd.SPIBus) {
	if _, err := spiBus.Write(startFrame[:]); err != nil {
		panic(err)
	}

	for i := 0; i < numLEDs; i++ {
		// calculate color
		ledBuffer := [4]byte{0xE0, 0x00, 0x00, 0x00}

		if _, err := spiBus.Write(ledBuffer[:]); err != nil {
			panic(err)
		}

	}
	// send end frame (NumLEDs/2) bits
	for i := 0; i < numLEDs; i+= 16 {
		if _, err := spiBus.Write(endFrame[:]); err != nil {
		panic(err)
		}
	}
}

func Renderer(numLEDs int, blender *Blender) {
	// setup embd spi port
	if err := embd.InitSPI(); err != nil {
		panic(err)
	}
	defer embd.CloseSPI()


	spiBus := embd.NewSPIBus(embd.SPIMode0, 1, 4000000, 8, 0)
	defer spiBus.Close()

	fmt.Println("SPI initialized")

	// REG

	startFrame := make([]byte, 4)
	endFrame := make([]byte, int(math.Ceil(float64(numLEDs)/16.0)))
	for i:=0; i < len(endFrame); i++ {
		endFrame[i] = byte(0xFF)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for sig := range signals {
			fmt.Println("Received signal = ", sig)
			off(numLEDs, startFrame, endFrame, spiBus)
			spiBus.Close()
			os.Exit(1)
		}
	}()


	start := time.Now()
	frameCount := 0

	// send data in 4kb blocks
	// total length
	totalLen := len(startFrame) + len(endFrame) + (numLEDs*4)
	ledData := make([]byte, 4096)
	ledData2 := make([]byte, totalLen - 4096)

	// copy start and endframe into right place
	copy(ledData, startFrame)
	copy(ledData2[:(numLEDs*4)+len(startFrame) - 4096], endFrame)
	for {
		blender.Redraw()

		usedBytes := len(startFrame)

		buff := &ledData
		// go byte by byte, but ever 3rd byte we must inject an extra global brightness value
		for i := 0; i < len(blender.Data); i++ {
			if i%3 == 0 {
				(*buff)[usedBytes] = 0xFF	
				usedBytes++
			}
			(*buff)[usedBytes] = blender.Data[i]
			usedBytes++
			if usedBytes == len(*buff) {
				// switch to second buffer
				usedBytes = 0
				buff = &ledData2
			} else if (i == len(blender.Data)-1) {
				// last byte....now do end frame
				copy((*buff)[usedBytes:], endFrame)
				usedBytes += len(endFrame)
				break
				
			}
		}
		// now send it
		if _, err := spiBus.Write(ledData[:]); err != nil {
			panic(err)
		}
		if _, err := spiBus.Write(ledData2[:]); err != nil {
			panic(err)
		}
		// start of frame....copy startFrame into ledData buffer
		//dataStartTime := time.Now()
		/*copy(ledData, startFrame)
		usedBytes := len(startFrame)
		// go byte by byte, but ever 3rd byte we must inject an extra global brightness value
		for i := 0; i < len(blender.Data); i++ {
			if i%3 == 0 {
				ledData[usedBytes] = 0xFF	
				usedBytes++
			}
			ledData[usedBytes] = blender.Data[i]
			usedBytes++
			if usedBytes == len(ledData) {
				// send what we have
				if _, err := spiBus.Write(ledData[:]); err != nil {
					panic(err)
				}
				usedBytes = 0
			} else if (i == len(blender.Data)-1) {
				// last byte....now do end frame
				copy(ledData[usedBytes:], endFrame)
				usedBytes += len(endFrame)
				// now send it
				if _, err := spiBus.Write(ledData[0:usedBytes]); err != nil {
					panic(err)
				}
			}
		}*/
		//dataElapsed := time.Since(dataStartTime)
		//fmt.Println(fmt.Println("Time:", dataElapsed.Seconds()))

		frameCount++

		if frameCount >= 60 {
			elapsed := time.Since(start)
			start = time.Now()
			// compute average
			fps := 1.0 / (elapsed.Seconds() / float64(frameCount))
			fmt.Println("FPS : ", fps)
			frameCount = 0
		}
		// yield to other processes
		runtime.Gosched()
	}
}