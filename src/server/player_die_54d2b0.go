package server

import (
	"encoding/binary"

	"github.com/opennox/libs/object"
)

const (
	playerDieKotrMode54D2B0       = uint32(0x0010)
	playerDieArenaMode54D2B0      = uint32(0x0100)
	playerDieElimMode54D2B0       = uint32(0x0400)
	playerDieCoopMode54D2B0       = uint32(0x0800)
	playerDieQuestMode54D2B0      = uint32(0x1000)
	playerDieOnlineMode54D2B0     = uint32(0x2000)
	playerDieElectricDamage54D2B0 = uint32(16)
	playerDieElectricSound54D2B0  = 299
	playerDieMaleSound54D2B0      = 321
	playerDieFemaleSound54D2B0    = 331
	playerDieTrapSpellCount54D2B0 = 5
)

// PlayerDieRuntime54D2B0 contains the services used by the pointer-width
// independent slices of GAME.EXE 0054D2B0. Unsupported is called before any
// player state is changed.
type PlayerDieRuntime54D2B0 struct {
	GameFlag           func(uint32) bool
	Frame              func() uint32
	TickRate           func() uint32
	PlayerByIndex      func(uint32) *Player
	ObjectByNetCode    func(uint32) *Object
	InformText         func(int, [14]byte)
	ResetAbility       func(*Object, int32)
	GameplayHasRivals  func() bool
	PrepareAnkhType    func()
	CancelPendingSave  func()
	Audio              func(int, *Object)
	SetPlayerState     func(*Object, PlayerState) bool
	RemoveActionShadow func(*Object)
	DropAllItems       func(*Object) int32
	NotifyPlayerDied   func(*Object)
	ProtectMana        func(uint32, int16)
	SetBuffFlags       func(*Object, uint32)
	CancelAbilities    func(*Object)
	CancelSpells       func(*Object)
	CancelTrade        func(*TradeSession)
	Unsupported        func(string, *Object)
}

func playerDieUnsupported54D2B0(runtime PlayerDieRuntime54D2B0, reason string, unit *Object) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, unit)
	}
	return false
}

func playerDieRuntimeReady54D2B0(runtime PlayerDieRuntime54D2B0) bool {
	return runtime.GameFlag != nil &&
		runtime.PrepareAnkhType != nil &&
		runtime.Audio != nil &&
		runtime.SetPlayerState != nil &&
		runtime.RemoveActionShadow != nil &&
		runtime.DropAllItems != nil &&
		runtime.NotifyPlayerDied != nil &&
		runtime.ProtectMana != nil &&
		runtime.SetBuffFlags != nil &&
		runtime.CancelAbilities != nil &&
		runtime.CancelSpells != nil
}

func playerDieOnlineRuntimeReady54D2B0(runtime PlayerDieRuntime54D2B0) bool {
	return runtime.Frame != nil &&
		runtime.TickRate != nil &&
		runtime.PlayerByIndex != nil &&
		runtime.ObjectByNetCode != nil &&
		runtime.InformText != nil &&
		runtime.ResetAbility != nil
}

func playerDieRecentAggressor54D2B0(player *Player, primary, victim *Object, runtime PlayerDieRuntime54D2B0) *Object {
	pending, playerIndex, frame := player.LastAggressorState()
	if pending == 0 || runtime.Frame()-frame >= 10*runtime.TickRate() {
		return nil
	}
	aggressor := runtime.PlayerByIndex(playerIndex)
	if aggressor == nil || aggressor.Active == 0 || aggressor.PlayerUnit == nil {
		return nil
	}
	assist := runtime.ObjectByNetCode(aggressor.NetCodeVal)
	if assist == primary || assist == victim {
		return nil
	}
	return assist
}

func playerDieOnlinePacket54D2B0(unit *Object, update *PlayerUpdateData, primary, assist *Object) [14]byte {
	var packet [14]byte
	if primary != nil && primary.Class().Has(object.ClassPlayer) {
		binary.LittleEndian.PutUint16(packet[2:], uint16(primary.NetCode))
	}
	if assist != nil && assist.Class().Has(object.ClassPlayer) {
		binary.LittleEndian.PutUint16(packet[4:], uint16(assist.NetCode))
	}
	binary.LittleEndian.PutUint16(packet[6:], uint16(unit.NetCode))

	var sourceID uint16
	var sourceKind byte
	source := unit.Obj130
	if source != nil {
		switch {
		case source.Class().Has(object.ClassMonster):
			sourceID = source.TypeInd
			sourceKind = 1
		case source.Class().Has(object.ClassPlayer):
			sourceID = uint16(update.Field75)
			sourceKind = byte(update.Field76)
		case source.ObjOwner != nil && source.ObjOwner.Class().Has(object.ClassPlayer):
			sourceID = uint16(update.Field75)
			sourceKind = byte(update.Field76)
		case source.ObjOwner != nil && source.ObjOwner.Class().Has(object.ClassMonster):
			sourceID = source.ObjOwner.TypeInd
			sourceKind = 1
		}
	}
	if sourceKind == 0 {
		sourceID = uint16(update.Field75)
		sourceKind = byte(update.Field76)
	}
	binary.LittleEndian.PutUint16(packet[8:], sourceID)
	packet[10] = sourceKind
	return packet
}

// PlayerDieNative54D2B0 restores GAME.EXE 0054D2B0 for offline solo
// cooperative play and hosted online play without a competitive scoring
// opponent. Competitive scoring, Elimination cleanup, and Quest lives remain
// separately admitted branches. The complete gate is evaluated before
// PrepareAnkhType, so a 64-bit build never partially executes an unsupported
// PE32 callback.
func PlayerDieNative54D2B0(unit *Object, runtime PlayerDieRuntime54D2B0) bool {
	if unit == nil || !unit.ObjClass.Has(object.ClassPlayer) || unit.UpdateData == nil {
		return playerDieUnsupported54D2B0(runtime, "non-player unit", unit)
	}
	update := unit.UpdateDataPlayer()
	player := update.Player
	if player == nil || player.PlayerUnit != unit {
		return playerDieUnsupported54D2B0(runtime, "invalid player binding", unit)
	}
	if unit.HealthData == nil || unit.HealthData.Cur != 0 || !unit.ObjFlags.Has(object.FlagDead) {
		return playerDieUnsupported54D2B0(runtime, "player is not at lethal state", unit)
	}
	if !playerDieRuntimeReady54D2B0(runtime) {
		return playerDieUnsupported54D2B0(runtime, "missing native death service", unit)
	}
	coop := runtime.GameFlag(playerDieCoopMode54D2B0)
	online := runtime.GameFlag(playerDieOnlineMode54D2B0)
	if runtime.GameFlag(playerDieQuestMode54D2B0) {
		return playerDieUnsupported54D2B0(runtime, "quest mode", unit)
	}
	if !coop && !online {
		return playerDieUnsupported54D2B0(runtime, "unsupported game mode", unit)
	}
	if coop && runtime.CancelPendingSave == nil {
		return playerDieUnsupported54D2B0(runtime, "missing cooperative death service", unit)
	}
	if runtime.GameFlag(playerDieElimMode54D2B0) {
		return playerDieUnsupported54D2B0(runtime, "elimination mode", unit)
	}
	if online && !playerDieOnlineRuntimeReady54D2B0(runtime) {
		return playerDieUnsupported54D2B0(runtime, "missing online death service", unit)
	}
	if runtime.GameFlag(playerDieArenaMode54D2B0) || runtime.GameFlag(playerDieKotrMode54D2B0) {
		if runtime.GameplayHasRivals == nil {
			return playerDieUnsupported54D2B0(runtime, "missing competitive death service", unit)
		}
		if runtime.GameplayHasRivals() {
			return playerDieUnsupported54D2B0(runtime, "competitive scoring", unit)
		}
	}
	if update.Trade70 != nil && runtime.CancelTrade == nil {
		return playerDieUnsupported54D2B0(runtime, "missing shop death service", unit)
	}

	runtime.PrepareAnkhType()
	if coop {
		runtime.CancelPendingSave()
	}
	if online {
		source := unit.Obj130
		var primary *Object
		if source != nil {
			primary = source.FindOwnerChainPlayer()
		}
		assist := playerDieRecentAggressor54D2B0(player, primary, unit, runtime)
		packet := playerDieOnlinePacket54D2B0(unit, update, primary, assist)
		runtime.InformText(14, packet)
		if packet[10] == 2 && binary.LittleEndian.Uint16(packet[8:]) == 2 {
			runtime.ResetAbility(primary, 1)
		}
		update.Field76 = 0
	}

	sound := playerDieMaleSound54D2B0
	if unit.Field131 == playerDieElectricDamage54D2B0 {
		sound = playerDieElectricSound54D2B0
	} else if player.Info().IsFemale() {
		sound = playerDieFemaleSound54D2B0
	}
	runtime.Audio(sound, unit)

	unit.ObjFlags |= object.FlagDead
	runtime.SetPlayerState(unit, PlayerState3)
	update.Field47_0 = 0
	update.SpellCastStart = 0
	for i := 0; i < playerDieTrapSpellCount54D2B0; i++ {
		update.TrapSpells[i] = 0
	}
	update.TrapSpellsCnt &^= 0xff

	unit.ObjFlags |= object.FlagShort
	runtime.RemoveActionShadow(unit)
	runtime.DropAllItems(unit)
	runtime.NotifyPlayerDied(unit)

	update.ManaCur = 0
	runtime.ProtectMana(player.ProtUnitManaCur, 0)
	runtime.SetBuffFlags(unit, 0)
	runtime.CancelAbilities(unit)
	player.Field3600 = 0
	runtime.CancelSpells(unit)
	runtime.SetBuffFlags(unit, 0)
	for i := range unit.BuffsDur {
		unit.BuffsDur[i] = 0
		unit.BuffsPower[i] = 0
	}
	if update.Trade70 != nil {
		runtime.CancelTrade(update.Trade70)
	}
	update.Trade70 = nil
	return true
}
