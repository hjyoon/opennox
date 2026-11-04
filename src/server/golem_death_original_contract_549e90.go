package server

import (
	"github.com/opennox/libs/types"
	"unsafe"
)

// GolemDeathState549E90 retains the original shared DWORD and signed divisor.
// Neither scalar is a native pointer slot, and both remain live across calls.
type GolemDeathState549E90 struct{ Index, Divisor *uint32 }

// GolemOrigin549E90 shares the already checked native source-reference carrier.
// A malformed nil source is retained as a source, never as a small Go pointer.
type GolemOrigin549E90 = SkeletonOrigin54A310

type GolemRandomSite549E90 uint8

const (
	GolemPartCount549E90 GolemRandomSite549E90 = iota
	GolemDecaySeconds549E90
)

// GolemDeathRuntime549E90 replaces only named original world services. Table
// callbacks are direct packed-data reads with the original DWORD arithmetic.
// MechName's argument is a slot cursor, RockName's is a signed DWORD index;
// Mass reads an unsigned byte at the live shared index, not the cached name slot.
type GolemDeathRuntime549E90 struct {
	RandomInt          func(int32, int32, GolemRandomSite549E90) int32
	RandomFloat        func(float32, float32) float64
	Flags              func() uint32
	FPS                func() uint32
	MechName, RockName func(int32) unsafe.Pointer
	Mass               func(int32) byte
	New                func(unsafe.Pointer) *Object
	Point              func(float32, GolemOrigin549E90, *types.Pointf)
	Create             func(*Object, *Object, types.Pointf, uint32) uint32
	Raise              func(*Object, float32) uint32
	Force              func(GolemOrigin549E90, *Object, float32) uint32
	Audio              func(uint32, *Object, uint32, uint32) uint32
	PointFX            func(uint32, GolemOrigin549E90) uint32
	Decay              func(*Object, uint32) uint32
	RoundToNearest     bool
}
