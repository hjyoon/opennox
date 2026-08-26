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

func handleObjectOutOfSightNative48EA70(data []byte, hooks objectOutOfSightHooks48EA70) int {
	state, ok := decodeObjectOutOfSightState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 3
	}
	dr := findObjectOutOfSightDrawable48EA70(hooks.firstDrawable(), state.Code)
	if dr == nil || isOneShotObjectOutOfSightDrawable48EA70(dr, hooks.animateDrawFunc) {
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

func (c *Client) handleObjectOutOfSightPacketNative48EA70(data []byte) int {
	return handleObjectOutOfSightNative48EA70(data, objectOutOfSightHooks48EA70{
		connected:       nox_client_isConnected,
		firstDrawable:   c.Objs.FirstList1,
		localDrawable:   c.ClientPlayerUnit,
		deactivate:      c.Nox_xxx_cliDestroyObj_45A9A0,
		deleteDrawable:  c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
		animateDrawFunc: legacy.Get_nox_thing_animate_draw(),
	})
}
