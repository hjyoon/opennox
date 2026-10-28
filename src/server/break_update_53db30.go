package server

const breakUpdateDelaySeconds53DB30 = uint32(2)

// BreakUpdate53DB30 restores GAME.EXE 0053DB30 without indexing a
// native-width Object as an array of PE32 words.
func (s *Server) BreakUpdate53DB30(source *Object) {
	if source == nil {
		return
	}
	BreakUpdateRaw53DB30(source, s.stateUpdateRuntime53DB30())
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
	h := s.stateUpdateRuntime53DB30()
	h.DelayedDelete = func(o *Object) {
		if runtime.DelayedDelete != nil {
			runtime.DelayedDelete(o)
		}
	}
	BreakAndRemoveUpdateRaw53DC30(source, h)
}
