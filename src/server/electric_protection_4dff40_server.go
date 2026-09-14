package server

import "unsafe"

const electricProtectionBalanceKey4DFF40 = "ElectricitySpellProtection"

// ElectricProtectionRuntime4DFF40 carries the original LightningProtectEngage
// callback identity. Protection sums matching modifiers without calling it.
type ElectricProtectionRuntime4DFF40 struct {
	ElectricProtectEngage unsafe.Pointer
}

// ElectricProtection4DFF40 restores GAME.EXE 004DFF40 on native-width objects.
func (s *Server) ElectricProtection4DFF40(unit *Object, runtime ElectricProtectionRuntime4DFF40) float64 {
	return elementalProtectionNative4DFE40(unit, runtime.ElectricProtectEngage,
		func(key string, index int32) float64 { return s.Balance.FloatInd(key, int(index)) },
		ENCHANT_PROTECT_FROM_ELECTRICITY, electricProtectionBalanceKey4DFF40)
}
