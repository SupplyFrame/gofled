package main

import (
	"fmt"
	"net"
	"bytes"
	"encoding/json"
)

type Command uint32
const (
	CmdFrame		Command	= 0
	CmdTransition 	Command = 1
	CmdOverlay	 	Command = 2
	CmdBlendMode  	Command = 3
	CmdSetActive	Command = 4
	CmdMeta			Command = 5
	CmdClosing 		Command = 99
)

type FrameSource struct {
	// ID of the source, incremental just for internal use
	ID 			int
	// the current slice of frame data
	current		[]byte `json:"-"`

	// the src ip that is sending us this data
	clientIP	net.Addr

	// true when this source is available for general use
	active 		bool

	// channels for input commands
	commands	chan BlenderCommand `json:"-"`

	// meta parameters
	name 		string
	fps			int
	blendMode 	string
	author		string

	// blending parameters
	amount		float64 // amount of this channel that we blend in, 0.0 to 1.0
}

var nextSourceId = 1

func (src *FrameSource) AddFrame(frame []byte) {
	// store the new frame
	src.current = frame
}

func (src *FrameSource) StartTransition() {
	// indicates to the renderer that this source wants to render a transition in its frames
	fmt.Println("Start transition")
}

func (src *FrameSource) SetMeta(meta map[string] interface{}) {
	// pull out parameters from the map and write them into our framesource
	if val, ok := meta["name"]; ok {
		src.name = val.(string)
	}
	if val, ok := meta["fps"]; ok {
		src.fps = int(val.(float64))
	}
	if val, ok := meta["blendMode"]; ok {
		src.blendMode = val.(string)
	}
	if val, ok := meta["author"]; ok {
		src.author = val.(string)
	}
	if val, ok := meta["active"]; ok {
		src.active = val.(bool)
	}
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
		case CmdOverlay:
			// this source needs to be overlayed on top of all other sources immediately
			// data specifies duration for the overlay to be a priority, after which the source disappears again
			var overlaySettings map[string]interface{}
			if err := json.Unmarshal(data, &overlaySettings); err != nil {
				fmt.Println("Error unmarshalling overlay settings : ", err.Error())
				return
			}
			fmt.Println("Making src ", src.ID, " overlay for ", overlaySettings["duration"].(float64))

			// send command to blender through src.commands to notify it that this is now required for overlay
			// then setup a go func to wait duration seconds and then send another command to remove it
			src.commands <- BlenderCommand{ Src: src, Type: "overlay", Data: overlaySettings }
		case CmdBlendMode:
			// specifies the blend mode for this source
			n := bytes.Index(data, []byte{0})
			s := string(data[:n])
			switch s {
			case "add":
				src.blendMode = s
			}
		case CmdSetActive:
			// set the active flag for this source
			// read first byte in data slice as bool
			if data[0] > 0 {
				src.active = true
			} else {
				src.active = false
			}
		case CmdMeta:
			var meta map[string]interface{}
			if err := json.Unmarshal(data, &meta); err != nil {
				fmt.Println("Error unmarshalling meta data : ", err.Error())
				return
			}
			src.SetMeta(meta)
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

		clientIP: clientIP,

		active: true,
	}
	nextSourceId++
	return source
}