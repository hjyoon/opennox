package server

// SelfDestructUpdateRuntime53CC90 supplies the deletion service used once an
// object's unsigned age exceeds two server ticks.
type SelfDestructUpdateRuntime53CC90 struct {
	DelayedDelete func(*Object)
}

// SelfDestructUpdate53CC90 restores GAME.EXE 0053CC90 without reading the
// native-width Object through PE32 field offsets. The strict comparison and
// uint32 subtraction preserve the original three-tick lifetime and wraparound.
func (s *Server) SelfDestructUpdate53CC90(source *Object, runtime SelfDestructUpdateRuntime53CC90) {
	if source == nil || runtime.DelayedDelete == nil {
		return
	}
	LifetimeEffectUpdateRaw53CC90(LifetimeEffectSelfDestruct53CC90, source, LifetimeEffectRuntime53CC90{Frame: s.Frame, DelayedDelete: runtime.DelayedDelete})
}
