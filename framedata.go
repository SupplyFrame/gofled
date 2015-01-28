package main
/*
import (
	"fmt"
	"time"
)*/

type FrameData struct {
	data 		[]byte
	id		int
	transition	bool
}

func NewFrameData(numLEDs int) FrameData {
	return FrameData{make([]byte, (numLEDs*3)), 0, false}
}