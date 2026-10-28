package server

import (
	"github.com/opennox/libs/object"
)

// ToggleTriggerRawRuntime53B060 keeps native object/script identities while
// leaving the shared wire/template record at its original 60-byte size.
type ToggleTriggerRawRuntime53B060 struct {
	Frame, FPS     func() uint32
	MarkAnimFrame  func(*Object, uint32)
	AudioEvent     func(uint32, *Object)
	CollideTarget  func(*Object, *TriggerUpdateData) *Object
	ScriptCallback func(*ScriptCallback, *Object, *Object, ScriptEventType)
}

func ToggleUpdateRaw53B060(unit *Object, h ToggleTriggerRawRuntime53B060) uint8 {
	flags := unit.ObjFlags
	return toggleUpdateState53B060(unit, (*ToggleUpdateData)(unit.UpdateData), flags, h)
}

func toggleUpdateState53B060(unit *Object, update *ToggleUpdateData, flags object.Flags, h ToggleTriggerRawRuntime53B060) uint8 {
	if !flags.Has(object.FlagEnabled) {
		update.Flags &^= 9
		return uint8(update.Flags)
	}
	if update.Flags&8 == 0 {
		update.State = 0
		unit.Field34 = h.Frame()
		h.MarkAnimFrame(unit, 0)
	}
	switch update.State {
	case 0:
		if h.Frame() > unit.Field34 && update.Flags&1 != 0 {
			h.AudioEvent(update.SoundActivate, unit)
			h.MarkAnimFrame(unit, 1)
			h.ScriptCallback(&update.ScriptActivate, h.CollideTarget(unit, update), unit, NoxEventToggleXXX)
			update.State = 3
			frame, fps := h.Frame(), h.FPS()
			unit.Field34 = frame + fps
		}
	case 1:
		if h.Frame() > unit.Field34 && update.Flags&1 != 0 {
			h.AudioEvent(update.SoundDeactivate, unit)
			h.MarkAnimFrame(unit, 0)
			h.ScriptCallback(&update.ScriptDeactivate, nil, unit, NoxEventToggleYYY)
			if update.Flags&2 != 0 {
				update.State = 5
			} else {
				update.State = 0
			}
			frame, fps := h.Frame(), h.FPS()
			unit.Field34 = frame + fps
		}
	case 3:
		if h.Frame() > unit.Field34 && update.Flags&1 == 0 {
			update.State = 1
		}
	}
	word := update.Flags
	if word&1 != 0 {
		word |= 4
	} else {
		word &^= 4
	}
	update.Flags = word
	update.Flags = word&^1 | 8
	return uint8(word)
}

type TriggerUpdateRawRuntime53B1B0 struct {
	ToggleTriggerRawRuntime53B060
	ImmediateType func(*Object) bool
}

func TriggerUpdateRaw53B1B0(unit *Object, h TriggerUpdateRawRuntime53B1B0) uint8 {
	return triggerUpdateState53B1B0(unit, (*TriggerUpdateData)(unit.UpdateData), h)
}

func triggerUpdateState53B1B0(unit *Object, update *TriggerUpdateData, h TriggerUpdateRawRuntime53B1B0) uint8 {
	delay := uint32(0)
	if !h.ImmediateType(unit) {
		delay = h.FPS()
	}
	if !unit.ObjFlags.Has(object.FlagEnabled) {
		update.Flags &^= 9
		return uint8(update.Flags)
	}
	if update.Flags&8 == 0 {
		update.State = 0
		unit.Field34 = h.Frame()
		h.MarkAnimFrame(unit, 0)
	}
	switch update.State {
	case 0:
		if update.Flags&1 != 0 {
			h.AudioEvent(update.SoundActivate, unit)
			h.MarkAnimFrame(unit, 1)
			h.ScriptCallback(&update.ScriptActivate, h.CollideTarget(unit, update), unit, NoxEventTriggerActivated)
			update.State = 1
			unit.Field34 = h.Frame() + delay
		}
	case 1:
		if h.Frame() > unit.Field34 && update.Flags&1 == 0 {
			h.AudioEvent(update.SoundDeactivate, unit)
			h.MarkAnimFrame(unit, 0)
			h.ScriptCallback(&update.ScriptDeactivate, nil, unit, NoxEventTriggerDeactivated)
			if update.Flags&2 != 0 {
				update.State = 5
			} else {
				update.State = 0
			}
		}
	}
	word := update.Flags
	if word&1 != 0 {
		word |= 4
	} else {
		word &^= 4
	}
	update.Flags = word
	word = word&^1 | 8
	update.Flags = word
	return uint8(word)
}

// TriggerImmediateTypeRaw53B1B0 preserves the two original shared DWORD type
// caches. A zero Trigger cache refreshes both; comparison uses full indices.
type TriggerTypeCacheRuntime53B1B0 struct {
	Load   func(int) uint32
	Store  func(int, uint32)
	Lookup func(string) uint32
}

func TriggerImmediateTypeRaw53B1B0(unit *Object, h TriggerTypeCacheRuntime53B1B0) bool {
	if h.Load(0) == 0 {
		trigger := h.Lookup("Trigger")
		h.Store(0, trigger)
		plate := h.Lookup("PressurePlate")
		h.Store(1, plate)
	}
	index := uint32(unit.TypeInd)
	return index == h.Load(0) || index == h.Load(1)
}
