package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// PlayerResetRuntime4EFF10 supplies services that remain owned by the legacy
// runtime. Object, UpdateData, and Player pointers stay native-width; mana,
// protection tokens, flags, indices, and the return value retain their
// original fixed-width contracts.
type PlayerResetRuntime4EFF10 struct {
	AwardBeastScrolls     func(*Player)
	AwardSpells           func(*Player)
	CancelAbilities       func(*Object)
	ReadValues            func(*Object, int32)
	AwardWarriorAbilities func(*Player)
	ProtectMana           func(uint32, uint16)
	SetHealthMaximum      func(*Object)
	SetPlayerState        func(*Object, PlayerState)
	ClearBuffs            func(*Object)
	CancelSpells          func(*Object)
	RemovePoison          func(*Object)
	ResetPlayerRuntime    func(*Object)
	ReportTotalHealth     func(uint8, *Object)
	ReportTotalMana       func(uint8, *Object)
}

func playerResetNative4EFF10(unit *Object, runtime PlayerResetRuntime4EFF10) int32 {
	return playerReset4EFF10(playerResetHooks4EFF10[
		*Object,
		*PlayerUpdateData,
		*Player,
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
		awardBeastScrolls:     runtime.AwardBeastScrolls,
		awardSpells:           runtime.AwardSpells,
		storePlayerLevel:      func(player *Player, value uint8) { player.Level = value },
		cancelAbilities:       runtime.CancelAbilities,
		readValues:            runtime.ReadValues,
		awardWarriorAbilities: runtime.AwardWarriorAbilities,
		loadManaMaximum:       func(update *PlayerUpdateData) uint16 { return update.ManaMax },
		storeManaCurrent:      func(update *PlayerUpdateData, value uint16) { update.ManaCur = value },
		storeManaPrevious:     func(update *PlayerUpdateData, value uint16) { update.ManaPrev = value },
		loadManaToken:         func(player *Player) uint32 { return player.ProtUnitManaCur },
		protectMana:           runtime.ProtectMana,
		storeTrapSpell: func(update *PlayerUpdateData, index int, value uint32) {
			update.TrapSpells[index] = value
		},
		storeTrapCountLow: func(update *PlayerUpdateData, value uint8) {
			update.TrapSpellsCnt = update.TrapSpellsCnt&^uint32(0xff) | uint32(value)
		},
		setHealthMaximum: runtime.SetHealthMaximum,
		loadObjectFlags:  func(unit *Object) uint32 { return uint32(unit.ObjFlags) },
		storeObjectField541: func(unit *Object, value uint8) {
			unit.Field541 = value
		},
		storeObjectFlags: func(unit *Object, value uint32) {
			unit.ObjFlags = object.Flags(value)
		},
		setPlayerState:     runtime.SetPlayerState,
		clearBuffs:         runtime.ClearBuffs,
		cancelSpells:       runtime.CancelSpells,
		removePoison:       runtime.RemovePoison,
		resetPlayerRuntime: runtime.ResetPlayerRuntime,
		loadPlayerIndex:    func(player *Player) uint8 { return player.PlayerInd },
		reportTotalHealth:  runtime.ReportTotalHealth,
		reportTotalMana:    runtime.ReportTotalMana,
		storeObject130: func(unit, value *Object) {
			unit.Obj130 = value
		},
		storePlayerMarker3664: func(player *Player, value uint32) {
			player.field3664 = value
		},
		storePlayerMarker3660: func(player *Player, value uint32) {
			player.field3660 = value
		},
	})
}

// PlayerReset4EFF10 binds GAME.EXE 004EFF10 to native-width Object,
// PlayerUpdateData, and Player storage without changing the original load,
// callback, store, or return order.
func (s *Server) PlayerReset4EFF10(unit *Object, runtime PlayerResetRuntime4EFF10) int32 {
	return playerResetNative4EFF10(unit, runtime)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(PlayerUpdateData{}.TrapSpells[0])]
	_ = [1]struct{}{}[4-unsafe.Sizeof(PlayerUpdateData{}.TrapSpellsCnt)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.ProtUnitManaCur)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.field3660)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.field3664)]
)
