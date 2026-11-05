package server

import "github.com/opennox/libs/types"

// SkullRawState54F9A0 is the six original live dwords at 00834B08.
// It contains numeric type indices and the shared reachable marker only.
type SkullRawState54F9A0 struct {
	Reachable       uint32
	MercArcherArrow uint32
	ProjectileTrap1 uint32
	ProjectileTrap2 uint32
	UpdateTrap1     uint32
	UpdateTrap2     uint32
}

// SkullRawRuntime54F9A0 describes the original world boundary independently
// of the raw native implementation. Enumeration must invoke its supplied
// candidate callback for every declared candidate, even after a hit.
// Pointer admission and ordinary dispatcher normalization are separate.
type SkullRawRuntime54F9A0 struct {
	Resolve            func(string) uint32
	New                func(uint32) *Object
	Create             func(*Object, *Object, types.Pointf, uint32)
	Balance            func() float64
	Audio              func(uint32, *Object)
	TrapFX             func(*types.Pointf, uint32)
	Enumerate          func(types.Rectf, *Object, func(*Object, *Object))
	Enemy              func(*Object, *Object) uint32
	MapCheck           func(*Object, *Object) uint32
	Front              func(*Object, int32, *Object) uint32
	FPS                func() uint32
	DirectionComponent func(int16, uint32) float32
	// Reference control only; ordinary startup uses PC53 rounding toward zero.
	RoundToNearest bool
}
