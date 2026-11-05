package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// Keep the original Y-before-X reads across the fault boundary. In
// particular, a partial source can have mapped Y and unmapped X, and an
// invalid cloud must not be read before either source coordinate.
//
//go:noinline
func polypRawY54CB10(source *Object) float32 { return source.PosVec.Y }

//go:noinline
func polypRawX54CB10(source *Object) float32 { return source.PosVec.X }

// PolypDeathNative54CB10 preserves the original unguarded raw C contract.
// Dispatcher admission remains separate; failed allocation still reaches
// source audio and deletion without reading the source or update record.
func (source *Object) PolypDeathNative54CB10(state *PolypDeathState54CB10, rt PolypDeathRawRuntime54CB10) {
	id := state.TypeIndex
	if id == 0 {
		id = rt.Resolve()
		state.TypeIndex = id
	}
	cloud := rt.New(id)
	if cloud != nil {
		y := polypRawY54CB10(source)
		x := polypRawX54CB10(source)
		data := (*ToxicCloudUpdateData)(cloud.UpdateData)
		rt.Create(cloud, nil, types.Ptf(x, y), 0)
		lifetime := rt.Balance()
		fps := rt.FPS()
		var value float32
		if rt.RoundToNearest {
			value = float32(lifetime * float64(int32(fps)))
		} else {
			value = experienceSpill32TowardZero(experienceMul53TowardZero(lifetime, float64(int32(fps))))
		}
		duration := netFXTruncateSignedDWORD419A70(value)
		if rt.RoundToNearest {
			rounded := math.RoundToEven(float64(value))
			if math.IsNaN(rounded) || rounded < -2147483648 || rounded >= 2147483648 {
				duration = math.MinInt32
			} else {
				duration = int32(rounded)
			}
		}
		data.Duration = duration
	}
	rt.Audio(284, source)
	rt.DelayedDelete(source)
}
