package server

const (
	poisonSetTimerActive4EEA90   = uint16(1000)
	poisonSetTimerInactive4EEA90 = uint16(0)
)

type poisonSetHooks4EEA90[O, H, U, P, I any] struct {
	loadUnitArg       func() O
	loadCurrent       func(O) uint8
	loadValueArg      func() int32
	loadHealth        func(O) H
	frame             func() uint32
	storeHealthFrame  func(H, uint32)
	loadClass         func(O) uint32
	storePoison       func(O, uint8)
	loadUpdateData    func(O) U
	loadPlayer        func(U) P
	needPlayerStatus  func(P, uint32)
	unsetPlayerStatus func(P, uint32)
	gameFlag          func(uint32) int32
	loadSubClass      func(O) uint32
	playerInfoByIndex func(int32) I
	loadPlayerUnit    func(I) O
	loadOwner         func(O) O
	reportPoison      func(O, O, int32)
	storePoisonTimer  func(O, uint16)
}

// setPoison4EEA90 preserves GAME.EXE 004EEA90. The whole int32 argument
// controls transition stamping, status/network state, and the timer, while
// only its low byte is stored as the object's poison value.
func setPoison4EEA90[O, H, U, P, I comparable](hooks poisonSetHooks4EEA90[O, H, U, P, I]) {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return
	}
	current := hooks.loadCurrent(unit)
	value := hooks.loadValueArg()
	if current == 0 && value > 0 {
		health := hooks.loadHealth(unit)
		var nilHealth H
		if health != nilHealth {
			frame := hooks.frame()
			hooks.storeHealthFrame(health, frame)
		}
	}

	class := hooks.loadClass(unit)
	active := value != 0
	hooks.storePoison(unit, uint8(value))
	if uint8(class)&poisonClearPlayerClassLow4EE8F0 != 0 {
		update := hooks.loadUpdateData(unit)
		player := hooks.loadPlayer(update)
		if active {
			hooks.needPlayerStatus(player, poisonClearPlayerStatus4EE8F0)
		} else {
			hooks.unsetPlayerStatus(player, poisonClearPlayerStatus4EE8F0)
		}
	} else if uint8(class)&poisonClearMonsterClassLow4EE8F0 != 0 {
		poisonSetMonsterReport4EEA90(hooks, unit, active)
	}

	if active {
		hooks.storePoisonTimer(unit, poisonSetTimerActive4EEA90)
	} else {
		hooks.storePoisonTimer(unit, poisonSetTimerInactive4EEA90)
	}
}

func poisonSetMonsterReport4EEA90[O, H, U, P, I comparable](
	hooks poisonSetHooks4EEA90[O, H, U, P, I],
	unit O,
	active bool,
) {
	var receiver O
	if hooks.gameFlag(poisonClearQuestGameFlag4EE8F0) == 1 &&
		uint8(hooks.loadSubClass(unit))&poisonClearQuestMonsterLow4EE8F0 != 0 {
		info := hooks.playerInfoByIndex(poisonClearQuestPlayerIndex4EE8F0)
		var nilInfo I
		if info == nilInfo {
			return
		}
		receiver = hooks.loadPlayerUnit(info)
		var nilObject O
		if receiver == nilObject {
			return
		}
	} else {
		if int8(uint8(hooks.loadSubClass(unit))) >= 0 {
			return
		}
		receiver = hooks.loadOwner(unit)
		var nilObject O
		if receiver == nilObject {
			return
		}
	}
	state := int32(0)
	if active {
		state = 1
	}
	hooks.reportPoison(receiver, unit, state)
}
