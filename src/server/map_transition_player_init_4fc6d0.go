package server

const (
	mapTransitionQuestFlag4FC6D0  = int32(0x1000)
	mapTransitionOnlineFlag4FC6D0 = int32(0x2000)
	mapTransitionChatFlag4FC6D0   = int32(0x80)
	mapTransitionHostIndex4FC6D0  = uint8(31)
	mapTransitionInvuln4FC6D0     = int32(0x17)
)

type mapTransitionPlayerInitHooks4FC6D0[Unit comparable, UpdateData, Player any] struct {
	loadMapInitState  func() int32
	loadMapEntryState func() int32
	firstPlayerUnit   func() Unit
	nextPlayerUnit    func(Unit) Unit
	hasGame           func(int32) int32

	loadQuestStage        func() int32
	loadRestorePredicate  func() int32
	loadRestoreReady      func() int32
	loadQueuedRestore     func() int32
	sendQuestStage        func(int32)
	sendQuestRestore      func(int32, int32)
	storeQueuedRestore    func(int32)
	markQuestReady        func(int32)
	finishQuestTransition func()
	fadeBegin             func(int32, int32)

	loadDataRoot       func() string
	formatTempSavePath func(string) string
	loadDeleteTempFile func() func(string)

	loadUpdateData      func(Unit) UpdateData
	loadPlayer          func(UpdateData) Player
	loadPlayerField4792 func(Player) int32
	loadUpdateField138  func(UpdateData) int32
	loadPlayerIndex     func(Player) uint8
	loadPlayerField3680 func(Player) uint8

	savePlayerData    func(string, uint8) int32
	preparePlayerData func(uint8) int32
	sendGauntlet      func(uint8, int32)
	restorePlayerData func(string, uint8) int32
	finishPlayerData  func(uint8)
	applyEnchant      func(Unit, int32, int32, int32)
}

// mapTransitionPlayerInit4FC6D0 preserves GAME.EXE 004FC6D0. The routine
// uses exact-one state gates, preserves the original short-circuit decisions,
// and traverses player units through live FirstUnit/NextUnit callbacks. In the
// temporary-save loop it caches UpdateData but reloads Player and its low-byte
// index before every callback that consumed the original PE32 fields.
func mapTransitionPlayerInit4FC6D0[Unit comparable, UpdateData, Player any](hooks mapTransitionPlayerInitHooks4FC6D0[Unit, UpdateData, Player]) {
	if hooks.loadMapInitState() != 1 {
		if hooks.loadMapEntryState() != 1 {
			return
		}
	}

	var nilUnit Unit
	if hooks.firstPlayerUnit() == nilUnit {
		return
	}

	if hooks.hasGame(mapTransitionQuestFlag4FC6D0) != 0 {
		if hooks.loadQuestStage() == 1 {
			hooks.sendQuestStage(255)
			hooks.markQuestReady(1)
			hooks.finishQuestTransition()
		} else if hooks.loadRestorePredicate() == 0 || hooks.loadRestoreReady() != 0 {
			if hooks.loadQueuedRestore() == 1 {
				hooks.sendQuestRestore(255, 1)
				hooks.storeQueuedRestore(0)
				hooks.markQuestReady(1)
				hooks.finishQuestTransition()
			} else {
				root := hooks.loadDataRoot()
				path := hooks.formatTempSavePath(root)
				unit := hooks.firstPlayerUnit()
				if unit != nilUnit {
					deleteTempFile := hooks.loadDeleteTempFile()
					for unit != nilUnit {
						ud := hooks.loadUpdateData(unit)
						player := hooks.loadPlayer(ud)
						if hooks.loadPlayerField4792(player) == 1 && hooks.loadUpdateField138(ud) == 0 {
							player = hooks.loadPlayer(ud)
							index := hooks.loadPlayerIndex(player)
							if hooks.savePlayerData(path, index) != 0 {
								player = hooks.loadPlayer(ud)
								index = hooks.loadPlayerIndex(player)
								prepared := hooks.preparePlayerData(index)

								player = hooks.loadPlayer(ud)
								index = hooks.loadPlayerIndex(player)
								hooks.sendGauntlet(index, 1)

								player = hooks.loadPlayer(ud)
								index = hooks.loadPlayerIndex(player)
								restored := hooks.restorePlayerData(path, index)
								if restored == 0 && prepared == 0 {
									player = hooks.loadPlayer(ud)
									index = hooks.loadPlayerIndex(player)
									hooks.sendGauntlet(index, 0)
								}
								deleteTempFile(path)
							}
						}

						player = hooks.loadPlayer(ud)
						index := hooks.loadPlayerIndex(player)
						hooks.finishPlayerData(index)
						unit = hooks.nextPlayerUnit(unit)
					}
				}
				hooks.sendQuestRestore(255, 0)
				hooks.markQuestReady(1)
				hooks.finishQuestTransition()
			}
		} else {
			hooks.sendQuestStage(255)
			hooks.markQuestReady(1)
			hooks.finishQuestTransition()
		}
	} else {
		hooks.fadeBegin(1, 1)
	}

	if hooks.hasGame(mapTransitionOnlineFlag4FC6D0) == 0 {
		return
	}
	if hooks.hasGame(mapTransitionChatFlag4FC6D0) != 0 {
		return
	}

	for unit := hooks.firstPlayerUnit(); unit != nilUnit; unit = hooks.nextPlayerUnit(unit) {
		ud := hooks.loadUpdateData(unit)
		player := hooks.loadPlayer(ud)
		if hooks.loadPlayerIndex(player) != mapTransitionHostIndex4FC6D0 && hooks.loadPlayerField3680(player)&1 == 0 {
			hooks.applyEnchant(unit, mapTransitionInvuln4FC6D0, 0, 5)
		}
	}
}
