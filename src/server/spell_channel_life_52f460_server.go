package server

import "math"

// SpellChannelLifeRuntime52F460 supplies the two legacy-owned effects and the
// level-indexed balance lookup. Every object passed to an effect is native-width.
type SpellChannelLifeRuntime52F460 struct {
	AddMana     func(*Object, int16)
	ClearDamage func(*Object, int32)
	Coefficient func(uint32) float64
}

// SpellChannelLifeUpdate52F460 runs the Channel Life duration callback without
// interpreting a native DurSpell or Object through PE32 byte offsets.
//
//go:noinline
func (s *Server) SpellChannelLifeUpdate52F460(record *DurSpell, runtime SpellChannelLifeRuntime52F460) int32 {
	return spellChannelLifeUpdate52F460(record, channelLifeHooks52F460[*DurSpell, *Object]{
		loadTarget:  func(record *DurSpell) *Object { return record.Target48 },
		loadFlags:   func(target *Object) uint32 { return uint32(target.ObjFlags) },
		loadMode:    func(record *DurSpell) uint32 { return record.Flag20 },
		addMana:     runtime.AddMana,
		clearDamage: runtime.ClearDamage,
		loadCaster:  func(record *DurSpell) *Object { return record.Caster16 },
		testBuff:    func(caster *Object, buff int32) int32 { return caster.UnitBuffTest4FF350(buff) },
		loadClass:   func(target *Object) uint8 { return uint8(target.ObjClass) },
		positionDelta: func(target *Object, record *DurSpell) int32 {
			return s.PositionDelta4FEA70(target, &record.Pos)
		},
		maxMana:      PlayerGetMaxMana4EECB0,
		currentMana:  UnitGetOldMana4EEC80,
		getHP:        UnitGetHP4EE780,
		loadFraction: func(record *DurSpell) float32 { return math.Float32frombits(uint32(record.Field72)) },
		loadLevel:    func(record *DurSpell) uint32 { return record.Level },
		coefficient:  runtime.Coefficient,
		storeFraction: func(record *DurSpell, fraction float32) {
			record.Field72 = int32(math.Float32bits(fraction))
		},
	})
}
