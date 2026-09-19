package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type deathmatchLowScoreTestTeam5095E0 struct {
	name  string
	score int32
}

type deathmatchLowScoreTestPlayer5095E0 struct {
	name  string
	flags uint32
	score int32
}

type deathmatchLowScoreTestUpdate5095E0 struct {
	player *deathmatchLowScoreTestPlayer5095E0
}

type deathmatchLowScoreTestObject5095E0 struct {
	name   string
	team   bool
	update *deathmatchLowScoreTestUpdate5095E0
}

type deathmatchLowScoreTestWorld5095E0 struct {
	teams   []*deathmatchLowScoreTestTeam5095E0
	players []*deathmatchLowScoreTestObject5095E0
	events  []string
}

func deathmatchLowScoreIndex5095E0[T comparable](values []T, current T) int {
	for i, value := range values {
		if value == current {
			return i
		}
	}
	return -1
}

func (w *deathmatchLowScoreTestWorld5095E0) hooks() deathmatchLowScoreWinnerHooks5095E0[
	*deathmatchLowScoreTestTeam5095E0,
	*deathmatchLowScoreTestObject5095E0,
	*deathmatchLowScoreTestUpdate5095E0,
	*deathmatchLowScoreTestPlayer5095E0,
] {
	return deathmatchLowScoreWinnerHooks5095E0[
		*deathmatchLowScoreTestTeam5095E0,
		*deathmatchLowScoreTestObject5095E0,
		*deathmatchLowScoreTestUpdate5095E0,
		*deathmatchLowScoreTestPlayer5095E0,
	]{
		firstTeam: func() *deathmatchLowScoreTestTeam5095E0 {
			w.events = append(w.events, "first-team")
			if len(w.teams) == 0 {
				return nil
			}
			return w.teams[0]
		},
		nextTeam: func(team *deathmatchLowScoreTestTeam5095E0) *deathmatchLowScoreTestTeam5095E0 {
			w.events = append(w.events, "next-team:"+team.name)
			i := deathmatchLowScoreIndex5095E0(w.teams, team) + 1
			if i <= 0 || i >= len(w.teams) {
				return nil
			}
			return w.teams[i]
		},
		loadTeamScore: func(team *deathmatchLowScoreTestTeam5095E0) int32 {
			w.events = append(w.events, fmt.Sprintf("team-score:%s:%d", team.name, team.score))
			return team.score
		},
		firstPlayer: func() *deathmatchLowScoreTestObject5095E0 {
			w.events = append(w.events, "first-player")
			if len(w.players) == 0 {
				return nil
			}
			return w.players[0]
		},
		loadUpdate: func(unit *deathmatchLowScoreTestObject5095E0) *deathmatchLowScoreTestUpdate5095E0 {
			w.events = append(w.events, "update:"+unit.name)
			return unit.update
		},
		hasTeam: func(unit *deathmatchLowScoreTestObject5095E0) bool {
			w.events = append(w.events, fmt.Sprintf("has-team:%s:%t", unit.name, unit.team))
			return unit.team
		},
		loadPlayer: func(update *deathmatchLowScoreTestUpdate5095E0) *deathmatchLowScoreTestPlayer5095E0 {
			w.events = append(w.events, "player:"+update.player.name)
			return update.player
		},
		loadPlayerFlags: func(player *deathmatchLowScoreTestPlayer5095E0) uint32 {
			w.events = append(w.events, fmt.Sprintf("flags:%s:%08x", player.name, player.flags))
			return player.flags
		},
		loadPlayerScore: func(player *deathmatchLowScoreTestPlayer5095E0) int32 {
			w.events = append(w.events, fmt.Sprintf("player-score:%s:%d", player.name, player.score))
			return player.score
		},
		nextPlayer: func(unit *deathmatchLowScoreTestObject5095E0) *deathmatchLowScoreTestObject5095E0 {
			w.events = append(w.events, "next-player:"+unit.name)
			i := deathmatchLowScoreIndex5095E0(w.players, unit) + 1
			if i <= 0 || i >= len(w.players) {
				return nil
			}
			return w.players[i]
		},
		setGameFlags: func(flags uint32) {
			w.events = append(w.events, fmt.Sprintf("set-flags:%08x", flags))
		},
		sendTeamWinner: func(team *deathmatchLowScoreTestTeam5095E0, flag uint8) int32 {
			name := "nil"
			if team != nil {
				name = team.name
			}
			w.events = append(w.events, fmt.Sprintf("team-winner:%s:%02x", name, flag))
			return 0x11111111
		},
		sendPlayerWinner: func(unit *deathmatchLowScoreTestObject5095E0, flag uint8) int32 {
			w.events = append(w.events, fmt.Sprintf("player-winner:%s:%02x", unit.name, flag))
			return 0x22222222
		},
	}
}

func deathmatchLowScorePlayer5095E0(name string, score int32) *deathmatchLowScoreTestObject5095E0 {
	return &deathmatchLowScoreTestObject5095E0{
		name: name,
		update: &deathmatchLowScoreTestUpdate5095E0{
			player: &deathmatchLowScoreTestPlayer5095E0{name: name, score: score},
		},
	}
}

func TestDeathmatchLowScoreWinner5095E0CandidateRules(t *testing.T) {
	tests := []struct {
		name       string
		teams      []int32
		players    []int32
		wantResult int32
		wantLast   string
	}{
		{name: "no candidates", wantResult: 0, wantLast: "set-flags:00000008"},
		{name: "one team", teams: []int32{7}, wantResult: 0x11111111, wantLast: "team-winner:t0:01"},
		{name: "equal teams tie", teams: []int32{7, 7}, wantResult: 0x11111111, wantLast: "team-winner:nil:01"},
		{name: "lower team clears tie", teams: []int32{7, 7, 6}, wantResult: 0x11111111, wantLast: "team-winner:t2:01"},
		{name: "player below team", teams: []int32{7}, players: []int32{6}, wantResult: 0x22222222, wantLast: "player-winner:p0:01"},
		{name: "first player equal team replaces it", teams: []int32{7}, players: []int32{7}, wantResult: 0x22222222, wantLast: "player-winner:p0:01"},
		{name: "second equal player ties", teams: []int32{7}, players: []int32{7, 7}, wantResult: 0x11111111, wantLast: "team-winner:nil:01"},
		{name: "lower player clears tie", teams: []int32{7, 7}, players: []int32{6}, wantResult: 0x22222222, wantLast: "player-winner:p0:01"},
		{name: "signed negative score", teams: []int32{0}, players: []int32{-1}, wantResult: 0x22222222, wantLast: "player-winner:p0:01"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			world := &deathmatchLowScoreTestWorld5095E0{}
			for i, score := range tc.teams {
				world.teams = append(world.teams, &deathmatchLowScoreTestTeam5095E0{name: fmt.Sprintf("t%d", i), score: score})
			}
			for i, score := range tc.players {
				world.players = append(world.players, deathmatchLowScorePlayer5095E0(fmt.Sprintf("p%d", i), score))
			}
			got := deathmatchLowScoreWinner5095E0(world.hooks())
			if got != tc.wantResult {
				t.Fatalf("result = %08x, want %08x; events: %q", uint32(got), uint32(tc.wantResult), world.events)
			}
			if gotLast := world.events[len(world.events)-1]; gotLast != tc.wantLast {
				t.Fatalf("last event = %q, want %q; events: %q", gotLast, tc.wantLast, world.events)
			}
			flagIndex := deathmatchLowScoreIndex5095E0(world.events, "set-flags:00000008")
			if flagIndex < 0 || flagIndex != len(world.events)-1 && flagIndex != len(world.events)-2 {
				t.Fatalf("flag order in events: %q", world.events)
			}
		})
	}
}

func TestDeathmatchLowScoreWinner5095E0SkipsTeamedAndObserverFields(t *testing.T) {
	teamed := deathmatchLowScorePlayer5095E0("teamed", math.MinInt32)
	teamed.team = true
	observer := deathmatchLowScorePlayer5095E0("observer", math.MinInt32)
	observer.update.player.flags = deathmatchLowScoreObserverFlag5095E0
	world := &deathmatchLowScoreTestWorld5095E0{players: []*deathmatchLowScoreTestObject5095E0{teamed, observer}}

	if got := deathmatchLowScoreWinner5095E0(world.hooks()); got != 0 {
		t.Fatalf("result = %08x, want 0; events: %q", uint32(got), world.events)
	}
	want := []string{
		"first-team",
		"first-player",
		"update:teamed",
		"has-team:teamed:true",
		"next-player:teamed",
		"update:observer",
		"has-team:observer:false",
		"player:observer",
		"flags:observer:00000001",
		"next-player:observer",
		"set-flags:00000008",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", world.events, want)
	}
}
