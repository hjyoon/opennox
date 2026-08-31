package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// playerAbilityRuntimeTick4FBEE0 is the active native-width replacement for
// GAME.EXE 004FBEE0. Player and execution-record pointers remain native Go
// pointers; only cooldown cells retain the original signed 32-bit width.
//
//go:noinline
func (a *serverAbilities) playerAbilityRuntimeTick4FBEE0() {
	playerAbilityRuntimeTick4FBEE0(playerAbilityRuntimeTickHooks4FBEE0[
		*server.Player,
		*server.Object,
		*server.ExecAbilityClass,
	]{
		firstPlayer: a.s.Players.First,
		nextPlayer:  a.s.Players.Next,
		loadPlayerUnit: func(player *server.Player) *server.Object {
			return player.PlayerUnit
		},
		loadPlayerClass: func(player *server.Player) uint8 {
			return uint8(player.PlayerClass())
		},
		loadPlayerIndex: func(player *server.Player) uint8 {
			return player.PlayerInd
		},
		loadCooldown: a.s.Abils.PlayerAbilityCooldownAt,
		storeCooldown: func(index uint8, ability server.Ability, cooldown int32) {
			a.s.Abils.SetPlayerAbilityCooldownAt(index, ability, cooldown)
		},
		reportState:  a.netAbilReportState,
		loadExecHead: a.s.Abils.ExecHead,
		storeExecHead: func(head *server.ExecAbilityClass) {
			a.s.Abils.SetExecHead(head)
		},
		loadExecUnit: func(exec *server.ExecAbilityClass) *server.Object {
			return exec.Unit
		},
		loadExecAbility: func(exec *server.ExecAbilityClass) server.Ability {
			return exec.Abil
		},
		loadExecFrame: func(exec *server.ExecAbilityClass) uint32 {
			return exec.Frame
		},
		loadExecNext: func(exec *server.ExecAbilityClass) *server.ExecAbilityClass {
			return exec.Next
		},
		loadExecPrev: func(exec *server.ExecAbilityClass) *server.ExecAbilityClass {
			return exec.Prev
		},
		storeExecNext: func(exec, next *server.ExecAbilityClass) {
			exec.Next = next
		},
		storeExecPrev: func(exec, prev *server.ExecAbilityClass) {
			exec.Prev = prev
		},
		loadUnitFlags: func(unit *server.Object) object.Flags {
			return unit.ObjFlags
		},
		loadFrame: a.s.Frame,
		loadEndingSound: func(ability server.Ability, slot int32) int32 {
			return int32(a.getSound(ability, int(slot)))
		},
		audio: func(soundID int32, unit *server.Object, kind int32, code uint32) {
			a.s.Audio.EventObj(sound.ID(soundID), unit, int(kind), code)
		},
		reportActive: func(unit *server.Object, ability server.Ability, active uint8) {
			a.netAbilReportActive(unit, ability, active != 0)
		},
		setPlayerState: func(unit *server.Object, state uint8) {
			nox_xxx_playerSetState_4FA020(unit, server.PlayerState(state))
		},
		freeExec: func(exec *server.ExecAbilityClass) {
			*exec = server.ExecAbilityClass{}
		},
	})
}
