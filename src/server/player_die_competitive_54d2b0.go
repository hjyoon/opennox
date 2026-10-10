package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/object"
)

// PlayerDieCompetitiveRuntime54D2B0 supplies the three already-native scoring
// roots and the elimination cleanup services. The original limited entry keeps
// its admission contract; the complete entry must supply every service before
// changing player state. Callbacks may change live mode/player fields later.
type PlayerDieCompetitiveRuntime54D2B0 struct {
	Arena         func(victim, killer, assist *Object, tracking uint32)
	Kotr          func(victim, killer *Object)
	Elimination   func(victim, killer *Object)
	GameDataLimit func(uint16) uint16
	RemoveSpawned func(*Object)
}

func playerDieCompetitiveRuntimeReady54D2B0(rt *PlayerDieCompetitiveRuntime54D2B0) bool {
	return rt != nil && rt.Arena != nil && rt.Kotr != nil && rt.Elimination != nil &&
		rt.GameDataLimit != nil && rt.RemoveSpawned != nil
}

// PlayerDieCompetitiveNative54D2B0 admits competitive death only with a complete
// native service set. It shares the cooperative/online/Quest body and tail;
// neither the PE32 fallback nor an integer object pointer is used here.
func PlayerDieCompetitiveNative54D2B0(unit *Object, runtime PlayerDieRuntime54D2B0, competitive PlayerDieCompetitiveRuntime54D2B0) bool {
	return playerDieNative54D2B0(unit, runtime, &competitive)
}

// The original fourth Arena argument is the nonzero aggressor Player identity,
// not the assisting object. A duplicate killer/victim or failed net-code lookup
// clears the object but does not clear a valid active aggressor record.
func playerDieTrackedAggressor54D2B0(player *Player, primary, victim *Object, runtime PlayerDieRuntime54D2B0) (*Object, uint32) {
	if player.Field3600 == 0 {
		return nil, 0
	}
	frame := runtime.Frame()
	elapsed := frame - player.field3608
	if elapsed >= 10*runtime.TickRate() {
		return nil, 0
	}
	aggressor := runtime.PlayerByIndex(player.Field3604)
	if aggressor == nil || *(*uint32)(unsafe.Pointer(&aggressor.Active)) == 0 || aggressor.PlayerUnit == nil {
		return nil, 0
	}
	assist := runtime.ObjectByNetCode(aggressor.NetCodeVal)
	if assist == primary || assist == victim {
		assist = nil
	}
	return assist, 1
}

// 0054D35A tests the whole DWORD. Player.Active is its named low byte followed
// by the original three reserved bytes, not a narrowed DWORD truth value.
var _ = [1]struct{}{}[4-(unsafe.Offsetof(Player{}.Field2096Buf)-unsafe.Offsetof(Player{}.Active))]

func playerDieNative54D2B0(unit *Object, runtime PlayerDieRuntime54D2B0, competitive *PlayerDieCompetitiveRuntime54D2B0) bool {
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
	quest := runtime.GameFlag(playerDieQuestMode54D2B0)
	if quest && !playerDieQuestRuntimeReady54D2B0(runtime) {
		return playerDieUnsupported54D2B0(runtime, "missing quest death service", unit)
	}
	competitiveMode := competitive != nil && (runtime.GameFlag(playerDieArenaMode54D2B0) ||
		runtime.GameFlag(playerDieKotrMode54D2B0) || runtime.GameFlag(playerDieElimMode54D2B0))
	if !coop && !online && !quest && !competitiveMode {
		return playerDieUnsupported54D2B0(runtime, "unsupported game mode", unit)
	}
	if coop && runtime.CancelPendingSave == nil {
		return playerDieUnsupported54D2B0(runtime, "missing cooperative death service", unit)
	}
	if competitive == nil && runtime.GameFlag(playerDieElimMode54D2B0) {
		return playerDieUnsupported54D2B0(runtime, "elimination mode", unit)
	}
	if online && !playerDieOnlineRuntimeReady54D2B0(runtime) {
		return playerDieUnsupported54D2B0(runtime, "missing online death service", unit)
	}
	if competitive != nil {
		if !playerDieCompetitiveRuntimeReady54D2B0(competitive) {
			return playerDieUnsupported54D2B0(runtime, "missing competitive death service", unit)
		}
	} else if runtime.GameFlag(playerDieArenaMode54D2B0) || runtime.GameFlag(playerDieKotrMode54D2B0) {
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
	if runtime.GameFlag(playerDieCoopMode54D2B0) {
		runtime.CancelPendingSave()
	}
	source := unit.Obj130
	var primary *Object
	if source != nil {
		primary = source.FindOwnerChainPlayer()
	}
	var assist *Object
	var tracking uint32
	if competitive != nil {
		assist, tracking = playerDieTrackedAggressor54D2B0(update.Player, primary, unit, runtime)
	} else {
		assist = playerDieRecentAggressor54D2B0(update.Player, primary, unit, runtime)
	}
	if runtime.GameFlag(playerDieOnlineMode54D2B0) {
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
	} else {
		female := false
		if competitive != nil {
			// 0054D4F2 reloads the cached update's Player before reading
			// its gender byte. Info/IsFemale's nil defaults would skip
			// the original invalid-binding fault and run later services.
			female = (*PlayerInfo)(unsafe.Pointer(&update.Player.info)).isFemale != 0
		} else {
			female = update.Player.Info().IsFemale()
		}
		if female {
			sound = playerDieFemaleSound54D2B0
		}
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
	rivals := runtime.GameplayHasRivals()
	if competitive != nil {
		if rivals {
			switch {
			case runtime.GameFlag(playerDieArenaMode54D2B0):
				competitive.Arena(unit, primary, assist, tracking)
			case runtime.GameFlag(playerDieKotrMode54D2B0):
				competitive.Kotr(unit, primary)
			case runtime.GameFlag(playerDieElimMode54D2B0):
				competitive.Elimination(unit, primary)
			}
		}
		if runtime.GameFlag(playerDieElimMode54D2B0) && competitive.GameDataLimit(uint16(playerDieElimMode54D2B0)) != 0 {
			// 0054D5F4 calls Get again before the Player load at 0054D5F9.
			// 0054D60D is JL: the death DWORD is compared as signed int32,
			// while the limit is zero-extended from its WORD return.
			limit := competitive.GameDataLimit(uint16(playerDieElimMode54D2B0))
			if int32(update.Player.Field2140) >= int32(limit) {
				competitive.RemoveSpawned(unit)
			}
		}
	}

	unit.ObjFlags |= object.FlagShort
	runtime.RemoveActionShadow(unit)
	if !runtime.GameFlag(playerDieQuestMode54D2B0) {
		runtime.DropAllItems(unit)
	}
	runtime.NotifyPlayerDied(unit)

	player = update.Player
	update.ManaCur = 0
	runtime.ProtectMana(player.ProtUnitManaCur, 0)
	runtime.SetBuffFlags(unit, 0)
	runtime.CancelAbilities(unit)
	update.Player.Field3600 = 0
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
	playerDieQuestNative54D2B0(unit, update, runtime)
	return true
}
