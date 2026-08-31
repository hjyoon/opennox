package opennox

import (
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

type playerExecuteAbilityNativeDeps4FBB70 struct {
	gameFlag      func(uint32) int32
	isActive      func(*server.Object, server.Ability) int32
	isActiveVal   func(*server.Object, server.Ability) int32
	reportText    func(uint8, uint8, *int32)
	loadCooldown  func(uint8, server.Ability) int32
	storeCooldown func(uint8, server.Ability, int32)
	loadDelay     func(server.Ability) int32
	reportState   func(*server.Object, server.Ability, uint8)
	loadDuration  func(server.Ability) int32
	allocExec     func() *server.ExecAbilityClass
	loadFrame     func() uint32
	loadExecHead  func() *server.ExecAbilityClass
	storeExecHead func(*server.ExecAbilityClass)
	invoke        func(*server.Object, server.Ability)
	loadSound     func(server.Ability, int32) int32
	audio         func(int32, *server.Object, int32, uint32)
}

func playerExecuteAbilityNative4FBB70(
	unit *server.Object,
	ability server.Ability,
	deps playerExecuteAbilityNativeDeps4FBB70,
) {
	playerExecuteAbility4FBB70(unit, ability, playerExecuteAbilityHooks4FBB70[
		*server.Object,
		*server.PlayerUpdateData,
		*server.Player,
		*server.Object,
		*server.ExecAbilityClass,
	]{
		loadFlags: func(unit *server.Object) object.Flags {
			return unit.ObjFlags
		},
		loadClassLow: func(unit *server.Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdateData: func(unit *server.Object) *server.PlayerUpdateData {
			return unit.UpdateDataPlayer()
		},
		loadPlayer: func(update *server.PlayerUpdateData) *server.Player {
			return update.Player
		},
		loadPlayerClassLow: func(player *server.Player) uint8 {
			if player == nil {
				panic("nox_xxx_playerExecuteAbil_4FBB70: nil player")
			}
			return uint8(player.PlayerClass())
		},
		gameFlag: deps.gameFlag,
		firstItem: func(unit *server.Object) *server.Object {
			return unit.InvFirstItem
		},
		nextItem: func(item *server.Object) *server.Object {
			return item.InvNextItem
		},
		loadItemClass: func(item *server.Object) object.Class {
			return item.ObjClass
		},
		isActive:    deps.isActive,
		isActiveVal: deps.isActiveVal,
		loadState: func(update *server.PlayerUpdateData) uint8 {
			return uint8(update.State)
		},
		loadSpellLevel: func(player *server.Player, ability server.Ability) uint32 {
			return player.SpellLvl[ability]
		},
		loadBerserkBlock: func(player *server.Player) uint32 {
			return player.Field3656
		},
		loadPlayerIndex: func(player *server.Player) uint8 {
			return player.PlayerInd
		},
		reportText:    deps.reportText,
		loadCooldown:  deps.loadCooldown,
		storeCooldown: deps.storeCooldown,
		loadDelay:     deps.loadDelay,
		reportState:   deps.reportState,
		loadDuration:  deps.loadDuration,
		allocExec:     deps.allocExec,
		loadFrame:     deps.loadFrame,
		storeExecUnit: func(exec *server.ExecAbilityClass, unit *server.Object) {
			exec.Unit = unit
		},
		storeExecAbility: func(exec *server.ExecAbilityClass, ability server.Ability) {
			exec.Abil = ability
		},
		storeExecFrame: func(exec *server.ExecAbilityClass, frame uint32) {
			exec.Frame = frame
		},
		loadExecHead: deps.loadExecHead,
		storeExecNext: func(exec, next *server.ExecAbilityClass) {
			exec.Next = next
		},
		storeExecActive: func(exec *server.ExecAbilityClass, active uint32) {
			exec.Active = active
		},
		storeExecPrev: func(exec, prev *server.ExecAbilityClass) {
			exec.Prev = prev
		},
		storeExecHead: deps.storeExecHead,
		invoke:        deps.invoke,
		loadSound:     deps.loadSound,
		audio:         deps.audio,
	})
}

func (a *serverAbilities) abilityRuntimeUnit4FBB70(fallback *server.Object, index uint8) *server.Object {
	if player := a.s.Players.ByIndRaw(ntype.PlayerInd(index)); player != nil && player.PlayerUnit != nil {
		return player.PlayerUnit
	}
	return fallback
}

// playerExecuteAbility4FBB70 is the active native-width replacement for
// GAME.EXE 004FBB70. Cooldowns remain represented by the existing Go
// per-unit runtime map; resolving each observed PE32 player index separately
// retains the original read/write reload boundary without storing pointers in
// 32-bit integer cells.
//
//go:noinline
func (a *serverAbilities) playerExecuteAbility4FBB70(unit *server.Object, ability server.Ability) {
	playerExecuteAbilityNative4FBB70(unit, ability, playerExecuteAbilityNativeDeps4FBB70{
		gameFlag: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		isActive: func(unit *server.Object, ability server.Ability) int32 {
			if a.s.Abils.IsActive(unit, ability) {
				return 1
			}
			return 0
		},
		isActiveVal: func(unit *server.Object, ability server.Ability) int32 {
			if a.s.Abils.IsActiveVal(unit, ability) {
				return 1
			}
			return 0
		},
		reportText: func(index, kind uint8, code *int32) {
			a.s.NetInformTextMsg(ntype.PlayerInd(index), kind, int(*code))
		},
		loadCooldown: func(index uint8, ability server.Ability) int32 {
			runtime := a.s.Abils.GetFor(a.abilityRuntimeUnit4FBB70(unit, index))
			return int32(runtime.Cooldowns[ability])
		},
		storeCooldown: func(index uint8, ability server.Ability, value int32) {
			runtime := a.s.Abils.GetFor(a.abilityRuntimeUnit4FBB70(unit, index))
			runtime.Cooldowns[ability] = int(value)
		},
		loadDelay: func(ability server.Ability) int32 {
			return int32(a.getDelay(ability))
		},
		reportState: a.netAbilReportState,
		loadDuration: func(ability server.Ability) int32 {
			return int32(a.getDuration(ability))
		},
		allocExec: func() *server.ExecAbilityClass {
			return new(server.ExecAbilityClass)
		},
		loadFrame: a.s.Frame,
		loadExecHead: func() *server.ExecAbilityClass {
			return a.s.Abils.GetFor(unit).ExecList
		},
		storeExecHead: func(exec *server.ExecAbilityClass) {
			a.s.Abils.GetFor(unit).ExecList = exec
		},
		invoke: a.nox_xxx_playerInvokeAbility_4FBAF0,
		loadSound: func(ability server.Ability, slot int32) int32 {
			return int32(a.getSound(ability, int(slot)))
		},
		audio: func(soundID int32, unit *server.Object, kind int32, code uint32) {
			a.s.Audio.EventObj(sound.ID(soundID), unit, int(kind), code)
		},
	})
}
