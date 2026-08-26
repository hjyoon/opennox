package server

import "github.com/opennox/libs/object"

type ToggleUpdateRuntime53B060 struct {
	CollideTarget  func(*Object) *Object
	AudioEvent     func(uint32, *Object)
	ScriptCallback func(*ScriptCallback, *Object, *Object, ScriptEventType)
}

// ToggleUpdate53B060 ports GAME.EXE 0053B060 to native Object pointers while
// retaining the original fixed-width 60-byte update record.
func (s *Server) ToggleUpdate53B060(unit *Object, runtime ToggleUpdateRuntime53B060) uint8 {
	update := unit.UpdateDataToggle()
	if !unit.ObjFlags.Has(object.FlagEnabled) {
		update.Flags &^= 0x9
		return uint8(update.Flags)
	}
	if update.Flags&0x8 == 0 {
		update.State = 0
		unit.Frame134 = s.Frame()
		unit.MarkAnimFrame(0)
	}
	switch update.State {
	case 0:
		if s.Frame() > unit.Frame134 && update.Flags&0x1 != 0 {
			runtime.AudioEvent(update.SoundActivate, unit)
			unit.MarkAnimFrame(1)
			runtime.ScriptCallback(&update.ScriptActivate, runtime.CollideTarget(unit), unit, NoxEventToggleXXX)
			update.State = 3
			unit.Frame134 = s.Frame() + s.TickRate()
		}
	case 1:
		if s.Frame() > unit.Frame134 && update.Flags&0x1 != 0 {
			runtime.AudioEvent(update.SoundDeactivate, unit)
			unit.MarkAnimFrame(0)
			runtime.ScriptCallback(&update.ScriptDeactivate, nil, unit, NoxEventToggleYYY)
			if update.Flags&0x2 != 0 {
				update.State = 5
			} else {
				update.State = 0
			}
			unit.Frame134 = s.Frame() + s.TickRate()
		}
	case 3:
		if s.Frame() > unit.Frame134 && update.Flags&0x1 == 0 {
			update.State = 1
		}
	}
	flags := update.Flags
	if flags&0x1 != 0 {
		flags |= 0x4
	} else {
		flags &^= 0x4
	}
	update.Flags = flags&^0x1 | 0x8
	return uint8(flags)
}
