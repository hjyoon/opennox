package server

import (
	"fmt"
	"reflect"
	"testing"
)

type checkVictoryTeam509A60 struct {
	name  string
	score int32
}

type checkVictoryPlayer509A60 struct {
	name   string
	flags  uint32
	score  int32
	deaths uint32
}

type checkVictoryUpdate509A60 struct {
	player *checkVictoryPlayer509A60
}

type checkVictoryObject509A60 struct {
	name   string
	teamID uint8
	update *checkVictoryUpdate509A60
}

type checkVictoryWorld509A60 struct {
	flags      uint32
	gameFlags  uint16
	limit      uint16
	rivals     bool
	teams      []*checkVictoryTeam509A60
	teamByID   map[uint8]*checkVictoryTeam509A60
	players    []*checkVictoryObject509A60
	events     []string
	winner     string
	setFlags   []uint32
	rivalCalls int
}

func checkVictoryIndex509A60[T comparable](values []T, current T) int {
	for i, value := range values {
		if value == current {
			return i
		}
	}
	return -1
}

func (w *checkVictoryWorld509A60) hooks() checkVictoryHooks509A60[
	*checkVictoryTeam509A60, *checkVictoryObject509A60, *checkVictoryUpdate509A60, *checkVictoryPlayer509A60,
] {
	return checkVictoryHooks509A60[
		*checkVictoryTeam509A60, *checkVictoryObject509A60, *checkVictoryUpdate509A60, *checkVictoryPlayer509A60,
	]{
		hasGameFlag: func(flag uint32) bool {
			w.events = append(w.events, fmt.Sprintf("has-flag:%03x", flag))
			return w.flags&flag != 0
		},
		loadGameFlags: func() uint16 {
			w.events = append(w.events, fmt.Sprintf("game-flags:%04x", w.gameFlags))
			return w.gameFlags
		},
		loadScoreLimit: func(flags uint16) uint16 {
			w.events = append(w.events, fmt.Sprintf("limit:%04x:%d", flags, w.limit))
			return w.limit
		},
		firstTeam: func() *checkVictoryTeam509A60 {
			w.events = append(w.events, "first-team")
			if len(w.teams) == 0 {
				return nil
			}
			return w.teams[0]
		},
		nextTeam: func(team *checkVictoryTeam509A60) *checkVictoryTeam509A60 {
			w.events = append(w.events, "next-team:"+team.name)
			i := checkVictoryIndex509A60(w.teams, team) + 1
			if i <= 0 || i >= len(w.teams) {
				return nil
			}
			return w.teams[i]
		},
		loadTeamScore: func(team *checkVictoryTeam509A60) int32 {
			w.events = append(w.events, fmt.Sprintf("team-score:%s:%d", team.name, team.score))
			return team.score
		},
		firstPlayer: func() *checkVictoryObject509A60 {
			w.events = append(w.events, "first-player")
			if len(w.players) == 0 {
				return nil
			}
			return w.players[0]
		},
		nextPlayer: func(unit *checkVictoryObject509A60) *checkVictoryObject509A60 {
			w.events = append(w.events, "next-player:"+unit.name)
			i := checkVictoryIndex509A60(w.players, unit) + 1
			if i <= 0 || i >= len(w.players) {
				return nil
			}
			return w.players[i]
		},
		loadUpdate: func(unit *checkVictoryObject509A60) *checkVictoryUpdate509A60 {
			w.events = append(w.events, "update:"+unit.name)
			return unit.update
		},
		loadPlayer: func(update *checkVictoryUpdate509A60) *checkVictoryPlayer509A60 {
			w.events = append(w.events, "player:"+update.player.name)
			return update.player
		},
		loadPlayerFlags: func(player *checkVictoryPlayer509A60) uint32 {
			w.events = append(w.events, fmt.Sprintf("player-flags:%s:%x", player.name, player.flags))
			return player.flags
		},
		loadPlayerScore: func(player *checkVictoryPlayer509A60) int32 {
			w.events = append(w.events, fmt.Sprintf("player-score:%s:%d", player.name, player.score))
			return player.score
		},
		loadPlayerDeaths: func(player *checkVictoryPlayer509A60) uint32 {
			w.events = append(w.events, fmt.Sprintf("player-deaths:%s:%d", player.name, player.deaths))
			return player.deaths
		},
		hasTeam: func(unit *checkVictoryObject509A60) bool {
			w.events = append(w.events, fmt.Sprintf("has-team:%s:%t", unit.name, unit.teamID != 0))
			return unit.teamID != 0
		},
		loadObjectTeamID: func(unit *checkVictoryObject509A60) uint8 {
			w.events = append(w.events, fmt.Sprintf("team-id:%s:%d", unit.name, unit.teamID))
			return unit.teamID
		},
		teamByID: func(id uint8) *checkVictoryTeam509A60 {
			w.events = append(w.events, fmt.Sprintf("team-by-id:%d", id))
			return w.teamByID[id]
		},
		gameplayHasRivals: func() bool {
			w.rivalCalls++
			w.events = append(w.events, fmt.Sprintf("rivals:%t", w.rivals))
			return w.rivals
		},
		setGameFlags: func(flags uint32) {
			w.setFlags = append(w.setFlags, flags)
			w.events = append(w.events, fmt.Sprintf("set-flags:%x", flags))
		},
		sendTeamWinner: func(team *checkVictoryTeam509A60, arg uint8) int32 {
			name := "nil"
			if team != nil {
				name = team.name
			}
			w.winner = fmt.Sprintf("team:%s:%d", name, arg)
			w.events = append(w.events, "winner:"+w.winner)
			return -1
		},
		sendPlayerWinner: func(unit *checkVictoryObject509A60, arg uint8) int32 {
			name := "nil"
			if unit != nil {
				name = unit.name
			}
			w.winner = fmt.Sprintf("player:%s:%d", name, arg)
			w.events = append(w.events, "winner:"+w.winner)
			return -2
		},
	}
}

func checkVictoryObjectWithPlayer509A60(name string, deaths uint32, score int32) *checkVictoryObject509A60 {
	return &checkVictoryObject509A60{name: name, update: &checkVictoryUpdate509A60{
		player: &checkVictoryPlayer509A60{name: name, deaths: deaths, score: score},
	}}
}

func TestCheckVictory509A60EliminationCandidates(t *testing.T) {
	team1 := &checkVictoryTeam509A60{name: "team1"}
	team2 := &checkVictoryTeam509A60{name: "team2"}
	makeWorld := func(players ...*checkVictoryObject509A60) *checkVictoryWorld509A60 {
		return &checkVictoryWorld509A60{
			flags:     checkVictoryEliminationFlag509A60,
			gameFlags: 0x4567,
			limit:     3,
			rivals:    true,
			teamByID: map[uint8]*checkVictoryTeam509A60{
				1: team1,
				2: team2,
			},
			players: players,
		}
	}
	unteamed := func(name string) *checkVictoryObject509A60 {
		return checkVictoryObjectWithPlayer509A60(name, 2, 0)
	}
	teamed := func(name string, id uint8) *checkVictoryObject509A60 {
		unit := unteamed(name)
		unit.teamID = id
		return unit
	}

	tests := []struct {
		name          string
		world         *checkVictoryWorld509A60
		wantWinner    string
		wantCompleted bool
		wantRivals    int
	}{
		{name: "sole unteamed player", world: makeWorld(unteamed("p1")), wantWinner: "player:p1:0", wantCompleted: true, wantRivals: 1},
		{name: "one common team", world: makeWorld(teamed("p1", 1), teamed("p2", 1)), wantWinner: "team:team1:0", wantCompleted: true, wantRivals: 1},
		{name: "different teams conflict", world: makeWorld(teamed("p1", 1), teamed("p2", 2))},
		{name: "team and unteamed conflict", world: makeWorld(teamed("p1", 1), unteamed("p2"))},
		{name: "two unteamed conflict", world: makeWorld(unteamed("p1"), unteamed("p2"))},
		{name: "no surviving candidate sends nil", world: func() *checkVictoryWorld509A60 {
			observer := unteamed("observer")
			observer.update.player.flags = checkVictoryObserverFlag509A60
			eliminated := checkVictoryObjectWithPlayer509A60("eliminated", 3, 0)
			return makeWorld(observer, eliminated)
		}(), wantWinner: "player:nil:0", wantCompleted: true, wantRivals: 1},
		{name: "unknown team remains nil", world: makeWorld(teamed("unknown", 9), unteamed("p1")), wantWinner: "player:p1:0", wantCompleted: true, wantRivals: 1},
		{name: "gameplay no longer has rivals", world: func() *checkVictoryWorld509A60 {
			world := makeWorld(unteamed("p1"))
			world.rivals = false
			return world
		}(), wantRivals: 1},
		{name: "zero limit", world: func() *checkVictoryWorld509A60 {
			world := makeWorld(unteamed("p1"))
			world.limit = 0
			return world
		}()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			checkVictory509A60(tc.world.hooks())
			if tc.world.winner != tc.wantWinner {
				t.Fatalf("winner = %q, want %q; events: %q", tc.world.winner, tc.wantWinner, tc.world.events)
			}
			completed := reflect.DeepEqual(tc.world.setFlags, []uint32{checkVictoryCompleteFlag509A60})
			if completed != tc.wantCompleted {
				t.Fatalf("completed = %t, want %t; flags: %x; events: %q", completed, tc.wantCompleted, tc.world.setFlags, tc.world.events)
			}
			if tc.world.rivalCalls != tc.wantRivals {
				t.Fatalf("rival calls = %d, want %d; events: %q", tc.world.rivalCalls, tc.wantRivals, tc.world.events)
			}
		})
	}
}

func TestCheckVictory509A60EliminationReadOrder(t *testing.T) {
	observer := checkVictoryObjectWithPlayer509A60("observer", 0, 0)
	observer.update.player.flags = checkVictoryObserverFlag509A60
	survivor := checkVictoryObjectWithPlayer509A60("survivor", 1, 0)
	world := &checkVictoryWorld509A60{
		flags:     checkVictoryEliminationFlag509A60,
		gameFlags: 0x0401,
		limit:     2,
		rivals:    true,
		players:   []*checkVictoryObject509A60{observer, survivor},
	}

	checkVictory509A60(world.hooks())
	want := []string{
		"has-flag:400", "game-flags:0401", "limit:0401:2", "first-player",
		"update:observer", "player:observer", "player-flags:observer:1", "next-player:observer",
		"update:survivor", "player:survivor", "player-flags:survivor:0", "player-deaths:survivor:1",
		"has-team:survivor:false", "next-player:survivor", "rivals:true", "set-flags:8",
		"winner:player:survivor:0",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", world.events, want)
	}
}

func TestCheckVictory509A60ScoreModes(t *testing.T) {
	tests := []struct {
		name       string
		flags      uint32
		teams      []int32
		players    []int32
		observer   bool
		teamed     bool
		limit      uint16
		wantWinner string
	}{
		{name: "coop team skips", flags: checkVictoryCoopTeamFlag509A60, limit: 1},
		{name: "zero limit skips", limit: 0},
		{name: "first qualifying team wins", teams: []int32{4, 5, 6}, players: []int32{9}, limit: 5, wantWinner: "team:t1:0"},
		{name: "team has precedence", teams: []int32{5}, players: []int32{9}, limit: 5, wantWinner: "team:t0:0"},
		{name: "teamed player remains eligible", teams: []int32{4}, players: []int32{5}, teamed: true, limit: 5, wantWinner: "player:p0:0"},
		{name: "observer is skipped", players: []int32{9}, observer: true, limit: 5},
		{name: "signed negative scores stay below limit", teams: []int32{-1}, players: []int32{-1}, limit: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			world := &checkVictoryWorld509A60{flags: tc.flags, gameFlags: 0x0102, limit: tc.limit, rivals: true}
			for i, score := range tc.teams {
				world.teams = append(world.teams, &checkVictoryTeam509A60{name: fmt.Sprintf("t%d", i), score: score})
			}
			for i, score := range tc.players {
				unit := checkVictoryObjectWithPlayer509A60(fmt.Sprintf("p%d", i), 0, score)
				if tc.observer {
					unit.update.player.flags = checkVictoryObserverFlag509A60
				}
				if tc.teamed {
					unit.teamID = 1
				}
				world.players = append(world.players, unit)
			}

			checkVictory509A60(world.hooks())
			if world.winner != tc.wantWinner {
				t.Fatalf("winner = %q, want %q; events: %q", world.winner, tc.wantWinner, world.events)
			}
			if tc.wantWinner == "" && len(world.setFlags) != 0 {
				t.Fatalf("flags = %x, want none; events: %q", world.setFlags, world.events)
			}
			if tc.wantWinner != "" && !reflect.DeepEqual(world.setFlags, []uint32{checkVictoryCompleteFlag509A60}) {
				t.Fatalf("flags = %x, want completion; events: %q", world.setFlags, world.events)
			}
		})
	}
}

func TestCheckVictory509A60EliminationPrecedesCoopTeam(t *testing.T) {
	unit := checkVictoryObjectWithPlayer509A60("p1", 0, 0)
	world := &checkVictoryWorld509A60{
		flags:     checkVictoryEliminationFlag509A60 | checkVictoryCoopTeamFlag509A60,
		gameFlags: 0x0600,
		limit:     1,
		rivals:    true,
		players:   []*checkVictoryObject509A60{unit},
	}
	checkVictory509A60(world.hooks())
	if world.winner != "player:p1:0" {
		t.Fatalf("winner = %q; events: %q", world.winner, world.events)
	}
}
