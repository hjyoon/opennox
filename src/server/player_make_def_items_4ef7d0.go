package server

const (
	playerMakeDefItemsObjectFlagMask4EF7D0 uint32 = 0xffeb3fe7
	playerMakeDefItemsKeepItemFlag4EF7D0   uint32 = 0x00000100
	playerMakeDefItemsArmorClass4EF7D0     uint32 = 0x02000000
	playerMakeDefItemsArmorDelete4EF7D0    uint32 = 0x00000808
)

type playerMakeDefItemsResultKind4EF7D0 uint8

const (
	playerMakeDefItemsResultPlayer4EF7D0 playerMakeDefItemsResultKind4EF7D0 = iota
	playerMakeDefItemsResultObject4EF7D0
	playerMakeDefItemsResultByte4EF7D0
)

// playerMakeDefItemsResult4EF7D0 retains the source of GAME.EXE's final AL
// value. Pointer truncation belongs only at the restored uint8 C ABI boundary;
// native-width pointers stay intact throughout the semantic and server layers.
type playerMakeDefItemsResult4EF7D0[O, P comparable] struct {
	kind   playerMakeDefItemsResultKind4EF7D0
	object O
	player P
	value  uint8
}

type playerMakeDefItemsHooks4EF7D0[O, U, P, I, H, M comparable] struct {
	loadUnitArg         func() O
	loadUpdateData      func(O) U
	loadPlayer          func(U) P
	loadPlayerInfo      func(P) I
	loadRestoreStatsArg func() int32
	removePoison        func(O)
	setHealthMaximum    func(O)
	refreshMana         func(O)
	cancelAbilities     func(O)
	resetCamping        func(O)
	storeQuestExit      func(U, O)
	storeQuestWarpGate  func(U, O)
	storeUnitField541   func(O, uint8)
	storeUpdateField21  func(U, uint32)
	storeUpdateField19  func(U, uint16)
	storeHealthCursor   func(U, uint16)
	loadHealthData      func(O) H
	loadHealthCurrent   func(H) uint16
	storeHealthSample   func(U, int, uint16)
	loadObjectFlags     func(O) uint32
	storeObjectFlags    func(O, uint32)
	setPlayerState      func(O, PlayerState)
	clearBuffs          func(O)
	storeUpdateField47  func(U, uint8)
	storeSpellCastStart func(U, uint32)
	storeTrapSpell      func(U, int, uint32)
	storeTrapCountLow   func(U, uint8)
	storeHarpoonBolt    func(U, O)
	storeHarpoonTarget  func(U, O)
	storeUpdateField67  func(U, uint32)
	resetPlayerRuntime  func(O)
	storeObject130      func(O, O)
	gameFlag            func(uint32) bool
	loadPlayerIndex     func(P) uint8
	markPlayerObjects   func(uint8)
	loadPlayerDone      func(P) uint32
	reportTotalHealth   func(uint8, O)
	reportTotalMana     func(uint8, O)
	loadKeepItemsArg    func() int32
	sendRespawn         func(O, uint8) uint8
	loadFirstItem       func(O) O
	loadNextItem        func(O) O
	itemMustDelete      func(O) bool
	loadItemFlags       func(O) uint32
	loadItemClass       func(O) uint32
	loadArmorEquipFlags func(O) uint32
	delayedDelete       func(O)
	modifierID          func(string) uint32
	modifierDesc        func(uint32) M
	modifierBase        func(M) uint32
	loadPlayerClass     func(P) uint8
	loadArmorEquip      func(P) uint32
	loadArmorEquipLow   func(P) uint8
	loadInfoByte        func(I, uint8) uint8
	respawnItem         func(O, string, *[4]M, int32, int32) O
	loadMode2048Type    func(uint8) string
	questDefaultsReady  func() int32
	loadMode4096Type    func(uint8) string
	storePlayerDone     func(P, uint32)
}

// playerMakeDefItems4EF7D0 preserves GAME.EXE 004EF7D0's observable load,
// store, callback, short-circuit, caching, and return-value order. In
// particular, UpdateData and the initial PlayerInfo address are cached at
// entry, while Player itself is reloaded at every original machine load.
// HealthData is likewise reloaded for every one of the 32 history samples.
//
// The inventory successor is cached before any predicate or deletion callback.
// TrapSpellsCnt is cleared through its low byte only. The result describes the
// value left in AL without narrowing a native pointer before the ABI boundary.
func playerMakeDefItems4EF7D0[O, U, P, I, H, M comparable](
	h playerMakeDefItemsHooks4EF7D0[O, U, P, I, H, M],
) playerMakeDefItemsResult4EF7D0[O, P] {
	unit := h.loadUnitArg()
	update := h.loadUpdateData(unit)
	initialPlayer := h.loadPlayer(update)
	initialInfo := h.loadPlayerInfo(initialPlayer)
	restoreStats := h.loadRestoreStatsArg()
	if restoreStats != 0 {
		h.removePoison(unit)
		h.setHealthMaximum(unit)
		h.refreshMana(unit)
	}

	h.cancelAbilities(unit)
	h.resetCamping(unit)
	var zeroObject O
	h.storeQuestExit(update, zeroObject)
	h.storeQuestWarpGate(update, zeroObject)
	h.storeUnitField541(unit, 0)
	h.storeUpdateField21(update, 0)
	h.storeUpdateField19(update, 0)
	h.storeHealthCursor(update, 0)
	for i := 0; i < 32; i++ {
		health := h.loadHealthData(unit)
		current := h.loadHealthCurrent(health)
		h.storeHealthSample(update, i, current)
	}
	flags := h.loadObjectFlags(unit)
	h.storeObjectFlags(unit, flags&playerMakeDefItemsObjectFlagMask4EF7D0)
	h.setPlayerState(unit, PlayerState13)
	h.clearBuffs(unit)
	h.storeUpdateField47(update, 0)
	h.storeSpellCastStart(update, 0)
	for i := 0; i < 5; i++ {
		h.storeTrapSpell(update, i, 0)
	}
	h.storeTrapCountLow(update, 0)
	h.storeHarpoonBolt(update, zeroObject)
	h.storeHarpoonTarget(update, zeroObject)
	h.storeUpdateField67(update, 0)
	h.resetPlayerRuntime(unit)
	h.storeObject130(unit, zeroObject)

	if h.gameFlag(0x2000) {
		player := h.loadPlayer(update)
		h.markPlayerObjects(h.loadPlayerIndex(player))
	}

	player := h.loadPlayer(update)
	var zeroPlayer P
	if player == zeroPlayer {
		return playerMakeDefItemsResult4EF7D0[O, P]{
			kind:   playerMakeDefItemsResultPlayer4EF7D0,
			player: player,
		}
	}
	if h.loadPlayerDone(player) != 0 {
		return playerMakeDefItemsResult4EF7D0[O, P]{
			kind:   playerMakeDefItemsResultPlayer4EF7D0,
			player: player,
		}
	}

	h.reportTotalHealth(h.loadPlayerIndex(player), unit)
	player = h.loadPlayer(update)
	h.reportTotalMana(h.loadPlayerIndex(player), unit)

	var result playerMakeDefItemsResult4EF7D0[O, P]
	keepItems := h.loadKeepItemsArg()
	if keepItems != 0 {
		result.kind = playerMakeDefItemsResultByte4EF7D0
		result.value = h.sendRespawn(unit, 0)
	} else {
		for item := h.loadFirstItem(unit); item != zeroObject; {
			next := h.loadNextItem(item)
			remove := h.itemMustDelete(item)
			if !remove {
				itemFlags := h.loadItemFlags(item)
				remove = itemFlags&playerMakeDefItemsKeepItemFlag4EF7D0 == 0
				if !remove {
					itemClass := h.loadItemClass(item)
					if itemClass&playerMakeDefItemsArmorClass4EF7D0 != 0 {
						remove = h.loadArmorEquipFlags(item)&playerMakeDefItemsArmorDelete4EF7D0 != 0
					}
				}
			}
			if remove {
				h.delayedDelete(item)
			}
			item = next
		}

		h.sendRespawn(unit, 1)
		userColorID := h.modifierID("UserColor1")
		userColor := h.modifierDesc(userColorID)
		userColorBase := h.modifierBase(userColor)

		shirt := h.gameFlag(0x0a00)
		if !shirt {
			player = h.loadPlayer(update)
			shirt = h.loadPlayerClass(player) != 0
		}
		if shirt {
			player = h.loadPlayer(update)
			if h.loadArmorEquip(player)&0x400 == 0 {
				var attrs [4]M
				attrs[1] = h.modifierDesc(userColorBase + uint32(h.loadInfoByte(initialInfo, 84)))
				attrs[2] = h.modifierDesc(userColorBase + uint32(h.loadInfoByte(initialInfo, 85)))
				h.respawnItem(unit, "StreetShirt", &attrs, 1, 0)
			}
		}

		player = h.loadPlayer(update)
		if h.loadArmorEquipLow(player)&0x04 == 0 {
			var attrs [4]M
			attrs[1] = h.modifierDesc(userColorBase + uint32(h.loadInfoByte(initialInfo, 83)))
			h.respawnItem(unit, "StreetPants", &attrs, 1, 0)
		}

		player = h.loadPlayer(update)
		if h.loadArmorEquipLow(player)&0x01 == 0 {
			var attrs [4]M
			attrs[0] = h.modifierDesc(userColorBase + uint32(h.loadInfoByte(initialInfo, 87)))
			attrs[1] = h.modifierDesc(userColorBase + uint32(h.loadInfoByte(initialInfo, 86)))
			h.respawnItem(unit, "StreetSneakers", &attrs, 1, 0)
		}

		if h.gameFlag(0x0800) {
			var attrs [4]M
			attrs[0] = h.modifierDesc(h.modifierID("ArmorQuality1"))
			player = h.loadPlayer(update)
			if h.loadPlayerClass(player) == 0 {
				attrs[1] = h.modifierDesc(h.modifierID("Material1"))
			}
			player = h.loadPlayer(update)
			class := h.loadPlayerClass(player)
			item := h.respawnItem(unit, h.loadMode2048Type(class), &attrs, 1, 0)
			result.kind = playerMakeDefItemsResultObject4EF7D0
			result.object = item
		} else if h.gameFlag(0x1000) && h.questDefaultsReady() >= 0 {
			var attrs [4]M
			player = h.loadPlayer(update)
			if h.loadPlayerClass(player) == 1 {
				attrs[2] = h.modifierDesc(h.modifierID("Replenishment1"))
			}
			player = h.loadPlayer(update)
			class := h.loadPlayerClass(player)
			item := h.respawnItem(unit, h.loadMode4096Type(class), &attrs, 1, 0)
			result.kind = playerMakeDefItemsResultObject4EF7D0
			result.object = item
		} else {
			player = h.loadPlayer(update)
			class := h.loadPlayerClass(player)
			result.kind = playerMakeDefItemsResultByte4EF7D0
			result.value = class
			switch class {
			case 0:
				h.respawnItem(unit, "Longsword", nil, 1, 0)
				result.kind = playerMakeDefItemsResultObject4EF7D0
				result.object = h.respawnItem(unit, "WoodenShield", nil, 1, 0)
			case 1:
				result.kind = playerMakeDefItemsResultObject4EF7D0
				result.object = h.respawnItem(unit, "WizardRobe", nil, 1, 0)
			}
		}
	}

	player = h.loadPlayer(update)
	if player != zeroPlayer {
		h.storePlayerDone(player, 1)
	}
	return result
}
