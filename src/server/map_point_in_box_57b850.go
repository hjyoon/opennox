package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const mapPointInBoxScaleBits57B850 = uint32(0x3f350481)

// The second diagonal retains its two corner sums; the other diagonals
// reload binary32 stores. Preserve startup PC53/ToZero and unordered gates.
func mapPointInBox57B850(source types.Pointf, shape *Shape, target types.Pointf) bool {
	add, sub, mul := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	k := float64(math.Float32frombits(mapPointInBoxScaleBits57B850))
	lt := spill(add(float64(shape.Box.LeftTop), memory(source.X)))
	lby := spill(add(float64(shape.Box.LeftBottom), memory(source.Y)))
	lbx := add(float64(shape.Box.LeftBottom2), memory(source.X))
	lty := add(float64(shape.Box.LeftTop2), memory(source.Y))
	rtx := spill(add(float64(shape.Box.RightTop), memory(source.X)))
	rby := spill(add(float64(shape.Box.RightBottom), memory(source.Y)))
	first := mul(sub(add(sub(float64(lby), memory(lt)), memory(target.X)), memory(target.Y)), k)
	if first >= 0 {
		return false
	}
	second := mul(sub(add(sub(lty, lbx), memory(target.X)), memory(target.Y)), k)
	if !(second > 0) {
		return false
	}
	third := mul(sub(sub(add(float64(rby), memory(rtx)), memory(target.X)), memory(target.Y)), k)
	if !(third > 0) {
		return false
	}
	fourth := mul(sub(sub(add(float64(lby), memory(lt)), memory(target.X)), memory(target.Y)), k)
	return !(fourth >= 0)
}

func MapPointInBox57B850(source *types.Pointf, shape *Shape, target *types.Pointf) bool {
	return mapPointInBox57B850(*source, shape, *target)
}
