package opennox

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

type cloudDrawableHooks4CE200 struct {
	frame        func() uint32
	typeID       func(string) int
	random       func(int, int) int
	spawn        func(int, image.Point) *client.Drawable
	activate     func(*client.Drawable)
	riseUpdate   unsafe.Pointer
	decay        func(*client.Drawable, int)
	updateList   func(*client.Drawable)
	sightDestroy func(*client.Drawable)
}

func (c *Client) cloudDrawableHooks4CE200() cloudDrawableHooks4CE200 {
	return cloudDrawableHooks4CE200{
		frame:        c.srv.Frame,
		typeID:       c.Things.IndByID,
		random:       c.srv.Rand.Other.Int,
		spawn:        c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate:     c.Objs.List34Add,
		riseUpdate:   legacy.Get_sub_4CE340(),
		decay:        c.Objs.TransparentDecay,
		updateList:   c.Objs.List5Add,
		sightDestroy: c.Objs.List6Add,
	}
}

// updateCloudDrawable4CE1D0 replaces the CloudUpdateDraw and
// SmallCloudUpdateDraw PE32 callbacks. Both callbacks received the source
// drawable through an int, truncating it on 64-bit hosts.
func updateCloudDrawable4CE1D0(source *client.Drawable, radius int, hooks cloudDrawableHooks4CE200) int {
	if hooks.frame()&1 == 0 {
		return 1
	}
	spawnCloudParticles4CE200(source, 1, radius, hooks)
	return 1
}

func spawnCloudParticles4CE200(source *client.Drawable, count, radius int, hooks cloudDrawableHooks4CE200) {
	greenPuff := hooks.typeID("GreenPuff")
	greenSmoke := hooks.typeID("GreenSmoke")
	for range count {
		angle := uint8(hooks.random(0, 255))
		distance := hooks.random(0, radius)
		direction := sincosTable16[angle]
		pos := source.PosVec.Add(image.Pt(distance*direction.X/16, distance*direction.Y/16))
		typ := greenPuff
		if hooks.random(0, 10) < 3 {
			typ = greenSmoke
		}
		particle := hooks.spawn(typ, pos)
		if particle == nil {
			continue
		}
		particle.ZVal = 0
		hooks.activate(particle)
		effect := particle.UnionEffect()
		effect.Field_108 = effect.Field_108&^0xff | uint32(byte(hooks.random(1, 3)))
		particle.Field_115 = hooks.riseUpdate
		hooks.decay(particle, hooks.random(10, 32))
		hooks.updateList(particle)
		hooks.sightDestroy(particle)
	}
}

// updateCloudParticleRise4CE340 replaces the secondary PE32 callback stored
// in Field_115 by sub_4CE200. Its speed is the low byte of the effect union.
func updateCloudParticleRise4CE340(particle *client.Drawable) int {
	particle.ZVal += uint16(byte(particle.UnionEffect().Field_108))
	return 1
}
