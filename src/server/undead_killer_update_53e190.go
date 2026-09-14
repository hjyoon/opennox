package server

// UndeadKillerUpdate53E190 follows GAME.EXE 0053E190 using the native-width
// collision-data spell pointer. Its frame subtraction retains dword wrapping.
func UndeadKillerUpdate53E190(obj *Object, frame uint32, delayedDelete func(*Object)) {
	var record *DurSpell
	if obj.CollideData != nil {
		record = (*UndeadKillerCollideData)(obj.CollideData).Spell
	}
	if (record != nil && record.Flags88&1 != 0) || frame-obj.Field34 > 70 {
		delayedDelete(obj)
	}
}
