package server

import "github.com/opennox/libs/object"

type CastShockRuntime52C5A0 struct {
	GlyphTypeCache  *uint32
	LookupTypeID    func(string) uint32
	BalanceFloat    func(string) float64
	BalanceFloatInd func(string, int32) float64
	BuffApply       func(*Object, int32, int16, int8)
}

// CastShock52C5A0 uses native SpellAcceptArg.Obj, Object.TypeInd and Damage
// links instead of packed PE32 offsets. Runtime services are not consulted
// before the original target gate. Missing required arg/target/callback links
// still fault; known damage callbacks use the native-width registry, whose
// existing policy rejects unrestored PE32 callbacks on wider hosts.
func CastShock52C5A0(caster, context *Object, arg *SpellAcceptArg, power int32, r CastShockRuntime52C5A0) int32 {
	return castShock52C5A0(caster, context, power, castShockHooks52C5A0[*Object]{
		target:         func() *Object { return arg.Obj },
		loadGlyphType:  func() uint32 { return *r.GlyphTypeCache },
		storeGlyphType: func(id uint32) { *r.GlyphTypeCache = id },
		lookupType:     r.LookupTypeID,
		typeIndex:      func(unit *Object) uint16 { return unit.TypeInd },
		balance:        r.BalanceFloat,
		balanceIndex:   r.BalanceFloatInd,
		floatToInt:     aiPathFloatToInt419A70,
		damage: func(target, source, weapon *Object, damage, typ int32) bool {
			return objDamage.Get(target.Damage)(target, source, weapon, damage, object.DamageType(typ))
		},
		apply: r.BuffApply,
	})
}
