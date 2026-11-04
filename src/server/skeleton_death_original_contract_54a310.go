package server

import (
	"github.com/opennox/libs/types"
	"unsafe"
)

// SkeletonDeathState54A310 keeps the original DWORD cache and signed divisor
// separate from native pointer storage. Both scalars remain live across calls.
type SkeletonDeathState54A310 struct{ Index, Divisor *uint32 }

type SkeletonDeathNames54A310 struct {
	Skull, Sword, WoodenShield, SteelShield, Power, Material unsafe.Pointer
}

type SkeletonRandomSite54A310 uint8

const (
	SkeletonRoll54A310 SkeletonRandomSite54A310 = iota
	SkeletonSkullSeconds54A310
	SkeletonPartCount54A310
	SkeletonBoneSeconds54A310
)

// SkeletonDeathRuntime54A310 supplies only original external world services.
// Names stay opaque until their consuming service; origin stays a pointer.
// RandomSite represents the original RNG caller metadata, not another draw.
// Apply receives the four initialized modifier pointers; the original fifth
// stack DWORD was uninitialized. Native Field16 is explicitly zero.
type SkeletonDeathRuntime54A310 struct {
	Names                   SkeletonDeathNames54A310
	RandomInt               func(int32, int32, SkeletonRandomSite54A310) int32
	RandomFloat             func(float32, float32) float64
	Flags                   func() uint32
	FPS                     func() uint32
	BoneName                func(int32) unsafe.Pointer
	New                     func(unsafe.Pointer) *Object
	Point                   func(float32, *types.Pointf, *types.Pointf)
	Create                  func(*Object, *Object, types.Pointf, uint32) uint32
	Raise                   func(*Object, float32) uint32
	Force                   func(*types.Pointf, *Object, float32) uint32
	Audio                   func(uint32, *Object, uint32, uint32) uint32
	PointFX                 func(uint32, *types.Pointf) uint32
	Decay                   func(*Object, uint32) uint32
	ModifierID              func(unsafe.Pointer) int32
	ModifierDesc            func(int32) unsafe.Pointer
	Apply                   func(*Object, *ModifierInitData) uint32
	ArmorSound, WeaponSound func(*Object) uint32
	RoundToNearest          bool
}
