package server

import "github.com/opennox/libs/object"

// TrapDoorUpdateRawRuntime53DE80 supplies the original reached services.
// Status/on-off production adapters retain Object's synchronization work.
type TrapDoorUpdateRawRuntime53DE80 struct {
	Frame, FPS   func() uint32
	SetOnOff     func(*Object, bool)
	UnsetXStatus func(*Object, uint32)
	SetXStatus   func(*Object, uint32)
	AudioEvent   func(uint32, *Object)
}

// TrapDoorUpdateRaw53DE80 preserves the required-pointer C entry. The enabled
// opening branch never dereferences the cached collision record, even if nil.
func TrapDoorUpdateRaw53DE80(unit *Object, h TrapDoorUpdateRawRuntime53DE80) {
	flags := unit.ObjFlags
	data := (*TrapDoorCollideData)(unit.CollideData)
	if flags.Has(object.FlagEnabled) {
		status := uint8(unit.Field5)
		switch {
		case status&2 != 0:
			h.UnsetXStatus(unit, 2)
			h.SetXStatus(unit, 8)
		case status&4 != 0:
			frame := h.Frame()
			if frame >= data.NextFrame {
				h.UnsetXStatus(unit, 4)
				h.SetXStatus(unit, 8)
			}
		default:
			data.Activated = 0
		}
		return
	}
	deadline := data.NextFrame
	if deadline == 0 || h.Frame() < deadline {
		return
	}
	h.SetOnOff(unit, true)
	h.UnsetXStatus(unit, 2)
	h.SetXStatus(unit, 4)
	fps := h.FPS()
	data.NextFrame = data.NextFrame + 5*fps
	h.AudioEvent(874, unit)
}
