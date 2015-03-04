package main

import "math"

type blendFunc func(byte, byte) byte

func blend(f blendFunc, dst byte, src byte, amount float64) byte {
	v := f(dst, src)
	d := byte((float64(dst)*(1-amount)) + (float64(v)*amount))
	return d
}

var blendMap = map[string]blendFunc {
	"replace": blendReplace,
	"add": blendAdd,
	"darken": blendDarken,
	"multiply": blendMultiply,
	"colorBurn": blendColorBurn,
	"linearBurn": blendLinearBurn,
	"screen": blendScreen,
	"colorDodge": blendColorDodge,
	"overlay": blendOverlay,
	"difference": blendDifference,
	"exclusion": blendExclusion,
	"subtract": blendSubtract,
}

func validBlendFunc(name string) bool {
	if _, ok := blendMap[name]; ok {
		return ok
	}
	return false
}

func getBlendFunc(name string) blendFunc {
	if val, ok := blendMap[name]; ok {
		return val
	}
	return blendAdd
}

func blendReplace(dst byte, src byte) byte {
	if src == 0 {
		return dst
	} else {
		return src
	}
}

func blendAdd(dst byte, src byte) byte {
	dst += src
	if dst > 255 {
		dst = 255
	}
	return dst
}
func blendDarken(dst byte, src byte) byte {
	if src < dst {
		return src
	}
	return dst
}
func blendMultiply(dst byte, src byte) byte {
	return byte((uint32(src) * uint32(dst)) / 255)
}
func blendColorBurn(dst byte, src byte) byte {
	if src == 0 {
		return src
	}
	d := uint32(dst)
	s := uint32(src)
	val := 255 - (( 255 - d)* 255 / s)
	if val > 0 {
		return byte(val)
	}
	return 0
}
func blendLinearBurn(dst byte, src byte) byte {
	d := uint32(dst)
	s := uint32(src)
	if s+d < 255 {
		return 0
	}
	return byte(s + d - 255)
}
func blendScreen(dst byte, src byte) byte {
	d := uint32(dst)
	s := uint32(src)
	return byte(s + d - s*d/255)
}

func blendColorDodge(dst byte, src byte) byte {
	d := uint32(dst)
	s := uint32(src)
	if s == 255 {
		return byte(s)
	}
	val := (d * 255 / (255 - s))
	if 255 < val {
		return 255
	}
	return byte(val)
}

func blendOverlay(dst byte, src byte) byte {
	d := uint32(dst)
	s := uint32(src)
	if d < 128 {
		return byte(2 * s * d / 255)
	}
	return byte(255 - 2*(255-s)*(255-d)/255)
}
func blendDifference(dst byte, src byte) byte {
	d := float64(dst)
	s := float64(src)
	
	return byte(math.Abs(s - d))
}
func blendExclusion(dst byte, src byte) byte {
	d := float64(dst)
	s := float64(src)

	return byte(s + d - s*d/128)
}
func blendSubtract(dst byte, src byte) byte {
	d := float64(dst)
	s := float64(src)

	if d-s < 0.0 {
		return 0.0
	}
	return byte(d-s)
}