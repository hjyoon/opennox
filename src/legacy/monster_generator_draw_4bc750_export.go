package legacy

/*
#include "monster_generator_draw_4bc750.h"
#include "client__draw__mgendraw.h"
#include "client__draw__parse__parse.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

// The existing C parser allocates this native-aligned conditional animation
// record; it is not a packed PE32 record or a collection of uint32 pointers.
type monsterGeneratorDrawDataNative4BC750 struct {
	size   uint32
	images [5]*noxrender.ImageHandle
	count  [5]uint8
	delay  [5]uint8
	kind   [5]uint32
}

func monsterGeneratorDrawNativeHooks4BC750(vp *noxrender.Viewport, dr *client.Drawable) monsterGeneratorDrawHooks4BC750[*monsterGeneratorDrawDataNative4BC750, *noxrender.ImageHandle, noxrender.ImageHandle] {
	return monsterGeneratorDrawHooks4BC750[*monsterGeneratorDrawDataNative4BC750, *noxrender.ImageHandle, noxrender.ImageHandle]{
		flags: func() uint32 { return dr.Flags70Val },
		data: func() *monsterGeneratorDrawDataNative4BC750 {
			return (*monsterGeneratorDrawDataNative4BC750)(dr.DrawData)
		},
		images: func(d *monsterGeneratorDrawDataNative4BC750, state int) *noxrender.ImageHandle {
			return d.images[state]
		},
		count: func(d *monsterGeneratorDrawDataNative4BC750, state int) uint8 { return d.count[state] },
		kind:  func(d *monsterGeneratorDrawDataNative4BC750, state int) uint32 { return d.kind[state] },
		delay: func(d *monsterGeneratorDrawDataNative4BC750, state int) uint8 { return d.delay[state] },
		delayIndex: func(d *monsterGeneratorDrawDataNative4BC750, state int) int32 {
			return int32(uint32(uintptr(unsafe.Pointer(&d.delay[state]))))
		},
		netcode: func() uint32 { return dr.NetCode32 },
		frame:   gameFrame,
		slave:   func() uint32 { return dr.AnimFrameSlave },
		random: func(min, max int32, source string, line int32) int32 {
			return nox_common_randomIntMinMax_415FF0(min, max, internCStr(source), line)
		},
		class:    func() uint32 { return uint32(dr.ObjClass) },
		setClass: func(v uint32) { dr.ObjClass = object.Class(v) },
		objFlags: func() uint32 { return uint32(dr.ObjFlags) },
		setFlags: func(v uint32) { dr.ObjFlags = object.Flags(v) },
		timer:    func() uint32 { return dr.UnionEffect().Field_108 },
		setTimer: func(v uint32) { dr.UnionEffect().Field_108 = v },
		setState: func(v uint32) { dr.Flags70Val = v },
		image: func(table *noxrender.ImageHandle, frame int32) noxrender.ImageHandle {
			// Keep the actual indexed load and nil-image propagation. A new
			// metadata bound would suppress original random/slave accesses.
			return *(*noxrender.ImageHandle)(unsafe.Add(unsafe.Pointer(table), int(frame)*int(unsafe.Sizeof(noxrender.ImageHandle(nil)))))
		},
		draw: func(img noxrender.ImageHandle) { Nox_xxx_drawObject_4C4770_draw(vp, dr, img) },
	}
}

var monsterGeneratorDrawCall4BC750 = func(vp *noxrender.Viewport, dr *client.Drawable) int32 {
	return monsterGeneratorDraw4BC750(monsterGeneratorDrawNativeHooks4BC750(vp, dr))
}

// DrawMonsterGenerator4BC750 shares the complete body with the public C
// callback, including callers which bypass the normal world-draw dispatch.
func DrawMonsterGenerator4BC750(vp *noxrender.Viewport, dr *client.Drawable) int32 {
	return monsterGeneratorDrawCall4BC750(vp, dr)
}

//export nox_client_monsterGeneratorDraw_native_4BC750
func nox_client_monsterGeneratorDraw_native_4BC750(vp *C.nox_draw_viewport_t, dr *C.nox_drawable) C.int {
	return C.int(monsterGeneratorDrawCall4BC750((*noxrender.Viewport)(unsafe.Pointer(vp)), (*client.Drawable)(unsafe.Pointer(dr))))
}

func monsterGeneratorDrawCEntry4BC750(vp *noxrender.Viewport, dr *client.Drawable) int32 {
	return int32(C.nox_thing_monster_gen_draw((*C.int)(vp.C()), (*C.nox_drawable)(dr.C())))
}

func monsterGeneratorDrawCLayout4BC750() [6]uintptr {
	return [6]uintptr{C.sizeof_nox_cond_animate_draw_data_t,
		unsafe.Offsetof(C.nox_cond_animate_draw_data_t{}.size),
		unsafe.Offsetof(C.nox_cond_animate_draw_data_t{}.images),
		unsafe.Offsetof(C.nox_cond_animate_draw_data_t{}.frame_count),
		unsafe.Offsetof(C.nox_cond_animate_draw_data_t{}.frame_delay),
		unsafe.Offsetof(C.nox_cond_animate_draw_data_t{}.animation_kind)}
}
