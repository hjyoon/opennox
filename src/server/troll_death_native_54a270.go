package server

import (
	"github.com/opennox/libs/types"
	"math"
)

// TrollDeathNative54A270 preserves the unguarded original C contract. The
// constructor may fail before the source is read; the cached update pointer
// is stored only after all world effects, even if those effects replace it.
func (unit *Object) TrollDeathNative54A270(state *TrollDeathState54A270, rt TrollDeathRuntime54A270) int32 {
	id := state.TypeIndex
	if id == 0 {
		id = rt.Resolve()
		state.TypeIndex = id
	}
	cloud := rt.New(id)
	if cloud == nil {
		return 1
	}
	data := (*ToxicCloudUpdateData)(cloud.UpdateData)
	y := unit.PosVec.Y
	x := unit.PosVec.X
	rt.Create(cloud, unit, types.Ptf(x, y), 0)
	rt.Audio(644, unit, 0, 0)
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
	return 1
}
