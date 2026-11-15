package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

// PullPushNativeRuntime52BFA0 supplies the original external service boundaries.
// The ordinary selector retains the existing balance, radial and audio bindings.
type PullPushNativeRuntime52BFA0 struct {
	Balance   func(string) float64
	PushUnits func(types.Pointf, float32, float32, float32, *Object, int, int)
	CastSound func(int32) sound.ID
	Audio     func(sound.ID, *Object, int, uint32)
}

func PullCastNative52BFA0(id int32, owner, caster *Object, level int32, rt PullPushNativeRuntime52BFA0) int32 {
	return pullCast52BFA0(id, owner, caster, level, pullCastDeps52BFA0{
		balance: rt.Balance, pushUnits: rt.PushUnits,
		castAudio: func(id int32, owner *Object) { rt.Audio(rt.CastSound(id), owner, 0, 0) },
	})
}

func PushCastNative52C000(id int32, owner, caster *Object, level int32, rt PullPushNativeRuntime52BFA0) int32 {
	return pushCast52C000(id, owner, caster, level, pushCastDeps52C000{
		balance: rt.Balance, pushUnits: rt.PushUnits,
		castAudio: func(id int32, owner *Object) { rt.Audio(rt.CastSound(id), owner, 0, 0) },
	})
}
