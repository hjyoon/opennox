package server

const (
	playerSubStaminaMonsterClass4F7D30 = uint32(2)
	playerSubStaminaPlayerClass4F7D30  = uint32(4)
)

type playerSubStaminaHooks4F7D30[O, U, P, M any] struct {
	loadClass           func(O) uint32
	loadPlayerUpdate    func(O) U
	loadMonsterUpdate   func(O) M
	loadAmount          func() int32
	loadPlayerStamina   func(U) uint8
	storePlayerStamina  func(U, uint8)
	loadPlayer          func(U) P
	loadPlayerIndex     func(P) uint8
	reportStamina       func(uint8, O)
	loadMonsterStamina  func(M) uint8
	storeMonsterStamina func(M, uint8)
}

// playerSubStamina4F7D30 preserves GAME.EXE 004F7D30. The complete class
// dword is loaded once, but the original TEST AL instructions inspect only
// its low byte and give Player precedence over Monster. Both branches compare
// zero-extended stamina against the full signed amount, then subtract only the
// amount's low byte. Player stamina is stored before Player and PlayerInd are
// read, and the report result is deliberately discarded.
func playerSubStamina4F7D30[O, U, P, M any](
	unit O,
	hooks playerSubStaminaHooks4F7D30[O, U, P, M],
) int32 {
	class := uint8(hooks.loadClass(unit))
	if class&uint8(playerSubStaminaPlayerClass4F7D30) != 0 {
		update := hooks.loadPlayerUpdate(unit)
		amount := hooks.loadAmount()
		stamina := hooks.loadPlayerStamina(update)
		if int32(stamina) < amount {
			return 0
		}
		hooks.storePlayerStamina(update, stamina-uint8(amount))
		player := hooks.loadPlayer(update)
		index := hooks.loadPlayerIndex(player)
		hooks.reportStamina(index, unit)
		return 1
	}
	if class&uint8(playerSubStaminaMonsterClass4F7D30) == 0 {
		return 1
	}

	update := hooks.loadMonsterUpdate(unit)
	amount := hooks.loadAmount()
	stamina := hooks.loadMonsterStamina(update)
	if int32(stamina) < amount {
		return 0
	}
	hooks.storeMonsterStamina(update, stamina-uint8(amount))
	return 1
}
