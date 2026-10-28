package server

import "github.com/opennox/libs/types"

// ArachnaphobiaRawRuntime53DA60 keeps the original shared DWORD type cache
// separate from native object pointers. Services are resolved only if reached.
type ArachnaphobiaRawRuntime53DA60 struct {
	Frame, FPS         func() uint32
	LoadTypeIndex      func() uint32
	StoreTypeIndex     func(uint32)
	TypeIndexByID      func(string) uint32
	NewObjectByTypeInd func(int32) *Object
	CreateAt           func(*Object, *Object, types.Pointf)
	RandomInt          func(int, int) int
	DelayedDelete      func(*Object)
}

// ArachnaphobiaUpdateRaw53DA60 is the required-pointer C entry contract.
func ArachnaphobiaUpdateRaw53DA60(source *Object, h ArachnaphobiaRawRuntime53DA60) {
	ageFrame := h.Frame()
	if source.Field34 < ageFrame {
		index := h.LoadTypeIndex()
		if index == 0 {
			index = h.TypeIndexByID("SmallSpider")
			h.StoreTypeIndex(index)
		}
		spider := h.NewObjectByTypeInd(int32(index))
		if spider != nil {
			h.CreateAt(spider, source.ObjOwner, source.PosVec)
		}
		delay := uint32(h.RandomInt(1, 5))
		source.Field34 = h.Frame() + delay
		ageFrame = h.Frame()
	}
	if ageFrame-source.Field32 > arachnaphobiaLifetimeSeconds53DA60*h.FPS() {
		h.DelayedDelete(source)
	}
}

// StateUpdateRawRuntime53DB30 exposes the original status/list services.
// Production adapters use Object.SetXStatus/UnsetXStatus, including sync work.
type StateUpdateRawRuntime53DB30 struct {
	Frame, FPS          func() uint32
	UnsetXStatus        func(*Object, uint32)
	SetXStatus          func(*Object, uint32)
	RemoveFromUpdatable func(*Object)
	DelayedDelete       func(*Object)
}

func BreakUpdateRaw53DB30(source *Object, h StateUpdateRawRuntime53DB30) {
	stateUpdateRaw53DB30(source, 0, h)
}

func OpenUpdateRaw53DBB0(source *Object, h StateUpdateRawRuntime53DB30) {
	stateUpdateRaw53DB30(source, 1, h)
}

func BreakAndRemoveUpdateRaw53DC30(source *Object, h StateUpdateRawRuntime53DB30) {
	stateUpdateRaw53DB30(source, 2, h)
}

func stateUpdateRaw53DB30(source *Object, kind int, h StateUpdateRawRuntime53DB30) {
	status := source.Field5
	if kind != 2 {
		switch {
		case status&2 != 0:
			status = 2
		case status&4 != 0:
			status = 4
		case status&8 != 0:
			status = 8
		default:
			return
		}
	}
	switch status {
	case 2:
		if source.ObjFlags&0x8000 == 0 {
			return
		}
		h.UnsetXStatus(source, 2)
		h.SetXStatus(source, 4)
		frame, fps := h.Frame(), h.FPS()
		if kind == 0 {
			flags := source.ObjFlags | 0x40
			source.Field34 = frame + breakUpdateDelaySeconds53DB30*fps
			source.ObjFlags = flags
		} else {
			source.Field34 = frame + breakUpdateDelaySeconds53DB30*fps
			if kind == 2 {
				source.ObjFlags |= 0x40
			}
		}
	case 4:
		if h.Frame() > source.Field34 {
			h.UnsetXStatus(source, 4)
			h.SetXStatus(source, 8)
		}
	case 8:
		h.RemoveFromUpdatable(source)
		if kind == 2 {
			h.DelayedDelete(source)
		}
	}
}

func (s *Server) stateUpdateRuntime53DB30() StateUpdateRawRuntime53DB30 {
	return StateUpdateRawRuntime53DB30{
		Frame: s.Frame, FPS: s.TickRate,
		UnsetXStatus: (*Object).UnsetXStatus, SetXStatus: (*Object).SetXStatus,
		RemoveFromUpdatable: s.Objs.RemoveFromUpdatable,
	}
}
