package server

const playerAdjustStaminaPlayerClass4F7DB0 = uint32(4)

type playerAdjustStaminaHooks4F7DB0[O, U, P any] struct {
	loadClass       func(O) uint32
	loadUpdate      func(O) U
	loadAmount      func() uint8
	loadStamina     func(U) uint8
	storeStamina    func(U, uint8)
	loadPlayer      func(U) P
	loadPlayerIndex func(P) uint8
	reportStamina   func(uint8, O)
}

// playerAdjustStamina4F7DB0 preserves GAME.EXE 004F7DB0. The original TEST
// reads only the class low byte. Its Player branch caches UpdateData, reads the
// amount and stamina low bytes, stores their wrapping difference, then reads
// Player and PlayerInd from the cached update before reporting the live unit.
func playerAdjustStamina4F7DB0[O, U, P any](
	unit O,
	hooks playerAdjustStaminaHooks4F7DB0[O, U, P],
) {
	if uint8(hooks.loadClass(unit))&uint8(playerAdjustStaminaPlayerClass4F7DB0) == 0 {
		return
	}
	update := hooks.loadUpdate(unit)
	amount := hooks.loadAmount()
	stamina := hooks.loadStamina(update)
	hooks.storeStamina(update, stamina-amount)
	player := hooks.loadPlayer(update)
	index := hooks.loadPlayerIndex(player)
	hooks.reportStamina(index, unit)
}
