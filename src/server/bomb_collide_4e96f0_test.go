package server

import (
	"reflect"
	"testing"
)

type bombCollideTestObject4E96F0 struct {
	name   string
	class  uint8
	flags  uint32
	update *bombCollideTestUpdate4E96F0
}

type bombCollideTestUpdate4E96F0 struct {
	name   string
	block  string
	target *bombCollideTestObject4E96F0
}

type bombCollideTestState4E96F0 struct {
	events    []string
	mode      int32
	first     *bombCollideTestObject4E96F0
	sameTeam  int32
	onScript  func()
	onSame    func()
	damageObj *bombCollideTestObject4E96F0
	damage    int32
}

func (s *bombCollideTestState4E96F0) hooks() bombCollideHooks4E96F0[
	*bombCollideTestObject4E96F0,
	*bombCollideTestUpdate4E96F0,
	string,
] {
	return bombCollideHooks4E96F0[
		*bombCollideTestObject4E96F0,
		*bombCollideTestUpdate4E96F0,
		string,
	]{
		loadUpdateData: func(obj *bombCollideTestObject4E96F0) *bombCollideTestUpdate4E96F0 {
			s.events = append(s.events, "update:"+obj.name)
			return obj.update
		},
		gameModeCoop: func() int32 {
			s.events = append(s.events, "mode")
			return s.mode
		},
		firstPlayerUnit: func() *bombCollideTestObject4E96F0 {
			s.events = append(s.events, "first")
			return s.first
		},
		loadFlags: func(obj *bombCollideTestObject4E96F0) uint32 {
			if obj == nil {
				s.events = append(s.events, "flags:nil")
				panic("nil flags")
			}
			s.events = append(s.events, "flags:"+obj.name)
			return obj.flags
		},
		collisionBlock: func(update *bombCollideTestUpdate4E96F0) string {
			s.events = append(s.events, "block:"+update.name)
			return update.block
		},
		scriptCallback: func(block string, caller, trigger *bombCollideTestObject4E96F0) {
			s.events = append(s.events, "script:"+block+":"+caller.name+":"+trigger.name)
			if s.onScript != nil {
				s.onScript()
			}
		},
		classLow: func(obj *bombCollideTestObject4E96F0) uint8 {
			s.events = append(s.events, "class:"+obj.name)
			return obj.class
		},
		unitsOnSameTeam: func(first, second *bombCollideTestObject4E96F0) int32 {
			s.events = append(s.events, "same:"+first.name+":"+second.name)
			if s.onSame != nil {
				s.onSame()
			}
			return s.sameTeam
		},
		storeCollideUnit: func(update *bombCollideTestUpdate4E96F0, other *bombCollideTestObject4E96F0) {
			s.events = append(s.events, "store:"+update.name+":"+other.name)
			update.target = other
		},
		damageClear: func(obj *bombCollideTestObject4E96F0, damage int32) {
			s.events = append(s.events, "damage:"+obj.name)
			s.damageObj = obj
			s.damage = damage
		},
	}
}

func TestBombCollide4E96F0CoopSuppressionUsesCachedUpdateFirst(t *testing.T) {
	bomb := &bombCollideTestObject4E96F0{name: "bomb"}
	other := &bombCollideTestObject4E96F0{name: "other", class: 6}
	first := &bombCollideTestObject4E96F0{name: "first", flags: 2}
	state := &bombCollideTestState4E96F0{mode: 1, first: first}
	bombCollide4E96F0(bomb, other, struct{ unread int }{unread: 9}, state.hooks())
	want := []string{"update:bomb", "mode", "first", "flags:first"}
	if !reflect.DeepEqual(state.events, want) || state.damageObj != nil {
		t.Fatalf("events/damage = (%#v, %p), want (%#v, nil)", state.events, state.damageObj, want)
	}
}

func TestBombCollide4E96F0CoopWithoutMonsterContinuesAfterFirstFlags(t *testing.T) {
	update := &bombCollideTestUpdate4E96F0{name: "cached", block: "collision"}
	bomb := &bombCollideTestObject4E96F0{name: "bomb", update: update}
	other := &bombCollideTestObject4E96F0{name: "other", class: 4}
	first := &bombCollideTestObject4E96F0{name: "first", flags: 0x80000000}
	state := &bombCollideTestState4E96F0{mode: 1, first: first}
	bombCollide4E96F0(bomb, other, nil, state.hooks())
	wantPrefix := []string{"update:bomb", "mode", "first", "flags:first", "block:cached", "script:collision:other:bomb"}
	if len(state.events) < len(wantPrefix) || !reflect.DeepEqual(state.events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("events = %#v, want prefix %#v", state.events, wantPrefix)
	}
	if update.target != other || state.damageObj != bomb || state.damage != 999 {
		t.Fatalf("target/damage = (%p, %p, %d), want (%p, %p, 999)", update.target, state.damageObj, state.damage, other, bomb)
	}
}

func TestBombCollide4E96F0OnlyExactCoopOneReadsFirstPlayer(t *testing.T) {
	update := &bombCollideTestUpdate4E96F0{name: "cached", block: "collision"}
	bomb := &bombCollideTestObject4E96F0{name: "bomb", update: update}
	other := &bombCollideTestObject4E96F0{name: "other", class: 2}
	state := &bombCollideTestState4E96F0{mode: 2}
	bombCollide4E96F0(bomb, other, nil, state.hooks())
	for _, event := range state.events {
		if event == "first" || event == "flags:nil" {
			t.Fatalf("noncanonical mode read first player: %#v", state.events)
		}
	}
	if update.target != other || state.damageObj != bomb || state.damage != 999 {
		t.Fatalf("target/damage = (%p, %p, %d), want (%p, %p, 999)", update.target, state.damageObj, state.damage, other, bomb)
	}
}

func TestBombCollide4E96F0NilFirstPlayerFaultOrder(t *testing.T) {
	update := &bombCollideTestUpdate4E96F0{name: "cached", block: "collision"}
	bomb := &bombCollideTestObject4E96F0{name: "bomb", update: update}
	state := &bombCollideTestState4E96F0{mode: 1}
	defer func() {
		if recover() == nil {
			t.Fatal("nil first player returned without a panic")
		}
		want := []string{"update:bomb", "mode", "first", "flags:nil"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	}()
	bombCollide4E96F0(bomb, nil, nil, state.hooks())
}

func TestBombCollide4E96F0ScriptPrecedesNilAndLiveTargetGates(t *testing.T) {
	update := &bombCollideTestUpdate4E96F0{name: "cached", block: "collision"}
	bomb := &bombCollideTestObject4E96F0{name: "bomb", update: update}
	state := &bombCollideTestState4E96F0{}
	hooks := state.hooks()
	hooks.scriptCallback = func(block string, caller, trigger *bombCollideTestObject4E96F0) {
		state.events = append(state.events, "script:nil")
		if block != "collision" || caller != nil || trigger != bomb {
			t.Fatalf("script args = (%q, %p, %p)", block, caller, trigger)
		}
	}
	bombCollide4E96F0(bomb, nil, nil, hooks)
	want := []string{"update:bomb", "mode", "block:cached", "script:nil"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
}

func TestBombCollide4E96F0CachesUpdateAndReadsPostScriptFields(t *testing.T) {
	cached := &bombCollideTestUpdate4E96F0{name: "cached", block: "collision"}
	replacement := &bombCollideTestUpdate4E96F0{name: "replacement", block: "wrong"}
	bomb := &bombCollideTestObject4E96F0{name: "bomb", update: cached}
	other := &bombCollideTestObject4E96F0{name: "other"}
	state := &bombCollideTestState4E96F0{}
	state.onScript = func() {
		bomb.update = replacement
		other.class = 6
		other.flags = 0
	}
	bombCollide4E96F0(bomb, other, nil, state.hooks())
	want := []string{
		"update:bomb", "mode", "block:cached", "script:collision:other:bomb",
		"class:other", "same:bomb:other", "flags:other", "store:cached:other", "damage:bomb",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
	if cached.target != other || replacement.target != nil || state.damageObj != bomb || state.damage != 999 {
		t.Fatalf("cached/replacement/damage = (%p, %p, %p, %d)", cached.target, replacement.target, state.damageObj, state.damage)
	}
}

func TestBombCollide4E96F0TargetFiltersStopAtOriginalGate(t *testing.T) {
	tests := []struct {
		name     string
		class    uint8
		flags    uint32
		sameTeam int32
		last     string
	}{
		{name: "non-unit", class: 1, last: "class:other"},
		{name: "same-team-nonzero", class: 2, sameTeam: -1, last: "same:bomb:other"},
		{name: "destroyed", class: 4, flags: 0x8000, last: "flags:other"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			update := &bombCollideTestUpdate4E96F0{name: "cached", block: "collision"}
			bomb := &bombCollideTestObject4E96F0{name: "bomb", update: update}
			other := &bombCollideTestObject4E96F0{name: "other", class: tc.class, flags: tc.flags}
			state := &bombCollideTestState4E96F0{sameTeam: tc.sameTeam}
			bombCollide4E96F0(bomb, other, nil, state.hooks())
			if got := state.events[len(state.events)-1]; got != tc.last {
				t.Fatalf("last event = %q, want %q; all=%#v", got, tc.last, state.events)
			}
			if update.target != nil || state.damageObj != nil {
				t.Fatalf("rejected path stored/damaged = (%p, %p)", update.target, state.damageObj)
			}
		})
	}
}

func TestBombCollide4E96F0FlagsAreLoadedAfterSameTeamCallback(t *testing.T) {
	update := &bombCollideTestUpdate4E96F0{name: "cached", block: "collision"}
	bomb := &bombCollideTestObject4E96F0{name: "bomb", update: update}
	other := &bombCollideTestObject4E96F0{name: "other", class: 2}
	state := &bombCollideTestState4E96F0{}
	state.onSame = func() { other.flags = 0x8000 }
	bombCollide4E96F0(bomb, other, nil, state.hooks())
	if update.target != nil || state.damageObj != nil {
		t.Fatalf("post-team destroyed target stored/damaged = (%p, %p)", update.target, state.damageObj)
	}
	last := state.events[len(state.events)-2:]
	if want := []string{"same:bomb:other", "flags:other"}; !reflect.DeepEqual(last, want) {
		t.Fatalf("last events = %#v, want %#v", last, want)
	}
}
