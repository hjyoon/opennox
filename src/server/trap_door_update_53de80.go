package server

import "github.com/opennox/libs/object"

type TrapDoorUpdateRuntime53DE80 struct {
	AudioEvent func(uint32, *Object)
}

// TrapDoorUpdate53DE80 restores GAME.EXE 0053DE80 without reading the
// native-width CollideData pointer through Object's original PE32 slot.
func (s *Server) TrapDoorUpdate53DE80(unit *Object, runtime TrapDoorUpdateRuntime53DE80) {
	if unit == nil || unit.CollideData == nil {
		return
	}
	data := (*TrapDoorCollideData)(unit.CollideData)
	if unit.ObjFlags.Has(object.FlagEnabled) {
		switch {
		case uint8(unit.Field5)&0x2 != 0:
			unit.UnsetXStatus(0x2)
			unit.SetXStatus(0x8)
		case uint8(unit.Field5)&0x4 != 0:
			if s.Frame() >= data.NextFrame {
				unit.UnsetXStatus(0x4)
				unit.SetXStatus(0x8)
			}
		default:
			data.Activated = 0
		}
		return
	}
	if data.NextFrame == 0 || s.Frame() < data.NextFrame {
		return
	}
	unit.SetOnOff(true)
	unit.UnsetXStatus(0x2)
	unit.SetXStatus(0x4)
	data.NextFrame += 5 * s.TickRate()
	if runtime.AudioEvent != nil {
		runtime.AudioEvent(874, unit)
	}
}
