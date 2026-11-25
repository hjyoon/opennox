package legacy

/*
#include "GAME1_1.h"
int sub_427C80(int4*, int4*);
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func mapSegmentAxisOverlaps(a0, a1, b0, b1 int32) bool {
	if a1-a0 < 0 {
		a0, a1 = a1, a0
	}
	if b0-b1 > 0 {
		b0, b1 = b1, b0
	}
	return a1 >= b0 && b1 >= a0
}

func mapSegmentsIntersect(a, b *[4]int32) bool {
	if !mapSegmentAxisOverlaps(a[0], a[2], b[0], b[2]) ||
		!mapSegmentAxisOverlaps(a[1], a[3], b[1], b[3]) {
		return false
	}

	ax, ay := a[2]-a[0], a[3]-a[1]
	bx, by := b[0]-b[2], b[1]-b[3]
	dx, dy := a[0]-b[0], a[1]-b[1]
	det := bx*ay - ax*by
	first, second := by*dx-bx*dy, ax*dy-ay*dx
	lo, hi := min(0, det), max(0, det)
	return first >= lo && first <= hi && second >= lo && second <= hi
}

//export sub_427C80
func sub_427C80(first, second *C.int4) C.int {
	a, b := (*[4]int32)(unsafe.Pointer(first)), (*[4]int32)(unsafe.Pointer(second))
	return C.int(bool2int(mapSegmentsIntersect(a, b)))
}

func divFloat64TowardZero(a, b float64) float64 {
	value := float64(a / b)
	residual := math.FMA(-value, b, a)
	if residual != 0 && !math.IsNaN(residual) {
		oppositeSign := math.Signbit(residual) != math.Signbit(b)
		if value > 0 && oppositeSign || value < 0 && !oppositeSign {
			return math.Nextafter(value, 0)
		}
	}
	return value
}

func sqrtFloat64TowardZero(a float64) float64 {
	root := math.Sqrt(a)
	if math.FMA(root, root, -a) > 0 {
		return math.Nextafter(root, 0)
	}
	return root
}

func projectMapPointOntoSegment(point *[2]int32, segment *[4]int32, radius float32) bool {
	spill := func(a float64) float64 { return float64(storeFloat32TowardZero(a)) }

	x0, y0 := spill(float64(segment[0])), spill(float64(segment[1]))
	dx := spill(addFloat64TowardZero(float64(segment[2]), -x0))
	dyLive := addFloat64TowardZero(float64(segment[3]), -y0)
	dy := spill(dyLive)
	square := addFloat64TowardZero(mulFloat64TowardZero(dyLive, dy), mulFloat64TowardZero(dx, dx))
	length := spill(sqrtFloat64TowardZero(square))
	nx := spill(divFloat64TowardZero(dx, length))
	nyLive := divFloat64TowardZero(dy, length)
	ny := spill(nyLive)

	px := spill(addFloat64TowardZero(float64(point[0]), -x0))
	py := spill(addFloat64TowardZero(float64(point[1]), -y0))
	perp := addFloat64TowardZero(mulFloat64TowardZero(px, -nyLive), mulFloat64TowardZero(py, nx))
	if spill(math.Abs(perp)) > float64(radius) {
		return false
	}
	projection := addFloat64TowardZero(mulFloat64TowardZero(py, ny), mulFloat64TowardZero(px, nx))
	stored := spill(projection)
	if projection > length || !(stored >= 0) {
		return false
	}
	x := storeFloat32TowardZero(addFloat64TowardZero(mulFloat64TowardZero(stored, nx), x0))
	y := storeFloat32TowardZero(addFloat64TowardZero(mulFloat64TowardZero(stored, ny), y0))
	point[0], point[1] = server.X87TruncateSignedDWORD419A70(x), server.X87TruncateSignedDWORD419A70(y)
	return true
}

//export sub_427DF0
func sub_427DF0(point *C.int2, segment *C.int, radius C.float) C.int {
	xy := (*[2]int32)(unsafe.Pointer(point))
	line := (*[4]int32)(unsafe.Pointer(segment))
	return C.int(bool2int(projectMapPointOntoSegment(xy, line, float32(radius))))
}
