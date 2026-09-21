package opennox

import (
	"image"
	"math"
	"sync"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

var (
	drawableUpdateCallbacksOnce49BD70 sync.Once
	drawableUpdateCallbacks49BD70     map[unsafe.Pointer]struct{}
)

// isDrawableUpdateCallback49BD70 identifies callbacks whose ABI is
// client-update, not draw. Calling one through Drawable.DrawFuncPtr passes the
// viewport as its first argument and corrupts the drawable pointer on 64-bit
// hosts. Keep this list in sync with the native update dispatchers below.
func isDrawableUpdateCallback49BD70(fn unsafe.Pointer) bool {
	if fn == nil {
		return false
	}
	drawableUpdateCallbacksOnce49BD70.Do(func() {
		callbacks := []unsafe.Pointer{
			legacy.Get_nox_xxx_updDrawDBall_4CDF80(),
			legacy.Get_sub_4CE0A0(),
			legacy.Get_nox_xxx_updDrawDBallCharge_4CE0C0(),
			legacy.Get_nox_xxx_updDrawMagic_4CDD80(),
			legacy.Get_nox_xxx_updDrawVortexSource_4CC950(),
			legacy.Get_sub_4CCD00(),
			legacy.Get_nox_xxx_updDrawFist_4CCDB0(),
			legacy.Get_nox_xxx_updDrawColorlight_4CE390(),
			legacy.Get_nox_xxx_updDrawUndeadKiller_4CCCF0(),
			legacy.Get_nox_xxx_updDrawMonsterGen_4BC920(),
			legacy.Get_nox_xxx_updDrawCloud_4CE1D0(),
			legacy.Get_sub_4CE360(),
			legacy.Get_sub_4CA650(),
			legacy.Get_sub_4CD400(),
			legacy.Get_sub_4CCE70(),
			legacy.Get_sub_4CD090(),
			legacy.Get_sub_4CD0C0(),
			legacy.Get_sub_4CD0F0(),
			legacy.Get_sub_4CD120(),
			legacy.Get_sub_4CD450(),
			legacy.Get_sub_4CD690(),
			legacy.Get_nox_xxx_updDrawManabombCharge_4CCAC0(),
			legacy.Get_nox_xxx_updDrawTeleportWake_4CD8D0(),
			legacy.Get_nox_xxx_updDrawSparkleTrail_4CDBF0(),
			legacy.Get_nox_xxx_updDrawMagicMissile_4CD9E0(),
			legacy.Get_sub_4CA720(),
			legacy.Get_sub_4CE340(),
			legacy.Get_nox_xxx_sprite_4CA540(),
		}
		drawableUpdateCallbacks49BD70 = make(map[unsafe.Pointer]struct{}, len(callbacks))
		for _, callback := range callbacks {
			if callback != nil {
				drawableUpdateCallbacks49BD70[callback] = struct{}{}
			}
		}
	})
	_, ok := drawableUpdateCallbacks49BD70[fn]
	return ok
}

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
	case legacy.Get_nox_xxx_updDrawColorlight_4CE390():
		return c.updateColorLightDrawable4CE390(vp, dr)
	case legacy.Get_nox_xxx_updDrawUndeadKiller_4CCCF0(), legacy.Get_nox_xxx_updDrawMonsterGen_4BC920():
		return 1
	case legacy.Get_nox_xxx_updDrawCloud_4CE1D0():
		return updateCloudDrawable4CE1D0(dr, 75, c.cloudDrawableHooks4CE200())
	case legacy.Get_sub_4CE360():
		return updateCloudDrawable4CE1D0(dr, 35, c.cloudDrawableHooks4CE200())
	case legacy.Get_sub_4CA650():
		return c.updateLinearOrb4CA650(dr)
	case legacy.Get_sub_4CD400():
		return updateCharmRay4CD400(vp, dr, c.drainHealRayHooks4CD450())
	case legacy.Get_sub_4CCE70():
		return c.updateFireballDrawable4CCE70(dr, 5)
	case legacy.Get_sub_4CD090():
		return c.updateFireballDrawable4CCE70(dr, 4)
	case legacy.Get_sub_4CD0C0():
		return c.updateFireballDrawable4CCE70(dr, 3)
	case legacy.Get_sub_4CD0F0():
		return c.updateFireballDrawable4CCE70(dr, 2)
	case legacy.Get_sub_4CD120():
		return c.updateFireballDrawable4CCE70(dr, 1)
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
		// All callbacks accepted by the thing parser, plus the dynamically
		// assigned mana-bomb callback, are dispatched above. Do not call an
		// unknown legacy callback: many of them still pass newly allocated
		// drawable pointers through PE32-sized ints and crash on 64-bit hosts.
		return 1
	}
}

func (c *Client) callDrawableSecondaryUpdate49BD70(vp *noxrender.Viewport, dr *client.Drawable) {
	switch fn := dr.Field_115; fn {
	case nil:
		return
	case legacy.Get_sub_4CE340():
		updateCloudParticleRise4CE340(dr)
	case legacy.Get_nox_xxx_sprite_4CA540():
		c.updateClientPredictLinear4CA540(vp, dr)
	default:
		// Unknown legacy callbacks cannot be called safely on a native-width
		// build.
		return
	}
}
