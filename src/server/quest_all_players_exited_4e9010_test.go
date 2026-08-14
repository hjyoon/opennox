package server

import (
	"fmt"
	"reflect"
	"testing"
)

type questAllPlayersExitedTestPlayer4E9010 struct {
	name  string
	index uint8
	state uint32
}

type questAllPlayersExitedTestUpdate4E9010 struct {
	name   string
	player *questAllPlayersExitedTestPlayer4E9010
	exit   *questAllPlayersExitedTestObject4E9010
}

type questAllPlayersExitedTestObject4E9010 struct {
	name   string
	update *questAllPlayersExitedTestUpdate4E9010
	next   *questAllPlayersExitedTestObject4E9010
}

type questAllPlayersExitedTestState4E9010 struct {
	events      []string
	first       *questAllPlayersExitedTestObject4E9010
	gameHost    int32
	noRendering int32
	onLoadExit  func(*questAllPlayersExitedTestUpdate4E9010)
}

func (s *questAllPlayersExitedTestState4E9010) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *questAllPlayersExitedTestState4E9010) hooks() questAllPlayersExitedHooks4E9010[
	*questAllPlayersExitedTestObject4E9010,
	*questAllPlayersExitedTestUpdate4E9010,
	*questAllPlayersExitedTestPlayer4E9010,
] {
	return questAllPlayersExitedHooks4E9010[
		*questAllPlayersExitedTestObject4E9010,
		*questAllPlayersExitedTestUpdate4E9010,
		*questAllPlayersExitedTestPlayer4E9010,
	]{
		firstUnit: func() *questAllPlayersExitedTestObject4E9010 {
			s.event("first")
			return s.first
		},
		nextUnit: func(unit *questAllPlayersExitedTestObject4E9010) *questAllPlayersExitedTestObject4E9010 {
			s.event("next:%s", unit.name)
			return unit.next
		},
		loadUpdateData: func(unit *questAllPlayersExitedTestObject4E9010) *questAllPlayersExitedTestUpdate4E9010 {
			s.event("update:%s", unit.name)
			return unit.update
		},
		gameHost: func() int32 {
			s.event("host")
			return s.gameHost
		},
		noRendering: func() int32 {
			s.event("render")
			return s.noRendering
		},
		loadPlayer: func(update *questAllPlayersExitedTestUpdate4E9010) *questAllPlayersExitedTestPlayer4E9010 {
			s.event("player:%s", update.name)
			return update.player
		},
		loadPlayerIndex: func(player *questAllPlayersExitedTestPlayer4E9010) uint8 {
			s.event("index:%s", player.name)
			return player.index
		},
		loadQuestState: func(player *questAllPlayersExitedTestPlayer4E9010) uint32 {
			s.event("state:%s", player.name)
			return player.state
		},
		loadQuestExit: func(update *questAllPlayersExitedTestUpdate4E9010) *questAllPlayersExitedTestObject4E9010 {
			s.event("exit:%s", update.name)
			if s.onLoadExit != nil {
				s.onLoadExit(update)
			}
			return update.exit
		},
	}
}

func questAllPlayersExitedTestUnit4E9010(
	name string,
	index uint8,
	state uint32,
	inExit bool,
) *questAllPlayersExitedTestObject4E9010 {
	player := &questAllPlayersExitedTestPlayer4E9010{name: name, index: index, state: state}
	update := &questAllPlayersExitedTestUpdate4E9010{name: name, player: player}
	unit := &questAllPlayersExitedTestObject4E9010{name: name, update: update}
	if inExit {
		update.exit = &questAllPlayersExitedTestObject4E9010{name: name + "-exit"}
	}
	return unit
}

func TestQuestAllPlayersExited4E9010EmptyList(t *testing.T) {
	state := &questAllPlayersExitedTestState4E9010{}
	if got := questAllPlayersExited4E9010(state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"first"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestQuestAllPlayersExited4E9010DedicatedHostSkipAndPlayerReload(t *testing.T) {
	host := questAllPlayersExitedTestUnit4E9010("host", 31, 1, false)
	inactive := questAllPlayersExitedTestUnit4E9010("inactive", 7, 0, false)
	ready := questAllPlayersExitedTestUnit4E9010("ready", 8, 0x80000000, true)
	host.next, inactive.next = inactive, ready
	state := &questAllPlayersExitedTestState4E9010{
		first: host, gameHost: -1, noRendering: 2,
	}
	if got := questAllPlayersExited4E9010(state.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"first",
		"update:host", "host", "render", "player:host", "index:host", "next:host",
		"update:inactive", "host", "render", "player:inactive", "index:inactive",
		"player:inactive", "state:inactive", "next:inactive",
		"update:ready", "host", "render", "player:ready", "index:ready",
		"player:ready", "state:ready", "exit:ready", "next:ready",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestQuestAllPlayersExited4E9010HostAndRenderingChecksShortCircuit(t *testing.T) {
	for _, tc := range []struct {
		name        string
		gameHost    int32
		noRendering int32
		want        []string
	}{
		{
			name: "not-host", gameHost: 0, noRendering: 1,
			want: []string{"update:p", "host", "player:p", "state:p", "next:p"},
		},
		{
			name: "rendering", gameHost: 1, noRendering: 0,
			want: []string{"update:p", "host", "render", "player:p", "state:p", "next:p"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := questAllPlayersExitedTestUnit4E9010("p", 31, 0, false)
			state := &questAllPlayersExitedTestState4E9010{
				first: unit, gameHost: tc.gameHost, noRendering: tc.noRendering,
			}
			if got := questAllPlayersExited4E9010(state.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			got := state.events[1:]
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("event tail = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestQuestAllPlayersExited4E9010MissingExitReturnsBeforeNext(t *testing.T) {
	unit := questAllPlayersExitedTestUnit4E9010("p", 4, 3, false)
	state := &questAllPlayersExitedTestState4E9010{first: unit}
	if got := questAllPlayersExited4E9010(state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"first", "update:p", "host", "player:p", "state:p", "exit:p"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestQuestAllPlayersExited4E9010UsesLiveSuccessorAfterExitRead(t *testing.T) {
	first := questAllPlayersExitedTestUnit4E9010("first", 1, 1, true)
	skipped := questAllPlayersExitedTestUnit4E9010("skipped", 2, 1, false)
	last := questAllPlayersExitedTestUnit4E9010("last", 3, 1, true)
	first.next, skipped.next = skipped, last
	state := &questAllPlayersExitedTestState4E9010{
		first: first,
		onLoadExit: func(update *questAllPlayersExitedTestUpdate4E9010) {
			if update == first.update {
				first.next = last
			}
		},
	}
	if got := questAllPlayersExited4E9010(state.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	for _, event := range state.events {
		if event == "update:skipped" {
			t.Fatal("followed successor captured before the exit callback")
		}
	}
}

func TestQuestAllPlayersExited4E9010LaterMissingExitOverridesReadyPlayer(t *testing.T) {
	ready := questAllPlayersExitedTestUnit4E9010("ready", 1, 1, true)
	missing := questAllPlayersExitedTestUnit4E9010("missing", 2, 1, false)
	ready.next = missing
	state := &questAllPlayersExitedTestState4E9010{first: ready}
	if got := questAllPlayersExited4E9010(state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestQuestAllPlayersExited4E9010RequiresActivePlayer(t *testing.T) {
	inactive := questAllPlayersExitedTestUnit4E9010("inactive", 1, 0, true)
	state := &questAllPlayersExitedTestState4E9010{first: inactive}
	if got := questAllPlayersExited4E9010(state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}
