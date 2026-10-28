package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
)

// ProjectileTrailRuntime53AEC0 keeps random sparks on the native object path.
type ProjectileTrailRuntime53AEC0 struct {
	RandomFloat     func(float32, float32) float32
	RandomFloatWide func(float32, float32) float64
	CreateSpark     func(types.Pointf, int, int, types.Pointf, float32, *Object)
}

// ProjectileTrailUpdate53AEC0 restores GAME.EXE 0053AEC0 without passing the
// projectile as a 32-bit integer owner to the legacy Spark factory.
func ProjectileTrailUpdate53AEC0(projectile *Object, runtime ProjectileTrailRuntime53AEC0) {
	if projectile == nil {
		return
	}
	random := runtime.RandomFloatWide
	if random == nil {
		random = func(lo, hi float32) float64 { return float64(runtime.RandomFloat(lo, hi)) }
	}
	ProjectileTrailUpdateRaw53AEC0(projectile, ProjectileTrailRawRuntime53AEC0{
		DirectionFloat: projectileTrailDirection53AEC0, RandomFloat: random,
		CreateSpark: func(p types.Pointf, kind, life int, v types.Pointf, z float32, owner *Object) *Object {
			runtime.CreateSpark(p, kind, life, v, z, owner)
			return nil
		},
	})
}

func projectileTrailDirection53AEC0(index int32) float32 {
	if index >= 0 && index < 512 {
		return sincosDirTable[index/2][index%2]
	}
	return *(*float32)(memmap.Ptr(uintptr(int64(0x5b6658) + int64(index)*4)))
}
