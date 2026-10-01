package server

type CastStunRuntime52C2C0 struct {
	BuffApply   func(*Object, int32, int16, int8)
	Attribution func(*Object, *Object)
}

// CastStun52C2C0 follows native SpellAcceptArg and Player links without
// interpreting their packed PE32 offsets. Invalid required links still fault;
// the nil-tolerant PlayerClass accessor would incorrectly choose Warrior.
func (s *Server) CastStun52C2C0(caster *Object, arg *SpellAcceptArg, power int32, r CastStunRuntime52C2C0) int32 {
	return castStun52C2C0(caster, power, castStunHooks52C2C0[*Object]{
		target:      func() *Object { return arg.Obj },
		balance:     s.Balance.Float,
		floatToInt:  aiPathFloatToInt419A70,
		classLow:    func(unit *Object) uint8 { return uint8(unit.ObjClass) },
		playerClass: func(unit *Object) uint8 { return (*PlayerUpdateData)(unit.UpdateData).Player.Info().playerClass },
		mass:        func(unit *Object) float32 { return unit.Mass },
		apply:       r.BuffApply,
		attribution: r.Attribution,
	})
}
