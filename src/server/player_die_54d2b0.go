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
	Quest              *PlayerDieQuestRuntime54D2B0
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
		runtime.Frame != nil && runtime.TickRate != nil &&
		runtime.PlayerByIndex != nil && runtime.ObjectByNetCode != nil &&
		runtime.GameplayHasRivals != nil &&
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
	if player.Field3600 == 0 {
		return nil
	}
	frame := runtime.Frame()
	elapsed := frame - player.field3608
	if elapsed >= 10*runtime.TickRate() {
		return nil
	}
	aggressor := runtime.PlayerByIndex(player.Field3604)
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
// opponent, including Quest lives and the ordered Quest penalty branch.
// Competitive scoring and Elimination cleanup remain separately admitted
// branches. The complete gate is evaluated before
// PrepareAnkhType, so a 64-bit build never partially executes an unsupported
// PE32 callback.
func PlayerDieNative54D2B0(unit *Object, runtime PlayerDieRuntime54D2B0) bool {
	return playerDieNative54D2B0(unit, runtime, nil)
}
