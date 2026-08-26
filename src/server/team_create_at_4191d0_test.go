package server

import (
	"reflect"
	"testing"
)

func TestTeamCreateAt4191D0NilValueStillLoadsGameBall(t *testing.T) {
	var calls []string
	teamCreateAt4191D0(1, 0, 1, 42, 0, teamCreateAtHooks4191D0[int, int, int, int]{
		loadGameBallType: func() uint16 {
			calls = append(calls, "game-ball")
			return 77
		},
	})
	if want := []string{"game-ball"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestTeamCreateAt4191D0ExistingMembershipStopsBeforeLink(t *testing.T) {
	var calls []string
	teamCreateAt4191D0(3, 11, 1, 42, 0, teamCreateAtHooks4191D0[int, int, int, int]{
		loadGameBallType: func() uint16 {
			calls = append(calls, "game-ball")
			return 77
		},
		findTeam: func(id uint8) int {
			calls = append(calls, "find")
			return 21
		},
		containsTeam: func(value int, id uint8) bool {
			calls = append(calls, "contains")
			return true
		},
	})
	if want := []string{"game-ball", "find", "contains"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestTeamCreateAt4191D0PreservesAttachAndPlayerOrder(t *testing.T) {
	var calls []string
	teamCreateAt4191D0(3, 11, 1, 42, 5, teamCreateAtHooks4191D0[int, int, int, int]{
		loadGameBallType: func() uint16 {
			calls = append(calls, "game-ball")
			return 77
		},
		findTeam: func(id uint8) int {
			calls = append(calls, "find")
			return 0
		},
		createTeam: func(id uint8) int {
			calls = append(calls, "create")
			return 21
		},
		linkTeam: func(value, team int) bool {
			calls = append(calls, "link")
			return true
		},
		loadTeamID: func(team int) uint8 {
			calls = append(calls, "team-id")
			return 3
		},
		clientNetCode: func() uint32 {
			calls = append(calls, "client-code")
			return 42
		},
		selectLocalTeam: func(id uint8) {
			calls = append(calls, "select")
		},
		afterAttach: func(team, value int, active int32, netCode uint32, flags int32, gameBallType uint16) {
			calls = append(calls, "notify")
			if team != 21 || value != 11 || active != 1 || netCode != 42 || flags != 5 || gameBallType != 77 {
				t.Fatalf("notification args = (%d, %d, %d, %d, %d, %d)", team, value, active, netCode, flags, gameBallType)
			}
		},
		commitMemberCount: func(team int) {
			calls = append(calls, "commit")
		},
		firstPlayerUnit: func() int {
			calls = append(calls, "first-player")
			return 101
		},
		nextPlayerUnit: func(unit int) int {
			calls = append(calls, "next-player")
			if unit == 101 {
				return 102
			}
			return 0
		},
		loadNetCode: func(unit int) uint32 {
			calls = append(calls, "net-code")
			if unit == 102 {
				return 42
			}
			return 7
		},
		loadPlayer: func(unit int) int {
			calls = append(calls, "player")
			return 31
		},
		loadPlayerIndex: func(player int) uint8 {
			calls = append(calls, "player-index")
			return 9
		},
		resetPlayer: func(playerIndex uint8) {
			calls = append(calls, "reset")
		},
		rebuildMasks: func(playerIndex uint8) {
			calls = append(calls, "rebuild")
		},
		markUpdate: func(unit int) {
			switch unit {
			case 102:
				calls = append(calls, "mark-player")
			case 202:
				calls = append(calls, "mark-owned-2")
			case 203:
				calls = append(calls, "mark-owned-4")
			default:
				t.Fatalf("unexpected marked object %d", unit)
			}
		},
		firstOwned: func(unit int) int {
			calls = append(calls, "first-owned")
			return 201
		},
		nextOwned: func(unit int) int {
			calls = append(calls, "next-owned")
			if unit < 203 {
				return unit + 1
			}
			return 0
		},
		loadClassLow: func(unit int) uint8 {
			calls = append(calls, "class")
			return map[int]uint8{201: 1, 202: 2, 203: 4}[unit]
		},
	})
	want := []string{
		"game-ball", "find", "create", "link", "team-id", "client-code",
		"select", "notify", "commit", "first-player", "net-code",
		"next-player", "net-code", "player", "player-index", "reset",
		"rebuild", "mark-player", "first-owned", "class", "next-owned",
		"class", "mark-owned-2", "next-owned", "class", "mark-owned-4",
		"next-owned",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls =\n%v\nwant =\n%v", calls, want)
	}
}
