package server

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type playerMakeDefItemsObject4EF7D0 string
type playerMakeDefItemsUpdate4EF7D0 string
type playerMakeDefItemsPlayer4EF7D0 string
type playerMakeDefItemsInfo4EF7D0 string
type playerMakeDefItemsHealth4EF7D0 string
type playerMakeDefItemsModifier4EF7D0 string

type playerMakeDefItemsRespawnCall4EF7D0 struct {
	typeID string
	attrs  *[4]playerMakeDefItemsModifier4EF7D0
	a4     int32
	a5     int32
	result playerMakeDefItemsObject4EF7D0
}

type playerMakeDefItemsWorld4EF7D0 struct {
	events  []string
	faultAt int
	onEvent func(string)

	unit         playerMakeDefItemsObject4EF7D0
	update       playerMakeDefItemsUpdate4EF7D0
	player       playerMakeDefItemsPlayer4EF7D0
	infoByPlayer map[playerMakeDefItemsPlayer4EF7D0]playerMakeDefItemsInfo4EF7D0
	restoreStats int32
	keepItems    int32
	sendResults  map[uint8]uint8
	gameFlags    map[uint32]bool
	questReady   int32

	healthByUnit   map[playerMakeDefItemsObject4EF7D0]playerMakeDefItemsHealth4EF7D0
	healthCurrent  map[playerMakeDefItemsHealth4EF7D0]uint16
	healthSamples  [32]uint16
	healthCursor   uint16
	objectFlags    uint32
	objectField541 uint8
	object130      playerMakeDefItemsObject4EF7D0
	questExit      playerMakeDefItemsObject4EF7D0
	questWarp      playerMakeDefItemsObject4EF7D0
	updateField21  uint32
	updateField19  uint16
	updateField47  uint8
	spellStart     uint32
	trapSpells     [5]uint32
	trapCount      uint32
	harpoonBolt    playerMakeDefItemsObject4EF7D0
	harpoonTarget  playerMakeDefItemsObject4EF7D0
	updateField67  uint32

	playerIndex map[playerMakeDefItemsPlayer4EF7D0]uint8
	playerDone  map[playerMakeDefItemsPlayer4EF7D0]uint32
	playerClass map[playerMakeDefItemsPlayer4EF7D0]uint8
	armorEquip  map[playerMakeDefItemsPlayer4EF7D0]uint32
	infoBytes   map[playerMakeDefItemsInfo4EF7D0]map[uint8]uint8

	firstItem  playerMakeDefItemsObject4EF7D0
	nextItem   map[playerMakeDefItemsObject4EF7D0]playerMakeDefItemsObject4EF7D0
	predicate  map[playerMakeDefItemsObject4EF7D0]bool
	itemFlags  map[playerMakeDefItemsObject4EF7D0]uint32
	itemClass  map[playerMakeDefItemsObject4EF7D0]uint32
	armorFlags map[playerMakeDefItemsObject4EF7D0]uint32
	deleted    []playerMakeDefItemsObject4EF7D0

	modifierIDs  map[string]uint32
	modifierBase map[playerMakeDefItemsModifier4EF7D0]uint32
	mode2048Type map[uint8]string
	mode4096Type map[uint8]string
	respawns     []playerMakeDefItemsRespawnCall4EF7D0
}

func (w *playerMakeDefItemsWorld4EF7D0) record(event string) {
	w.events = append(w.events, event)
	if w.onEvent != nil {
		w.onEvent(event)
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func playerMakeDefItemsAttrsString4EF7D0(attrs *[4]playerMakeDefItemsModifier4EF7D0) string {
	if attrs == nil {
		return "nil"
	}
	parts := make([]string, len(attrs))
	for i, mod := range attrs {
		parts[i] = string(mod)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (w *playerMakeDefItemsWorld4EF7D0) hooks() playerMakeDefItemsHooks4EF7D0[
	playerMakeDefItemsObject4EF7D0,
	playerMakeDefItemsUpdate4EF7D0,
	playerMakeDefItemsPlayer4EF7D0,
	playerMakeDefItemsInfo4EF7D0,
	playerMakeDefItemsHealth4EF7D0,
	playerMakeDefItemsModifier4EF7D0,
] {
	return playerMakeDefItemsHooks4EF7D0[
		playerMakeDefItemsObject4EF7D0,
		playerMakeDefItemsUpdate4EF7D0,
		playerMakeDefItemsPlayer4EF7D0,
		playerMakeDefItemsInfo4EF7D0,
		playerMakeDefItemsHealth4EF7D0,
		playerMakeDefItemsModifier4EF7D0,
	]{
		loadUnitArg: func() playerMakeDefItemsObject4EF7D0 {
			value := w.unit
			w.record("arg:" + string(value))
			return value
		},
		loadUpdateData: func(unit playerMakeDefItemsObject4EF7D0) playerMakeDefItemsUpdate4EF7D0 {
			value := w.update
			w.record("update:" + string(unit) + "=" + string(value))
			return value
		},
		loadPlayer: func(update playerMakeDefItemsUpdate4EF7D0) playerMakeDefItemsPlayer4EF7D0 {
			value := w.player
			w.record("player:" + string(update) + "=" + string(value))
			return value
		},
		loadPlayerInfo: func(player playerMakeDefItemsPlayer4EF7D0) playerMakeDefItemsInfo4EF7D0 {
			value := w.infoByPlayer[player]
			w.record("info:" + string(player) + "=" + string(value))
			return value
		},
		loadRestoreStatsArg: func() int32 {
			value := w.restoreStats
			w.record(fmt.Sprintf("restore:%d", value))
			return value
		},
		removePoison: func(unit playerMakeDefItemsObject4EF7D0) {
			w.record("remove-poison:" + string(unit))
		},
		setHealthMaximum: func(unit playerMakeDefItemsObject4EF7D0) {
			w.record("health-max:" + string(unit))
		},
		refreshMana: func(unit playerMakeDefItemsObject4EF7D0) {
			w.record("mana-max:" + string(unit))
		},
		cancelAbilities: func(unit playerMakeDefItemsObject4EF7D0) {
			w.record("cancel-abilities:" + string(unit))
		},
		resetCamping: func(unit playerMakeDefItemsObject4EF7D0) {
			w.record("reset-camping:" + string(unit))
		},
		storeQuestExit: func(update playerMakeDefItemsUpdate4EF7D0, value playerMakeDefItemsObject4EF7D0) {
			w.record("quest-exit:" + string(update) + "=" + string(value))
			w.questExit = value
		},
		storeQuestWarpGate: func(update playerMakeDefItemsUpdate4EF7D0, value playerMakeDefItemsObject4EF7D0) {
			w.record("quest-warp:" + string(update) + "=" + string(value))
			w.questWarp = value
		},
		storeUnitField541: func(unit playerMakeDefItemsObject4EF7D0, value uint8) {
			w.record(fmt.Sprintf("unit-541:%s=%d", unit, value))
			w.objectField541 = value
		},
		storeUpdateField21: func(update playerMakeDefItemsUpdate4EF7D0, value uint32) {
			w.record(fmt.Sprintf("update-21:%s=%d", update, value))
			w.updateField21 = value
		},
		storeUpdateField19: func(update playerMakeDefItemsUpdate4EF7D0, value uint16) {
			w.record(fmt.Sprintf("update-19:%s=%d", update, value))
			w.updateField19 = value
		},
		storeHealthCursor: func(update playerMakeDefItemsUpdate4EF7D0, value uint16) {
			w.record(fmt.Sprintf("health-cursor:%s=%d", update, value))
			w.healthCursor = value
		},
		loadHealthData: func(unit playerMakeDefItemsObject4EF7D0) playerMakeDefItemsHealth4EF7D0 {
			value := w.healthByUnit[unit]
			w.record("health:" + string(unit) + "=" + string(value))
			return value
		},
		loadHealthCurrent: func(health playerMakeDefItemsHealth4EF7D0) uint16 {
			value := w.healthCurrent[health]
			w.record(fmt.Sprintf("health-current:%s=%#04x", health, value))
			return value
		},
		storeHealthSample: func(update playerMakeDefItemsUpdate4EF7D0, index int, value uint16) {
			w.record(fmt.Sprintf("health-sample:%s:%d=%#04x", update, index, value))
			w.healthSamples[index] = value
		},
		loadObjectFlags: func(unit playerMakeDefItemsObject4EF7D0) uint32 {
			value := w.objectFlags
			w.record(fmt.Sprintf("object-flags:%s=%#08x", unit, value))
			return value
		},
		storeObjectFlags: func(unit playerMakeDefItemsObject4EF7D0, value uint32) {
			w.record(fmt.Sprintf("store-object-flags:%s=%#08x", unit, value))
			w.objectFlags = value
		},
		setPlayerState: func(unit playerMakeDefItemsObject4EF7D0, state PlayerState) {
			w.record(fmt.Sprintf("state:%s=%d", unit, state))
		},
		clearBuffs: func(unit playerMakeDefItemsObject4EF7D0) {
			w.record("clear-buffs:" + string(unit))
		},
		storeUpdateField47: func(update playerMakeDefItemsUpdate4EF7D0, value uint8) {
			w.record(fmt.Sprintf("update-47:%s=%d", update, value))
			w.updateField47 = value
		},
		storeSpellCastStart: func(update playerMakeDefItemsUpdate4EF7D0, value uint32) {
			w.record(fmt.Sprintf("spell-start:%s=%d", update, value))
			w.spellStart = value
		},
		storeTrapSpell: func(update playerMakeDefItemsUpdate4EF7D0, index int, value uint32) {
			w.record(fmt.Sprintf("trap:%s:%d=%d", update, index, value))
			w.trapSpells[index] = value
		},
		storeTrapCountLow: func(update playerMakeDefItemsUpdate4EF7D0, value uint8) {
			w.record(fmt.Sprintf("trap-count-low:%s=%d", update, value))
			w.trapCount = w.trapCount&^0xff | uint32(value)
		},
		storeHarpoonBolt: func(update playerMakeDefItemsUpdate4EF7D0, value playerMakeDefItemsObject4EF7D0) {
			w.record("harpoon-bolt:" + string(update) + "=" + string(value))
			w.harpoonBolt = value
		},
		storeHarpoonTarget: func(update playerMakeDefItemsUpdate4EF7D0, value playerMakeDefItemsObject4EF7D0) {
			w.record("harpoon-target:" + string(update) + "=" + string(value))
			w.harpoonTarget = value
		},
		storeUpdateField67: func(update playerMakeDefItemsUpdate4EF7D0, value uint32) {
			w.record(fmt.Sprintf("update-67:%s=%d", update, value))
			w.updateField67 = value
		},
		resetPlayerRuntime: func(unit playerMakeDefItemsObject4EF7D0) {
			w.record("reset-runtime:" + string(unit))
		},
		storeObject130: func(unit, value playerMakeDefItemsObject4EF7D0) {
			w.record("object-130:" + string(unit) + "=" + string(value))
			w.object130 = value
		},
		gameFlag: func(flag uint32) bool {
			value := w.gameFlags[flag]
			w.record(fmt.Sprintf("game-flag:%#x=%t", flag, value))
			return value
		},
		loadPlayerIndex: func(player playerMakeDefItemsPlayer4EF7D0) uint8 {
			value := w.playerIndex[player]
			w.record(fmt.Sprintf("player-index:%s=%d", player, value))
			return value
		},
		markPlayerObjects: func(index uint8) {
			w.record(fmt.Sprintf("mark-player-objects:%d", index))
		},
		loadPlayerDone: func(player playerMakeDefItemsPlayer4EF7D0) uint32 {
			value := w.playerDone[player]
			w.record(fmt.Sprintf("player-done:%s=%d", player, value))
			return value
		},
		reportTotalHealth: func(index uint8, unit playerMakeDefItemsObject4EF7D0) {
			w.record(fmt.Sprintf("report-health:%d:%s", index, unit))
		},
		reportTotalMana: func(index uint8, unit playerMakeDefItemsObject4EF7D0) {
			w.record(fmt.Sprintf("report-mana:%d:%s", index, unit))
		},
		loadKeepItemsArg: func() int32 {
			value := w.keepItems
			w.record(fmt.Sprintf("keep-items:%d", value))
			return value
		},
		sendRespawn: func(unit playerMakeDefItemsObject4EF7D0, keep uint8) uint8 {
			value := w.sendResults[keep]
			w.record(fmt.Sprintf("send-respawn:%s:%d=%#02x", unit, keep, value))
			return value
		},
		loadFirstItem: func(unit playerMakeDefItemsObject4EF7D0) playerMakeDefItemsObject4EF7D0 {
			value := w.firstItem
			w.record("first-item:" + string(unit) + "=" + string(value))
			return value
		},
		loadNextItem: func(item playerMakeDefItemsObject4EF7D0) playerMakeDefItemsObject4EF7D0 {
			value := w.nextItem[item]
			w.record("next-item:" + string(item) + "=" + string(value))
			return value
		},
		itemMustDelete: func(item playerMakeDefItemsObject4EF7D0) bool {
			value := w.predicate[item]
			w.record(fmt.Sprintf("item-predicate:%s=%t", item, value))
			return value
		},
		loadItemFlags: func(item playerMakeDefItemsObject4EF7D0) uint32 {
			value := w.itemFlags[item]
			w.record(fmt.Sprintf("item-flags:%s=%#x", item, value))
			return value
		},
		loadItemClass: func(item playerMakeDefItemsObject4EF7D0) uint32 {
			value := w.itemClass[item]
			w.record(fmt.Sprintf("item-class:%s=%#x", item, value))
			return value
		},
		loadArmorEquipFlags: func(item playerMakeDefItemsObject4EF7D0) uint32 {
			value := w.armorFlags[item]
			w.record(fmt.Sprintf("armor-flags:%s=%#x", item, value))
			return value
		},
		delayedDelete: func(item playerMakeDefItemsObject4EF7D0) {
			w.record("delete:" + string(item))
			w.deleted = append(w.deleted, item)
		},
		modifierID: func(name string) uint32 {
			value := w.modifierIDs[name]
			w.record(fmt.Sprintf("modifier-id:%s=%d", name, value))
			return value
		},
		modifierDesc: func(id uint32) playerMakeDefItemsModifier4EF7D0 {
			value := playerMakeDefItemsModifier4EF7D0(fmt.Sprintf("mod-%d", id))
			w.record(fmt.Sprintf("modifier-desc:%d=%s", id, value))
			return value
		},
		modifierBase: func(modifier playerMakeDefItemsModifier4EF7D0) uint32 {
			value := w.modifierBase[modifier]
			w.record(fmt.Sprintf("modifier-base:%s=%d", modifier, value))
			return value
		},
		loadPlayerClass: func(player playerMakeDefItemsPlayer4EF7D0) uint8 {
			value := w.playerClass[player]
			w.record(fmt.Sprintf("player-class:%s=%d", player, value))
			return value
		},
		loadArmorEquip: func(player playerMakeDefItemsPlayer4EF7D0) uint32 {
			value := w.armorEquip[player]
			w.record(fmt.Sprintf("armor-equip:%s=%#x", player, value))
			return value
		},
		loadArmorEquipLow: func(player playerMakeDefItemsPlayer4EF7D0) uint8 {
			value := uint8(w.armorEquip[player])
			w.record(fmt.Sprintf("armor-equip-low:%s=%#x", player, value))
			return value
		},
		loadInfoByte: func(info playerMakeDefItemsInfo4EF7D0, offset uint8) uint8 {
			value := w.infoBytes[info][offset]
			w.record(fmt.Sprintf("info-byte:%s:%d=%d", info, offset, value))
			return value
		},
		respawnItem: func(unit playerMakeDefItemsObject4EF7D0, typeID string, attrs *[4]playerMakeDefItemsModifier4EF7D0, a4, a5 int32) playerMakeDefItemsObject4EF7D0 {
			result := playerMakeDefItemsObject4EF7D0("made:" + typeID)
			var copied *[4]playerMakeDefItemsModifier4EF7D0
			if attrs != nil {
				value := *attrs
				copied = &value
			}
			w.record(fmt.Sprintf("respawn-item:%s:%s:%s:%d:%d=%s", unit, typeID, playerMakeDefItemsAttrsString4EF7D0(attrs), a4, a5, result))
			w.respawns = append(w.respawns, playerMakeDefItemsRespawnCall4EF7D0{typeID: typeID, attrs: copied, a4: a4, a5: a5, result: result})
			return result
		},
		loadMode2048Type: func(class uint8) string {
			value := w.mode2048Type[class]
			w.record(fmt.Sprintf("mode-2048-type:%d=%s", class, value))
			return value
		},
		questDefaultsReady: func() int32 {
			value := w.questReady
			w.record(fmt.Sprintf("quest-defaults-ready=%d", value))
			return value
		},
		loadMode4096Type: func(class uint8) string {
			value := w.mode4096Type[class]
			w.record(fmt.Sprintf("mode-4096-type:%d=%s", class, value))
			return value
		},
		storePlayerDone: func(player playerMakeDefItemsPlayer4EF7D0, value uint32) {
			w.record(fmt.Sprintf("store-player-done:%s=%d", player, value))
			w.playerDone[player] = value
		},
	}
}

func newPlayerMakeDefItemsWorld4EF7D0() *playerMakeDefItemsWorld4EF7D0 {
	const (
		unit   = playerMakeDefItemsObject4EF7D0("unit-a")
		update = playerMakeDefItemsUpdate4EF7D0("update-a")
		player = playerMakeDefItemsPlayer4EF7D0("player-a")
		info   = playerMakeDefItemsInfo4EF7D0("info-a")
		health = playerMakeDefItemsHealth4EF7D0("health-a")
	)
	items := []playerMakeDefItemsObject4EF7D0{
		"predicate-delete", "flag-delete", "armor-delete", "armor-keep", "plain-keep",
	}
	w := &playerMakeDefItemsWorld4EF7D0{
		unit:           unit,
		update:         update,
		player:         player,
		infoByPlayer:   map[playerMakeDefItemsPlayer4EF7D0]playerMakeDefItemsInfo4EF7D0{player: info},
		restoreStats:   1,
		sendResults:    map[uint8]uint8{0: 0xa0, 1: 0xa1},
		gameFlags:      map[uint32]bool{0x2000: true, 0x0a00: true, 0x0800: true},
		questReady:     0,
		healthByUnit:   map[playerMakeDefItemsObject4EF7D0]playerMakeDefItemsHealth4EF7D0{unit: health},
		healthCurrent:  map[playerMakeDefItemsHealth4EF7D0]uint16{health: 0x1234},
		healthCursor:   0xeeee,
		objectFlags:    0xffffffff,
		objectField541: 0xff,
		object130:      "old-130",
		questExit:      "old-exit",
		questWarp:      "old-warp",
		updateField21:  0xffffffff,
		updateField19:  0xffff,
		updateField47:  0xff,
		spellStart:     0xffffffff,
		trapSpells:     [5]uint32{1, 2, 3, 4, 5},
		trapCount:      0xaabbccdd,
		harpoonBolt:    "old-bolt",
		harpoonTarget:  "old-target",
		updateField67:  0xffffffff,
		playerIndex:    map[playerMakeDefItemsPlayer4EF7D0]uint8{player: 7},
		playerDone:     map[playerMakeDefItemsPlayer4EF7D0]uint32{player: 0},
		playerClass:    map[playerMakeDefItemsPlayer4EF7D0]uint8{player: 0},
		armorEquip:     map[playerMakeDefItemsPlayer4EF7D0]uint32{player: 0},
		infoBytes: map[playerMakeDefItemsInfo4EF7D0]map[uint8]uint8{
			info: {83: 3, 84: 4, 85: 5, 86: 6, 87: 7},
		},
		firstItem: items[0],
		nextItem: map[playerMakeDefItemsObject4EF7D0]playerMakeDefItemsObject4EF7D0{
			items[0]: items[1], items[1]: items[2], items[2]: items[3], items[3]: items[4], items[4]: "",
		},
		predicate: map[playerMakeDefItemsObject4EF7D0]bool{items[0]: true},
		itemFlags: map[playerMakeDefItemsObject4EF7D0]uint32{
			items[1]: 0, items[2]: 0x100, items[3]: 0x100, items[4]: 0x100,
		},
		itemClass: map[playerMakeDefItemsObject4EF7D0]uint32{
			items[2]: 0x02000000, items[3]: 0x02000000, items[4]: 0,
		},
		armorFlags: map[playerMakeDefItemsObject4EF7D0]uint32{items[2]: 0x808, items[3]: 0},
		modifierIDs: map[string]uint32{
			"UserColor1": 100, "ArmorQuality1": 200, "Material1": 300, "Replenishment1": 400,
		},
		modifierBase: map[playerMakeDefItemsModifier4EF7D0]uint32{"mod-100": 1000},
		mode2048Type: map[uint8]string{0: "Sword", 1: "StaffWooden", 2: "StaffWooden"},
		mode4096Type: map[uint8]string{0: "Sword", 1: "SulphorousFlareWand", 2: "Bow"},
	}
	return w
}

func playerMakeDefItemsFullEvents4EF7D0() []string {
	events := []string{
		"arg:unit-a", "update:unit-a=update-a", "player:update-a=player-a", "info:player-a=info-a", "restore:1",
		"remove-poison:unit-a", "health-max:unit-a", "mana-max:unit-a", "cancel-abilities:unit-a", "reset-camping:unit-a",
		"quest-exit:update-a=", "quest-warp:update-a=", "unit-541:unit-a=0", "update-21:update-a=0", "update-19:update-a=0", "health-cursor:update-a=0",
	}
	for i := 0; i < 32; i++ {
		events = append(events,
			"health:unit-a=health-a",
			"health-current:health-a=0x1234",
			fmt.Sprintf("health-sample:update-a:%d=0x1234", i),
		)
	}
	events = append(events,
		"object-flags:unit-a=0xffffffff", "store-object-flags:unit-a=0xffeb3fe7", "state:unit-a=13", "clear-buffs:unit-a",
		"update-47:update-a=0", "spell-start:update-a=0",
		"trap:update-a:0=0", "trap:update-a:1=0", "trap:update-a:2=0", "trap:update-a:3=0", "trap:update-a:4=0",
		"trap-count-low:update-a=0", "harpoon-bolt:update-a=", "harpoon-target:update-a=", "update-67:update-a=0",
		"reset-runtime:unit-a", "object-130:unit-a=", "game-flag:0x2000=true", "player:update-a=player-a", "player-index:player-a=7", "mark-player-objects:7",
		"player:update-a=player-a", "player-done:player-a=0", "player-index:player-a=7", "report-health:7:unit-a",
		"player:update-a=player-a", "player-index:player-a=7", "report-mana:7:unit-a", "keep-items:0", "first-item:unit-a=predicate-delete",
		"next-item:predicate-delete=flag-delete", "item-predicate:predicate-delete=true", "delete:predicate-delete",
		"next-item:flag-delete=armor-delete", "item-predicate:flag-delete=false", "item-flags:flag-delete=0x0", "delete:flag-delete",
		"next-item:armor-delete=armor-keep", "item-predicate:armor-delete=false", "item-flags:armor-delete=0x100", "item-class:armor-delete=0x2000000", "armor-flags:armor-delete=0x808", "delete:armor-delete",
		"next-item:armor-keep=plain-keep", "item-predicate:armor-keep=false", "item-flags:armor-keep=0x100", "item-class:armor-keep=0x2000000", "armor-flags:armor-keep=0x0",
		"next-item:plain-keep=", "item-predicate:plain-keep=false", "item-flags:plain-keep=0x100", "item-class:plain-keep=0x0",
		"send-respawn:unit-a:1=0xa1", "modifier-id:UserColor1=100", "modifier-desc:100=mod-100", "modifier-base:mod-100=1000",
		"game-flag:0xa00=true", "player:update-a=player-a", "armor-equip:player-a=0x0",
		"info-byte:info-a:84=4", "modifier-desc:1004=mod-1004", "info-byte:info-a:85=5", "modifier-desc:1005=mod-1005",
		"respawn-item:unit-a:StreetShirt:[,mod-1004,mod-1005,]:1:0=made:StreetShirt",
		"player:update-a=player-a", "armor-equip-low:player-a=0x0", "info-byte:info-a:83=3", "modifier-desc:1003=mod-1003",
		"respawn-item:unit-a:StreetPants:[,mod-1003,,]:1:0=made:StreetPants",
		"player:update-a=player-a", "armor-equip-low:player-a=0x0", "info-byte:info-a:87=7", "modifier-desc:1007=mod-1007", "info-byte:info-a:86=6", "modifier-desc:1006=mod-1006",
		"respawn-item:unit-a:StreetSneakers:[mod-1007,mod-1006,,]:1:0=made:StreetSneakers",
		"game-flag:0x800=true", "modifier-id:ArmorQuality1=200", "modifier-desc:200=mod-200",
		"player:update-a=player-a", "player-class:player-a=0", "modifier-id:Material1=300", "modifier-desc:300=mod-300",
		"player:update-a=player-a", "player-class:player-a=0", "mode-2048-type:0=Sword",
		"respawn-item:unit-a:Sword:[mod-200,mod-300,,]:1:0=made:Sword",
		"player:update-a=player-a", "store-player-done:player-a=1",
	)
	return events
}

func TestPlayerMakeDefItems4EF7D0ExactOrderAndState(t *testing.T) {
	w := newPlayerMakeDefItemsWorld4EF7D0()
	got := playerMakeDefItems4EF7D0(w.hooks())
	if got.kind != playerMakeDefItemsResultObject4EF7D0 || got.object != "made:Sword" {
		t.Fatalf("result = %#v, want object made:Sword", got)
	}
	if want := playerMakeDefItemsFullEvents4EF7D0(); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events mismatch\n got: %q\nwant: %q", w.events, want)
	}
	for i, sample := range w.healthSamples {
		if sample != 0x1234 {
			t.Fatalf("health sample %d = %#x, want 0x1234", i, sample)
		}
	}
	if w.objectFlags != playerMakeDefItemsObjectFlagMask4EF7D0 || w.trapCount != 0xaabbcc00 {
		t.Fatalf("flags/trap count = %#x/%#x", w.objectFlags, w.trapCount)
	}
	if got, want := w.deleted, []playerMakeDefItemsObject4EF7D0{"predicate-delete", "flag-delete", "armor-delete"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("deleted = %v, want %v", got, want)
	}
	if w.playerDone["player-a"] != 1 {
		t.Fatalf("player completion = %d, want 1", w.playerDone["player-a"])
	}
}

func TestPlayerMakeDefItems4EF7D0EarlyPlayerResults(t *testing.T) {
	t.Run("nil player", func(t *testing.T) {
		w := newPlayerMakeDefItemsWorld4EF7D0()
		w.gameFlags = map[uint32]bool{}
		w.player = ""
		got := playerMakeDefItems4EF7D0(w.hooks())
		if got.kind != playerMakeDefItemsResultPlayer4EF7D0 || got.player != "" {
			t.Fatalf("result = %#v, want nil player source", got)
		}
		if strings.Contains(strings.Join(w.events, "|"), "player-done:") || strings.Contains(strings.Join(w.events, "|"), "keep-items:") {
			t.Fatalf("nil player reached later loads: %v", w.events)
		}
	})

	t.Run("already complete", func(t *testing.T) {
		w := newPlayerMakeDefItemsWorld4EF7D0()
		w.gameFlags = map[uint32]bool{}
		w.playerDone["player-a"] = 9
		got := playerMakeDefItems4EF7D0(w.hooks())
		if got.kind != playerMakeDefItemsResultPlayer4EF7D0 || got.player != "player-a" {
			t.Fatalf("result = %#v, want player-a source", got)
		}
		if strings.Contains(strings.Join(w.events, "|"), "report-health:") {
			t.Fatalf("completed player reached reports: %v", w.events)
		}
	})
}

func TestPlayerMakeDefItems4EF7D0KeepItemsReturnsSendByte(t *testing.T) {
	w := newPlayerMakeDefItemsWorld4EF7D0()
	w.gameFlags = map[uint32]bool{}
	w.restoreStats = 0
	w.keepItems = 1
	got := playerMakeDefItems4EF7D0(w.hooks())
	if got.kind != playerMakeDefItemsResultByte4EF7D0 || got.value != 0xa0 {
		t.Fatalf("result = %#v, want send byte 0xa0", got)
	}
	joined := strings.Join(w.events, "|")
	for _, forbidden := range []string{"remove-poison:", "first-item:", "modifier-id:", "game-flag:0xa00"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("keep-items path contains %q: %v", forbidden, w.events)
		}
	}
	if w.playerDone["player-a"] != 1 {
		t.Fatalf("player completion = %d, want 1", w.playerDone["player-a"])
	}
}

func TestPlayerMakeDefItems4EF7D0DefaultEquipmentBranches(t *testing.T) {
	tests := []struct {
		name       string
		class      uint8
		flags      map[uint32]bool
		questReady int32
		wantTypes  []string
		wantKind   playerMakeDefItemsResultKind4EF7D0
		wantObject playerMakeDefItemsObject4EF7D0
		wantByte   uint8
	}{
		{name: "quest wizard", class: 1, flags: map[uint32]bool{0x1000: true}, questReady: 0, wantTypes: []string{"SulphorousFlareWand"}, wantKind: playerMakeDefItemsResultObject4EF7D0, wantObject: "made:SulphorousFlareWand"},
		{name: "quest rejected falls through", class: 2, flags: map[uint32]bool{0x1000: true}, questReady: -1, wantKind: playerMakeDefItemsResultByte4EF7D0, wantByte: 2},
		{name: "classic warrior", class: 0, flags: map[uint32]bool{}, wantTypes: []string{"Longsword", "WoodenShield"}, wantKind: playerMakeDefItemsResultObject4EF7D0, wantObject: "made:WoodenShield"},
		{name: "classic wizard", class: 1, flags: map[uint32]bool{}, wantTypes: []string{"WizardRobe"}, wantKind: playerMakeDefItemsResultObject4EF7D0, wantObject: "made:WizardRobe"},
		{name: "classic conjurer residue", class: 2, flags: map[uint32]bool{}, wantKind: playerMakeDefItemsResultByte4EF7D0, wantByte: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerMakeDefItemsWorld4EF7D0()
			w.restoreStats = 0
			w.gameFlags = test.flags
			w.questReady = test.questReady
			w.playerClass["player-a"] = test.class
			w.firstItem = ""
			w.armorEquip["player-a"] = 0x405
			got := playerMakeDefItems4EF7D0(w.hooks())
			if got.kind != test.wantKind ||
				(test.wantKind == playerMakeDefItemsResultObject4EF7D0 && got.object != test.wantObject) ||
				(test.wantKind == playerMakeDefItemsResultByte4EF7D0 && got.value != test.wantByte) {
				t.Fatalf("result = %#v, want kind/object/byte %d/%q/%d", got, test.wantKind, test.wantObject, test.wantByte)
			}
			var types []string
			for _, call := range w.respawns {
				types = append(types, call.typeID)
			}
			if !reflect.DeepEqual(types, test.wantTypes) {
				t.Fatalf("respawn types = %v, want %v", types, test.wantTypes)
			}
		})
	}
}

func TestPlayerMakeDefItems4EF7D0CachesInfoButReloadsPlayerAndHealth(t *testing.T) {
	w := newPlayerMakeDefItemsWorld4EF7D0()
	w.restoreStats = 0
	w.gameFlags = map[uint32]bool{0x0a00: true}
	w.firstItem = ""
	w.armorEquip["player-a"] = 0x405
	w.infoByPlayer["player-b"] = "info-b"
	w.playerIndex["player-b"] = 9
	w.playerDone["player-b"] = 0
	w.playerClass["player-b"] = 2
	w.armorEquip["player-b"] = 0x005
	w.infoBytes["info-b"] = map[uint8]uint8{84: 44, 85: 55}
	w.healthCurrent["health-b"] = 0x5678
	w.onEvent = func(event string) {
		switch event {
		case "info:player-a=info-a":
			w.healthByUnit["unit-a"] = "health-b"
		case "health-sample:update-a:0=0x5678":
			w.healthCurrent["health-b"] = 0x9abc
		case "report-health:7:unit-a":
			w.player = "player-b"
		}
	}
	got := playerMakeDefItems4EF7D0(w.hooks())
	if got.kind != playerMakeDefItemsResultByte4EF7D0 || got.value != 2 {
		t.Fatalf("result = %#v, want live player-b class byte 2", got)
	}
	if w.healthSamples[0] != 0x5678 || w.healthSamples[1] != 0x9abc || w.healthSamples[31] != 0x9abc {
		t.Fatalf("health samples do not use live HealthData/current: first=%#x second=%#x last=%#x", w.healthSamples[0], w.healthSamples[1], w.healthSamples[31])
	}
	joined := strings.Join(w.events, "|")
	if !strings.Contains(joined, "report-mana:9:unit-a") || !strings.Contains(joined, "info-byte:info-a:84=4") || strings.Contains(joined, "info-byte:info-b:") {
		t.Fatalf("cached/live load contract mismatch: %v", w.events)
	}
	if w.playerDone["player-b"] != 1 || w.playerDone["player-a"] != 0 {
		t.Fatalf("completion used wrong live player: a=%d b=%d", w.playerDone["player-a"], w.playerDone["player-b"])
	}
}

func testPlayerMakeDefItemsFaultPrefixes4EF7D0(
	t *testing.T,
	configure func(*playerMakeDefItemsWorld4EF7D0),
) {
	t.Helper()
	base := newPlayerMakeDefItemsWorld4EF7D0()
	configure(base)
	playerMakeDefItems4EF7D0(base.hooks())
	want := append([]string(nil), base.events...)
	if len(want) == 0 {
		t.Fatal("fault-prefix path has no observable event")
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event_%03d", faultAt), func(t *testing.T) {
			w := newPlayerMakeDefItemsWorld4EF7D0()
			configure(w)
			w.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				playerMakeDefItems4EF7D0(w.hooks())
			}()
			if !reflect.DeepEqual(w.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", w.events, want[:faultAt])
			}
		})
	}
}

func TestPlayerMakeDefItems4EF7D0EveryObservableFaultPrefix(t *testing.T) {
	testPlayerMakeDefItemsFaultPrefixes4EF7D0(t, func(*playerMakeDefItemsWorld4EF7D0) {})
}

func TestPlayerMakeDefItems4EF7D0EverySparsePathFaultPrefix(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*playerMakeDefItemsWorld4EF7D0)
	}{
		{
			name: "nil player",
			configure: func(w *playerMakeDefItemsWorld4EF7D0) {
				w.restoreStats = 0
				w.gameFlags = map[uint32]bool{}
				w.player = ""
			},
		},
		{
			name: "already complete",
			configure: func(w *playerMakeDefItemsWorld4EF7D0) {
				w.restoreStats = 0
				w.gameFlags = map[uint32]bool{}
				w.playerDone["player-a"] = 1
			},
		},
		{
			name: "keep inventory",
			configure: func(w *playerMakeDefItemsWorld4EF7D0) {
				w.restoreStats = 0
				w.gameFlags = map[uint32]bool{}
				w.keepItems = 1
			},
		},
		{
			name: "quest wizard",
			configure: func(w *playerMakeDefItemsWorld4EF7D0) {
				w.restoreStats = 0
				w.gameFlags = map[uint32]bool{0x1000: true}
				w.playerClass["player-a"] = 1
				w.armorEquip["player-a"] = 0x405
				w.firstItem = ""
				w.questReady = 0
			},
		},
		{
			name: "quest rejected classic conjurer",
			configure: func(w *playerMakeDefItemsWorld4EF7D0) {
				w.restoreStats = 0
				w.gameFlags = map[uint32]bool{0x1000: true}
				w.playerClass["player-a"] = 2
				w.armorEquip["player-a"] = 0x405
				w.firstItem = ""
				w.questReady = -1
			},
		},
		{
			name: "classic warrior",
			configure: func(w *playerMakeDefItemsWorld4EF7D0) {
				w.restoreStats = 0
				w.gameFlags = map[uint32]bool{}
				w.playerClass["player-a"] = 0
				w.armorEquip["player-a"] = 0x405
				w.firstItem = ""
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testPlayerMakeDefItemsFaultPrefixes4EF7D0(t, test.configure)
		})
	}
}
