/*

The Blender selects from the various available sources
Depending on flags set on each source it will pick one or more sources and render them into the active buffer in a format ready for sending to the LEDs

We need to trigger this process everytime one of the selected sources is updated.


Each source is being filled by its socket or websocket client, everytime the socket receives some data, we add a FrameData object to the source
When the source has buffered enough frames it starts spitting them out as the current data object

When the Blender selects a source as one it will use it should add the 'notify' channel to the source.

	select {
		case <- notify:
			// one of our sources notified us to update, so reblend the sources
		default:
			// no activity, so do nothing
	}

	Then when the source receives data it should push onto the notify channel using a non blocking send.
	
	// notify any interested parties
	if notify != nil {
		select { 
			case notify <- true:
				// we've notified the channel that a source has updated
			default:
				// the notify channel is already alerted
		}
	}

	


*/
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
}

// A Blender manages FrameSources and creates a ready to send byte array of resulting rendering
type Blender struct {
	broker *Broker
	Sources map[int] *FrameSource 	// a map of all available sources
	active []*FrameSource		// the sources this blender has selected to be active in the order they should be blended together

	joining chan *FrameSource 		// a channel for adding sources
	leaving chan *FrameSource 		// a channel for removing sources

	Data []byte						// the rendered sources

	update chan bool				// blocking channel used to trigger updates
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
				default:
					fmt.Println("Unknown blender command : ", cmd.Type)
				}
			case s := <-b.joining:
				// tell the source about our command channel
				s.commands = b.commands
				// tell the source about our update channel (non blocking only)
				s.update = b.update
				// store the source
				b.Sources[s.ID] = s

				fmt.Println("Adding source ", s.ID)
				b.broker.messages <- &Message{ID: strconv.Itoa(s.ID), Type: "add-source"}
				// if length is now 1, do a selectsource
				//if len(b.Sources) == 1 {
					b.SelectSources()
				//}
			case s := <-b.leaving:
				// release the commands and update channels
				s.commands = nil
				s.update = nil

				fmt.Println("Removing source")
				// delete the source from our map
				delete(b.Sources, s.ID)
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
	if b.active[0].Ready == false {
		return
	}

	// wipe the data array and then start blending through the active sources
	copy(b.Data, b.active[0].current)

	// use the blend mode specified by each source
	if (len(b.active) > 1) {
		// loop over each source and blend the pixels into the b.Data array
		for i := 0; i < len(b.Data); i++ {
			v := b.Data[i]
			for s:=1; s < len(b.active); s++ {
				if b.active[s].Ready==false {
					continue
				}
				if v + b.active[s].current[i] > 255 {
					v = 255
					break
				}
				v += b.active[s].current[i]
			}
			b.Data[i] = v
		}
		
	}
}
func (b *Blender) RefreshSources(dst chan *Message) {
	for _, src := range b.Sources {
		b.broker.messages <- &Message{ID: strconv.Itoa(src.ID), Type: "add-source", Dest: dst}
	}
}
func (b *Blender) SelectSources() {
	// select a new source
	switch len(b.Sources) {
	case 0:
		// no sources available, show placeholder graphics from a dummy source?
		b.active = nil
	case 1:
		// only one source available, use it
		b.active = nil
		for _,src := range b.Sources {
			b.active = append(b.active, src)
			fmt.Println("Source ", src.ID, " active")
			break
		}
	default:
		// many sources available.....how do we select one?
		for _, src := range b.Sources {
			b.active = append(b.active, src)
		}
		/*i := int(float32(len(b.Sources)) * rand.Float32())
		fmt.Println("Source ", i, " active")
		for k, _ := range b.Sources {
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
		Sources: make(map[int] *FrameSource),
		active: make([]*FrameSource, 0, 50),
		joining: make(chan *FrameSource),
		leaving: make(chan *FrameSource),
		Data: make([]byte, numLEDs*3),
		update: make(chan bool, 1), // non-buffered channel
		commands: make(chan BlenderCommand, 60), // buffered channel
	}
	return v;
}