package main

/*
#include "lib/spislave.h"
#cgo LDFLAGS: -L. -lspislave -lprussdrv -lpthread
*/
import "C"

import "unsafe"
import "reflect"
import "time"
import "fmt"

func main() {
	fmt.Println("Initializing SPI")

	var sharedMem *C.uchar = C.spiinit()

	fmt.Println("SPI ready")
	length := 11
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(sharedMem)),
		Len: length,
		Cap: length,
	}
	data := *(*[]C.uchar)(unsafe.Pointer(&hdr))
	
	fmt.Println("Ready to send data")
	
	fmt.Println("PRU flag = ", data[6], " Ready flag = ", data[5])
	data[5] = 1 // ready!	
	go func() {
		for ; data[5] != 0; {
			fmt.Println("PRU flag = ", data[6], " Ready flag = ", data[5])
			time.Sleep(500*time.Millisecond)
		}
		C.spiclose()
	}()

	go func() {
		time.Sleep(5*time.Second)
		data[5] = 0
	}()
	time.Sleep(10*time.Second)
//	C.spiclose()
//	time.Sleep(5*time.Second)
}
