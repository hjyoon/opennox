package opennox

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

type objectOutOfSightState48EA70 struct {
	Code uint16
}

type objectOutOfSightHooks48EA70 struct {
	connected       func() bool
	firstDrawable   func() *client.Drawable
	localDrawable   func() *client.Drawable
	deactivate      func(*client.Drawable)
	deleteDrawable  func(*client.Drawable)
	animateDrawFunc unsafe.Pointer
	animationKind   func(unsafe.Pointer) uint32
}

type objectInShadowsHooks48EA70 struct {
	connected       func() bool
	firstDrawable   func() *client.Drawable
	localDrawable   func() *client.Drawable
	fadeObjects     func() bool
	tickRate        func() uint32
	transparentFade func(*client.Drawable, int)
	deactivate      func(*client.Drawable)
	deleteDrawable  func(*client.Drawable)
	animateDrawFunc unsafe.Pointer
	animationKind   func(unsafe.Pointer) uint32
}

func decodeObjectOutOfSightState48EA70(data []byte) (objectOutOfSightState48EA70, bool) {
	if len(data) < 3 {
		return objectOutOfSightState48EA70{}, false
	}
	return objectOutOfSightState48EA70{Code: binary.LittleEndian.Uint16(data[1:3])}, true
}

// findObjectOutOfSightDrawable48EA70 mirrors the two GAME.EXE net-code lists:
// bit 15 selects the 0x20400000-class list, while the remaining bits are the ID.
func findObjectOutOfSightDrawable48EA70(first *client.Drawable, code uint16) *client.Drawable {
	wantStatic := nox_xxx_netTestHighBit_578B70(code)
	wantID := uint32(nox_xxx_netClearHighBit_578B30(code))
	for dr := first; dr != nil; dr = dr.NextPtr {
		isStatic := dr.Class()&0x20400000 != 0
		if isStatic == wantStatic && dr.NetCode32 == wantID {
			return dr
		}
	}
	return nil
}

func isOneShotObjectOutOfSightDrawable48EA70(dr *client.Drawable, animateDraw unsafe.Pointer) bool {
	if dr == nil {
		return false
	}
	return dr.DrawFuncPtr == animateDraw && dr.DrawData != nil && *(*uint32)(unsafe.Add(dr.DrawData, 12)) == 1
}

// A nil reader preserves the PE32 service contract used by the original
// branch fixtures. Native packet adapters always provide the compiler-aligned
// parser layout; they must not interpret the high dword of images as kind.
func isOneShotObjectOutOfSightWithReader48EA70(dr *client.Drawable, animateDraw unsafe.Pointer, readKind func(unsafe.Pointer) uint32) bool {
	if readKind == nil {
		return isOneShotObjectOutOfSightDrawable48EA70(dr, animateDraw)
	}
	return dr != nil && dr.DrawFuncPtr == animateDraw && dr.DrawData != nil && readKind(dr.DrawData) == 1
}

func handleObjectOutOfSightNative48EA70(data []byte, hooks objectOutOfSightHooks48EA70) int {
	state, ok := decodeObjectOutOfSightState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 3
	}
	dr := findObjectOutOfSightDrawable48EA70(hooks.firstDrawable(), state.Code)
	if dr == nil || isOneShotObjectOutOfSightWithReader48EA70(dr, hooks.animateDrawFunc, hooks.animationKind) {
		return 3
	}
	if dr == hooks.localDrawable() {
		return 3
	}
	if nox_xxx_netTestHighBit_578B70(state.Code) {
		hooks.deactivate(dr)
	} else {
		hooks.deleteDrawable(dr)
	}
	return 3
}

func handleObjectInShadowsNative48EA70(data []byte, hooks objectInShadowsHooks48EA70) int {
	state, ok := decodeObjectOutOfSightState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 3
	}
	dr := findObjectOutOfSightDrawable48EA70(hooks.firstDrawable(), state.Code)
	if dr == nil {
		return 3
	}

	// GAME.EXE writes all three visibility counters before selecting the fade
	// or immediate-removal path. These are native fields 120..122, not byte
	// offsets 480..488 in a widened Drawable.
	dr.Field_121 = 1
	dr.Field_120 = 1
	dr.Field_122 = 1

	if hooks.fadeObjects() {
		if dr != hooks.localDrawable() {
			hooks.transparentFade(dr, int(hooks.tickRate()))
		}
		return 3
	}
	if isOneShotObjectOutOfSightWithReader48EA70(dr, hooks.animateDrawFunc, hooks.animationKind) || dr == hooks.localDrawable() {
		return 3
	}
	if nox_xxx_netTestHighBit_578B70(state.Code) {
		hooks.deactivate(dr)
	} else {
		hooks.deleteDrawable(dr)
	}
	return 3
}

func (c *Client) handleObjectOutOfSightPacketNative48EA70(data []byte) int {
	return handleObjectOutOfSightNative48EA70(data, objectOutOfSightHooks48EA70{
		connected:       nox_client_isConnected,
		firstDrawable:   c.Objs.FirstList1,
		localDrawable:   c.ClientPlayerUnit,
		deactivate:      c.Nox_xxx_cliDestroyObj_45A9A0,
		deleteDrawable:  c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
		animateDrawFunc: legacy.Get_nox_thing_animate_draw(),
		animationKind:   animateDrawKindNative48EA70,
	})
}

func (c *Client) handleObjectInShadowsPacketNative48EA70(data []byte) int {
	return handleObjectInShadowsNative48EA70(data, objectInShadowsHooks48EA70{
		connected:       nox_client_isConnected,
		firstDrawable:   c.Objs.FirstList1,
		localDrawable:   c.ClientPlayerUnit,
		fadeObjects:     func() bool { return legacy.Get_nox_client_fadeObjects_80836() != 0 },
		tickRate:        c.Server.TickRate,
		transparentFade: c.Objs.TransparentDecay,
		deactivate:      c.Nox_xxx_cliDestroyObj_45A9A0,
		deleteDrawable:  c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
		animateDrawFunc: legacy.Get_nox_thing_animate_draw(),
		animationKind:   animateDrawKindNative48EA70,
	})
}
