package server

import (
	"math"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

type armorCreateNativeDeps54C950 struct {
	findDefinition func(uint16) *Modifier
	gameFlag       func(uint32) int32
	loadBalance    func(string) float32
	floatToInt     func(float32) int32
}

func armorCreateNative54C950(obj *Object, deps armorCreateNativeDeps54C950) {
	armorCreate54C950(obj, armorCreateHooks54C950[*Object, *Modifier, *HealthData]{
		loadTypeInd: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		findDefinition: deps.findDefinition,
		loadHealth: func(obj *Object) *HealthData {
			return obj.HealthData
		},
		loadDurability: func(definition *Modifier) uint16 {
			return uint16(definition.Durability52)
		},
		storeCurrent: func(health *HealthData, value uint16) {
			health.Cur = value
		},
		storeMaximum: func(health *HealthData, value uint16) {
			health.Max = value
		},
		gameFlag:    deps.gameFlag,
		loadBalance: deps.loadBalance,
		loadCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		loadMaximum: func(health *HealthData) uint16 {
			return health.Max
		},
		floatToInt: deps.floatToInt,
	})
}

// armorCreateRoundFloat32ToInt32_54C950 models nox_float2int at 00419A70:
// x87 FISTP under the default round-to-nearest-even mode. Invalid and
// out-of-range conversions produce the integer-indefinite INT32_MIN value.
func armorCreateRoundFloat32ToInt32_54C950(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

// ArmorCreate54C950 initializes armor durability without sending the live
// native-width Object, Modifier, or HealthData pointers through the PE32 C
// callback. The legacy function address remains registered as thing.bin
// callback identity by the legacy package.
func (s *Server) ArmorCreate54C950(obj *Object) {
	armorCreateNative54C950(obj, armorCreateNativeDeps54C950{
		findDefinition: func(typeInd uint16) *Modifier {
			return s.Modif.Nox_xxx_equipClothFindDefByTT413270(int(typeInd))
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		loadBalance: func(key string) float32 {
			return float32(s.Balance.Float(key))
		},
		floatToInt: armorCreateRoundFloat32ToInt32_54C950,
	})
}
