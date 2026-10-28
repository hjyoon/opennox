package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
)

// MagicMissileUpdateRuntime53BDA0 supplies the world operations used by the
// thing.bin update callback. The missile, owner, and target stay native-width.
type MagicMissileUpdateRuntime53BDA0 struct {
	Frame          func() uint32
	FPS            func() uint32
	SearchTarget   func(missile, owner *Object) *Object
	Collide        func(missile *Object)
	CollideResult  func(missile *Object) int32
	DirectionFloat func(index int32) float32
}

// MagicMissileUpdate53BDA0 follows GAME.EXE 0053BDA0 without reading PE32
// object or update-data pointer offsets from native-width records.
func MagicMissileUpdate53BDA0(missile *Object, runtime MagicMissileUpdateRuntime53BDA0) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	if missile.UpdateDataMissile().Owner == nil {
		magicMissileCollision53BDA0(missile, runtime)
		return
	}
	MagicMissileUpdateRaw53BDA0(missile, runtime)
}

// MagicMissileUpdateRaw53BDA0 retains the original required pointers, cached
// update/owner and DWORD result. The registered Go API keeps its old preflights.
func MagicMissileUpdateRaw53BDA0(missile *Object, runtime MagicMissileUpdateRuntime53BDA0) int32 {
	update := missile.UpdateDataMissile()
	owner := update.Owner
	if owner.ObjFlags.Has(object.FlagDestroyed) {
		return magicMissileCollision53BDA0(missile, runtime)
	}
	frame := runtime.Frame()
	created := missile.Field32
	if frame-created > 3*runtime.FPS() {
		return magicMissileCollision53BDA0(missile, runtime)
	}

	if target := update.Target; target != nil && target.ObjFlags.HasAny(object.FlagDestroyed|object.FlagDead) {
		update.Target = nil
	}
	if update.Target == nil && runtime.Frame()&7 == 0 {
		update.Target = runtime.SearchTarget(missile, owner)
		if update.Target == missile.ObjOwner {
			update.Target = nil
		}
	}
	read := runtime.DirectionFloat
	if read == nil {
		read = magicMissileDirectionFloat53BDA0
	}
	if target := update.Target; target != nil {
		direction := uint16(missile.Direction2)
		index := int32(int16(direction)) * 2
		cos := objectForceMemory32_52DF80(read(index))
		sin := objectForceMemory32_52DF80(read(index + 1))
		dy := objectForceSub53_52DF80(float64(target.PosVec.Y), objectForceMemory32_52DF80(missile.PosVec.Y))
		y := objectForceMul53_52DF80(dy, cos)
		dx := objectForceSub53_52DF80(float64(target.PosVec.X), objectForceMemory32_52DF80(missile.PosVec.X))
		x := objectForceMul53_52DF80(dx, sin)
		if objectForceSub53_52DF80(y, x) >= 0 {
			direction += 42
			for int16(direction) >= 256 {
				direction -= 256
			}
		} else {
			// C0 also selects this branch for unordered comparisons. Both loops
			// use signed WORDs, rather than narrowing the input to a byte.
			direction -= 42
			for int16(direction) < 0 {
				direction += 256
			}
		}
		missile.Direction2 = Dir16(direction)
	}
	index := int32(int16(missile.Direction2)) * 2
	cos := objectForceMemory32_52DF80(read(index))
	missile.ForceVec.X = experienceSpill32TowardZero(objectForceMul53_52DF80(cos, objectForceMemory32_52DF80(missile.SpeedCur)))
	sin := objectForceMemory32_52DF80(read(index + 1))
	y := objectForceMul53_52DF80(sin, objectForceMemory32_52DF80(missile.SpeedCur))
	missile.Float28 = math.Float32frombits(1061997773)
	missile.ForceVec.Y = experienceSpill32TowardZero(y)
	return index * 4
}

func magicMissileCollision53BDA0(missile *Object, runtime MagicMissileUpdateRuntime53BDA0) int32 {
	if runtime.CollideResult != nil {
		return runtime.CollideResult(missile)
	}
	runtime.Collide(missile)
	return 0
}

func magicMissileDirectionFloat53BDA0(index int32) float32 {
	if index >= 0 && index < 512 {
		return sincosDirTable[index/2][index&1]
	}
	// Raw signed-WORD indices retain the original address, with no byte mask.
	return *(*float32)(memmap.Ptr(uintptr(int64(0x5b6658) + int64(index)*4)))
}
