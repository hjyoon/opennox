package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// PlayerInputAttackRuntime4F9C70 supplies the two state-changing services
// still owned by the outer server package.
type PlayerInputAttackRuntime4F9C70 struct {
	SetState func(*Object, PlayerState) bool
	BuffOff  func(*Object, int32) int32
}

func playerAimsAtEnemyNative4F9DC0(s *Server, unit *Object) int32 {
	return playerAimsAtEnemy4F9DC0(unit, playerAimsAtEnemyHooks4F9DC0[*Object, *PlayerUpdateData]{
		loadUpdate: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadCursor: func(update *PlayerUpdateData) *Object {
			return update.CursorObj
		},
		isEnemy: s.IsEnemyTo,
		questMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
	})
}

// PlayerAimsAtEnemy4F9DC0 binds GAME.EXE 004F9DC0 to native-width Object and
// PlayerUpdateData pointers.
//
//go:noinline
func (s *Server) PlayerAimsAtEnemy4F9DC0(unit *Object) int32 {
	return playerAimsAtEnemyNative4F9DC0(s, unit)
}

func playerInputAttackNative4F9C70(
	s *Server,
	unit *Object,
	runtime PlayerInputAttackRuntime4F9C70,
) {
	playerInputAttack4F9C70(unit, playerInputAttackHooks4F9C70[
		*Object, *PlayerUpdateData, *Player, *Object, *WandUseData,
	]{
		aimsAtEnemy: s.PlayerAimsAtEnemy4F9DC0,
		loadUpdate: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadWeaponEquip: func(player *Player) uint32 {
			return player.WeaponEquip
		},
		actionState: s.PlayerActionState4FA2B0,
		loadWeapon: func(update *PlayerUpdateData) *Object {
			return update.EquippedWeapon
		},
		loadWeaponUseData: func(weapon *Object) *WandUseData {
			return (*WandUseData)(weapon.UseData.Ptr)
		},
		loadCharge: func(useData *WandUseData) uint8 {
			return useData.Charge
		},
		loadMaxCharge: func(useData *WandUseData) uint8 {
			return useData.MaxCharge
		},
		subStamina: s.PlayerSubStamina4F7D30,
		loadWeaponFlags: func(useData *WandUseData) uint32 {
			return useData.Flags
		},
		storeWeaponFlags: func(useData *WandUseData, flags uint32) {
			useData.Flags = flags
		},
		loadFrame: s.Frame,
		storeAttackFrame: func(unit *Object, frame uint32) {
			unit.Field34 = frame
		},
		storeAnimFrame: func(update *PlayerUpdateData, frame uint8) {
			update.Field59_0 = frame
		},
		setState: runtime.SetState,
		useByNetCode: func(owner, weapon *Object) int32 {
			return s.UseByNetCode53F8E0(owner, weapon)
		},
		loadState: func(update *PlayerUpdateData) PlayerState {
			return update.State
		},
		weaponStamina: WeaponStaminaByType4F7E80,
		adjustStamina: s.PlayerAdjustStamina4F7DB0,
		buffOff:       runtime.BuffOff,
		cancelDurSpell: func(spellID int32, caster *Object) int32 {
			return s.Spells.Dur.SpellCancelDurSpell4FEB10(spellID, caster)
		},
	})
}

// PlayerInputAttack4F9C70 binds GAME.EXE 004F9C70 to native-width Object,
// PlayerUpdateData, Player, weapon, and use-data pointers.
//
//go:noinline
func (s *Server) PlayerInputAttack4F9C70(
	unit *Object,
	runtime PlayerInputAttackRuntime4F9C70,
) {
	playerInputAttackNative4F9C70(s, unit, runtime)
}
