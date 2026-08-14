package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const mapPointInBoxScaleBits57B850 = uint32(0x3f350481)

// Keep every GAME.EXE x87 arithmetic instruction at an explicit binary64
// boundary. The Win32 executable uses the x87 53-bit precision mode here;
// separate calls also prevent contraction or reassociation on other ISAs.
//
//go:noinline
func mapPointInBoxAdd64_57B850(a, b float64) float64 { return a + b }

//go:noinline
func mapPointInBoxSub64_57B850(a, b float64) float64 { return a - b }

//go:noinline
func mapPointInBoxMul64_57B850(a, b float64) float64 { return a * b }

// mapPointInBox57B850 preserves GAME.EXE 0057B850. The left-top/left-bottom
// and right-top/right-bottom corner sums are spilled to binary32; the second
// diagonal keeps both source-relative corner sums in x87 precision. The first
// and fourth comparisons test C0 only, so unordered values pass those gates,
// while the middle comparisons require an ordered, strictly positive result.
func mapPointInBox57B850(source types.Pointf, shape *Shape, target types.Pointf) bool {
	scale := float64(math.Float32frombits(mapPointInBoxScaleBits57B850))

	leftTop := float32(mapPointInBoxAdd64_57B850(
		float64(shape.Box.LeftTop), float64(source.X),
	))
	leftBottom := float32(mapPointInBoxAdd64_57B850(
		float64(shape.Box.LeftBottom), float64(source.Y),
	))
	leftBottom2 := mapPointInBoxAdd64_57B850(
		float64(shape.Box.LeftBottom2), float64(source.X),
	)
	leftTop2 := mapPointInBoxAdd64_57B850(
		float64(shape.Box.LeftTop2), float64(source.Y),
	)
	rightTop := float32(mapPointInBoxAdd64_57B850(
		float64(shape.Box.RightTop), float64(source.X),
	))
	rightBottom := float32(mapPointInBoxAdd64_57B850(
		float64(shape.Box.RightBottom), float64(source.Y),
	))

	first := mapPointInBoxMul64_57B850(
		mapPointInBoxSub64_57B850(
			mapPointInBoxAdd64_57B850(
				mapPointInBoxSub64_57B850(float64(leftBottom), float64(leftTop)),
				float64(target.X),
			),
			float64(target.Y),
		),
		scale,
	)
	if first >= 0 { // unordered sets C0 and therefore continues
		return false
	}

	second := mapPointInBoxMul64_57B850(
		mapPointInBoxSub64_57B850(
			mapPointInBoxAdd64_57B850(
				mapPointInBoxSub64_57B850(leftTop2, leftBottom2),
				float64(target.X),
			),
			float64(target.Y),
		),
		scale,
	)
	if !(second > 0) { // zero and unordered both fail C0|C3
		return false
	}

	third := mapPointInBoxMul64_57B850(
		mapPointInBoxSub64_57B850(
			mapPointInBoxSub64_57B850(
				mapPointInBoxAdd64_57B850(float64(rightBottom), float64(rightTop)),
				float64(target.X),
			),
			float64(target.Y),
		),
		scale,
	)
	if !(third > 0) { // zero and unordered both fail C0|C3
		return false
	}

	fourth := mapPointInBoxMul64_57B850(
		mapPointInBoxSub64_57B850(
			mapPointInBoxSub64_57B850(
				mapPointInBoxAdd64_57B850(float64(leftBottom), float64(leftTop)),
				float64(target.X),
			),
			float64(target.Y),
		),
		scale,
	)
	return !(fourth >= 0) // unordered sets C0 and therefore succeeds
}

// MapPointInBox57B850 exposes the native-pointer form used by the CGo bridge.
func MapPointInBox57B850(source *types.Pointf, shape *Shape, target *types.Pointf) bool {
	return mapPointInBox57B850(*source, shape, *target)
}
