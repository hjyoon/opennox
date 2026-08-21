package server

import (
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

var (
	playerMakeDefItemsMode2048Types4EF7D0 = [...]string{"Sword", "StaffWooden", "StaffWooden"}
	playerMakeDefItemsMode4096Types4EF7D0 = [...]string{"Sword", "SulphorousFlareWand", "Bow"}
)

// PlayerMakeDefItemsRuntime4EF7D0 supplies services still owned by the outer
// server or legacy package. Every pointer crossing this boundary retains its
// native width; only the public result is the original function's low AL byte.
type PlayerMakeDefItemsRuntime4EF7D0 struct {
	RemovePoison       func(*Object)
	SetHealthMaximum   func(*Object)
	RefreshMana        func(*Object)
	CancelAbilities    func(*Object)
	SetPlayerState     func(*Object, PlayerState)
	ClearBuffs         func(*Object)
	ResetPlayerRuntime func(*Object)
	ReportTotalHealth  func(uint8, *Object)
	ReportTotalMana    func(uint8, *Object)
	SendRespawn        func(*Object, uint8) uint8
	DelayedDelete      func(*Object)
	RespawnItem        func(*Object, string, *ModifierInitData, int32, int32) *Object
	QuestDefaultsReady func() int32
}

type playerMakeDefItemsNativeDeps4EF7D0 struct {
	removePoison       func(*Object)
	setHealthMaximum   func(*Object)
	refreshMana        func(*Object)
	cancelAbilities    func(*Object)
	resetCamping       func(*Object)
	setPlayerState     func(*Object, PlayerState)
	clearBuffs         func(*Object)
	resetPlayerRuntime func(*Object)
	gameFlag           func(uint32) bool
	markPlayerObjects  func(uint8)
	reportTotalHealth  func(uint8, *Object)
	reportTotalMana    func(uint8, *Object)
	sendRespawn        func(*Object, uint8) uint8
	armorEquipFlags    func(*Object) uint32
	delayedDelete      func(*Object)
	modifierID         func(string) uint32
	modifierDesc       func(uint32) *ModifierEff
	respawnItem        func(*Object, string, *ModifierInitData, int32, int32) *Object
	questDefaultsReady func() int32
}

func playerMakeDefItemsNative4EF7D0(
	unit *Object,
	restoreStats int32,
	keepItems int32,
	deps playerMakeDefItemsNativeDeps4EF7D0,
) playerMakeDefItemsResult4EF7D0[*Object, *Player] {
	return playerMakeDefItems4EF7D0(playerMakeDefItemsHooks4EF7D0[
		*Object,
		*PlayerUpdateData,
		*Player,
		*PlayerInfo,
		*HealthData,
		*ModifierEff,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerInfo: func(player *Player) *PlayerInfo {
			return player.Info()
		},
		loadRestoreStatsArg: func() int32 {
			return restoreStats
		},
		removePoison:     deps.removePoison,
		setHealthMaximum: deps.setHealthMaximum,
		refreshMana:      deps.refreshMana,
		cancelAbilities:  deps.cancelAbilities,
		resetCamping:     deps.resetCamping,
		storeQuestExit: func(update *PlayerUpdateData, value *Object) {
			update.QuestExit = value
		},
		storeQuestWarpGate: func(update *PlayerUpdateData, value *Object) {
			update.QuestWarpGate = value
		},
		storeUnitField541: func(unit *Object, value uint8) {
			unit.Field541 = value
		},
		storeUpdateField21: func(update *PlayerUpdateData, value uint32) {
			update.Field21 = value
		},
		storeUpdateField19: func(update *PlayerUpdateData, value uint16) {
			update.Field19_1 = value
		},
		storeHealthCursor: func(update *PlayerUpdateData, value uint16) {
			update.HealthSampleCur = value
		},
		loadHealthData: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadHealthCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		storeHealthSample: func(update *PlayerUpdateData, index int, value uint16) {
			update.HealthSamples[index] = value
		},
		loadObjectFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		storeObjectFlags: func(unit *Object, value uint32) {
			unit.ObjFlags = object.Flags(value)
		},
		setPlayerState: deps.setPlayerState,
		clearBuffs:     deps.clearBuffs,
		storeUpdateField47: func(update *PlayerUpdateData, value uint8) {
			update.Field47_0 = value
		},
		storeSpellCastStart: func(update *PlayerUpdateData, value uint32) {
			update.SpellCastStart = value
		},
		storeTrapSpell: func(update *PlayerUpdateData, index int, value uint32) {
			update.TrapSpells[index] = value
		},
		storeTrapCountLow: func(update *PlayerUpdateData, value uint8) {
			update.TrapSpellsCnt = update.TrapSpellsCnt&^0xff | uint32(value)
		},
		storeHarpoonBolt: func(update *PlayerUpdateData, value *Object) {
			update.HarpoonBolt = value
		},
		storeHarpoonTarget: func(update *PlayerUpdateData, value *Object) {
			update.HarpoonTarg = value
		},
		storeUpdateField67: func(update *PlayerUpdateData, value uint32) {
			update.Field67 = value
		},
		resetPlayerRuntime: deps.resetPlayerRuntime,
		storeObject130: func(unit, value *Object) {
			unit.Obj130 = value
		},
		gameFlag: deps.gameFlag,
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		markPlayerObjects: deps.markPlayerObjects,
		loadPlayerDone: func(player *Player) uint32 {
			return player.Field4700
		},
		reportTotalHealth: deps.reportTotalHealth,
		reportTotalMana:   deps.reportTotalMana,
		loadKeepItemsArg: func() int32 {
			return keepItems
		},
		sendRespawn: deps.sendRespawn,
		loadFirstItem: func(unit *Object) *Object {
			return unit.InvFirstItem
		},
		loadNextItem: func(item *Object) *Object {
			return item.InvNextItem
		},
		itemMustDelete: func(item *Object) bool {
			if uint32(item.ObjClass)&playerMakeDefItemsArmorClass4EF7D0 == 0 {
				return true
			}
			return deps.armorEquipFlags(item)&0x0c0d == 0
		},
		loadItemFlags: func(item *Object) uint32 {
			return uint32(item.ObjFlags)
		},
		loadItemClass: func(item *Object) uint32 {
			return uint32(item.ObjClass)
		},
		loadArmorEquipFlags: deps.armorEquipFlags,
		delayedDelete:       deps.delayedDelete,
		modifierID:          deps.modifierID,
		modifierDesc:        deps.modifierDesc,
		modifierBase: func(modifier *ModifierEff) uint32 {
			return modifier.ind4
		},
		loadPlayerClass: func(player *Player) uint8 {
			return player.info[66]
		},
		loadArmorEquip: func(player *Player) uint32 {
			return player.ArmorEquip
		},
		loadArmorEquipLow: func(player *Player) uint8 {
			return uint8(player.ArmorEquip)
		},
		loadInfoByte: func(info *PlayerInfo, offset uint8) uint8 {
			switch offset {
			case 83:
				return info.Colors.Pants
			case 84:
				return info.Colors.Shirt1
			case 85:
				return info.Colors.Shirt2
			case 86:
				return info.Colors.Shoes1
			case 87:
				return info.Colors.Shoes2
			default:
				panic("unexpected PlayerInfo byte")
			}
		},
		respawnItem: func(unit *Object, typeID string, attrs *[4]*ModifierEff, a4, a5 int32) *Object {
			if attrs == nil {
				return deps.respawnItem(unit, typeID, nil, a4, a5)
			}
			data := &ModifierInitData{Modifiers: *attrs}
			return deps.respawnItem(unit, typeID, data, a4, a5)
		},
		loadMode2048Type: func(class uint8) string {
			return playerMakeDefItemsMode2048Types4EF7D0[class]
		},
		questDefaultsReady: deps.questDefaultsReady,
		loadMode4096Type: func(class uint8) string {
			return playerMakeDefItemsMode4096Types4EF7D0[class]
		},
		storePlayerDone: func(player *Player, value uint32) {
			player.Field4700 = value
		},
	})
}

func playerMakeDefItemsServerDeps4EF7D0(
	s *Server,
	runtime PlayerMakeDefItemsRuntime4EF7D0,
) playerMakeDefItemsNativeDeps4EF7D0 {
	return playerMakeDefItemsNativeDeps4EF7D0{
		removePoison:       runtime.RemovePoison,
		setHealthMaximum:   runtime.SetHealthMaximum,
		refreshMana:        runtime.RefreshMana,
		cancelAbilities:    runtime.CancelAbilities,
		resetCamping:       s.Sub_4D7E50,
		setPlayerState:     runtime.SetPlayerState,
		clearBuffs:         runtime.ClearBuffs,
		resetPlayerRuntime: runtime.ResetPlayerRuntime,
		gameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		markPlayerObjects: func(index uint8) {
			s.Sub4DE4D0(int(index))
		},
		reportTotalHealth: runtime.ReportTotalHealth,
		reportTotalMana:   runtime.ReportTotalMana,
		sendRespawn:       runtime.SendRespawn,
		armorEquipFlags:   s.Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70,
		delayedDelete:     runtime.DelayedDelete,
		modifierID: func(name string) uint32 {
			return uint32(s.Modif.Nox_xxx_modifGetIdByName413290(name))
		},
		modifierDesc: func(id uint32) *ModifierEff {
			return s.Modif.Nox_xxx_modifGetDescById413330(int(id))
		},
		respawnItem:        runtime.RespawnItem,
		questDefaultsReady: runtime.QuestDefaultsReady,
	}
}

func playerMakeDefItemsResultLow4EF7D0(
	result playerMakeDefItemsResult4EF7D0[*Object, *Player],
) uint8 {
	switch result.kind {
	case playerMakeDefItemsResultPlayer4EF7D0:
		return uint8(uintptr(unsafe.Pointer(result.player)))
	case playerMakeDefItemsResultObject4EF7D0:
		return uint8(uintptr(unsafe.Pointer(result.object)))
	default:
		return result.value
	}
}

// PlayerMakeDefItems4EF7D0 binds GAME.EXE 004EF7D0 to native-width Object,
// PlayerUpdateData, Player, PlayerInfo, HealthData, inventory, and modifier
// pointers. restoreStats and keepItems retain the original fixed-width int32
// inputs; the return value is the original public low-AL byte.
func (s *Server) PlayerMakeDefItems4EF7D0(
	unit *Object,
	restoreStats int32,
	keepItems int32,
	runtime PlayerMakeDefItemsRuntime4EF7D0,
) uint8 {
	result := playerMakeDefItemsNative4EF7D0(
		unit,
		restoreStats,
		keepItems,
		playerMakeDefItemsServerDeps4EF7D0(s, runtime),
	)
	return playerMakeDefItemsResultLow4EF7D0(result)
}
