package server

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// AbilityRewardRuntime4FB9C0 supplies the protection and line-message
// operations that remain owned by the root/legacy runtime. Object and Player
// pointers stay native-width; only the original protection token and scalar
// arguments retain their fixed 32-bit representation.
type AbilityRewardRuntime4FB9C0 struct {
	AwardProtection func(uint32, int32, int32)
	SendLineMessage func(*Object, string) bool
}

type abilityRewardNativeDeps4FB9C0 struct {
	loadString       func(string, string, int) string
	sendLineMessage  func(*Object, string)
	primaryMessage   func(*Object, string, uint8)
	awardProtection  func(uint32, int32, int32)
	reportAbility    func(*Object, int32, int32)
	gameFlagsCheck   func(uint32) int32
	rewardNotify     func(*Object, int32, *Object, int32)
	checkPlayerState func(*Object) int32
	firstPlayerUnit  func() *Object
	nextPlayerUnit   func(*Object) *Object
}

func abilityRewardServNative4FB9C0(
	unit *Object,
	ability, rewardArg int32,
	deps abilityRewardNativeDeps4FB9C0,
) int32 {
	return abilityRewardServ4FB9C0(
		ability,
		rewardArg,
		abilityRewardHooks4FB9C0[
			*Object,
			*PlayerUpdateData,
			*Player,
			string,
		]{
			loadUnitArg: func() *Object {
				return unit
			},
			loadClassLow: func(unit *Object) uint8 {
				return uint8(unit.ObjClass)
			},
			loadUpdateData: func(unit *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(unit.UpdateData)
			},
			loadPlayer: func(update *PlayerUpdateData) *Player {
				return update.Player
			},
			loadAbilityLevel: func(player *Player, ability int32) uint32 {
				return player.SpellLvl[ability]
			},
			storeAbilityLevel: func(player *Player, ability int32, level uint32) {
				player.SpellLvl[ability] = level
			},
			loadProtection: func(player *Player) uint32 {
				return player.Prot4636
			},
			loadString:       deps.loadString,
			sendLineMessage:  deps.sendLineMessage,
			primaryMessage:   deps.primaryMessage,
			awardProtection:  deps.awardProtection,
			reportAbility:    deps.reportAbility,
			gameFlagsCheck:   deps.gameFlagsCheck,
			rewardNotify:     deps.rewardNotify,
			checkPlayerState: deps.checkPlayerState,
			firstPlayerUnit:  deps.firstPlayerUnit,
			nextPlayerUnit:   deps.nextPlayerUnit,
		},
	)
}

func abilityRewardReportNative4D8060(
	s *Server,
	unit *Object,
	ability, rewardArg int32,
) int32 {
	// GAME.EXE 004D8060 dereferences unit before its class gate.
	if uint8(unit.ObjClass)&abilityRewardPlayerClass4FB9C0 == 0 {
		return int32(uintptr(unsafe.Pointer(unit)))
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	packet := [3]byte{
		byte(netmsg.MSG_REPORT_ABILITY_AWARD),
		byte(ability),
		byte(player.SpellLvl[ability]),
	}
	if rewardArg != 0 {
		packet[2] |= 0x80
	}
	player = update.Player
	return int32(s.NetSendPacketXxx1(int(player.PlayerInd), packet[:], nil, 1))
}

func abilityRewardServerDeps4FB9C0(
	s *Server,
	runtime AbilityRewardRuntime4FB9C0,
) abilityRewardNativeDeps4FB9C0 {
	return abilityRewardNativeDeps4FB9C0{
		loadString: func(key, path string, line int) string {
			_ = line // retained by the generic provenance contract
			return s.Strings().GetStringInFile(strman.ID(key), path)
		},
		sendLineMessage: func(unit *Object, message string) {
			_ = runtime.SendLineMessage(unit, message)
		},
		primaryMessage: func(unit *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(unit, strman.ID(message), value)
		},
		awardProtection: runtime.AwardProtection,
		reportAbility: func(unit *Object, ability, rewardArg int32) {
			abilityRewardReportNative4D8060(s, unit, ability, rewardArg)
		},
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		rewardNotify: func(recipient *Object, kind int32, source *Object, ability int32) {
			spellGrantRewardNotifyNative4FAD50(s, recipient, kind, source, ability)
		},
		checkPlayerState: func(unit *Object) int32 {
			if s.Players.CheckXxx(unit) {
				return 1
			}
			return 0
		},
		firstPlayerUnit: s.Players.FirstUnit,
		nextPlayerUnit:  s.Players.NextUnit,
	}
}

// AbilityRewardServ4FB9C0 binds GAME.EXE 004FB9C0 to native-width Object,
// PlayerUpdateData, Player, and active-player-unit pointers.
func (s *Server) AbilityRewardServ4FB9C0(
	unit *Object,
	ability, rewardArg int32,
	runtime AbilityRewardRuntime4FB9C0,
) int32 {
	return abilityRewardServNative4FB9C0(
		unit,
		ability,
		rewardArg,
		abilityRewardServerDeps4FB9C0(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.NetCode)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.SpellLvl[0])]
	_ = [1]struct{}{}[137-len(Player{}.SpellLvl)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.Prot4636)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(Player{}.PlayerInd)]
)
