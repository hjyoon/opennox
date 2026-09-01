package server

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/sound"
)

// BeastGuideAwardRuntime4FAE80 supplies the protection, relation-table, and
// line-message services still owned by the root/legacy runtime. All pointers
// remain native-width; guide IDs, levels, and protection tokens retain their
// original fixed 32-bit representation.
type BeastGuideAwardRuntime4FAE80 struct {
	AwardProtection func(uint32, int32, int32)
	RelatedGuides   func(int32) []int32
	SendLineMessage func(*Object, string) bool
}

type beastGuideAwardNativeDeps4FAE80 struct {
	loadString       func(string, string, int) string
	sendLineMessage  func(*Object, string)
	awardProtection  func(uint32, int32, int32)
	audio            func(uint32, *Object, int32, uint32)
	rewardNotify     func(*Object, int32, *Object, int32)
	relatedGuides    func(int32) []int32
	firstPlayer      func() *Player
	nextPlayer       func(*Player) *Player
	reportGuideAward func(*Object, int32, int32, int32)
}

func beastGuideAwardNative4FAE80(
	unit *Object,
	guide, notify int32,
	deps beastGuideAwardNativeDeps4FAE80,
) int32 {
	return beastGuideAward4FAE80(
		guide,
		notify,
		beastGuideAwardHooks4FAE80[
			*Object,
			*PlayerUpdateData,
			*Player,
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
			loadGuideLevel: func(player *Player, guide int32) uint32 {
				return player.BeastScrollLvl[guide]
			},
			storeGuideLevel: func(player *Player, guide int32, level uint32) {
				player.BeastScrollLvl[guide] = level
			},
			loadProtection: func(player *Player) uint32 {
				return player.Prot4640
			},
			loadString:      deps.loadString,
			sendLineMessage: deps.sendLineMessage,
			awardProtection: deps.awardProtection,
			audio:           deps.audio,
			rewardNotify:    deps.rewardNotify,
			relatedGuides:   deps.relatedGuides,
			firstPlayer:     deps.firstPlayer,
			nextPlayer:      deps.nextPlayer,
			loadPlayerUnit: func(player *Player) *Object {
				return player.PlayerUnit
			},
			reportGuideAward: deps.reportGuideAward,
		},
	)
}

func beastGuideAwardReportNative4D8000(
	s *Server,
	unit *Object,
	guide, notify, shop int32,
) int32 {
	if uint8(unit.ObjClass)&beastGuideAwardPlayerClass4FAE80 == 0 {
		return int32(uintptr(unsafe.Pointer(unit)))
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	packet := [3]byte{
		byte(netmsg.MSG_REPORT_GUIDE_AWARD),
		byte(guide),
		byte(notify),
	}
	if shop != 0 {
		packet[2] |= 0x80
	}
	player := update.Player
	return int32(s.NetSendPacketXxx1(int(player.PlayerInd), packet[:], nil, 1))
}

func beastGuideAwardServerDeps4FAE80(
	s *Server,
	runtime BeastGuideAwardRuntime4FAE80,
) beastGuideAwardNativeDeps4FAE80 {
	return beastGuideAwardNativeDeps4FAE80{
		loadString: func(key, path string, line int) string {
			_ = line // retained by the generic provenance contract
			return s.Strings().GetStringInFile(strman.ID(key), path)
		},
		sendLineMessage: func(unit *Object, message string) {
			_ = runtime.SendLineMessage(unit, message)
		},
		awardProtection: runtime.AwardProtection,
		audio: func(id uint32, unit *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), code)
		},
		rewardNotify: func(recipient *Object, kind int32, source *Object, guide int32) {
			spellGrantRewardNotifyNative4FAD50(s, recipient, kind, source, guide)
		},
		relatedGuides: runtime.RelatedGuides,
		firstPlayer:   s.Players.First,
		nextPlayer:    s.Players.Next,
		reportGuideAward: func(unit *Object, guide, notify, shop int32) {
			beastGuideAwardReportNative4D8000(s, unit, guide, notify, shop)
		},
	}
}

// AwardBeastGuide4FAE80 binds GAME.EXE 004FAE80 to native-width Object,
// PlayerUpdateData, Player, and active-player-record pointers.
func (s *Server) AwardBeastGuide4FAE80(
	unit *Object,
	guide, notify int32,
	runtime BeastGuideAwardRuntime4FAE80,
) int32 {
	return beastGuideAwardNative4FAE80(
		unit,
		guide,
		notify,
		beastGuideAwardServerDeps4FAE80(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.BeastScrollLvl[0])]
	_ = [1]struct{}{}[41-len(Player{}.BeastScrollLvl)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.Prot4640)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(Player{}.PlayerInd)]
)
