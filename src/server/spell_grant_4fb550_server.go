package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// SpellGrantRuntime4FB550 supplies the protection, line-message, and shop
// services that remain owned by the root/legacy runtime. Every pointer stays
// native-width; the protection token and original scalar arguments retain the
// executable's fixed 32-bit representation.
type SpellGrantRuntime4FB550 struct {
	AwardProtection func(uint32, int32, int32)
	SendLineMessage func(*Object, string) bool
	ShopExit        func(*TradeSession)
}

type spellGrantNativeDeps4FB550 struct {
	gameFlagsCheck   func(uint32) int32
	loadString       func(string, string, int) string
	sendLineMessage  func(*Object, string)
	awardProtection  func(uint32, int32, int32)
	spellHasFlags    func(int32, int32) int32
	spellIsValid     func(int32) int32
	audio            func(uint32, *Object, int32, uint32)
	rewardNotify     func(*Object, int32, *Object, int32)
	checkPlayerState func(*Object) int32
	firstPlayer      func() *Player
	nextPlayer       func(*Player) *Player
	shopExit         func(*TradeSession)
	reportSpellAward func(*Object, int32, int32, int32)
}

func spellGrantToPlayerNative4FB550(
	unit *Object,
	spellID, notify, shop, override int32,
	deps spellGrantNativeDeps4FB550,
) int32 {
	return spellGrantToPlayer4FB550(
		spellID,
		notify,
		shop,
		override,
		spellGrantHooks4FB550[
			*Object,
			*PlayerUpdateData,
			*Player,
			*TradeSession,
			string,
		]{
			loadUnitArg: func() *Object {
				return unit
			},
			loadClassLow: func(unit *Object) uint8 {
				// Use the raw field: GAME.EXE has no nil-safe Class helper here.
				return uint8(unit.ObjClass)
			},
			loadUpdateData: func(unit *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(unit.UpdateData)
			},
			loadPlayer: func(update *PlayerUpdateData) *Player {
				return update.Player
			},
			loadSpellLevel: func(player *Player, spellID int32) uint32 {
				return player.SpellLvl[spellID]
			},
			storeSpellLevel: func(player *Player, spellID int32, level uint32) {
				player.SpellLvl[spellID] = level
			},
			loadProtection: func(player *Player) uint32 {
				return player.Prot4636
			},
			gameFlagsCheck:  deps.gameFlagsCheck,
			loadString:      deps.loadString,
			sendLineMessage: deps.sendLineMessage,
			awardProtection: deps.awardProtection,
			spellHasFlags:   deps.spellHasFlags,
			spellIsValid:    deps.spellIsValid,
			audio:           deps.audio,
			loadNotifyField: func(player *Player) uint32 {
				return player.Field4792
			},
			rewardNotify:     deps.rewardNotify,
			checkPlayerState: deps.checkPlayerState,
			firstPlayer:      deps.firstPlayer,
			nextPlayer:       deps.nextPlayer,
			loadPlayerUnit: func(player *Player) *Object {
				return player.PlayerUnit
			},
			loadTrade: func(update *PlayerUpdateData) *TradeSession {
				return update.Trade70
			},
			shopExit:         deps.shopExit,
			reportSpellAward: deps.reportSpellAward,
		},
	)
}

func spellGrantRewardNotifyNative4FAD50(
	s *Server,
	recipient *Object,
	kind int32,
	source *Object,
	spellID int32,
) {
	if recipient == nil || uint8(recipient.ObjClass)&spellGrantPlayerClass4FB550 == 0 {
		return
	}
	update := (*PlayerUpdateData)(recipient.UpdateData)
	packet := [5]byte{byte(netmsg.MSG_INFORM)}
	switch kind {
	case 0:
		packet[1] = 30
	case 1:
		packet[1] = 31
	case 2:
		packet[1] = 32
	default:
		return
	}
	packet[2] = byte(spellID)
	binary.LittleEndian.PutUint16(packet[3:], uint16(source.NetCode))
	player := update.Player
	s.NetSendPacketXxx0(int(player.PlayerInd), packet[:], nil, 1)
}

func spellGrantReportNative4D7F90(
	s *Server,
	unit *Object,
	spellID, notify, shop int32,
) {
	// GAME.EXE 004D7F90 dereferences unit before its class gate.
	if uint8(unit.ObjClass)&spellGrantPlayerClass4FB550 == 0 {
		return
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	packet := [4]byte{
		byte(netmsg.MSG_REPORT_SPELL_AWARD),
		byte(spellID),
		byte(player.SpellLvl[spellID]),
		byte(notify),
	}
	if shop != 0 {
		packet[3] |= 0x80
	}
	player = update.Player
	s.NetSendPacketXxx1(int(player.PlayerInd), packet[:], nil, 1)
}

func spellGrantServerDeps4FB550(
	s *Server,
	runtime SpellGrantRuntime4FB550,
) spellGrantNativeDeps4FB550 {
	return spellGrantNativeDeps4FB550{
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		loadString: func(key, path string, line int) string {
			_ = line // retained by the generic provenance contract
			return s.Strings().GetStringInFile(strman.ID(key), path)
		},
		sendLineMessage: func(unit *Object, message string) {
			_ = runtime.SendLineMessage(unit, message)
		},
		awardProtection: runtime.AwardProtection,
		spellHasFlags: func(spellID, flags int32) int32 {
			if s.Spells.HasFlags(spell.ID(spellID), things.SpellFlags(flags)) {
				return 1
			}
			return 0
		},
		spellIsValid: func(spellID int32) int32 {
			definition := s.Spells.DefByInd(spell.ID(spellID))
			if definition != nil && definition.IsValid() {
				return 1
			}
			return 0
		},
		audio: func(id uint32, unit *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), code)
		},
		rewardNotify: func(recipient *Object, kind int32, source *Object, spellID int32) {
			spellGrantRewardNotifyNative4FAD50(s, recipient, kind, source, spellID)
		},
		checkPlayerState: func(unit *Object) int32 {
			if s.Players.CheckXxx(unit) {
				return 1
			}
			return 0
		},
		firstPlayer: s.Players.First,
		nextPlayer:  s.Players.Next,
		shopExit:    runtime.ShopExit,
		reportSpellAward: func(unit *Object, spellID, notify, shop int32) {
			spellGrantReportNative4D7F90(s, unit, spellID, notify, shop)
		},
	}
}

// SpellGrantToPlayer4FB550 binds GAME.EXE 004FB550 to native-width Object,
// PlayerUpdateData, Player, active-player-list, and TradeSession pointers.
func (s *Server) SpellGrantToPlayer4FB550(
	unit *Object,
	spellID, notify, shop, override int32,
	runtime SpellGrantRuntime4FB550,
) int32 {
	return spellGrantToPlayerNative4FB550(
		unit,
		spellID,
		notify,
		shop,
		override,
		spellGrantServerDeps4FB550(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.SpellLvl[0])]
	_ = [1]struct{}{}[137-len(Player{}.SpellLvl)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.Prot4636)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.Field4792)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(Player{}.PlayerInd)]
)
