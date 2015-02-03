package main

import "time"
import "math"

type tweenFunc func(time.Duration, time.Duration, float64, float64) float64

func tween(f tweenFunc, updateRate time.Duration, duration time.Duration, from float64, to float64, dest *float64) {
	start := time.Now()
	for {
		// calculate time since we started
		elapsed := time.Since(start)
		// clamp elapsed to duration
		if elapsed > duration {
			elapsed = duration
		}
		delta := to - from
		// do the tween and store the updated value
		val := f(elapsed, duration, from, delta)
		*dest = val
		
		if elapsed >= duration {
			break
		}
		
		// apply sleep
		time.Sleep(updateRate)
	}
}

func linearTween(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	return from+(delta*(elapsed.Seconds()/duration.Seconds()))
}

func easeInQuad(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	r := elapsed.Seconds() / duration.Seconds()
	return delta*r*r + from
}
func easeOutQuad(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	r := elapsed.Seconds() / duration.Seconds()
	return -delta * r*(r-2) + from
}
func easeInOutQuad(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	r := elapsed.Seconds() / (duration.Seconds()/2)
	if r < 1 {
		return delta/2*r*r + from
	}
	r--
	return -delta/2 * (r*(r-2) - 1) + from
}
func easeInCubic(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	r := elapsed.Seconds() / duration.Seconds()
	return delta*r*r*r + from
}
func easeOutCubic(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	r := elapsed.Seconds() / duration.Seconds()
	r--
	return delta*(r*r*r + 1) + from
}
func easeInOutCubic(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	r := elapsed.Seconds() / (duration.Seconds()/2)
	if r < 1 {
		return delta/2*r*r*r + from
	}
	r = r-2
	return delta/2 * (r*r*r + 2) + from
}

func easeInSine(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	return -delta * math.Cos(elapsed.Seconds()/duration.Seconds() * (math.Pi/2)) + delta + from
}
func easeOutSine(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	return delta * math.Sin(elapsed.Seconds()/duration.Seconds() * (math.Pi/2)) + from
}
func easeInOutSine(elapsed time.Duration, duration time.Duration, from float64, delta float64) float64 {
	return -delta/2 * (math.Cos(math.Pi * elapsed.Seconds()/duration.Seconds())-1) + from
}
