package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

// BurnNativeRuntime52C3E0 provides declared external services for the original
// cast. The ordinary selector retains its existing runtime and cache bindings.
type BurnNativeRuntime52C3E0 struct {
	LoadGlyph  func() uint32
	StoreGlyph func(uint32)
	LoadFlame  func() uint32
	StoreFlame func(uint32)
	LookupType func(string) uint32
	TraceRay   func(types.Pointf, types.Pointf, MapTraceFlags) bool
	NewObject  func(uint32) *Object
	CreateAt   func(*Object, *Object, types.Pointf)
	Balance    func(string) float64
	SetDecay   func(*Object, uint32)
	Spark      func(types.Pointf, byte)
	CastSound  func(int32) sound.ID
	Audio      func(sound.ID, types.Pointf, int, uint32)
	Inform     func(uint8, byte, int32)
}

func BurnCastNative52C3E0(id int32, caster *Object, arg *SpellAcceptArg, rt BurnNativeRuntime52C3E0) int32 {
	return burnCast52C3E0(id, caster, arg, burnCastDeps52C3E0{
		loadGlyph: rt.LoadGlyph, storeGlyph: rt.StoreGlyph, loadFlame: rt.LoadFlame, storeFlame: rt.StoreFlame,
		lookupType: rt.LookupType, traceRay: rt.TraceRay, newObject: rt.NewObject, createAt: rt.CreateAt,
		balance: rt.Balance, setDecay: rt.SetDecay, spark: rt.Spark, castSound: rt.CastSound,
		audio: rt.Audio, inform: rt.Inform,
	})
}
