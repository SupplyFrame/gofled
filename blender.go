package main

import (
	"math/rand"
	"time"
	"fmt"
	"strconv"
	"github.com/fatih/structs"
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
	primaryActive *FrameSource 		// the primary active source, all other sources are blended on top
	active []*FrameSource			// the sources this blender has selected to be active in the order they should be blended together

	joining chan *FrameSource 		// a channel for adding sources
	leaving chan *FrameSource 		// a channel for removing sources

	transition *Transition

	data []byte						// the rendered sources

	commands chan BlenderCommand 	// a command channel, used to select sources, overlay sources, trigger redraws etc
}

// Start managing client connections and message broadcasts.
func (b *Blender) Start() {
	// initialize random seed
	rand.Seed(time.Now().UnixNano())

	go b.SourceSelector()

	go func() {
		for {
			select {
			case cmd := <-b.commands:
				// test for a command we know about
				switch cmd.Type {
				case "overlay":
					fmt.Println("Source ", cmd.Src.ID, " requested overlay")
					// read out duration parameter from data object
					duration := time.Duration(cmd.Data["duration"].(float64))
					go func() {
						fmt.Println("Activating overlay for src ", cmd.Src.ID)
						cmd.Src.amount = 0.0
						b.active = append(b.active, cmd.Src)

						// fade in
						tween(easeInQuad, 10*time.Millisecond, 500*time.Millisecond, 0.0, 1.0, &cmd.Src.amount)

						time.Sleep(duration)

						// fade out
						tween(easeOutQuad, 10*time.Millisecond, 500*time.Millisecond, 1.0, 0.0, &cmd.Src.amount)

						fmt.Println("Deactivating overlay for src ", cmd.Src.ID)
						for p, v := range b.active {
							if v == cmd.Src {
								b.active = append(b.active[:p], b.active[p+1:]...)
								break
							}
						}
					}()
				default:
					fmt.Println("Unknown blender command : ", cmd.Type)
				}
			case s := <-b.joining:
				// tell the source about our command channel
				s.commands = b.commands
				// store the source
				b.sources[s.ID] = s

				fmt.Println("Adding source ", s.ID)
				b.broker.messages <- &Message{ID: strconv.Itoa(s.ID), Type: "add-source", Data: structs.Map(s) }
				// if length is now 1, do a selectsource
				if len(b.sources) == 1 {
					b.RandomSource()
				}
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
						// trigger a source selection if we only have one source
						if len(b.sources) == 1 {
							b.RandomSource()
						}
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
	// check we have an active source
	if b.primaryActive == nil {
		return
	}
	if b.primaryActive.current == nil {
		return
	}
	if (b.transition != nil) {
		data := b.transition.Render()
		copy(b.data, data)
	} else {
		// copy active source over data array
		copy(b.data, b.primaryActive.current)
		// render other layers on top
		b.DrawActiveLayers(b.data)
	}
}

func (b *Blender) DrawActiveLayers(data []byte) {
	if (len(b.active) > 0) {
		// loop over each source and blend the pixels into the b.data array
		for i := 0; i < len(data); i++ {
			v := data[i]
			for s:=0; s < len(b.active); s++ {
				if b.active[s].current == nil {
					continue
				}
				f := getBlendFunc(b.active[s].blendMode)
				v = blend(f, v, b.active[s].current[i], b.active[s].amount)
			}
			data[i] = v
		}
	}
}

func (b *Blender) RefreshSources(dst chan *Message) {
	for _, src := range b.sources {
		b.broker.messages <- &Message{ID: strconv.Itoa(src.ID), Type: "add-source", Dest: dst, Data: structs.Map(src)}
	}
}

func (b *Blender) RandomSource() *FrameSource {
	// select a new random source

	currentSource := b.primaryActive

	// count number of active sources
	activeCount := 0
	for _,src := range b.sources {
		if src.active {
			activeCount++
		}
	}
	if activeCount == 0 {
		b.primaryActive = nil
		b.active = nil
		return nil
	}
	// use total count as a probability value so if we have 6 sources, we have a 1/6 chance of picking any specific source
	for i:=0; i < 50; i++ {
		target := rand.Intn(activeCount)
		matched := 0
		for _,src := range b.sources {
			// skip over inactive sources
			if !src.active {
				continue
			}
			if matched == target && src != currentSource {
				fmt.Println("Active source = ", src.ID)
				return src
			}
			matched++
		}
	}
	return currentSource
}

func (b *Blender) SourceSelector() {
	// run in a loop forever, select a source and run for 5 minutes before repeating
	for {
		src := b.RandomSource()
		if (b.primaryActive == nil) {
			b.primaryActive = src
		} else {
			// setup a transition from current active to new source
			b.transition = NewTransition(b, src)
		}
		time.Sleep(10*time.Second)
	}
}

func NewBlender(numLEDs int, broker *Broker) *Blender {
	v := &Blender{
		broker: broker,
		sources: make(map[int] *FrameSource),
		primaryActive: nil,
		active: make([]*FrameSource, 0, 50),
		joining: make(chan *FrameSource),
		leaving: make(chan *FrameSource),
		data: make([]byte, numLEDs*3),
		commands: make(chan BlenderCommand, 60), // buffered channel
	}
	return v;
}