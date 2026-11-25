package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guirank.h"

typedef const wchar2_t rank_const_text_t;
extern uint32_t dword_5d4594_1090120;
*/
import "C"

import (
	"math"
	"runtime"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// These fixed-size rows are also read by the legacy rank widget.
type clientRankPlayerRow struct {
	Name      [26]uint16
	TeamID    int32
	Class     byte
	_         [3]byte
	NetCode   uint32
	Score     uint32
	Ping      uint32
	Objective uint32
	Status    uint32
}

type clientRankTeamRow struct {
	Name  [22]uint16
	ID    uint32
	Score int32
	Color byte
	_     [3]byte
}

func clientRankPlayerAt(index byte) *clientRankPlayerRow {
	return memmap.PtrT[clientRankPlayerRow](0x5D4594, 1084132+80*uintptr(index))
}

func clientRankTeamAt(index byte) *clientRankTeamRow {
	return memmap.PtrT[clientRankTeamRow](0x5D4594, 1087204+56*uintptr(index))
}

//export sub_46DB80
func sub_46DB80() C.int {
	var result C.int
	file := internCStr("client_rank_list_native.go")
	for page := 0; page < 2; page++ {
		for offset := 1090060; offset <= 1090092; offset += 8 {
			if win := C.nox_rank_window_at(C.int(offset + 4*page)); win != nil {
				result = C.int(nox_window_call_field_94_fnc(win, 16399, 1, 0, file, 0))
			}
		}
	}
	return result
}

//export sub_46DC00
func sub_46DC00(win *C.nox_window, color C.uchar, text *C.rank_const_text_t) C.int {
	if win == nil || text == nil {
		return 0
	}
	nox_window_call_field_94_fnc(win, 16397, uintptr(unsafe.Pointer(text)), uintptr(color), internCStr("client_rank_list_native.go"), 0)
	runtime.KeepAlive(text)
	return 1
}

//export sub_46DCC0
func sub_46DCC0() *C.char {
	*memmap.PtrUint8(0x5D4594, 1090117) = 0
	*memmap.PtrUint8(0x5D4594, 1090118) = 0
	count := clientRankPlayerCount()
	switch {
	case C.dword_5d4594_1090120 == 5:
		clientRankStorePlayers(clientRankLadder, count)
	case noxflags.HasGame(noxflags.GameModeElimination):
		clientRankStoreTeams(true)
		clientRankStorePlayers(clientRankElimination, count)
	default:
		return sub_46E4E0()
	}
	return nil
}

//export sub_46E080
func sub_46E080(player *C.nox_playerInfo) C.int {
	if player == nil {
		return 0
	}
	netCode := asPlayerS(player).NetCodeVal
	switch {
	case noxflags.HasGame(noxflags.GameModeCTF):
		if netCode == uint32(memmap.Uint16(0x5D4594, 1090128)) {
			return 2
		}
		if netCode == uint32(memmap.Uint16(0x5D4594, 1090130)) {
			return 3
		}
	case noxflags.HasGame(noxflags.GameModeFlagBall):
		if netCode == uint32(memmap.Uint16(0x5D4594, 1090132)) {
			return 4
		}
	case noxflags.HasGame(noxflags.GameModeKOTR):
		drawable := nox_xxx_netSpriteByCodeDynamic_45A6F0(int32(netCode))
		if drawable != nil && bool(nox_client_drawable_testBuff_4356C0(drawable, 30)) {
			return 1
		}
	}
	return 0
}

//export sub_46E130
func sub_46E130(id C.int) C.int {
	for i := range memmap.Uint8(0x5D4594, 1090116) {
		if clientRankTeamAt(i).ID == uint32(id) {
			return 1
		}
	}
	return 0
}

//export sub_46E170
func sub_46E170(name *C.wchar2_t) *C.ushort {
	limit := memmap.Int32(0x5D4594, 1084036)
	if name == nil || limit == 0 {
		return nil
	}
	var width C.int
	nox_xxx_drawGetStringSize_43F840(nil, name, &width, nil, 0)
	if width == 0 {
		return nil
	}
	chars := unsafe.Slice(name, alloc.StrLen(name))
	for length := len(chars); length > 0 && int32(width)+5 > limit; {
		length--
		chars[length] = 0
		nox_xxx_drawGetStringSize_43F840(nil, name, &width, nil, 0)
	}
	return (*C.ushort)(name)
}

//export sub_46E1E0
func sub_46E1E0(code C.int) C.int {
	for i := range memmap.Uint8(0x5D4594, 1090117) {
		if clientRankPlayerAt(i).NetCode == uint32(code) {
			return 1
		}
	}
	return 0
}

//export sub_46E4E0
func sub_46E4E0() *C.char {
	clientRankStoreTeams(false)
	clientRankStorePlayers(clientRankLessons, clientRankPlayerCount())
	return nil
}

func clientRankPlayerCount() int {
	count := GetServer().S().Players.Count()
	if count != 0 && noxflags.HasGame(noxflags.GameHost) && noxflags.HasEngine(noxflags.EngineNoRendering) {
		count--
	}
	return count
}

func clientRankTeamID(code uint32) int32 {
	membership := nox_xxx_objGetTeamByNetCode_418C80(C.int(code))
	if membership != nil && nox_xxx_servObjectHasTeam_419130(membership) != 0 {
		if team := GetServer().S().Teams.ByID(server.TeamID(membership.id)); team != nil {
			return int32(team.ID())
		}
	}
	return -1
}

func clientRankStorePlayer(row *clientRankPlayerRow, player *server.Player, score uint32) {
	native := (*C.nox_playerInfo)(player.C())
	row.TeamID = clientRankTeamID(player.NetCodeVal)
	row.Class = byte(player.PlayerClass())
	row.NetCode, row.Score = player.NetCodeVal, score
	row.Ping = uint32(uint16(native.field_2148))
	row.Objective = uint32(sub_46E080(native))
	row.Status = player.Field3680
	alloc.StrCopyZero16S(row.Name[:], player.NameFinal[:])
	sub_46E170((*C.wchar2_t)(&row.Name[0]))
}

func clientRankStoreTeams(ascending bool) {
	teams := &GetServer().S().Teams
	limit := teams.Count()
	count := memmap.PtrUint8(0x5D4594, 1090116)
	*count = 0
	previous, boundary := clientRankBounds(ascending, math.MinInt32, math.MaxInt32)
	for int(*count) < limit {
		var next *server.Team
		nextScore := boundary
		for team := teams.First(); team != nil; team = teams.Next(team) {
			score := int64(team.Lessons)
			if sub_46E130(C.int(team.ID())) == 0 && clientRankScoreInRange(ascending, score, previous, nextScore) {
				next, nextScore = team, score
			}
		}
		if next == nil {
			break
		}
		row := clientRankTeamAt(*count)
		row.ID, row.Score, row.Color = uint32(next.ID()), next.Lessons, byte(next.ColorInd)
		alloc.StrCopyZero16P(row.Name[:], (*uint16)(next.C()))
		sub_46E170((*C.wchar2_t)(&row.Name[0]))
		(*count)++
		previous = nextScore
	}
}

type clientRankMode byte

const (
	clientRankLessons clientRankMode = iota
	clientRankElimination
	clientRankLadder
	rankObserverPenalty = 0xffff
	rankUnrankedLadder  = 0x08000000
	rankObserverLadder  = 0x80000000
)

func clientRankObserver(player *server.Player) bool {
	return player.Field3680&1 != 0 && player.Field3680&0x20 == 0
}

func (mode clientRankMode) adjust(player *server.Player, entering bool) {
	switch mode {
	case clientRankLadder:
		if clientRankObserver(player) {
			if entering {
				player.Field2108 |= rankObserverLadder
			} else {
				player.Field2108 &^= rankObserverLadder
			}
		} else if entering && player.Field2108 == 0 {
			player.Field2108 = rankUnrankedLadder
		} else if !entering && player.Field2108 == rankUnrankedLadder {
			player.Field2108 = 0
		}
	case clientRankElimination:
		if clientRankObserver(player) {
			if entering {
				player.Field2140 += rankObserverPenalty
			} else {
				player.Field2140 -= rankObserverPenalty
			}
		}
	case clientRankLessons:
		if clientRankObserver(player) {
			if entering {
				player.Lessons -= rankObserverPenalty
			} else {
				player.Lessons += rankObserverPenalty
			}
		}
	}
}

func (mode clientRankMode) sortScore(player *server.Player) int64 {
	switch mode {
	case clientRankLadder:
		return int64(player.Field2108)
	case clientRankElimination:
		return int64(int32(player.Field2140))
	default:
		return int64(player.Lessons)
	}
}

func (mode clientRankMode) displayScore(player *server.Player) (uint32, bool) {
	observer := clientRankObserver(player)
	switch mode {
	case clientRankLadder:
		score := player.Field2108
		if observer {
			return score + rankObserverLadder, false
		}
		if score == rankUnrankedLadder {
			return 0, false
		}
		return score, true
	case clientRankElimination:
		score := player.Field2140
		if observer {
			score -= rankObserverPenalty
		}
		return score, !observer
	default:
		score := uint32(player.Lessons)
		if observer {
			score += rankObserverPenalty
		}
		return score, !observer
	}
}

func clientRankBounds(ascending bool, minimum, maximum int64) (int64, int64) {
	if ascending {
		return minimum, maximum
	}
	return maximum, minimum
}

func clientRankScoreInRange(ascending bool, score, previous, next int64) bool {
	if ascending {
		return score >= previous && score < next
	}
	return score <= previous && score > next
}

func clientRankStorePlayers(mode clientRankMode, limit int) {
	players := &GetServer().S().Players
	for player := players.First(); player != nil; player = players.Next(player) {
		mode.adjust(player, true)
	}
	// The original scan compares ladder ranks unsigned and the two game
	// scores signed. The strict upper/lower bound also preserves tie order.
	ascending := mode != clientRankLessons
	previous, boundary := clientRankBounds(ascending, math.MinInt32, math.MaxInt32)
	if mode == clientRankLadder {
		previous, boundary = 0, math.MaxUint32
	} else if mode == clientRankElimination {
		previous = -1
	}
	rows := memmap.PtrUint8(0x5D4594, 1090117)
	playing := memmap.PtrUint8(0x5D4594, 1090118)
	for int(*rows) < limit {
		var next *server.Player
		nextScore := boundary
		for player := players.First(); player != nil; player = players.Next(player) {
			if player.PlayerInd == server.HostPlayerIndex && noxflags.HasEngine(noxflags.EngineNoRendering) {
				continue
			}
			score := mode.sortScore(player)
			if clientRankScoreInRange(ascending, score, previous, nextScore) && sub_46E1E0(C.int(player.NetCodeVal)) == 0 {
				next, nextScore = player, score
			}
		}
		if next == nil {
			break
		}
		score, counted := mode.displayScore(next)
		if counted {
			(*playing)++
		}
		clientRankStorePlayer(clientRankPlayerAt(*rows), next, score)
		(*rows)++
		previous = nextScore
	}
	for player := players.First(); player != nil; player = players.Next(player) {
		mode.adjust(player, false)
	}
}
