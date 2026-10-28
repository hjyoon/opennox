package server

type ToggleUpdateRuntime53B060 struct {
	CollideTarget  func(*Object) *Object
	AudioEvent     func(uint32, *Object)
	ScriptCallback func(*ScriptCallback, *Object, *Object, ScriptEventType)
}

// ToggleUpdate53B060 ports GAME.EXE 0053B060 to native Object pointers while
// retaining the original fixed-width 60-byte update record. Its cooldown
// belongs to Field34 (PE32 object+136), not the separate damage timestamp.
func (s *Server) ToggleUpdate53B060(unit *Object, runtime ToggleUpdateRuntime53B060) uint8 {
	update := unit.UpdateDataToggle()
	return toggleUpdateState53B060(unit, update, unit.ObjFlags, ToggleTriggerRawRuntime53B060{
		Frame: s.Frame, FPS: s.TickRate, MarkAnimFrame: (*Object).MarkAnimFrame,
		CollideTarget: func(o *Object, _ *TriggerUpdateData) *Object { return runtime.CollideTarget(o) },
		AudioEvent:    runtime.AudioEvent, ScriptCallback: runtime.ScriptCallback,
	})
}
