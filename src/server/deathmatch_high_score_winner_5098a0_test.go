package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type highScoreTeam5098A0 struct {
	name  string
	score int32
}

type highScorePlayer5098A0 struct {
	name  string
	flags uint32
	score int32
}

type highScoreUpdate5098A0 struct{ player *highScorePlayer5098A0 }

type highScoreObject5098A0 struct {
	name   string
	team   bool
	update *highScoreUpdate5098A0
}

type highScoreWorld5098A0 struct {
	teams   []*highScoreTeam5098A0
	players []*highScoreObject5098A0
	events  []string
}

func indexHighScore5098A0[T comparable](values []T, current T) int {
	for i, value := range values {
		if value == current {
			return i
		}
	}
	return -1
}

func (w *highScoreWorld5098A0) hooks() deathmatchHighScoreWinnerHooks5098A0[
	*highScoreTeam5098A0, *highScoreObject5098A0, *highScoreUpdate5098A0, *highScorePlayer5098A0,
] {
	return deathmatchHighScoreWinnerHooks5098A0[
		*highScoreTeam5098A0, *highScoreObject5098A0, *highScoreUpdate5098A0, *highScorePlayer5098A0,
	]{
		firstTeam: func() *highScoreTeam5098A0 {
			w.events = append(w.events, "first-team")
			if len(w.teams) == 0 {
				return nil
			}
			return w.teams[0]
		},
		nextTeam: func(team *highScoreTeam5098A0) *highScoreTeam5098A0 {
			w.events = append(w.events, "next-team:"+team.name)
			i := indexHighScore5098A0(w.teams, team) + 1
			if i <= 0 || i >= len(w.teams) {
				return nil
			}
			return w.teams[i]
		},
		loadTeamScore: func(team *highScoreTeam5098A0) int32 {
			w.events = append(w.events, fmt.Sprintf("team-score:%s:%d", team.name, team.score))
			return team.score
		},
		firstPlayer: func() *highScoreObject5098A0 {
			w.events = append(w.events, "first-player")
			if len(w.players) == 0 {
				return nil
			}
			return w.players[0]
		},
		loadUpdate: func(unit *highScoreObject5098A0) *highScoreUpdate5098A0 {
			w.events = append(w.events, "update:"+unit.name)
			return unit.update
		},
		hasTeam: func(unit *highScoreObject5098A0) bool {
			w.events = append(w.events, fmt.Sprintf("has-team:%s:%t", unit.name, unit.team))
			return unit.team
		},
		loadPlayer: func(update *highScoreUpdate5098A0) *highScorePlayer5098A0 {
			w.events = append(w.events, "player:"+update.player.name)
			return update.player
		},
		loadPlayerFlags: func(player *highScorePlayer5098A0) uint32 {
			w.events = append(w.events, fmt.Sprintf("flags:%s:%08x", player.name, player.flags))
			return player.flags
		},
		loadPlayerScore: func(player *highScorePlayer5098A0) int32 {
			w.events = append(w.events, fmt.Sprintf("player-score:%s:%d", player.name, player.score))
			return player.score
		},
		nextPlayer: func(unit *highScoreObject5098A0) *highScoreObject5098A0 {
			w.events = append(w.events, "next-player:"+unit.name)
			i := indexHighScore5098A0(w.players, unit) + 1
			if i <= 0 || i >= len(w.players) {
				return nil
			}
			return w.players[i]
		},
		setGameFlags: func(flags uint32) {
			w.events = append(w.events, fmt.Sprintf("set-flags:%08x", flags))
		},
		sendTeamWinner: func(team *highScoreTeam5098A0, flag uint8) int32 {
			name := "nil"
			if team != nil {
				name = team.name
			}
			w.events = append(w.events, fmt.Sprintf("team-winner:%s:%02x", name, flag))
			return 0x11111111
		},
		sendPlayerWinner: func(unit *highScoreObject5098A0, flag uint8) int32 {
			w.events = append(w.events, fmt.Sprintf("player-winner:%s:%02x", unit.name, flag))
			return 0x22222222
		},
	}
}

func highScorePlayerObject5098A0(name string, score int32) *highScoreObject5098A0 {
	return &highScoreObject5098A0{name: name, update: &highScoreUpdate5098A0{
		player: &highScorePlayer5098A0{name: name, score: score},
	}}
}

func TestDeathmatchHighScoreWinner5098A0CandidateRules(t *testing.T) {
	tests := []struct {
		name       string
		teams      []int32
		players    []int32
		wantResult int32
		wantLast   string
	}{
		{name: "no candidates", wantLast: "set-flags:00000008"},
		{name: "one team", teams: []int32{7}, wantResult: 0x11111111, wantLast: "team-winner:t0:01"},
		{name: "equal teams tie", teams: []int32{7, 7}, wantResult: 0x11111111, wantLast: "team-winner:nil:01"},
		{name: "higher team clears tie", teams: []int32{7, 7, 8}, wantResult: 0x11111111, wantLast: "team-winner:t2:01"},
		{name: "first player equal team replaces it", teams: []int32{7}, players: []int32{7}, wantResult: 0x22222222, wantLast: "player-winner:p0:01"},
		{name: "second equal player ties", teams: []int32{7}, players: []int32{7, 7}, wantResult: 0x11111111, wantLast: "team-winner:nil:01"},
		{name: "higher player clears tie", teams: []int32{7, 7}, players: []int32{8}, wantResult: 0x22222222, wantLast: "player-winner:p0:01"},
		{name: "signed negative score", teams: []int32{-2}, players: []int32{-1}, wantResult: 0x22222222, wantLast: "player-winner:p0:01"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			world := &highScoreWorld5098A0{}
			for i, score := range tc.teams {
				world.teams = append(world.teams, &highScoreTeam5098A0{name: fmt.Sprintf("t%d", i), score: score})
			}
			for i, score := range tc.players {
				world.players = append(world.players, highScorePlayerObject5098A0(fmt.Sprintf("p%d", i), score))
			}
			if got := deathmatchHighScoreWinner5098A0(world.hooks()); got != tc.wantResult {
				t.Fatalf("result = %08x, want %08x; events: %q", uint32(got), uint32(tc.wantResult), world.events)
			}
			if got := world.events[len(world.events)-1]; got != tc.wantLast {
				t.Fatalf("last event = %q, want %q; events: %q", got, tc.wantLast, world.events)
			}
		})
	}
}

func TestDeathmatchHighScoreWinner5098A0SkipsTeamedAndObserverFields(t *testing.T) {
	teamed := highScorePlayerObject5098A0("teamed", math.MaxInt32)
	teamed.team = true
	observer := highScorePlayerObject5098A0("observer", math.MaxInt32)
	observer.update.player.flags = deathmatchHighScoreObserverFlag5098A0
	world := &highScoreWorld5098A0{players: []*highScoreObject5098A0{teamed, observer}}

	if got := deathmatchHighScoreWinner5098A0(world.hooks()); got != 0 {
		t.Fatalf("result = %08x, want 0", uint32(got))
	}
	want := []string{
		"first-team", "first-player", "update:teamed", "has-team:teamed:true", "next-player:teamed",
		"update:observer", "has-team:observer:false", "player:observer", "flags:observer:00000001",
		"next-player:observer", "set-flags:00000008",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", world.events, want)
	}
}
