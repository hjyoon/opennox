package opennox

import (
	"encoding/binary"
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

// updateCharmRay4CD400 replaces the PE32 CharmUpdateDraw callback. The legacy
// routine passed the ray drawable through int before emitting an orb from each
// endpoint, which truncated the pointer on native 64-bit address spaces.
func updateCharmRay4CD400(vp *noxrender.Viewport, dr *client.Drawable, hooks drainHealRayHooks4CD450) int {
	typ := hooks.typeID("CharmOrb")
	spawnCharmRayOrb4CD150(vp, dr, typ, true, hooks)
	spawnCharmRayOrb4CD150(vp, dr, typ, false, hooks)
	return 1
}

func spawnCharmRayOrb4CD150(vp *noxrender.Viewport, dr *client.Drawable, typ int, reverse bool, hooks drainHealRayHooks4CD450) {
	if hooks.random(0, 100) >= 50 {
		return
	}
	ray := unsafe.Slice((*byte)(unsafe.Pointer(&dr.Union)), 13)
	var source, target image.Point
	if ray[0] != 0 {
		from := hooks.byCode(binary.LittleEndian.Uint16(ray[5:]))
		to := hooks.byCode(binary.LittleEndian.Uint16(ray[9:]))
		if from == nil || to == nil {
			return
		}
		source, target = from.PosVec, to.PosVec
	} else {
		source = image.Pt(int(binary.LittleEndian.Uint16(ray[5:])), int(binary.LittleEndian.Uint16(ray[7:])))
		target = image.Pt(int(binary.LittleEndian.Uint16(ray[9:])), int(binary.LittleEndian.Uint16(ray[11:])))
	}
	spawnAt, destination := source, target
	if reverse {
		spawnAt, destination = target, source
	}
	x := spawnAt.X + hooks.random(-20, 20)
	y := spawnAt.Y + hooks.random(-20, 20)
	z := int(int16(dr.ZVal))
	screenX := x + vp.Screen.Min.X - vp.World.Min.X
	screenY := y + vp.Screen.Min.Y - vp.World.Min.Y - z
	if screenX < 0 {
		x = vp.World.Min.X + vp.Screen.Min.X + 1
	}
	if screenY < 0 {
		y = vp.World.Min.Y + vp.Screen.Min.Y - z + 1
	}
	if screenX >= vp.Size.X {
		x = vp.Screen.Max.X + vp.World.Min.X - 1
	}
	if screenY >= vp.Size.Y {
		y = vp.Screen.Max.Y + vp.World.Min.Y - z - 1
	}
	orb := hooks.spawn(typ, image.Pt(x, y))
	if orb == nil {
		return
	}
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&orb.Union)), 13)
	binary.LittleEndian.PutUint16(payload[0:], uint16(destination.X))
	binary.LittleEndian.PutUint16(payload[2:], uint16(destination.Y))
	payload[11] = byte(hooks.random(6, 12))
	payload[12] = byte(hooks.random(3, 10))
}
