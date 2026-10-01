package server

type CastConfuseRuntime52C1E0 struct {
	BuffApply   func(*Object, int32, int16, int8)
	Attribution func(*Object, *Object)
}

// CastConfuse52C1E0 reads the host-width SpellAcceptArg.Obj, not the PE32
// first-dword target. It does not add a null-argument guard to the original.
func (s *Server) CastConfuse52C1E0(caster *Object, arg *SpellAcceptArg, power int32, r CastConfuseRuntime52C1E0) int32 {
	return castConfuse52C1E0(caster, power, castConfuseHooks52C1E0[*Object]{
		target:      func() *Object { return arg.Obj },
		balance:     s.Balance.Float,
		floatToInt:  aiPathFloatToInt419A70,
		apply:       r.BuffApply,
		attribution: r.Attribution,
	})
}
