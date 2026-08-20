package server

const (
	poisonClearPlayerClassLow4EE8F0     = uint8(0x04)
	poisonClearMonsterClassLow4EE8F0    = uint8(0x02)
	poisonClearQuestGameFlag4EE8F0      = uint32(0x00000800)
	poisonClearQuestMonsterLow4EE8F0    = uint8(0x10)
	poisonClearOwnedMonsterLow4EE8F0    = uint8(0x80)
	poisonClearPlayerStatus4EE8F0       = uint32(0x00000400)
	poisonClearQuestPlayerIndex4EE8F0   = int32(31)
	poisonClearFadeMessage4EE8F0        = "Health.c:PoisonFade"
	poisonClearNetworkInactive4EE8F0    = int32(0)
	poisonClearPriorityMessageArg4EE8F0 = uint8(0)
)

type poisonClearHooks4EE8F0[O, H, U, P, I any] struct {
	loadUnitArg       func() O
	loadAmountArg     func() int32
	loadPoison        func(O) uint8
	storePoison       func(O, uint8)
	loadHealth        func(O) H
	clearHealthFrame  func(H)
	loadClass         func(O) uint32
	loadUpdateData    func(O) U
	loadPlayer        func(U) P
	unsetPlayerStatus func(P, uint32)
	priorityMessage   func(O, string, uint8)
	gameFlag          func(uint32) int32
	loadSubClass      func(O) uint32
	playerInfoByIndex func(int32) I
	loadPlayerUnit    func(I) O
	loadOwner         func(O) O
	reportPoison      func(O, O, int32)
}

// updatePoison4EE8F0 preserves the signed comparison against the decrement
// argument and the low-byte subtraction used by GAME.EXE 004EE8F0.
func updatePoison4EE8F0[O, H, U, P, I comparable](hooks poisonClearHooks4EE8F0[O, H, U, P, I]) {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return
	}
	current := hooks.loadPoison(unit)
	amount := hooks.loadAmountArg()
	if int32(current) > amount {
		hooks.storePoison(unit, current-uint8(amount))
		return
	}
	poisonClearEffects4EE8F0(hooks, unit, true)
}

// removePoison4EE9D0 preserves the zero-poison early return and otherwise
// shares the clear/status path with 004EE8F0 without its Player fade message.
func removePoison4EE9D0[O, H, U, P, I comparable](hooks poisonClearHooks4EE8F0[O, H, U, P, I]) {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return
	}
	if hooks.loadPoison(unit) == 0 {
		return
	}
	poisonClearEffects4EE8F0(hooks, unit, false)
}

func poisonClearEffects4EE8F0[O, H, U, P, I comparable](
	hooks poisonClearHooks4EE8F0[O, H, U, P, I],
	unit O,
	emitFadeMessage bool,
) {
	health := hooks.loadHealth(unit)
	hooks.storePoison(unit, 0)
	var nilHealth H
	if health != nilHealth {
		hooks.clearHealthFrame(health)
	}

	class := hooks.loadClass(unit)
	if uint8(class)&poisonClearPlayerClassLow4EE8F0 != 0 {
		update := hooks.loadUpdateData(unit)
		player := hooks.loadPlayer(update)
		hooks.unsetPlayerStatus(player, poisonClearPlayerStatus4EE8F0)
		if emitFadeMessage {
			hooks.priorityMessage(unit, poisonClearFadeMessage4EE8F0, poisonClearPriorityMessageArg4EE8F0)
		}
		return
	}
	if uint8(class)&poisonClearMonsterClassLow4EE8F0 == 0 {
		return
	}

	if hooks.gameFlag(poisonClearQuestGameFlag4EE8F0) == 1 &&
		uint8(hooks.loadSubClass(unit))&poisonClearQuestMonsterLow4EE8F0 != 0 {
		info := hooks.playerInfoByIndex(poisonClearQuestPlayerIndex4EE8F0)
		var nilInfo I
		if info == nilInfo {
			return
		}
		receiver := hooks.loadPlayerUnit(info)
		var nilObject O
		if receiver == nilObject {
			return
		}
		hooks.reportPoison(receiver, unit, poisonClearNetworkInactive4EE8F0)
		return
	}

	if uint8(hooks.loadSubClass(unit))&poisonClearOwnedMonsterLow4EE8F0 == 0 {
		return
	}
	receiver := hooks.loadOwner(unit)
	var nilObject O
	if receiver == nilObject {
		return
	}
	hooks.reportPoison(receiver, unit, poisonClearNetworkInactive4EE8F0)
}
