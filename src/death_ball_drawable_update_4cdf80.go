package opennox

import (
	"image"
	"math"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

type deathBallDrawableHooks4CDF80 struct {
	typeID   func(string) int
	random   func(int, int) int
	frame    func() uint32
	spawn    func(int, image.Point) *client.Drawable
	activate func(*client.Drawable)
}

func (c *Client) deathBallDrawableHooks4CDF80() deathBallDrawableHooks4CDF80 {
	return deathBallDrawableHooks4CDF80{
		typeID:   c.Things.IndByID,
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	}
}

// updateDeathBallSparks4CDF80 replaces both PE32 drawable callbacks that
// passed the drawable pointer through an int before spawning sparks.
func updateDeathBallSparks4CDF80(dr *client.Drawable, count int, hooks deathBallDrawableHooks4CDF80) int {
	typ := hooks.typeID("DeathBallSpark")
	for range count {
		spark := hooks.spawn(typ, dr.PosVec)
		if spark == nil {
			continue
		}
		effect := spark.UnionEffect()
		effect.Field_108 = uint32(dr.PosVec.X) << 12
		effect.Field_109 = uint32(dr.PosVec.Y) << 12
		spark.Field_74_4 = byte(hooks.random(0, 255))
		effect.Field_110 = uint32(hooks.random(1000, 3000))
		effect.Field_112 = hooks.frame() + uint32(hooks.random(10, 40))
		effect.Field_111 = hooks.frame()
		spark.ZVal = 22
		spark.VelZ = int8(hooks.random(0, 4))
		hooks.activate(spark)
	}
	return 1
}

// updateDeathBallCharge4CE0C0 mirrors the CharmOrb shower around a charging
// DeathBall. All source and spawned drawables remain native-width Go pointers.
func updateDeathBallCharge4CE0C0(dr *client.Drawable, hooks deathBallDrawableHooks4CDF80) int {
	typ := hooks.typeID("CharmOrb")
	originX, originY := uint16(dr.PosVec.X), uint16(dr.PosVec.Y)
	for range 10 {
		angle := uint8(hooks.random(0, 255))
		radius := hooks.random(2, 8)
		direction := sincosTable16[angle]
		x := uint16(int(originX) + radius*direction.X)
		y := uint16(int(originY) + radius*direction.Y)
		if hooks.random(0, 100) >= 50 {
			continue
		}
		age := byte(hooks.random(6, 10))
		dy := hooks.random(-20, 20)
		dx := hooks.random(-20, 20)
		orb := hooks.spawn(typ, image.Pt(dx+int(x), dy+int(y)))
		if orb == nil {
			continue
		}
		effect := orb.UnionEffect()
		effect.Field_108 = uint32(originX) | uint32(originY)<<16
		// The original routine writes bytes at 443-446 of the PE32 union.
		effect.Field_110 = effect.Field_110&0x00ffffff | uint32(age)<<24
		effect.Field_111 = effect.Field_111&0xff000000 | uint32(byte(hooks.random(3, 10)))
		hooks.activate(orb)
	}
	return 1
}

// linearOrbStep4CA650 advances a CharmOrb toward the charge origin stored in
// its effect union. The old C callback read the drawable through a 32-bit int.
func linearOrbStep4CA650(dr *client.Drawable) (image.Point, bool) {
	effect := dr.UnionEffect()
	dest := image.Pt(int(uint16(effect.Field_108)), int(uint16(effect.Field_108>>16)))
	dx, dy := dest.X-dr.PosVec.X, dest.Y-dr.PosVec.Y
	distance := int(math.Sqrt(float64(int64(dx)*int64(dx) + int64(dy)*int64(dy))))
	speed := int(byte(effect.Field_110 >> 24))
	next := image.Pt(dr.PosVec.X+dx*speed/(distance+1), dr.PosVec.Y+dy*speed/(distance+1))
	if distance+1 <= 10 ||
		int64(dr.PosVec.X-dest.X)*int64(next.X-dest.X)+
			int64(dr.PosVec.Y-dest.Y)*int64(next.Y-dest.Y) < 0 {
		return next, true
	}
	return next, false
}

func (c *Client) updateLinearOrb4CA650(dr *client.Drawable) int {
	next, done := linearOrbStep4CA650(dr)
	if done {
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return 0
	}
	c.Nox_xxx_updateSpritePosition_49AA90(dr, next.X, next.Y)
	return 1
}

func (c *Client) callDrawableUpdate49BD70(vp *noxrender.Viewport, dr *client.Drawable) int {
	switch fn := dr.ClientUpdateFuncPtr; fn {
	case nil:
		return 1
	case legacy.Get_nox_xxx_updDrawDBall_4CDF80():
		return updateDeathBallSparks4CDF80(dr, 3, c.deathBallDrawableHooks4CDF80())
	case legacy.Get_sub_4CE0A0():
		return updateDeathBallSparks4CDF80(dr, 1, c.deathBallDrawableHooks4CDF80())
	case legacy.Get_nox_xxx_updDrawDBallCharge_4CE0C0():
		return updateDeathBallCharge4CE0C0(dr, c.deathBallDrawableHooks4CDF80())
	case legacy.Get_nox_xxx_updDrawMagic_4CDD80():
		return c.updateMagicDrawable4CDD80(dr)
	case legacy.Get_nox_xxx_updDrawVortexSource_4CC950():
		return c.updateVortexSourceDrawable4CC950(dr)
	case legacy.Get_sub_4CCD00():
		return c.updateMeteorDrawable4CCD00(dr)
	case legacy.Get_nox_xxx_updDrawFist_4CCDB0():
		return c.updateFistDrawable4CCDB0(dr)
	case legacy.Get_nox_xxx_updDrawCloud_4CE1D0():
		return updateCloudDrawable4CE1D0(dr, 75, c.cloudDrawableHooks4CE200())
	case legacy.Get_sub_4CE360():
		return updateCloudDrawable4CE1D0(dr, 35, c.cloudDrawableHooks4CE200())
	case legacy.Get_sub_4CA650():
		return c.updateLinearOrb4CA650(dr)
	case legacy.Get_sub_4CD450():
		return updateDrainHealRay4CD450(vp, dr, "HealOrb", c.drainHealRayHooks4CD450())
	case legacy.Get_sub_4CD690():
		return updateDrainHealRay4CD450(vp, dr, "DrainManaOrb", c.drainHealRayHooks4CD450())
	case legacy.Get_nox_xxx_updDrawManabombCharge_4CCAC0():
		return updateManaBombCharge4CCAC0(dr, c.manaBombDrawableHooks4CCAC0())
	case legacy.Get_nox_xxx_updDrawTeleportWake_4CD8D0():
		return c.updateTeleportWakeDrawable4CD8D0(dr)
	case legacy.Get_nox_xxx_updDrawSparkleTrail_4CDBF0():
		return c.updateSparkleTrailDrawable4CDBF0(dr)
	case legacy.Get_nox_xxx_updDrawMagicMissile_4CD9E0():
		return c.updateMagicMissileDrawable4CD9E0(dr)
	case legacy.Get_sub_4CA720():
		return c.updateManaBombOrb4CA720(dr)
	default:
		return ccall.CallIntPtr2(fn, vp.C(), dr.C())
	}
}

func (c *Client) callDrawableSecondaryUpdate49BD70(vp *noxrender.Viewport, dr *client.Drawable) {
	switch fn := dr.Field_115; fn {
	case nil:
		return
	case legacy.Get_sub_4CE340():
		updateCloudParticleRise4CE340(dr)
	default:
		ccall.CallVoidPtr2(fn, vp.C(), dr.C())
	}
}
