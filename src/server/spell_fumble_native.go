package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

// FumbleNativeRuntime52C060 supplies the original external services and cache cell.
// The ordinary selector retains its existing native inventory, force and audio bindings.
type FumbleNativeRuntime52C060 struct {
	ForceDrop      func(*Object, *Object) int32
	DropAll        func(*Object)
	LoadTypeCache  func() uint32
	LookupType     func(string) uint32
	StoreTypeCache func(uint32)
	ApplyForce     func(*Object, types.Pointf, float64)
	ClearOwner     func(*Object)
	EffectSound    func(int32) sound.ID
	Audio          func(sound.ID, *Object, int, uint32)
}

func FumbleCastNative52C060(id int32, caster *Object, arg *SpellAcceptArg, rt FumbleNativeRuntime52C060) int32 {
	return fumbleCast52C060(id, caster, arg, fumbleCastDeps52C060{
		forceDrop: rt.ForceDrop, dropAll: rt.DropAll,
		loadTypeCache: rt.LoadTypeCache, lookupType: rt.LookupType, storeTypeCache: rt.StoreTypeCache,
		applyForce: rt.ApplyForce, clearOwner: rt.ClearOwner,
		ballAudio:   func(target *Object) { rt.Audio(sound.ID(926), target, 0, 0) },
		effectAudio: func(id int32, target *Object) { rt.Audio(rt.EffectSound(id), target, 0, 0) },
	})
}
