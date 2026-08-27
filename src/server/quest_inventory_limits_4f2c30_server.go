package server

import "math"

// questInventoryRoundFloat32ToInt32_4F2C30 models nox_float2int at
// GAME.EXE 00419A70: x87 FISTP under the default round-to-nearest-even mode.
// Invalid and out-of-range conversions produce integer-indefinite INT32_MIN.
func questInventoryRoundFloat32ToInt32_4F2C30(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

// QuestInventoryLimits4F2C30 binds GAME.EXE 004F2C30 to native Server and
// Object values. Object pointers remain native-width while cached PE32 type
// identifiers and inventory counts preserve their exact 32-bit semantics.
//
//go:noinline
func (s *Server) QuestInventoryLimits4F2C30(owner *Object) int32 {
	return questInventoryLimits4F2C30(owner, &s.questInventoryLimits, questInventoryLimitsHooks4F2C30[*Object]{
		objectTypeID: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		isNil: func(owner *Object) bool {
			return owner == nil
		},
		loadClass: func(owner *Object) uint32 {
			return uint32(owner.ObjClass)
		},
		countInventory: func(owner *Object, typeID int32) int32 {
			return owner.CountInventoryWithType(typeID)
		},
		loadBalance: func(key string) float32 {
			return float32(s.Balance.Float(key))
		},
		floatToInt: questInventoryRoundFloat32ToInt32_4F2C30,
	})
}
