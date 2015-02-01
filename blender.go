package main

import (
	"math/rand"
	"time"
	"fmt"
	"strconv"
)

type BlenderCommand struct {
	Type string
	Src *FrameSource
	Data map[string]interface{}
}

// A Blender manages FrameSources and creates a ready to send byte array of resulting rendering
type Blender struct {
	broker *Broker
	sources map[int] *FrameSource 	// a map of all available sources
	active []*FrameSource		// the sources this blender has selected to be active in the order they should be blended together

	joining chan *FrameSource 		// a channel for adding sources
	leaving chan *FrameSource 		// a channel for removing sources

	data []byte						// the rendered sources

	commands chan BlenderCommand 	// a command channel, used to select sources, overlay sources, trigger redraws etc
}


// Start managing client connections and message broadcasts.
func (b *Blender) Start() {
	// initialize random seed
	rand.Seed(time.Now().UnixNano())

	go func() {
		for {
			select {
			case cmd := <-b.commands:
				// test for a command we know about
				switch cmd.Type {
				case "overlay":
					fmt.Println("Source " + cmd.Src.ID + " requested overlay")
					// read out duration parameter from data object
				default:
					fmt.Println("Unknown blender command : ", cmd.Type)
				}
			case s := <-b.joining:
				// tell the source about our command channel
				s.commands = b.commands
				// store the source
				b.sources[s.ID] = s

				fmt.Println("Adding source ", s.ID)
				b.broker.messages <- &Message{ID: strconv.Itoa(s.ID), Type: "add-source"}
				// if length is now 1, do a selectsource
				//if len(b.sources) == 1 {
					b.SelectSources()
				//}
			case s := <-b.leaving:
				// release the command channel
				s.commands = nil

				fmt.Println("Removing source")
				// delete the source from our map
				delete(b.sources, s.ID)
				b.broker.messages <- &Message{ID: strconv.Itoa(s.ID), Type: "del-source"}

				// find the source in the active list and delete it
				for p, v := range b.active {
					if v == s {
						b.active = append(b.active[:p], b.active[p+1:]...)
						// trigger a source selection
						b.SelectSources()
						break
					}
				}
			default:
				// do nothing
				time.Sleep(5*time.Millisecond)
			}
		}
	}()
}

func (b *Blender) Redraw() {
	// blend all active sources and write into data array for pushing out
	if len(b.active) == 0 {
		return // leave as is
	}
	if b.active[0].current == nil {
		return
	}

	// wipe the data array and then start blending through the active sources
	copy(b.data, b.active[0].current)

	if (len(b.active) > 1) {
		// loop over each source and blend the pixels into the b.data array
		for i := 0; i < len(b.data); i++ {
			v := b.data[i]
			for s:=1; s < len(b.active); s++ {
				if b.active[s].current == nil {
					continue
				}

				// TODO: this is a simple additive blend right now
				// but we need to use the blend mode specified by each source
				// so break this out into a switch on src.blendMode
				if v + b.active[s].current[i] > 255 {
					v = 255
					break
				}
				v += b.active[s].current[i]
			}
			b.data[i] = v
		}
		
	}
}
func (b *Blender) RefreshSources(dst chan *Message) {
	for _, src := range b.sources {
		b.broker.messages <- &Message{ID: strconv.Itoa(src.ID), Type: "add-source", Dest: dst}
	}
}
func (b *Blender) SelectSources() {
	// select a new source
	switch len(b.sources) {
	case 0:
		// no sources available, show placeholder graphics from a dummy source?
		b.active = nil
	case 1:
		// only one source available, use it
		b.active = nil
		for _,src := range b.sources {
			b.active = append(b.active, src)
			fmt.Println("Source ", src.ID, " active")
			break
		}
	default:
		// many sources available.....how do we select one?
		for _, src := range b.sources {
			b.active = append(b.active, src)
		}
		/*i := int(float32(len(b.sources)) * rand.Float32())
		fmt.Println("Source ", i, " active")
		for k, _ := range b.sources {
			if i == 0 {
				// found random key, make it active
				b.active = append(b.active, k)
			} else {
				i--
			}
		}*/
	}
}

func NewBlender(numLEDs int, broker *Broker) *Blender {
	v := &Blender{
		broker: broker,
		sources: make(map[int] *FrameSource),
		active: make([]*FrameSource, 0, 50),
		joining: make(chan *FrameSource),
		leaving: make(chan *FrameSource),
		data: make([]byte, numLEDs*3),
		commands: make(chan BlenderCommand, 60), // buffered channel
	}
	return v;
}