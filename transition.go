package main

/*
	transitions blend from one frame buffer to another
	transitions have an amount value which is the position through the transition from src to dst
	transition types:
		crossfade - lerps between src to dst using linear interpolation of each value
		dissolve - swaps out random values from src to dst in hard cuts, until all values are moved
		wipe - slides in the new values from one side, pushing the old values out to the other
*/

import (
	"time"
	"math/rand"
	"log"
)

type TransitionInterface interface {
	Start(transition *Transition)
	Draw(dst []byte, oldSrc []byte, newSrc []byte, elapsed time.Duration) []byte
	Duration() time.Duration
}

type Transition struct {
	BlenderPtr *Blender
	NewSrc *FrameSource
	StartTime time.Time
	Worker TransitionInterface
}
func (self *Transition) Init(blender *Blender, newSrc *FrameSource, worker TransitionInterface) {
	self.BlenderPtr = blender
	self.NewSrc = newSrc
	self.StartTime = time.Now()
	self.Worker = worker

	self.Worker.Start(self)
}
func (self *Transition) Draw(dst []byte, src []byte) (bool) {
	elapsed := time.Since(self.StartTime)
	self.Worker.Draw(dst, src, self.NewSrc.current, elapsed)

	complete := false
	if elapsed >= self.Worker.Duration() {
		complete = true
	}
	return complete
}

func (self *Transition) Render(buffer []byte) {
	complete := self.Draw(buffer, self.BlenderPtr.primaryActive.current)

	// render other layers on top
	self.BlenderPtr.DrawActiveLayers(buffer)

	if complete {
		// remove ourself from the blender
		self.BlenderPtr.transition = nil
		self.BlenderPtr.primaryActive = self.NewSrc
		self.BlenderPtr = nil
		self.Worker = nil
	}
}

type CrossFadeTransition struct {	
	duration time.Duration
}
func (self *CrossFadeTransition) Duration() time.Duration {
	return self.duration
}
func (self *CrossFadeTransition) Start(_ *Transition) {
	log.Println("CrossFadeTransition:Start")
}
func (self *CrossFadeTransition) Draw(data []byte, oldSrc []byte, newSrc []byte, elapsed time.Duration) []byte {
	// blend between old src and new src based on elapsed time
	// return a byte array resulting from the blend

	percent := elapsed.Seconds() / self.Duration().Seconds()
	// blend over the new source based on percentage elapsed
	for i := 0; i < len(data); i++ {
		oldV := byte(float64(oldSrc[i]) * (1-percent))
		newV := byte(float64(newSrc[i]) * (percent))
		data[i] = oldV + newV
	}
	return data
}

type WipeDir int
const (
	WIPE_UP WipeDir = iota
	WIPE_DOWN
	WIPE_LEFT
	WIPE_RIGHT
)
type WipeTransition struct {
	duration time.Duration

	margin float64
	direction WipeDir
	percent float64
}
func (self *WipeTransition) Duration() time.Duration {
	return self.duration
}
func (self *WipeTransition) Start(t *Transition) {
	log.Println("WipeTweenTransition:Start")
	// initialize a tween, query the transition object to find out details
	self.percent = 0

	startPos := -self.margin
	endPos := 1.0
	if self.direction == WIPE_LEFT || self.direction == WIPE_UP {
		startPos = 1.0+self.margin
		endPos = 0
	}
	self.percent = startPos

	go tween(easeInOutQuad, 10*time.Millisecond, self.Duration(), startPos, endPos, &self.percent)
}
func (self *WipeTransition) Draw(data []byte, oldSrc []byte, newSrc []byte, elapsed time.Duration) []byte {
	// blend between old src and new src based on elapsed time
	// return a byte array resulting from the blend

	// blend over the new source based on percentage elapsed
	for i := 0; i < len(data); i++ {

		x, y := XY(i)

		crossover := 0.0

		switch self.direction {
		case WIPE_RIGHT:
			crossover = (float64(x)/float64(ledWidth)) - self.percent
		case WIPE_LEFT:
			crossover = (float64(x)/float64(ledWidth)) - (self.percent)
		case WIPE_UP:
			crossover =  (float64(y)/float64(ledHeight)) - (self.percent)
		case WIPE_DOWN:
			crossover =  (float64(y)/float64(ledHeight)) - self.percent
		}

		if self.direction == WIPE_LEFT || self.direction == WIPE_UP {
			if crossover > 0 {
				data[i] = newSrc[i]
			} else if crossover < -self.margin {
				data[i] = oldSrc[i]
			} else {
				// crossover = 0.0 -> 0.3
				// at 0.3 we want all old, at 0.0 we want all new
				percentOld := (crossover / -self.margin)
				data[i] = byte((float64(oldSrc[i]) * (percentOld)) + (float64(newSrc[i]) * (1-percentOld)))
			}
		} else {
			if crossover < 0 {
				data[i] = newSrc[i]
			} else if crossover > self.margin {
				data[i] = oldSrc[i]
			} else {
				// crossover = 0.0 -> 0.3
				// at 0.3 we want all old, at 0.0 we want all new
				percentOld := (crossover / self.margin)
				data[i] = byte((float64(oldSrc[i]) * (percentOld)) + (float64(newSrc[i]) * (1-percentOld)))
			}
		}
		
	}
	return data
}

var TransitionTypes = []TransitionInterface {
	&WipeTransition{ duration: 2*time.Second, margin:0.1, direction:WIPE_LEFT },
	&WipeTransition{ duration: 2*time.Second, margin:0.1, direction:WIPE_UP },
	&WipeTransition{ duration: 2*time.Second, margin:0.1, direction:WIPE_DOWN },
	&WipeTransition{ duration: 2*time.Second, margin:0.1, direction:WIPE_RIGHT },
	&CrossFadeTransition{ duration: 3*time.Second },
}

func NewTransition(blender *Blender, newSrc *FrameSource) *Transition {
	// pick a random transition type
	r := rand.Intn(len(TransitionTypes))

	t := TransitionTypes[r]

	// create a transition
	transition := Transition{}
	transition.Init(blender, newSrc, t)

	return &transition
}