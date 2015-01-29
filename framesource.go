package main

import (
	"fmt"
	"net"
)

type Command uint32
const (
	CmdFrame		Command	= 0
	CmdTransition 	Command = 1
	CmdAttention 	Command = 2
	CmdClosing 		Command = 3
)

type FrameSource struct {
	ID 			int
	current		[]byte
	source		chan []byte
	clientIP	net.Addr
	Ready		bool
	update 		chan bool
	commands	chan BlenderCommand
}

var nextSourceId = 1

func (src *FrameSource) AddFrame(frame []byte) {
	select {
		case src.source <- frame: // pushed frame onto buffer since there's some space
			src.Ready = false
		default:
			// no space on the buffer
			// pop the first item off the buffer as current frame
			src.current = <- src.source
			//fmt.Println("Current frame updated to ", src.current)
			// updated current frame from buffer
			// now push the new frame onto the buffer, this should no longer block
			src.source <- frame
			src.Ready = true
			// trigger update
			if src.update != nil {
				select {
					case src.update <- true:
					default:
						// do nothing update already pending
				}
			}
	}	
}
func (src *FrameSource) StartTransition() {
	// indicates to the renderer that this source wants to render a transition in its frames
	fmt.Println("Start transition")
}

func (src *FrameSource) ParseCommand(cmd Command, data []byte) {
	// parse the type to determine what to do with the data which can be empty
	switch cmd {
		case CmdFrame:
			// its frame data, add it
			src.AddFrame(data)
		case CmdTransition:
			// source is requesting that we start a transition, byte data should be the name of a transition if specified at all
			if len(data) > 0 {
				// they specified the name of a transition, read it here as a string and look it up in our transition list
			}
			src.StartTransition()
		case CmdAttention:
			// this source needs attention, it wants to be promoted to the display
			// need some sort of timeout to avoid attention fighting
		case CmdClosing:
			// this source is about to end soon, use this to start a transition and tell the renderer to move to another source
		default:
			// unknown command
			fmt.Println("Unknown command : ", cmd, " : ", data)
	}
}

func NewFrameSource(numLEDs int, bufferLen int, clientIP net.Addr) *FrameSource {
	source := &FrameSource{
		ID: nextSourceId,
		current: make([]byte, numLEDs*3),
		source: make(chan []byte,bufferLen),
		clientIP: clientIP,
		Ready: false,
		update: nil,
		commands: nil,
	}
	nextSourceId++
	return source
}