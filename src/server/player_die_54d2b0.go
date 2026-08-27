package server

import "github.com/opennox/libs/object"

const (
	playerDieCoopMode54D2B0       = uint32(0x0800)
	playerDieQuestMode54D2B0      = uint32(0x1000)
	playerDieOnlineMode54D2B0     = uint32(0x2000)
	playerDieCompetitive54D2B0    = uint32(0x0510)
	playerDieElectricDamage54D2B0 = uint32(16)
	playerDieElectricSound54D2B0  = 299
	playerDieMaleSound54D2B0      = 321
	playerDieFemaleSound54D2B0    = 331
	playerDieTrapSpellCount54D2B0 = 5
)

// PlayerDieRuntime54D2B0 contains the services used by the pointer-width
// independent solo-cooperative slice of GAME.EXE 0054D2B0. Unsupported is
// called before any player state is changed.
type PlayerDieRuntime54D2B0 struct {
	GameFlag           func(uint32) bool
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
		runtime.CancelPendingSave != nil &&
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

// PlayerDieNative54D2B0 restores GAME.EXE 0054D2B0 for an offline solo
// cooperative player. Online kill attribution, competitive scoring, shops,
// and Quest lives remain separately admitted branches. The complete gate is
// evaluated before PrepareAnkhType, matching the migration rule that a 64-bit
// build must never partially execute an unsupported PE32 callback.
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
	if !runtime.GameFlag(playerDieCoopMode54D2B0) ||
		runtime.GameFlag(playerDieQuestMode54D2B0) ||
		runtime.GameFlag(playerDieOnlineMode54D2B0) ||
		runtime.GameFlag(playerDieCompetitive54D2B0) {
		return playerDieUnsupported54D2B0(runtime, "non-solo-cooperative mode", unit)
	}
	if update.Trade70 != nil {
		return playerDieUnsupported54D2B0(runtime, "active shop session", unit)
	}

	runtime.PrepareAnkhType()
	runtime.CancelPendingSave()

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
	update.Trade70 = nil
	return true
}
