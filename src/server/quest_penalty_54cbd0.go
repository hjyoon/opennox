package server

type questPenaltyHooks54CBD0[O, U, P any] struct {
	loadUpdate                      func(O) U
	loadPlayer                      func(U) P
	loadClass                       func(P) uint8
	getGold                         func(O) uint32
	subGold                         func(O, uint32)
	loseGems, loseWeapon, loseArmor func(O)
	loseSpell, loseGuide            func(O)
	loseAbility                     func(O) int8
}

// questPenalty54CBD0 preserves the ordered loss dispatcher in GAME.EXE
// 0054CBD0. Only the entry update is cached: its Player/class is read after
// the first armor loss. Each repeated helper remains a separate live call.
func questPenalty54CBD0[O, U, P any](unit O, h questPenaltyHooks54CBD0[O, U, P]) {
	update := h.loadUpdate(unit)
	gold := h.getGold(unit)
	h.subGold(unit, gold>>1)
	h.loseGems(unit)
	h.loseWeapon(unit)
	h.loseArmor(unit)
	if h.loadClass(h.loadPlayer(update)) == 0 {
		h.loseArmor(unit)
	}
	h.loseSpell(unit)
	h.loseSpell(unit)
	h.loseGuide(unit)
	h.loseGuide(unit)
	_ = h.loseAbility(unit)
}
