package server

const unusedManaTransferAmount4EAD50 = int16(1)

type unusedManaTransferHooks4EAD50[O, P, R, T any] struct {
	loadClassByte    func(O) uint8
	loadPlayerUpdate func(O) P
	loadManaCurrent  func(P) uint16
	loadSourceUpdate func(O) R
	loadManaMax      func(P) uint16
	loadSourceMana   func(R) int32
	hasTeam          func(O) int32
	loadObjectTeamID func(O) uint8
	findTeamByID     func(uint8) T
	loadTeamID       func(T) uint8
	teamContains     func(O, uint8) int32
	addPlayerMana    func(O, int16) uint16
	storeSourceMana  func(R, int32)
}

// unusedManaTransfer4EAD50 preserves the unreferenced GAME.EXE function at
// 004EAD50. It conditionally transfers one mana point from a cached source
// update record to a Player target. The source record is read only after the
// target class and current-mana loads, and its live value is decremented after
// addPlayerMana returns. The original function has no collision argument.
func unusedManaTransfer4EAD50[O comparable, P, R any, T comparable](
	source, target O,
	hooks unusedManaTransferHooks4EAD50[O, P, R, T],
) {
	var zeroObject O
	if target == zeroObject {
		return
	}
	if hooks.loadClassByte(target)&0x04 == 0 {
		return
	}

	playerUpdate := hooks.loadPlayerUpdate(target)
	current := hooks.loadManaCurrent(playerUpdate)
	sourceUpdate := hooks.loadSourceUpdate(source)
	maximum := hooks.loadManaMax(playerUpdate)
	if current >= maximum {
		return
	}
	if hooks.loadSourceMana(sourceUpdate) <= 0 {
		return
	}

	if hooks.hasTeam(source) != 0 {
		if hooks.hasTeam(target) == 0 {
			return
		}
		objectTeamID := hooks.loadObjectTeamID(target)
		team := hooks.findTeamByID(objectTeamID)
		var zeroTeam T
		if team == zeroTeam {
			return
		}
		if hooks.teamContains(source, hooks.loadTeamID(team)) == 0 {
			return
		}
	}

	_ = hooks.addPlayerMana(target, unusedManaTransferAmount4EAD50)
	liveMana := hooks.loadSourceMana(sourceUpdate)
	hooks.storeSourceMana(sourceUpdate, liveMana-1)
}
