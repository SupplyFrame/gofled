package main

import "math"

type blendFunc func(byte, byte) byte

func blend(f blendFunc, dst byte, src byte, amount float64) byte {
	srcVal := byte(float64(src) * amount)
	return f(dst, srcVal)	
}

func getBlendFunc(name string) blendFunc {
	switch name {
	case "add":
		return blendAdd
	case "darken":
		return blendDarken
	case "multiply":
		return blendMultiply
	case "colorBurn":
		return blendColorBurn
	case "linearBurn":
		return blendLinearBurn
	case "screen":
		return blendScreen
	case "colorDodge":
		return blendColorDodge
	case "overlay":
		return blendOverlay
	case "difference":
		return blendDifference
	case "exclusion":
		return blendExclusion
	case "subtract":
		return blendSubtract
	}
	return blendAdd
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