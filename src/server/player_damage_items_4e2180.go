package server

type playerDamageItemsHooks4E2180[O any] struct {
	loadClass        func(O) uint32
	loadSubclassLow  func(O) uint8
	loadPlayerArmor  func(O) float32
	loadMonsterArmor func(O) float32
	loadFirstItem    func(O) O
	loadFlags        func(O) uint32
	itemArmorValue   func(O) float64
	equipDamage      func(O, O, O, O, float32, int32)
	loadNextItem     func(O) O
}

// playerDamageItems4E2180 preserves GAME.EXE 004E2180. The armor denominator
// is captured once from the live unit update at entry. Each item's armor
// lookup and EquipDamage run before the next inventory link is read; neither
// health nor a zero/signed damage amount gates this loop. EquipDamage owns the
// no-health guard and the modifier/carry/durability order. The target and the
// selected unit update retain the original unguarded dereferences.
func playerDamageItems4E2180[O comparable](
	target, source, effective O,
	damage, typ int32,
	hooks playerDamageItemsHooks4E2180[O],
) {
	class := hooks.loadClass(target)
	var armor float32
	if class&4 != 0 {
		armor = hooks.loadPlayerArmor(target)
	} else {
		if class&2 == 0 || hooks.loadSubclassLow(target)&0x10 == 0 {
			return
		}
		armor = hooks.loadMonsterArmor(target)
	}

	var nilObject O
	for item := hooks.loadFirstItem(target); item != nilObject; item = hooks.loadNextItem(item) {
		if hooks.loadClass(item)&0x02000000 == 0 || hooks.loadFlags(item)&0x100 == 0 {
			continue
		}
		// 004E21E4 divides the lookup result by the cached binary32 armor;
		// 004E21ED multiplies by the signed damage and 004E21F1 spills to f32.
		portion := float32(hooks.itemArmorValue(item) / float64(armor) * float64(damage))
		hooks.equipDamage(item, target, source, effective, portion, typ)
	}
}
