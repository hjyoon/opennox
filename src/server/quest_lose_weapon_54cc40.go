package server

type questLoseWeaponHooks54CC40[O comparable, D, A any, M comparable] struct {
	updateData    func(O) D
	first, next   func(O) O
	flags, class  func(O) uint32
	subclass      func(O) uint32
	modifierData  func(O) A
	modifier      func(A, int) M
	playerClass   func(D) uint8
	classCanUse   func(O, uint8) int32
	delayedDelete func(O)
}

// questLoseWeapon54CC40 preserves GAME.EXE 0054CC40, including its cached
// player update, first eligible equipped weapon, all four modifier reads,
// and complete live spare-weapon walk. Only an exact CanUseItem result of
// one admits a spare; the decompiled C's nonzero comparison is not the oracle.
func questLoseWeapon54CC40[O comparable, D, A any, M comparable](unit O, h questLoseWeaponHooks54CC40[O, D, A, M]) {
	var zero O
	update := h.updateData(unit)
	weapon := h.first(unit)
	for weapon != zero {
		if h.flags(weapon)&0x100 != 0 && h.class(weapon)&0x1001000 != 0 && uint8(h.subclass(weapon))&2 == 0 {
			break
		}
		weapon = h.next(weapon)
	}
	if weapon == zero {
		return
	}
	subclass := h.subclass(weapon)
	if subclass&0x10000 != 0 {
		return
	}
	if subclass&0x104 != 0 {
		data := h.modifierData(weapon)
		var nilModifier M
		plain := true
		for slot := 0; slot != 4; slot++ {
			if h.modifier(data, slot) != nilModifier {
				plain = false
			}
		}
		if plain {
			return
		}
	}
	foundSpare := false
	for item := h.first(unit); item != zero; item = h.next(item) {
		if h.class(item)&0x1001000 != 0 && h.flags(item)&0x100 == 0 && h.classCanUse(item, h.playerClass(update)) == 1 {
			foundSpare = true
		}
	}
	if foundSpare {
		h.delayedDelete(weapon)
	}
}
