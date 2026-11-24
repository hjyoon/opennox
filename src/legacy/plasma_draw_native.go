package legacy

/*
#include "GAME3_1.h"
#include "client__draw__fx.h"
#include "client__draw__plasma.h"
extern uint32_t dword_5d4594_1316408, dword_5d4594_1316412;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const plasmaDrawSource = `C:\NoxPost\src\client\Draw\Plasma.c`

type plasmaPointMotion struct {
	tangent      int32
	angle        int32
	velocity     int32
	acceleration int32
	ticks        int32
}

type plasmaPathPoint struct {
	x, y   int32
	motion plasmaPointMotion
}

// The three 30-point lanes occupy 28 bytes per point in the original data blob.
func plasmaPaths() *[3][30]plasmaPathPoint {
	return memmap.PtrT[[3][30]plasmaPathPoint](0x5D4594, 1313884)
}

func plasmaRandom(min, max, line int32) int32 {
	return nox_common_randomIntMinMax_415FF0(min, max, internCStr(plasmaDrawSource), line)
}

func initPlasmaPaths() {
	blue := uint32(nox_color_rgb_4344A0(40, 180, 255))
	white := uint32(nox_color_rgb_4344A0(255, 255, 255))
	*memmap.PtrT[[3]uint32](0x5D4594, 1313828) = [3]uint32{blue, white, blue}
	*memmap.PtrT[[3]uint8](0x5D4594, 1313840) = [3]uint8{16, 24, 16}
	*memmap.PtrT[[3]uint32](0x5D4594, 1313844) = [3]uint32{8, 12, 8}
	*memmap.PtrT[[3]float32](0x5D4594, 1313856) = [3]float32{}
	paths := plasmaPaths()
	for lane := range paths {
		for index := range paths[lane] {
			paths[lane][index].motion = plasmaPointMotion{}
		}
	}
	*memmap.PtrUint32(0x5D4594, 1316404) = 1
}

func (m *plasmaPointMotion) advance(lane, index int, direction int32) {
	m.angle += m.velocity
	if m.angle >= 256 {
		m.angle -= 256
	} else if m.angle < 0 {
		m.angle += 255
	}
	if index == 0 {
		m.angle = direction
	}
	limit := int32(25)
	if lane == 1 {
		limit = 10
	}
	m.velocity = min(limit, max(-limit, m.velocity+m.acceleration))
	m.ticks--
	if m.ticks > 0 {
		return
	}
	m.ticks = plasmaRandom(10, 90, 378)
	m.acceleration = plasmaRandom(4, 8, 382)
	if plasmaRandom(0, 1, 385) != 0 {
		m.acceleration = -m.acceleration
	}
	if lane == 1 {
		m.tangent = plasmaRandom(40, 50, 392)
	} else {
		m.tangent = plasmaRandom(80, 110, 395)
		if index < 4 && plasmaRandom(0, 100, 402) > 90 {
			m.tangent = plasmaRandom(150, 200, 407)
			m.ticks *= 2
		}
	}
}

func plasmaTangent(angle int32, magnitude float32) plasmaCurvePoint {
	sinCos := memmap.PtrT[[2]float32](0x587000, uintptr(194136+8*angle))
	return plasmaCurvePoint{
		field_0: plasmaCurveFloatToInt(magnitude * sinCos[0]),
		field_4: plasmaCurveFloatToInt(magnitude * sinCos[1]),
	}
}

//export sub_4BA230
func sub_4BA230(direction, fromX, fromY, toX, toY C.int) C.int {
	if memmap.Uint32(0x5D4594, 1316404) == 0 {
		initPlasmaPaths()
	}
	sub_4BA670(direction, fromX, fromY, toX, toY)
	count := int(C.int(C.dword_5d4594_1316408))
	paths := plasmaPaths()
	for lane := range paths {
		for index := 0; index <= count; index++ {
			paths[lane][index].motion.advance(lane, index, int32(direction))
		}
	}
	colors := memmap.PtrT[[3]uint32](0x5D4594, 1313828)
	intensities := memmap.PtrT[[3]uint32](0x5D4594, 1313844)
	widths := memmap.PtrT[[3]uint8](0x5D4594, 1313840)
	for lane := range paths {
		thick := C.int(bool2int(lane == 1))
		for index := 0; index < count; index++ {
			current, next := paths[lane][index], paths[lane][index+1]
			from := plasmaCurvePoint{field_0: C.int(current.x), field_4: C.int(current.y)}
			to := plasmaCurvePoint{field_0: C.int(next.x), field_4: C.int(next.y)}
			magnitude := float32(current.motion.tangent)
			startTangent := plasmaTangent(current.motion.angle, magnitude)
			endTangent := plasmaTangent(next.motion.angle, magnitude)
			sub_4BE800(C.int(colors[lane]))
			sub_4BE810(1, C.int(colors[lane]), C.int(intensities[lane]), C.char(widths[lane]))
			sub_4BEAD0(&from, &to, &startTangent, &endTangent, 8, thick)
		}
	}
	return C.int(count)
}

func plasmaRememberSegment(from, to *plasmaCurvePoint, lane int) C.char {
	view := GetClient().Viewport()
	spark := memmap.PtrUint32(0x5D4594, 1316416)
	if *spark == 0 {
		*spark = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("CyanSpark")))
	}
	index := int(C.dword_5d4594_1316412)
	points := &plasmaPaths()[lane]
	points[index].x, points[index].y = int32(from.field_0), int32(from.field_4)
	points[index+1].x, points[index+1].y = int32(to.field_0), int32(to.field_4)
	frame := byte(gameFrame())
	C.dword_5d4594_1316412++
	if byte(gameFrame())&4 == 0 {
		return C.char(frame)
	}
	roll := plasmaRandom(0, 10, 135)
	if roll <= 5 {
		return C.char(roll)
	}
	x := to.field_0 + C.int(view.World.Min.X) - C.int(view.Screen.Min.X)
	y := to.field_4 + C.int(view.World.Min.Y) - C.int(view.Screen.Min.Y)
	return C.char(nox_xxx_drawEnergyBolt_499710(x, y, 8, C.int(*spark)))
}

//export sub_4BA8B0
func sub_4BA8B0(from, to, lane *C.int) C.char {
	return plasmaRememberSegment((*plasmaCurvePoint)(unsafe.Pointer(from)), (*plasmaCurvePoint)(unsafe.Pointer(to)), int(*lane))
}

// The endpoint fields are unaligned inside the native drawable union.
func drawableRayPayload(d *client.Drawable) *[13]byte {
	return (*[13]byte)(drawableEffectBytes(d)[:13])
}

func drawableRayEndpoints(ray *[13]byte) (from, to C.int2, ok bool) {
	source := binary.LittleEndian.Uint32(ray[5:9])
	target := binary.LittleEndian.Uint32(ray[9:13])
	if ray[0] == 0 {
		return C.int2{field_0: C.int(uint16(source)), field_4: C.int(source >> 16)},
			C.int2{field_0: C.int(uint16(target)), field_4: C.int(target >> 16)}, true
	}
	objects := GetClient().Cli().Objs
	src, dst := objects.ByNetCode(uint16(source)), objects.ByNetCode(uint16(target))
	if src == nil || dst == nil {
		return from, to, false
	}
	p1, p2 := src.Pos(), dst.Pos()
	return C.int2{field_0: C.int(p1.X), field_4: C.int(p1.Y)},
		C.int2{field_0: C.int(p2.X), field_4: C.int(p2.Y)}, true
}

func drawableRayScreen(vp *noxrender.Viewport, p C.int2) C.int2 {
	return C.int2{
		field_0: C.int(vp.Screen.Min.X) + p.field_0 - C.int(vp.World.Min.X),
		field_4: C.int(vp.Screen.Min.Y) + p.field_4 - C.int(vp.World.Min.Y) - 20,
	}
}

//export nox_thing_plasma_draw
func nox_thing_plasma_draw(view *C.int, drawable *nox_drawable) C.int {
	vp := (*noxrender.Viewport)(unsafe.Pointer(view))
	ray := drawableRayPayload(asDrawable(drawable))
	from, to, ok := drawableRayEndpoints(ray)
	if !ok {
		return 1
	}
	p1, p2 := drawableRayScreen(vp, from), drawableRayScreen(vp, to)
	direction := C.int(ray[1])
	if ray[0] != 0 {
		mouse := GetClient().GetMousePos()
		direction = C.int(server.DirFromVec(types.Pointf{
			X: float32(C.int(mouse.X) - p1.field_0),
			Y: float32(C.int(mouse.Y) - p1.field_4),
		}))
	}
	sub_4BA230(direction, p1.field_0, p1.field_4, p2.field_0, p2.field_4)
	return 1
}
