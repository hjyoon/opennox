package server

import "github.com/opennox/libs/types"

// TrollDeathState54A270 is the original fixed-width SmallToxicCloud type cache.
// It is independent of object pointers and remains live across world calls.
type TrollDeathState54A270 struct {
	TypeIndex uint32
}

// TrollDeathRuntime54A270 supplies only the original external world services.
// FPS represents the direct signed-DWORD global read after Balance returns.
// The extra Create word records the original caller's trailing zero; the
// native four-argument world constructor does not consume that word.
type TrollDeathRuntime54A270 struct {
	Resolve        func() uint32
	New            func(uint32) *Object
	Create         func(*Object, *Object, types.Pointf, uint32) uint32
	Audio          func(uint32, *Object, uint32, uint32) uint32
	Balance        func() float64
	FPS            func() uint32
	RoundToNearest bool
}
