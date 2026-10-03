package server

type ItemPreDamageRuntime4E13B0 struct {
	ApplyPreDamage func(*ModifierEff, *Object, *Object, *Object, *int32)
}

func ItemPreDamage4E13B0(target, source, weapon *Object, damage *int32, r ItemPreDamageRuntime4E13B0) int32 {
	return itemPreDamage4E13B0(target, source, weapon, damage, itemPreDamageHooks4E13B0[*Object, *ModifierInitData, *ModifierEff, *int32]{
		initData:       func(obj *Object) *ModifierInitData { return (*ModifierInitData)(obj.InitData) },
		modifier:       func(init *ModifierInitData, slot int) *ModifierEff { return init.Modifiers[slot] },
		hasPreDamage:   func(effect *ModifierEff) bool { return effect.AttackPreDmg64.Fnc != nil },
		applyPreDamage: r.ApplyPreDamage,
	})
}
