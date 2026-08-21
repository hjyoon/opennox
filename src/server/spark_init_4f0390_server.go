package server

// SparkInit4F0390 binds GAME.EXE 004F0390 to native-width Object and
// SparkUpdateData pointers while retaining the original fixed-width fields.
// There are deliberately no nil guards.
//
//go:noinline
func SparkInit4F0390(unit *Object) *SparkUpdateData {
	return sparkInit4F0390(unit, sparkInitHooks4F0390[*Object, *SparkUpdateData]{
		loadUpdateData: func(obj *Object) *SparkUpdateData {
			return (*SparkUpdateData)(obj.UpdateData)
		},
		storeLifetimeRemaining: func(update *SparkUpdateData, value uint32) {
			update.LifetimeRemaining = value
		},
		storeLifetimeInitial: func(update *SparkUpdateData, value uint32) {
			update.LifetimeInitial = value
		},
	})
}
