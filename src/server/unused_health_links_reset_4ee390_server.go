package server

// unusedHealthLinksResetNative4EE390 binds the reachable GAME.EXE 004EE390
// paths to native Object and HealthData pointers. The two HealthData link
// fields remain fixed-width ABI32 dwords because this routine only clears
// them; the following list routines are restored separately with
// pointer-width-safe state.
func unusedHealthLinksResetNativeWithState4EE390(state *healthLinksState4EE390, obj *Object) *HealthData {
	result := unusedHealthLinksReset4EE390(obj, unusedHealthLinksResetHooks4EE390[*Object, *HealthData]{
		loadHealth: func(obj *Object) *HealthData {
			return obj.HealthData
		},
		storeHealthPrevious: func(health *HealthData, previous *Object) {
			if previous != nil {
				panic("GAME.EXE 004EE390 attempted a native pointer store into the ABI32 previous link")
			}
			health.field12 = 0
			if state != nil {
				state.storePrevious(health, nil)
			}
		},
		storeHealthNext: func(health *HealthData, next *Object) {
			if next != nil {
				panic("GAME.EXE 004EE390 attempted a native pointer store into the ABI32 next link")
			}
			health.field8 = 0
			if state != nil {
				state.storeNext(health, nil)
			}
		},
		storeAbsoluteNullPrevious: func() {
			panic("GAME.EXE 004EE390 absolute null write at 0x0000000C")
		},
		loadHead: func() *Object {
			panic("GAME.EXE 004EE390 reached bytes after its absolute null write")
		},
		storeObjectPrevious: func(obj, previous *Object) {
			obj.ObjPrev = previous
		},
		storeHead: func(*Object) {
			panic("GAME.EXE 004EE390 reached bytes after its absolute null write")
		},
	})
	if result.kind == unusedHealthLinksResetHealth4EE390 {
		return result.health
	}
	return nil
}

func unusedHealthLinksResetNative4EE390(obj *Object) *HealthData {
	return unusedHealthLinksResetNativeWithState4EE390(nil, obj)
}

// UnusedHealthLinksReset4EE390 exposes the unreferenced original routine
// without inventing a C caller or registration. A non-nil object with nil
// HealthData intentionally panics at the original absolute-null-write point.
func (s *Server) UnusedHealthLinksReset4EE390(obj *Object) *HealthData {
	return unusedHealthLinksResetNativeWithState4EE390(&s.healthLinks, obj)
}
