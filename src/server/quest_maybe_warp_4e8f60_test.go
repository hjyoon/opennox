package server

import (
	"fmt"
	"reflect"
	"testing"
)

type questMaybeWarpTestPlayer4E8F60 struct {
	name  string
	index uint8
	state uint32
	stage uint32
}

type questMaybeWarpTestUpdate4E8F60 struct {
	name   string
	player *questMaybeWarpTestPlayer4E8F60
	gate   *questMaybeWarpTestObject4E8F60
}

type questMaybeWarpTestObject4E8F60 struct {
	name   string
	update *questMaybeWarpTestUpdate4E8F60
	next   *questMaybeWarpTestObject4E8F60
}

type questMaybeWarpTestState4E8F60 struct {
	events      []string
	stage       uint32
	threshold   uint32
	first       *questMaybeWarpTestObject4E8F60
	gameHost    int32
	noRendering int32
	onLoadStage func(*questMaybeWarpTestPlayer4E8F60)
}

func (s *questMaybeWarpTestState4E8F60) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *questMaybeWarpTestState4E8F60) hooks() questMaybeWarpHooks4E8F60[
	*questMaybeWarpTestObject4E8F60,
	*questMaybeWarpTestUpdate4E8F60,
	*questMaybeWarpTestPlayer4E8F60,
] {
	return questMaybeWarpHooks4E8F60[
		*questMaybeWarpTestObject4E8F60,
		*questMaybeWarpTestUpdate4E8F60,
		*questMaybeWarpTestPlayer4E8F60,
	]{
		currentQuestStage: func() uint32 {
			s.event("current:%08x", s.stage)
			return s.stage
		},
		nextStageThreshold: func(stage uint32) uint32 {
			s.event("threshold:%08x", stage)
			return s.threshold
		},
		firstUnit: func() *questMaybeWarpTestObject4E8F60 {
			s.event("first")
			return s.first
		},
		nextUnit: func(unit *questMaybeWarpTestObject4E8F60) *questMaybeWarpTestObject4E8F60 {
			s.event("next:%s", unit.name)
			return unit.next
		},
		loadUpdateData: func(unit *questMaybeWarpTestObject4E8F60) *questMaybeWarpTestUpdate4E8F60 {
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
		loadPlayer: func(update *questMaybeWarpTestUpdate4E8F60) *questMaybeWarpTestPlayer4E8F60 {
			s.event("player:%s", update.name)
			return update.player
		},
		loadPlayerIndex: func(player *questMaybeWarpTestPlayer4E8F60) uint8 {
			s.event("index:%s", player.name)
			return player.index
		},
		loadQuestState: func(player *questMaybeWarpTestPlayer4E8F60) uint32 {
			s.event("state:%s", player.name)
			return player.state
		},
		loadQuestWarpGate: func(update *questMaybeWarpTestUpdate4E8F60) *questMaybeWarpTestObject4E8F60 {
			s.event("gate:%s", update.name)
			return update.gate
		},
		loadQuestStage: func(player *questMaybeWarpTestPlayer4E8F60) uint32 {
			s.event("stage:%s", player.name)
			if s.onLoadStage != nil {
				s.onLoadStage(player)
			}
			return player.stage
		},
	}
}

func questMaybeWarpTestUnit4E8F60(
	name string,
	index uint8,
	state, stage uint32,
	inGate bool,
) *questMaybeWarpTestObject4E8F60 {
	player := &questMaybeWarpTestPlayer4E8F60{name: name, index: index, state: state, stage: stage}
	update := &questMaybeWarpTestUpdate4E8F60{name: name, player: player}
	unit := &questMaybeWarpTestObject4E8F60{name: name, update: update}
	if inGate {
		update.gate = &questMaybeWarpTestObject4E8F60{name: name + "-gate"}
	}
	return unit
}

func TestQuestMaybeWarp4E8F60ComputesThresholdBeforeEmptyList(t *testing.T) {
	state := &questMaybeWarpTestState4E8F60{stage: 0xfedcba98, threshold: 0x87654321}
	if got := questMaybeWarp4E8F60(state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"current:fedcba98", "threshold:fedcba98", "first"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestQuestMaybeWarp4E8F60DedicatedHostSkipAndPlayerReload(t *testing.T) {
	host := questMaybeWarpTestUnit4E8F60("host", 31, 1, ^uint32(0), false)
	inactive := questMaybeWarpTestUnit4E8F60("inactive", 7, 0, ^uint32(0), false)
	ready := questMaybeWarpTestUnit4E8F60("ready", 8, 2, 9, true)
	host.next, inactive.next = inactive, ready
	state := &questMaybeWarpTestState4E8F60{
		stage: 4, threshold: 9, first: host, gameHost: -1, noRendering: 2,
	}
	if got := questMaybeWarp4E8F60(state.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"current:00000004", "threshold:00000004", "first",
		"update:host", "host", "render", "player:host", "index:host", "next:host",
		"update:inactive", "host", "render", "player:inactive", "index:inactive",
		"player:inactive", "state:inactive", "next:inactive",
		"update:ready", "host", "render", "player:ready", "index:ready",
		"player:ready", "state:ready", "gate:ready", "stage:ready", "next:ready",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestQuestMaybeWarp4E8F60HostAndRenderingChecksShortCircuit(t *testing.T) {
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
			unit := questMaybeWarpTestUnit4E8F60("p", 31, 0, 0, false)
			state := &questMaybeWarpTestState4E8F60{
				first: unit, gameHost: tc.gameHost, noRendering: tc.noRendering,
			}
			if got := questMaybeWarp4E8F60(state.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			got := state.events[3:]
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("event tail = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestQuestMaybeWarp4E8F60MissingGateReturnsBeforeStageAndNext(t *testing.T) {
	unit := questMaybeWarpTestUnit4E8F60("p", 4, 0x80000000, ^uint32(0), false)
	state := &questMaybeWarpTestState4E8F60{first: unit, threshold: 1}
	if got := questMaybeWarp4E8F60(state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	wantTail := []string{"update:p", "host", "player:p", "state:p", "gate:p"}
	gotTail := state.events[len(state.events)-len(wantTail):]
	if !reflect.DeepEqual(gotTail, wantTail) {
		t.Fatalf("event tail = %q, want %q", gotTail, wantTail)
	}
}

func TestQuestMaybeWarp4E8F60UnsignedStageAndLiveSuccessor(t *testing.T) {
	low := questMaybeWarpTestUnit4E8F60("low", 1, 1, 0x7fffffff, true)
	skipped := questMaybeWarpTestUnit4E8F60("skipped", 2, 1, 0, false)
	high := questMaybeWarpTestUnit4E8F60("high", 3, 1, 0xffffffff, true)
	low.next, skipped.next = skipped, high
	state := &questMaybeWarpTestState4E8F60{
		first: low, threshold: 0x80000000,
		onLoadStage: func(player *questMaybeWarpTestPlayer4E8F60) {
			if player == low.update.player {
				low.next = high
			}
		},
	}
	if got := questMaybeWarp4E8F60(state.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	for _, event := range state.events {
		if event == "update:skipped" {
			t.Fatal("followed successor captured before the stage callback")
		}
	}
}

func TestQuestMaybeWarp4E8F60LaterMissingGateOverridesAllowedPlayer(t *testing.T) {
	allowed := questMaybeWarpTestUnit4E8F60("allowed", 1, 1, 12, true)
	missing := questMaybeWarpTestUnit4E8F60("missing", 2, 1, 12, false)
	allowed.next = missing
	state := &questMaybeWarpTestState4E8F60{first: allowed, threshold: 10}
	if got := questMaybeWarp4E8F60(state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestQuestMaybeWarp4E8F60RequiresReachedStage(t *testing.T) {
	a := questMaybeWarpTestUnit4E8F60("a", 1, 1, 8, true)
	b := questMaybeWarpTestUnit4E8F60("b", 2, 1, 9, true)
	a.next = b
	state := &questMaybeWarpTestState4E8F60{first: a, threshold: 10}
	if got := questMaybeWarp4E8F60(state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}
