package server

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
// pointer on 64-bit targets. The cooldown uses Field34 (PE32 object+136),
// independently of Frame134 (the damage timestamp at PE32 object+536).
func (s *Server) TriggerUpdate53B1B0(unit *Object, runtime TriggerUpdateRuntime53B1B0) uint8 {
	update := unit.UpdateDataTrigger()
	return triggerUpdateState53B1B0(unit, update, TriggerUpdateRawRuntime53B1B0{
		ImmediateType: runtime.ImmediateType,
		ToggleTriggerRawRuntime53B060: ToggleTriggerRawRuntime53B060{
			Frame: s.Frame, FPS: s.TickRate, MarkAnimFrame: (*Object).MarkAnimFrame,
			CollideTarget: func(o *Object, _ *TriggerUpdateData) *Object { return runtime.CollideTarget(o) },
			AudioEvent:    runtime.AudioEvent, ScriptCallback: runtime.ScriptCallback,
		},
	})
}
