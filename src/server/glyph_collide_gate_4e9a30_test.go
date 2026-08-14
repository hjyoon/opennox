package server

import (
	"fmt"
	"reflect"
	"testing"
)

type glyphCollideGateTestObject4E9A30 struct {
	name   string
	flags  uint32
	class  uint8
	parent *glyphCollideGateTestObject4E9A30
}

type glyphCollideGateTestState4E9A30 struct {
	events       []string
	flags        map[uint32]int32
	first        *glyphCollideGateTestObject4E9A30
	sameTeam     int32
	ability      int32
	onSameTeam   func()
	onFindParent func(*glyphCollideGateTestObject4E9A30)
}

func (s *glyphCollideGateTestState4E9A30) hooks() glyphCollideGateHooks4E9A30[*glyphCollideGateTestObject4E9A30] {
	return glyphCollideGateHooks4E9A30[*glyphCollideGateTestObject4E9A30]{
		gameFlag: func(flag uint32) int32 {
			s.events = append(s.events, fmt.Sprintf("game:%#x", flag))
			return s.flags[flag]
		},
		firstPlayerUnit: func() *glyphCollideGateTestObject4E9A30 {
			s.events = append(s.events, "first")
			return s.first
		},
		loadFlags: func(obj *glyphCollideGateTestObject4E9A30) uint32 {
			if obj == nil {
				s.events = append(s.events, "flags:nil")
				panic("nil flags")
			}
			s.events = append(s.events, "flags:"+obj.name)
			return obj.flags
		},
		unitsOnSameTeam: func(source, target *glyphCollideGateTestObject4E9A30) int32 {
			s.events = append(s.events, "same:"+objectName4E9A30(source)+":"+objectName4E9A30(target))
			if s.onSameTeam != nil {
				s.onSameTeam()
			}
			return s.sameTeam
		},
		findParent: func(obj *glyphCollideGateTestObject4E9A30) *glyphCollideGateTestObject4E9A30 {
			s.events = append(s.events, "parent:"+objectName4E9A30(obj))
			if s.onFindParent != nil {
				s.onFindParent(obj)
			}
			if obj == nil {
				return nil
			}
			return obj.parent
		},
		classLow: func(obj *glyphCollideGateTestObject4E9A30) uint8 {
			if obj == nil {
				s.events = append(s.events, "class:nil")
				panic("nil class")
			}
			s.events = append(s.events, "class:"+obj.name)
			return obj.class
		},
		abilityActive: func(obj *glyphCollideGateTestObject4E9A30, ability int32) int32 {
			s.events = append(s.events, fmt.Sprintf("ability:%s:%d", objectName4E9A30(obj), ability))
			return s.ability
		},
	}
}

func objectName4E9A30(obj *glyphCollideGateTestObject4E9A30) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func TestGlyphCollideGate4E9A30ExactCoopSuppression(t *testing.T) {
	first := &glyphCollideGateTestObject4E9A30{name: "first", flags: 0xa5000002}
	state := &glyphCollideGateTestState4E9A30{
		flags: map[uint32]int32{glyphCollideCoopFlag4E9A30: 1},
		first: first,
	}
	if got := glyphCollideGate4E9A30[*glyphCollideGateTestObject4E9A30](nil, nil, state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"game:0x800", "first", "flags:first"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
}

func TestGlyphCollideGate4E9A30OnlyExactCoopOneReadsFirstPlayer(t *testing.T) {
	state := &glyphCollideGateTestState4E9A30{
		flags: map[uint32]int32{glyphCollideCoopFlag4E9A30: 2},
	}
	if got := glyphCollideGate4E9A30[*glyphCollideGateTestObject4E9A30](nil, nil, state.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"game:0x800", "ability:nil:4"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
}

func TestGlyphCollideGate4E9A30NilFirstPlayerFaultOrder(t *testing.T) {
	state := &glyphCollideGateTestState4E9A30{
		flags: map[uint32]int32{glyphCollideCoopFlag4E9A30: 1},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil first player returned without a panic")
		}
		want := []string{"game:0x800", "first", "flags:nil"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	}()
	_ = glyphCollideGate4E9A30[*glyphCollideGateTestObject4E9A30](nil, nil, state.hooks())
}

func TestGlyphCollideGate4E9A30NilTargetStillChecksAbilityLast(t *testing.T) {
	for _, ability := range []int32{0, -7} {
		state := &glyphCollideGateTestState4E9A30{flags: make(map[uint32]int32), ability: ability}
		got := glyphCollideGate4E9A30[*glyphCollideGateTestObject4E9A30](nil, nil, state.hooks())
		wantResult := int32(1)
		if ability != 0 {
			wantResult = 0
		}
		if got != wantResult {
			t.Fatalf("ability %d result = %d, want %d", ability, got, wantResult)
		}
		want := []string{"game:0x800", "ability:nil:4"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("ability %d events = %#v, want %#v", ability, state.events, want)
		}
	}
}

func TestGlyphCollideGate4E9A30SameTeamSkipsParentModeButChecksAbility(t *testing.T) {
	source := &glyphCollideGateTestObject4E9A30{name: "source"}
	target := &glyphCollideGateTestObject4E9A30{name: "target"}
	for _, sameTeam := range []int32{1, -1} {
		state := &glyphCollideGateTestState4E9A30{
			flags:    map[uint32]int32{glyphCollideCoopTeamFlag4E9A30: 1},
			sameTeam: sameTeam,
		}
		if got := glyphCollideGate4E9A30(source, target, state.hooks()); got != 0 {
			t.Fatalf("same-team %d result = %d, want 0", sameTeam, got)
		}
		want := []string{"game:0x800", "same:source:target", "ability:target:4"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("same-team %d events = %#v, want %#v", sameTeam, state.events, want)
		}
	}
}

func TestGlyphCollideGate4E9A30ParentLookupAndClassReadOrder(t *testing.T) {
	sourceParent := &glyphCollideGateTestObject4E9A30{name: "source-parent", class: glyphCollidePlayerClass4E9A30}
	targetParent := &glyphCollideGateTestObject4E9A30{name: "target-parent", class: glyphCollidePlayerClass4E9A30}
	source := &glyphCollideGateTestObject4E9A30{name: "source", parent: sourceParent}
	target := &glyphCollideGateTestObject4E9A30{name: "target", parent: targetParent}
	state := &glyphCollideGateTestState4E9A30{
		flags: map[uint32]int32{glyphCollideCoopTeamFlag4E9A30: 9},
	}
	if got := glyphCollideGate4E9A30(source, target, state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"game:0x800", "same:source:target", "game:0x200",
		"parent:source", "parent:target", "class:source-parent", "class:target-parent",
		"ability:target:4",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
}

func TestGlyphCollideGate4E9A30SourceParentClassShortCircuitsTargetClass(t *testing.T) {
	sourceParent := &glyphCollideGateTestObject4E9A30{name: "source-parent", class: 0x82}
	targetParent := &glyphCollideGateTestObject4E9A30{name: "target-parent", class: glyphCollidePlayerClass4E9A30}
	source := &glyphCollideGateTestObject4E9A30{name: "source", parent: sourceParent}
	target := &glyphCollideGateTestObject4E9A30{name: "target", parent: targetParent}
	state := &glyphCollideGateTestState4E9A30{
		flags: map[uint32]int32{glyphCollideCoopTeamFlag4E9A30: 1},
	}
	if got := glyphCollideGate4E9A30(source, target, state.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"game:0x800", "same:source:target", "game:0x200",
		"parent:source", "parent:target", "class:source-parent", "ability:target:4",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
}

func TestGlyphCollideGate4E9A30ReadsLiveParentsAfterSameTeam(t *testing.T) {
	oldParent := &glyphCollideGateTestObject4E9A30{name: "old", class: 0}
	newSourceParent := &glyphCollideGateTestObject4E9A30{name: "new-source", class: glyphCollidePlayerClass4E9A30}
	newTargetParent := &glyphCollideGateTestObject4E9A30{name: "new-target", class: glyphCollidePlayerClass4E9A30}
	source := &glyphCollideGateTestObject4E9A30{name: "source", parent: oldParent}
	target := &glyphCollideGateTestObject4E9A30{name: "target", parent: oldParent}
	state := &glyphCollideGateTestState4E9A30{
		flags: map[uint32]int32{glyphCollideCoopTeamFlag4E9A30: 1},
	}
	state.onSameTeam = func() {
		source.parent = newSourceParent
		target.parent = newTargetParent
	}
	if got := glyphCollideGate4E9A30(source, target, state.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0 from live parents", got)
	}
}
