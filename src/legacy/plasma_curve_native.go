package legacy

/*
#include "GAME1_1.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
extern uint32_t dword_5d4594_1316408, dword_5d4594_1316412, dword_5d4594_1313880;
extern uint32_t dword_587000_180476, dword_587000_180480;
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

type plasmaCurvePoint = C.int2

func plasmaCurveCoefficients(a, b, c, d plasmaCurvePoint) (x, y [4]int32) {
	matrix := memmap.PtrT[[4][4]int32](0x581450, 9872)
	for i, row := range matrix {
		x[i] = row[0]*int32(a.field_0) + row[1]*int32(b.field_0) + row[2]*int32(c.field_0) + row[3]*int32(d.field_0)
		y[i] = row[0]*int32(a.field_4) + row[1]*int32(b.field_4) + row[2]*int32(c.field_4) + row[3]*int32(d.field_4)
	}
	return x, y
}

func plasmaCurveForwardDeltas(steps C.int, x, y [4]int32) (dx, dy [4]float32) {
	step := 1.0 / float64(steps)
	square, cube := float32(step*step), float32(step*step*step)
	sixCube := float32(6 * float64(cube))
	*memmap.PtrFloat32(0x587000, 180484) = float32(step)
	// These two coefficients are also entries in the original PE32 matrix.
	C.dword_587000_180480 = C.uint32_t(math.Float32bits(square))
	*memmap.PtrFloat32(0x587000, 180480) = square
	C.dword_587000_180476 = C.uint32_t(math.Float32bits(cube))
	*memmap.PtrFloat32(0x587000, 180476) = cube
	*memmap.PtrFloat32(0x587000, 180496) = square + square
	*memmap.PtrFloat32(0x587000, 180492) = sixCube
	*memmap.PtrFloat32(0x587000, 180508) = sixCube

	matrix := memmap.PtrT[[4][4]float32](0x587000, 180460)
	for i, row := range matrix {
		dx[i] = float32(float64(x[0])*float64(row[0]) + float64(x[1])*float64(row[1]) + float64(x[2])*float64(row[2]) + float64(x[3])*float64(row[3]))
		dy[i] = float32(float64(float32(y[0]))*float64(row[0]) + float64(float32(y[1]))*float64(row[1]) + float64(y[2])*float64(row[2]) + float64(y[3])*float64(row[3]))
	}
	return dx, dy
}

func addPlasmaCurveLine(from, to C.int2) {
	nox_client_drawAddPoint_49F500(int32(from.field_0), int32(from.field_4))
	nox_client_drawAddPoint_49F500(int32(to.field_0), int32(to.field_4))
}

func plasmaCurveFloatToInt(value float32) C.int {
	if !(value >= -2147483648 && value < 2147483648) {
		return -2147483648
	}
	return C.int(value)
}

//export sub_4BEAD0
func sub_4BEAD0(a, b, c, d *C.int2, steps, thick C.int) {
	x, y := plasmaCurveCoefficients(*a, *b, *c, *d)
	dx, dy := plasmaCurveForwardDeltas(steps, x, y)
	from := *a
	for i := C.int(0); i < steps; i++ {
		for j := 0; j < 3; j++ {
			dx[j] += dx[j+1]
			dy[j] += dy[j+1]
		}
		to := C.int2{field_0: plasmaCurveFloatToInt(dx[0]), field_4: plasmaCurveFloatToInt(dy[0])}
		addPlasmaCurveLine(from, to)
		nox_client_drawSetColor_434460(memmap.Int32(0x5D4594, 1316980))
		nox_client_drawLineFromPoints_49E4B0()
		if thick != 0 {
			// 004BED3D compares the same absolute difference twice.
			for _, side := range [...]C.int{-1, 1} {
				addPlasmaCurveLine(C.int2{field_0: to.field_0 + side, field_4: to.field_4},
					C.int2{field_0: from.field_0 + side, field_4: from.field_4})
				nox_client_drawLineFromPoints_49E4B0()
			}
		}
		if memmap.Uint32(0x5D4594, 1316988) != 0 {
			addPlasmaCurveLine(from, to)
			nox_draw_set54RGB32_434040(memmap.Int32(0x5D4594, 1316984))
			sub_434080(memmap.Int32(0x5D4594, 1316992))
			sub_49E4F0(int32(memmap.Uint8(0x5D4594, 1316996)))
		}
		from = to
	}
}

func plasmaSplineSamples(a, b, c, d plasmaCurvePoint, steps C.int, phase float32, emit func(*plasmaCurvePoint, *plasmaCurvePoint)) {
	xc, yc := plasmaCurveCoefficients(a, b, c, d)
	var x, y [4]float32
	for i := range x {
		x[i], y[i] = float32(xc[i]), float32(yc[i])
	}
	step := 1.0 / float64(steps)
	offset := step * float64(phase)
	increment := float32(step + offset)
	lanePhase := float32(offset)
	var t float32
	from := a
	for i := C.int(0); i < steps; i++ {
		value := min(1.0, float64(increment)+float64(t))
		f := float32(value)
		f2 := float64(f) * float64(f)
		sq, cube := float32(f2), float32(f2*float64(f))
		coordinate := func(coef [4]float32) C.int {
			result := float64(coef[1])*float64(sq) + float64(coef[0])*float64(cube) +
				float64(coef[2])*float64(f) + float64(coef[3])
			return plasmaCurveFloatToInt(float32(result))
		}
		to := C.int2{field_0: coordinate(x), field_4: coordinate(y)}
		t = f - lanePhase
		emit(&from, &to)
		from = to
	}
}

//export sub_4BEDE0
func sub_4BEDE0(a, b, c, d *C.int2, steps C.int, phase C.float, callback C.uintptr_t, arg unsafe.Pointer) {
	plasmaSplineSamples(*a, *b, *c, *d, steps, float32(phase), func(from, to *C.int2) {
		ccall.CallVoidPtr3(unsafe.Pointer(uintptr(callback)), unsafe.Pointer(from), unsafe.Pointer(to), arg)
	})
}

//export sub_4BA670
func sub_4BA670(direction, x1, y1, x2, y2 C.int) {
	dx, dy := x2-x1, y2-y1
	distance := int32(approximateDistance32(int32(dx), int32(dy)))
	count := distance/40 + 1
	if distance/40+2 >= 30 {
		count = 28
	}
	C.dword_5d4594_1316408 = C.uint32_t(count)

	sinCos := memmap.PtrT[[2]float32](0x587000, uintptr(194136+8*int32(direction)))
	unitX, unitY := float64(sinCos[0]), float64(sinCos[1])
	dy32 := float32(dy)
	C.dword_5d4594_1313880 = C.uint32_t(math.Float32bits(dy32))
	*memmap.PtrFloat32(0x5D4594, 1313880) = dy32
	magnitude := float32(math.Sqrt(float64(dy)*float64(dy32)+float64(dx)*float64(dx)) + 0.0099999998)
	dirX := float32(float64(dx) / float64(magnitude))
	dirY := float64(dy32) / float64(magnitude)
	*memmap.PtrFloat32(0x5D4594, 1313876) = dirX
	C.dword_5d4594_1313880 = C.uint32_t(math.Float32bits(float32(dirY)))
	*memmap.PtrFloat32(0x5D4594, 1313880) = float32(dirY)
	dot := dirY*unitY + float64(dirX)*unitX
	if dot < 0 {
		dot *= 0.2
	}
	length := (1 - dot) * float64(distance) * 2.3
	tangentX, tangentY := float32(unitX*length), float32(unitY*length)
	*memmap.PtrFloat32(0x5D4594, 1313868) = tangentX
	*memmap.PtrFloat32(0x5D4594, 1313872) = tangentY

	a, b := C.int2{field_0: x1, field_4: y1}, C.int2{field_0: x2, field_4: y2}
	c := C.int2{field_0: plasmaCurveFloatToInt(tangentX), field_4: plasmaCurveFloatToInt(tangentY)}
	d := C.int2{field_0: dx, field_4: dy}
	paths := plasmaPaths()
	for lane := range paths {
		progress := memmap.PtrFloat32(0x5D4594, 1313856+4*uintptr(lane))
		delta := 0.25
		if lane == 0 {
			delta = 0.2
		}
		*progress = float32(float64(*progress) + delta)
		segments := C.int(C.dword_5d4594_1316408)
		if *progress >= 1 {
			points := &paths[lane]
			for i := int(segments) + 1; i > 0; i-- {
				points[i].motion = points[i-1].motion
			}
			*progress = 0
			points[0].motion.ticks = 0
		}
		C.dword_5d4594_1316412 = 0
		plasmaSplineSamples(a, b, c, d, segments, *progress, func(from, to *plasmaCurvePoint) {
			plasmaRememberSegment(from, to, lane)
		})
	}
}
