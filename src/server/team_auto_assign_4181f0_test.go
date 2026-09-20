package server

import (
	"reflect"
	"testing"
)

type teamAutoAssignTestPlayer4181F0 struct {
	unit       int
	netCode    uint32
	status     uint32
	configured int
}

type teamAutoAssignTestTeam4181F0 struct {
	id      TeamID
	members int
}

type teamAutoAssignTestAttachment4181F0 struct {
	teamID  TeamID
	unit    int
	netCode uint32
	flags   int32
}

func teamAutoAssignTestHooks4181F0(
	players []teamAutoAssignTestPlayer4181F0,
	unitsWithTeam map[int]bool,
	teams []teamAutoAssignTestTeam4181F0,
	preferConfigured bool,
	attachments *[]teamAutoAssignTestAttachment4181F0,
) teamAutoAssignHooks4181F0[int, int, int] {
	return teamAutoAssignHooks4181F0[int, int, int]{
		firstPlayer: func() int {
			if len(players) == 0 {
				return 0
			}
			return 1
		},
		nextPlayer: func(player int) int {
			if player >= len(players) {
				return 0
			}
			return player + 1
		},
		loadUnit: func(player int) int {
			return players[player-1].unit
		},
		loadNetCode: func(player int) uint32 {
			return players[player-1].netCode
		},
		loadStatus: func(player int) uint32 {
			return players[player-1].status
		},
		hasTeam:       func(unit int) bool { return unitsWithTeam[unit] },
		clientNetCode: func() uint32 { return 22 },
		noRendering:   func() bool { return true },
		randomInt:     func(_, _ int) int { return 0 },
		preferConfiguredTeam: func() bool {
			return preferConfigured
		},
		configuredTeam: func(player int) int {
			return players[player-1].configured
		},
		firstTeam: func() int {
			if len(teams) == 0 {
				return 0
			}
			return 1
		},
		nextTeam: func(team int) int {
			if team >= len(teams) {
				return 0
			}
			return team + 1
		},
		teamMemberCount: func(team int) int {
			return teams[team-1].members
		},
		teamID: func(team int) TeamID {
			return teams[team-1].id
		},
		attach: func(teamID TeamID, unit int, netCode uint32, flags int32) {
			*attachments = append(*attachments, teamAutoAssignTestAttachment4181F0{
				teamID: teamID, unit: unit, netCode: netCode, flags: flags,
			})
			for i := range teams {
				if teams[i].id == teamID {
					teams[i].members++
					break
				}
			}
		},
	}
}

func TestTeamAutoAssign4181F0FiltersAndBalancesNativePlayers(t *testing.T) {
	players := []teamAutoAssignTestPlayer4181F0{
		{unit: 101, netCode: 11},
		{unit: 102, netCode: 22},
		{unit: 103, netCode: 33, status: 1},
		{unit: 104, netCode: 44},
		{unit: 105, netCode: 55, status: 0x21},
		{unit: 0, netCode: 66},
	}
	teams := []teamAutoAssignTestTeam4181F0{{id: 7}, {id: 8, members: 2}}
	var attachments []teamAutoAssignTestAttachment4181F0
	hooks := teamAutoAssignTestHooks4181F0(players, map[int]bool{104: true}, teams, false, &attachments)
	randomCalls := 0
	hooks.randomInt = func(minimum, maximum int) int {
		randomCalls++
		if minimum != 0 || maximum != 1 {
			t.Fatalf("random range = %d..%d", minimum, maximum)
		}
		return 0
	}

	teamAutoAssign4181F0(hooks)
	want := []teamAutoAssignTestAttachment4181F0{
		{teamID: 7, unit: 101, netCode: 11, flags: 1},
		{teamID: 7, unit: 105, netCode: 55, flags: 1},
	}
	if !reflect.DeepEqual(attachments, want) {
		t.Fatalf("attachments = %#v, want %#v", attachments, want)
	}
	if randomCalls != 100 {
		t.Fatalf("random calls = %d, want 100", randomCalls)
	}
}

func TestTeamAutoAssign4181F0UsesConfiguredTeam(t *testing.T) {
	players := []teamAutoAssignTestPlayer4181F0{
		{unit: 101, netCode: 11, configured: 2},
		{unit: 102, netCode: 12},
	}
	teams := []teamAutoAssignTestTeam4181F0{{id: 7}, {id: 8}}
	var attachments []teamAutoAssignTestAttachment4181F0
	hooks := teamAutoAssignTestHooks4181F0(players, nil, teams, true, &attachments)

	teamAutoAssign4181F0(hooks)
	want := []teamAutoAssignTestAttachment4181F0{
		{teamID: 8, unit: 101, netCode: 11, flags: 0},
	}
	if !reflect.DeepEqual(attachments, want) {
		t.Fatalf("attachments = %#v, want %#v", attachments, want)
	}
}

func TestTeamAutoAssign4181F0ResetsTeamsBeforeFilteringPlayers(t *testing.T) {
	players := []teamAutoAssignTestPlayer4181F0{{unit: 101, netCode: 11}}
	teams := []teamAutoAssignTestTeam4181F0{{id: 7}, {id: 8}}
	unitsWithTeam := map[int]bool{101: true}
	var attachments []teamAutoAssignTestAttachment4181F0
	hooks := teamAutoAssignTestHooks4181F0(players, unitsWithTeam, teams, false, &attachments)
	hooks.resetTeams = true
	var resetOrder []int
	hooks.resetTeam = func(team int) {
		resetOrder = append(resetOrder, team)
		unitsWithTeam[101] = false
	}

	teamAutoAssign4181F0(hooks)
	if !reflect.DeepEqual(resetOrder, []int{1, 2}) {
		t.Fatalf("reset order = %v, want [1 2]", resetOrder)
	}
	want := []teamAutoAssignTestAttachment4181F0{
		{teamID: 7, unit: 101, netCode: 11, flags: 1},
	}
	if !reflect.DeepEqual(attachments, want) {
		t.Fatalf("attachments = %#v, want %#v", attachments, want)
	}
}

func TestServerTeamsMemberCountPrefersNativeSidecar(t *testing.T) {
	team := &Team{field_48: 9}
	teams := &serverTeams{}
	if got := teams.MemberCount(team); got != 9 {
		t.Fatalf("legacy member count = %d, want 9", got)
	}
	teams.members = map[*Team]map[*ObjectTeam]struct{}{
		team: {new(ObjectTeam): {}, new(ObjectTeam): {}},
	}
	if got := teams.MemberCount(team); got != 2 {
		t.Fatalf("native member count = %d, want 2", got)
	}
}
