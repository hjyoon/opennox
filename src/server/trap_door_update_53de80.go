package server

type TrapDoorUpdateRuntime53DE80 struct {
	AudioEvent func(uint32, *Object)
}

// TrapDoorUpdate53DE80 restores GAME.EXE 0053DE80 without reading the
// native-width CollideData pointer through Object's original PE32 slot.
func (s *Server) TrapDoorUpdate53DE80(unit *Object, runtime TrapDoorUpdateRuntime53DE80) {
	if unit == nil || unit.CollideData == nil {
		return
	}
	TrapDoorUpdateRaw53DE80(unit, TrapDoorUpdateRawRuntime53DE80{
		Frame: s.Frame, FPS: s.TickRate,
		SetOnOff:     (*Object).SetOnOff,
		UnsetXStatus: (*Object).UnsetXStatus, SetXStatus: (*Object).SetXStatus,
		AudioEvent: func(id uint32, o *Object) {
			if runtime.AudioEvent != nil {
				runtime.AudioEvent(id, o)
			}
		},
	})
}
