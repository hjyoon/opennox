package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	blackPowderBurnFuseDelay53CCB0 = uint32(3)
	blackPowderBurnRadius53CCB0    = float32(15)
	blackPowderBurnDamage53CCB0    = 1
	blackPowderBurnLifetime53CCB0  = uint32(2)
)

// BlackPowderBurnUpdateRuntime53CCB0 supplies the damage and deletion effects
// reached by the native-width restoration of GAME.EXE 0053CCB0.
type BlackPowderBurnUpdateRuntime53CCB0 struct {
	DamageUnitsAround func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	DelayedDelete     func(*Object)
}

type blackPowderBurnUpdateDeps53CCB0 struct {
	frame             func() uint32
	fps               func() uint32
	damageUnitsAround func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	delayedDelete     func(*Object)
}

func blackPowderBurnUpdateNative53CCB0(source *Object, deps blackPowderBurnUpdateDeps53CCB0) {
	fuseFrame := source.Field34
	creationFrame := source.Field32
	if fuseFrame == creationFrame {
		source.Field34 = deps.frame() + blackPowderBurnFuseDelay53CCB0
		return
	}
	if fuseFrame == deps.frame() {
		deps.damageUnitsAround(
			source.PosVec,
			blackPowderBurnRadius53CCB0,
			blackPowderBurnRadius53CCB0,
			blackPowderBurnDamage53CCB0,
			object.DamageFlame,
			source,
			nil,
		)
		return
	}
	if deps.frame()-creationFrame >= blackPowderBurnLifetime53CCB0*deps.fps() {
		deps.delayedDelete(source)
	}
}

// BlackPowderBurnUpdate53CCB0 restores GAME.EXE 0053CCB0 without reading a
// native-width Object through PE32 offsets.
func (s *Server) BlackPowderBurnUpdate53CCB0(source *Object, runtime BlackPowderBurnUpdateRuntime53CCB0) {
	blackPowderBurnUpdateNative53CCB0(source, blackPowderBurnUpdateDeps53CCB0{
		frame:             s.Frame,
		fps:               s.TickRate,
		damageUnitsAround: runtime.DamageUnitsAround,
		delayedDelete:     runtime.DelayedDelete,
	})
}
