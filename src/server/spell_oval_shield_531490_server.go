package server

// SpellOvalShieldRuntime531490 contains the two buff effects still supplied
// by the legacy layer. Object arguments retain their full native pointer.
type SpellOvalShieldRuntime531490 struct {
	ApplyBuff func(*Object, int32, int16, int8)
	BuffOff   func(*Object, int32)
}

func spellOvalShieldServerHooks531490(s *Server, runtime SpellOvalShieldRuntime531490) spellOvalShieldHooks531490[*DurSpell, *Object] {
	return spellOvalShieldHooks531490[*DurSpell, *Object]{
		loadFPS:   s.TickRate,
		loadLevel: func(record *DurSpell) uint32 { return record.Level },
		loadTarget: func(record *DurSpell) *Object {
			return record.Target48
		},
		loadClass: func(target *Object) uint8 { return uint8(target.ObjClass) },
		loadFlags: func(target *Object) uint32 { return uint32(target.ObjFlags) },
		cancelOffensive: func(target *Object) {
			s.Spells.Dur.SpellDurationCancelOffensive4FF310(target)
		},
		applyBuff: runtime.ApplyBuff,
		loadFrame: s.Frame,
		storeFrame: func(record *DurSpell, frame uint32) {
			record.Frame68 = frame
		},
		testBuff: func(target *Object, buff int32) int32 {
			return target.UnitBuffTest4FF350(buff)
		},
		positionDelta: func(target *Object, record *DurSpell) int32 {
			return s.PositionDelta4FEA70(target, &record.Pos)
		},
		buffOff: runtime.BuffOff,
	}
}

// SpellOvalShieldCreate531490 binds the PE32 callback to native-width record
// and object fields.
//
//go:noinline
func (s *Server) SpellOvalShieldCreate531490(record *DurSpell, runtime SpellOvalShieldRuntime531490) int32 {
	return spellOvalShieldCreate531490(record, spellOvalShieldServerHooks531490(s, runtime))
}

// SpellOvalShieldUpdate5314F0 evaluates the target and saved starting point
// without interpreting a native record as a 32-bit byte array.
//
//go:noinline
func (s *Server) SpellOvalShieldUpdate5314F0(record *DurSpell) int32 {
	return spellOvalShieldUpdate5314F0(record, spellOvalShieldServerHooks531490(s, SpellOvalShieldRuntime531490{}))
}

// SpellOvalShieldDestroy531560 removes the shield buff from a non-nil target
// through the native-width buff boundary.
//
//go:noinline
func (s *Server) SpellOvalShieldDestroy531560(record *DurSpell, runtime SpellOvalShieldRuntime531490) {
	spellOvalShieldDestroy531560(record, spellOvalShieldServerHooks531490(s, runtime))
}
