package server

import (
	"github.com/opennox/libs/object"
)

// TriggerUpdateRuntime53B1B0 supplies the engine services that surround the
// fixed-width TriggerUpdateData state machine. Keeping them explicit makes the
// original callback order independently testable without C pointers.
type TriggerUpdateRuntime53B1B0 struct {
	ImmediateType  func(*Object) bool
	CollideTarget  func(*Object) *Object
	AudioEvent     func(uint32, *Object)
	ScriptCallback func(*ScriptCallback, *Object, *Object, ScriptEventType)
}

// TriggerUpdate53B1B0 ports GAME.EXE 0053B1B0 to native Object pointers.
// TriggerUpdateData remains the original 60-byte record; its transient object
// pointer is supplied by CollideTarget because PE32 Field4 cannot hold a native
// pointer on 64-bit targets.
func (s *Server) TriggerUpdate53B1B0(unit *Object, runtime TriggerUpdateRuntime53B1B0) uint8 {
	update := unit.UpdateDataTrigger()
	delay := s.TickRate()
	if runtime.ImmediateType(unit) {
		delay = 0
	}
	if unit.ObjFlags.Has(object.FlagEnabled) {
		if update.Flags&0x8 == 0 {
			update.State = 0
			unit.Frame134 = s.Frame()
			unit.MarkAnimFrame(0)
		}
		if update.State != 0 {
			if update.State == 1 && s.Frame() > unit.Frame134 && update.Flags&0x1 == 0 {
				runtime.AudioEvent(update.SoundDeactivate, unit)
				unit.MarkAnimFrame(0)
				runtime.ScriptCallback(&update.ScriptDeactivate, nil, unit, NoxEventTriggerDeactivated)
				if update.Flags&0x2 != 0 {
					update.State = 5
				} else {
					update.State = 0
				}
			}
		} else if update.Flags&0x1 != 0 {
			runtime.AudioEvent(update.SoundActivate, unit)
			unit.MarkAnimFrame(1)
			runtime.ScriptCallback(&update.ScriptActivate, runtime.CollideTarget(unit), unit, NoxEventTriggerActivated)
			update.State = 1
			unit.Frame134 = delay + s.Frame()
		}
		flags := update.Flags
		if flags&0x1 != 0 {
			flags |= 0x4
		} else {
			flags &^= 0x4
		}
		update.Flags = flags&^0x1 | 0x8
	} else {
		update.Flags &^= 0x9
	}
	return uint8(update.Flags)
}
