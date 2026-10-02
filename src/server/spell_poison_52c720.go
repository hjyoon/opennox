package server

type poisonCastHooks52C720[O comparable] struct {
	loadTarget              func() O
	activatePoison          func(O, int32, int32) int32
	recordPlayerAttribution func(O, O)
}

// castPoison52C720 restores GAME.EXE 0052C720..0052C74F. The entry target
// is passed to activation with signed spell power as both increment and cap.
// Discard that service's result, reload the target for player attribution,
// and return one even if poison was resisted or otherwise rejected. The
// spell ID, second object, fourth object and argument position are unused.
func castPoison52C720[O comparable](owner O, level int32, h poisonCastHooks52C720[O]) int32 {
	target := h.loadTarget()
	var nilObject O
	if target == nilObject {
		return 0
	}
	h.activatePoison(target, level, level)
	h.recordPlayerAttribution(owner, h.loadTarget())
	return 1
}

type PoisonCastRuntime52C720 struct {
	ActivatePoison          func(*Object, int32, int32) int32
	RecordPlayerAttribution func(*Object, *Object)
}

// CastPoison52C720 retains the instant-spell selector signature. Missing
// arg still faults on its first target load; only a nil target returns zero.
// No class, flag, spell-level or helper-result gate is added to this cast.
func (*Server) CastPoison52C720(_ int32, _ *Object, owner *Object, _ *Object, arg *SpellAcceptArg, level int32, runtime PoisonCastRuntime52C720) int32 {
	return castPoison52C720(owner, level, poisonCastHooks52C720[*Object]{
		loadTarget:              func() *Object { return arg.Obj },
		activatePoison:          runtime.ActivatePoison,
		recordPlayerAttribution: runtime.RecordPlayerAttribution,
	})
}
