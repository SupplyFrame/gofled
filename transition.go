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
	"math"
)

type TransitionFunc func(oldSrc *FrameSource, newSrc *FrameSource, elapsed time.Duration, total time.Duration) []byte

type Transition struct {
	blender *Blender
	newSrc *FrameSource
	elapsed time.Duration
	start time.Time
	duration time.Duration
	transition TransitionFunc
}

func (t *Transition) Render() []byte {
	// update elapsed
	t.elapsed = time.Since(t.start)
	newData := t.transition(t.blender.primaryActive, t.newSrc, t.elapsed, t.duration)

	// render other layers on top
	t.blender.DrawActiveLayers(newData)

	if t.elapsed >= t.duration {
		// remove ourself from the blender
		t.blender.transition = nil
		t.blender.primaryActive = t.newSrc
	}
	return newData
}

func CrossFadeTransition(oldSrc *FrameSource, newSrc *FrameSource, elapsed time.Duration, total time.Duration) []byte {
	// blend between old src and new src based on elapsed time
	// return a byte array resulting from the blend
	data := make([]byte, len(oldSrc.current))

	percent := elapsed.Seconds() / total.Seconds()
	// blend over the new source based on percentage elapsed
	for i := 0; i < len(data); i++ {
		oldV := byte(float64(oldSrc.current[i]) * (1-percent))
		newV := byte(float64(newSrc.current[i]) * (percent))
		data[i] = oldV + newV
	}
	return data
}
func WipeTransition(oldSrc *FrameSource, newSrc *FrameSource, elapsed time.Duration, total time.Duration) []byte {
	// blend between old src and new src based on elapsed time
	// return a byte array resulting from the blend
	data := make([]byte, len(oldSrc.current))

	percent := elapsed.Seconds() / total.Seconds()
	crossover := int(math.Floor(float64(ledWidth) * percent))
	// blend over the new source based on percentage elapsed
	for i := 0; i < len(data); i++ {
		x := (i % ledWidth)
		if x > crossover {
			data[i] = oldSrc.current[i]
		} else {
			data[i] = newSrc.current[i]
		}
	}
	return data
}

func NewTransition(blender *Blender, newSrc *FrameSource, duration time.Duration) *Transition {
	return &Transition{
		blender: blender,
		newSrc: newSrc,
		duration: duration,
		elapsed: 0,
		start: time.Now(),
		transition: WipeTransition,
	}
}