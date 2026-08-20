package server

import "unsafe"

func playerHPInitNative4EE730(unit *Object) {
	playerHPInit4EE730(playerHPInitHooks4EE730[*Object, *HealthData, *PlayerUpdateData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		storeSample: func(update *PlayerUpdateData, index int, value uint16) {
			update.HealthSamples[index] = value
		},
		storeCurrentSample: func(update *PlayerUpdateData, value uint16) {
			update.HealthSampleCur = value
		},
	})
}

// PlayerHPInit4EE730 initializes the native-width Player update record's HP
// samples with the exact live HealthData reload behavior of GAME.EXE 004EE730.
func PlayerHPInit4EE730(unit *Object) {
	playerHPInitNative4EE730(unit)
}

var (
	_ = [1]struct{}{}[12-unsafe.Offsetof(PlayerUpdateData{}.HealthSamples)]
	_ = [1]struct{}{}[64-unsafe.Sizeof(PlayerUpdateData{}.HealthSamples)]
	_ = [1]struct{}{}[76-unsafe.Offsetof(PlayerUpdateData{}.HealthSampleCur)]
)
