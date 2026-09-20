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
	age := deps.frame() - source.Field32

	// GAME.EXE stores the frame delta in a 32-bit int for this first branch.
	// A creation frame in the future therefore does not apply nearby damage.
	if int32(age) > deathBallFragmentDamageDelay53D220 {
		deps.damageUnitsAround(
			source.PosVec,
			deathBallFragmentOuterRadius53D220,
			deathBallFragmentInnerRadius53D220,
			deathBallFragmentDamage53D220,
			object.DamageCrush,
			source,
			nil,
		)
	}

	// gameFPS returns uint32, so the original second comparison promotes the
	// same delta back to unsigned. Keep both the strict boundary and wraparound.
	if age > deathBallFragmentLifetime53D220*deps.fps() {
		deps.delayedDelete(source)
	}
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
