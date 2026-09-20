package server

const breakUpdateDelaySeconds53DB30 = uint32(2)

// BreakUpdate53DB30 restores GAME.EXE 0053DB30 without indexing a
// native-width Object as an array of PE32 words.
func (s *Server) BreakUpdate53DB30(source *Object) {
	if source == nil {
		return
	}
	switch {
	case source.Field5&0x2 != 0:
		if uint32(source.ObjFlags)&0x8000 != 0 {
			source.UnsetXStatus(0x2)
			source.SetXStatus(0x4)
			source.ObjFlags |= 0x40
			source.Field34 = s.Frame() + breakUpdateDelaySeconds53DB30*s.TickRate()
		}
	case source.Field5&0x4 != 0:
		if s.Frame() > source.Field34 {
			source.UnsetXStatus(0x4)
			source.SetXStatus(0x8)
		}
	case source.Field5&0x8 != 0:
		s.Objs.RemoveFromUpdatable(source)
	}
}

// BreakAndRemoveUpdateRuntime53DC30 supplies the deletion service used after
// the exact terminal state is reached.
type BreakAndRemoveUpdateRuntime53DC30 struct {
	DelayedDelete func(*Object)
}

// BreakAndRemoveUpdate53DC30 restores GAME.EXE 0053DC30. Unlike BreakUpdate,
// the original dispatches on the complete status word rather than testing its
// low bits in priority order.
func (s *Server) BreakAndRemoveUpdate53DC30(source *Object, runtime BreakAndRemoveUpdateRuntime53DC30) {
	if source == nil {
		return
	}
	switch source.Field5 {
	case 0x2:
		if uint32(source.ObjFlags)&0x8000 != 0 {
			source.UnsetXStatus(0x2)
			source.SetXStatus(0x4)
			source.ObjFlags |= 0x40
			source.Field34 = s.Frame() + breakUpdateDelaySeconds53DB30*s.TickRate()
		}
	case 0x4:
		if s.Frame() > source.Field34 {
			source.UnsetXStatus(0x4)
			source.SetXStatus(0x8)
		}
	case 0x8:
		s.Objs.RemoveFromUpdatable(source)
		if runtime.DelayedDelete != nil {
			runtime.DelayedDelete(source)
		}
	}
}
