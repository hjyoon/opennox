package opennox

import (
	"encoding/binary"
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

type drainHealRayHooks4CD450 struct {
	typeID func(string) int
	random func(int, int) int
	byCode func(uint16) *client.Drawable
	spawn  func(int, image.Point) *client.Drawable
}

func (c *Client) drainHealRayHooks4CD450() drainHealRayHooks4CD450 {
	return drainHealRayHooks4CD450{
		typeID: c.Things.IndByID,
		random: c.srv.Rand.Other.Int,
		byCode: c.Objs.ByNetCode,
		spawn:  c.Nox_xxx_spriteLoadAdd_45A360_drawable,
	}
}

// Both legacy ray updates passed a drawable pointer through int and read
// PE32 offsets. Keep the packed ray coordinates/codes but resolve drawable
// positions and spawned orb payloads at their native-width layout.
func updateDrainHealRay4CD450(vp *noxrender.Viewport, dr *client.Drawable, orbName string, hooks drainHealRayHooks4CD450) int {
	if hooks.random(0, 100) >= 50 {
		return 1
	}
	ray := unsafe.Slice((*byte)(unsafe.Pointer(&dr.Union)), 13)
	var source, target image.Point
	if ray[0] != 0 {
		from := hooks.byCode(binary.LittleEndian.Uint16(ray[5:]))
		to := hooks.byCode(binary.LittleEndian.Uint16(ray[9:]))
		if from == nil || to == nil {
			return 1
		}
		source, target = from.PosVec, to.PosVec
	} else {
		source = image.Pt(int(binary.LittleEndian.Uint16(ray[5:])), int(binary.LittleEndian.Uint16(ray[7:])))
		target = image.Pt(int(binary.LittleEndian.Uint16(ray[9:])), int(binary.LittleEndian.Uint16(ray[11:])))
	}
	x := target.X + hooks.random(-20, 20)
	y := target.Y + hooks.random(-20, 20)
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
	orb := hooks.spawn(hooks.typeID(orbName), image.Pt(x, y))
	if orb != nil {
		payload := unsafe.Slice((*byte)(unsafe.Pointer(&orb.Union)), 13)
		binary.LittleEndian.PutUint16(payload[0:], uint16(source.X))
		binary.LittleEndian.PutUint16(payload[2:], uint16(source.Y))
		payload[11] = byte(hooks.random(6, 12))
		payload[12] = byte(hooks.random(3, 10))
	}
	return 1
}
