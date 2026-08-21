package server

func directionInitNative4F0490(unit *Object) int32 {
	return directionInit4F0490(unit, directionInitHooks4F0490[*Object, *DirectionInitData]{
		loadInitData: func(unit *Object) *DirectionInitData {
			return (*DirectionInitData)(unit.InitData)
		},
		directionToAngle: directionToAngleNative509E00,
		storeDirection2: func(unit *Object, angle uint16) {
			unit.Direction2 = Dir16(angle)
		},
		storeDirection1: func(unit *Object, angle uint16) {
			unit.Direction1 = Dir16(angle)
		},
	})
}

// DirectionInit4F0490 binds GAME.EXE 004F0490 to a native-width Object and
// the fixed-width DirectionInitData record. There are deliberately no nil
// guards.
//
//go:noinline
func DirectionInit4F0490(unit *Object) int32 {
	return directionInitNative4F0490(unit)
}
