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
//	"runtime"
	"unsafe"
	"reflect"
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
	
	start := time.Now()
	frameCount := 0

	for {
		blender.Redraw()

		//copy(data[11:], []C.uchar(blender.Data))
		for i:=0; i < len(blender.Data); i++ {
			data[11+i] = C.uchar(blender.Data[i])
		}
		data[5] = 1; // send!

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
		//runtime.Gosched()
		time.Sleep(10*time.Millisecond)
	}
}
