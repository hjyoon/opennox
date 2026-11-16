package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

// InversionNativeRuntime52BEB0 supplies the original declared external services.
// The ordinary selector retains its native map, ownership and audio bindings.
type InversionNativeRuntime52BEB0 struct {
	Balance     func(string) float64
	EachMissile func(types.Pointf, float32, func(*Object) bool)
	ChangeOwner func(*Object, *Object)
	CastSound   func(int32) sound.ID
	Audio       func(sound.ID, *Object, int, uint32)
}

func InversionCastNative52BEB0(id int32, owner, caster *Object, rt InversionNativeRuntime52BEB0) int32 {
	return inversionCast52BEB0(id, owner, caster, inversionCastDeps52BEB0{
		balance: rt.Balance, eachMissile: rt.EachMissile, changeOwner: rt.ChangeOwner,
		castAudio: func(id int32, caster *Object) { rt.Audio(rt.CastSound(id), caster, 0, 0) },
	})
}
