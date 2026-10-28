package server

// UndeadKillerUpdate53E190 follows GAME.EXE 0053E190 using the native-width
// collision-data spell pointer. Its frame subtraction retains dword wrapping.
func UndeadKillerUpdate53E190(obj *Object, frame uint32, delayedDelete func(*Object)) {
	data := (*UndeadKillerCollideData)(obj.CollideData)
	if data == nil {
		data = new(UndeadKillerCollideData)
	}
	undeadKillerUpdate53E190(obj, data, UndeadKillerUpdateRawRuntime53E190{
		Frame: func() uint32 { return frame }, DelayedDelete: delayedDelete,
	})
}

type UndeadKillerUpdateRawRuntime53E190 struct {
	Frame         func() uint32
	DelayedDelete func(*Object)
}

// UndeadKillerUpdateRaw53E190 requires the original collision record; its
// spell pointer may be nil. Cancellation precedes the wrapping age test.
func UndeadKillerUpdateRaw53E190(obj *Object, h UndeadKillerUpdateRawRuntime53E190) {
	undeadKillerUpdate53E190(obj, (*UndeadKillerCollideData)(obj.CollideData), h)
}

func undeadKillerUpdate53E190(obj *Object, data *UndeadKillerCollideData, h UndeadKillerUpdateRawRuntime53E190) {
	record := data.Spell
	if record != nil && uint8(record.Flags88)&1 != 0 {
		h.DelayedDelete(obj)
		return
	}
	if h.Frame()-obj.Field34 > 70 {
		h.DelayedDelete(obj)
	}
}
