// +build linux,arm

package main

/*
#include "lib/spislave.h"
#cgo LDFLAGS: -L. -lspislave -lprussdrv -lpthread
*/
import "C"

import (
	"time"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"unsafe"
	"reflect"
	"encoding/binary"
)

func Renderer(numLEDs int, blender *Blender) {
	// setup spi port
	fmt.Println("Initializing SPI")

	var sharedMem *C.uchar = C.spiinit()
	

	fmt.Println("SPI ready")
	length := 25000
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(sharedMem)),
		Len: length,
		Cap: length,
	}
	data := *(*[]C.uchar)(unsafe.Pointer(&hdr))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for sig := range signals {
			fmt.Println("Received signal = ", sig)
			C.spiclose()
			go func() {
				time.Sleep(3*time.Second)
				os.Exit(1)
			}()
		}
	}()

	fmt.Println("Data at PRU :", data[6])
	
	//start := time.Now()
	frameCount := 0

	// set number of leds....
	ledLenBuff := make([]byte, 4)
	binary.LittleEndian.PutUint32(ledLenBuff, uint32(numLEDs))
	// copy the bytes into the data array at position 7
	data[7] = C.uchar(ledLenBuff[0])
	data[8] = C.uchar(ledLenBuff[1])
	data[9] = C.uchar(ledLenBuff[2])
	data[10] = C.uchar(ledLenBuff[3])

	for {
		blender.Redraw()

		for y := 0; y < ledHeight; y++ {
			for x := 0; x < ledWidth; x++ {
				ledPos := 0
				i := y*(ledWidth*3)+x*3
				if y%2==0 {
					//even rows are normal address
					ledPos = i
				} else {
					// odd rows are reversed on x
					ledPos = y*(ledWidth*3) + ((ledWidth-1)*3) - x*3
				}
				data[11+ledPos] = C.uchar(blender.data[i])
				data[11+ledPos+1] = C.uchar(blender.data[i+1])
				data[11+ledPos+2] = C.uchar(blender.data[i+2])
			}
		}
		//copy(data[11:], []C.uchar(blender.Data))
		/*for i:=0; i < len(blender.data); i++ {
			// we need to reverse the ordering of led data every other row
			// so figure out our row position
			//y := i / ledWidth
			//x := i % ledWidth
			// then if we're on an even row
			//if y % 2 == 0 {
				data[11+i] = C.uchar(blender.data[i])
			//} else {
			//	data[11+y*(ledWidth*3)-x*3] = C.uchar(blender.data[i])
			//}
		}*/
		data[5] = 1; // send!

		frameCount++

		/*if frameCount >= 60 {
			elapsed := time.Since(start)
			start = time.Now()
			// compute average
			fps := 1.0 / (elapsed.Seconds() / float64(frameCount))
			fmt.Println("FPS : ", fps)
			frameCount = 0
		}*/
		// yield to other processes
		runtime.Gosched()
		time.Sleep(1*time.Microsecond)
	}
}
