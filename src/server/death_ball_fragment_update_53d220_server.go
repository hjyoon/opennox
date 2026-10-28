package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	deathBallFragmentDamageDelay53D220 = int32(10)
	deathBallFragmentOuterRadius53D220 = float32(25)
	deathBallFragmentInnerRadius53D220 = float32(0)
	deathBallFragmentDamage53D220      = 20
	deathBallFragmentLifetime53D220    = uint32(2)
)

// DeathBallFragmentUpdateRuntime53D220 supplies the two side effects reached
// by the native-width restoration of GAME.EXE 0053D220.
type DeathBallFragmentUpdateRuntime53D220 struct {
	DamageUnitsAround func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	DelayedDelete     func(*Object)
}

type deathBallFragmentUpdateNativeDeps53D220 struct {
	frame             func() uint32
	fps               func() uint32
	damageUnitsAround func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	delayedDelete     func(*Object)
}

func deathBallFragmentUpdateNative53D220(source *Object, deps deathBallFragmentUpdateNativeDeps53D220) {
	LifetimeEffectUpdateRaw53CC90(LifetimeEffectDeathBallFragment53D220, source, LifetimeEffectRuntime53CC90{
		Frame: deps.frame, FPS: deps.fps, DelayedDelete: deps.delayedDelete,
		DamageAround: func(p *types.Pointf, a, b float32, d int, typ object.DamageType, o *Object, x Obj) {
			deps.damageUnitsAround(*p, a, b, d, typ, o, x)
		},
	})
}

// DeathBallFragmentUpdate53D220 restores GAME.EXE 0053D220 without narrowing
// the fragment pointer to the PE32 integer accepted by the legacy C body.
func (s *Server) DeathBallFragmentUpdate53D220(source *Object, runtime DeathBallFragmentUpdateRuntime53D220) {
	deathBallFragmentUpdateNative53D220(source, deathBallFragmentUpdateNativeDeps53D220{
		frame:             s.Frame,
		fps:               s.TickRate,
		damageUnitsAround: runtime.DamageUnitsAround,
		delayedDelete:     runtime.DelayedDelete,
	})
}
