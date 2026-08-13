package opennox

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type targetSearchTestObject struct {
	ObjClass    object.Class
	ObjSubClass object.SubClass
	ObjFlags    object.Flags
	PosVec      types.Pointf
	Direction1  int16
}

type targetSearchTestArg = targetSearchArg4E6EA0[*targetSearchTestObject]
type targetSearchTestHooks = targetSearch4E6EA0Hooks[*targetSearchTestObject]

func targetSearchTestArg4E6EA0(source *targetSearchTestObject) *targetSearchTestArg {
	return &targetSearchTestArg{
		Field0:          1,
		Field4:          1,
		ClassAllow12:    object.ClassMonster,
		SubClassAllow20: object.SubClass(0x10),
		FlagsAllow28:    object.FlagActive,
		Source36:        source,
	}
}

func targetSearchTestHooks4E6EA0(each func(types.Pointf, float32, func(*targetSearchTestObject) bool)) targetSearchTestHooks {
	return targetSearchTestHooks{
		eachInCircle: each,
		class:        func(it *targetSearchTestObject) object.Class { return it.ObjClass },
		subClass:     func(it *targetSearchTestObject) object.SubClass { return it.ObjSubClass },
		flags:        func(it *targetSearchTestObject) object.Flags { return it.ObjFlags },
		position:     func(it *targetSearchTestObject) types.Pointf { return it.PosVec },
		directionInd: func(it *targetSearchTestObject) int16 { return it.Direction1 },
		sameTeam:     func(*targetSearchTestObject, *targetSearchTestObject) bool { return false },
		playerStatus: func(*targetSearchTestObject) uint32 { return 0 },
		isEnemy:      func(*targetSearchTestObject, *targetSearchTestObject) bool { return true },
		direction:    func(types.Pointf, int16, types.Pointf) uint32 { return 1 },
		canInteract:  func(*targetSearchTestObject, *targetSearchTestObject, int) bool { return true },
	}
}

func targetSearchTestCandidate4E6EA0(x, y float32) *targetSearchTestObject {
	return &targetSearchTestObject{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(0x10),
		ObjFlags:    object.FlagActive,
		PosVec:      types.Ptf(x, y),
	}
}

func TestTargetSearch4E6EA0NilArgument(t *testing.T) {
	called := false
	h := targetSearchTestHooks4E6EA0(func(types.Pointf, float32, func(*targetSearchTestObject) bool) {
		called = true
	})
	if got := targetSearch4E6EA0((*targetSearchTestObject)(nil), 5, nil, h); got != nil {
		t.Fatalf("nil argument returned %p, want nil", got)
	}
	if called {
		t.Fatal("nil argument entered the circle iterator")
	}
}

func TestTargetSearch4E6EA0ClosestAndFloat32Best(t *testing.T) {
	source := &targetSearchTestObject{PosVec: types.Ptf(10, 20)}
	first := targetSearchTestCandidate4E6EA0(11, 20+float32(math.Ldexp(1, -13)))
	tieAfterFloat32Spill := targetSearchTestCandidate4E6EA0(11, 20)
	arg := targetSearchTestArg4E6EA0(nil)
	h := targetSearchTestHooks4E6EA0(func(center types.Pointf, radius float32, visit func(*targetSearchTestObject) bool) {
		if center != source.PosVec || radius != 5 {
			t.Fatalf("iterator got center/radius %v/%v, want %v/5", center, radius, source.PosVec)
		}
		if arg.Source36 != source {
			t.Fatalf("source was not stored before iteration: %p", arg.Source36)
		}
		visit(first)
		visit(tieAfterFloat32Spill)
	})
	if got := targetSearch4E6EA0(source, 5, arg, h); got != first {
		t.Fatalf("float32-spilled tie selected %p, want first %p", got, first)
	}
}

func TestTargetSearch4E6EA0UnorderedBest(t *testing.T) {
	source := &targetSearchTestObject{}
	near := targetSearchTestCandidate4E6EA0(1, 0)
	unordered := targetSearchTestCandidate4E6EA0(float32(math.NaN()), 0)
	farAfterNaN := targetSearchTestCandidate4E6EA0(4, 0)
	h := targetSearchTestHooks4E6EA0(func(_ types.Pointf, _ float32, visit func(*targetSearchTestObject) bool) {
		visit(near)
		visit(unordered)
		visit(farAfterNaN)
	})
	if got := targetSearch4E6EA0(source, 10, targetSearchTestArg4E6EA0(nil), h); got != farAfterNaN {
		t.Fatalf("unordered best did not select the following candidate: got %p, want %p", got, farAfterNaN)
	}
}

func TestTargetSearch4E6EF0ReloadsSourceAroundCalls(t *testing.T) {
	sourceA := &targetSearchTestObject{PosVec: types.Ptf(1, 1), Direction1: 11}
	sourceB := &targetSearchTestObject{PosVec: types.Ptf(2, 2), Direction1: 22}
	sourceC := &targetSearchTestObject{PosVec: types.Ptf(3, 3), Direction1: 33}
	sourceD := &targetSearchTestObject{PosVec: types.Ptf(4, 4), Direction1: 44}
	sourceE := &targetSearchTestObject{PosVec: types.Ptf(100, 100), Direction1: 55}
	candidate := targetSearchTestCandidate4E6EA0(100, 100)
	arg := targetSearchTestArg4E6EA0(nil)
	arg.Field8 = 1
	var calls []string
	h := targetSearchTestHooks4E6EA0(func(_ types.Pointf, _ float32, visit func(*targetSearchTestObject) bool) {
		visit(candidate)
	})
	h.sameTeam = func(gotCandidate, gotSource *targetSearchTestObject) bool {
		calls = append(calls, "team")
		if gotCandidate != candidate || gotSource != sourceA {
			t.Fatalf("team got candidate/source %p/%p", gotCandidate, gotSource)
		}
		arg.Source36 = sourceB
		return false
	}
	h.isEnemy = func(gotSource, gotCandidate *targetSearchTestObject) bool {
		calls = append(calls, "enemy")
		if gotSource != sourceB || gotCandidate != candidate {
			t.Fatalf("enemy got source/candidate %p/%p", gotSource, gotCandidate)
		}
		arg.Source36 = sourceC
		return true
	}
	h.direction = func(pos types.Pointf, dir int16, target types.Pointf) uint32 {
		calls = append(calls, "direction")
		if pos != sourceC.PosVec || dir != sourceC.Direction1 || target != candidate.PosVec {
			t.Fatalf("direction got %v/%d/%v", pos, dir, target)
		}
		arg.Source36 = sourceD
		arg.Field0 = 8
		return 8
	}
	h.canInteract = func(gotSource, gotCandidate *targetSearchTestObject, flags int) bool {
		calls = append(calls, "interact")
		if gotSource != sourceD || gotCandidate != candidate || flags != 0 {
			t.Fatalf("interact got source/candidate/flags %p/%p/%d", gotSource, gotCandidate, flags)
		}
		arg.Source36 = sourceE
		return true
	}
	if got := targetSearch4E6EA0(sourceA, 1, arg, h); got != candidate {
		t.Fatalf("reloaded distance source selected %p, want %p", got, candidate)
	}
	if want := []string{"team", "enemy", "direction", "interact"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("call order = %v, want %v", calls, want)
	}
	if arg.Source36 != sourceE {
		t.Fatalf("final source = %p, want %p", arg.Source36, sourceE)
	}
}

func TestTargetSearch4E6EF0Filters(t *testing.T) {
	type behavior struct {
		sameTeam     bool
		playerStatus uint32
		enemy        bool
		direction    uint32
		interact     bool
	}
	tests := []struct {
		name            string
		setup           func(*targetSearchTestObject, *targetSearchTestArg, *behavior)
		selectCandidate bool
		calls           []string
	}{
		{name: "eligible", selectCandidate: true, calls: []string{"team", "direction", "interact"}},
		{name: "destroyed", setup: func(c *targetSearchTestObject, _ *targetSearchTestArg, _ *behavior) {
			c.ObjFlags |= object.FlagDestroyed
		}},
		{name: "same team", setup: func(_ *targetSearchTestObject, _ *targetSearchTestArg, b *behavior) { b.sameTeam = true }, calls: []string{"team"}},
		{name: "observing player", setup: func(c *targetSearchTestObject, a *targetSearchTestArg, b *behavior) {
			c.ObjClass = object.ClassPlayer
			a.ClassAllow12 = object.ClassPlayer
			b.playerStatus = 1
		}, calls: []string{"team", "player"}},
		{name: "not enemy", setup: func(_ *targetSearchTestObject, a *targetSearchTestArg, b *behavior) { a.Field8 = 1; b.enemy = false }, calls: []string{"team", "enemy"}},
		{name: "outside direction mask", setup: func(_ *targetSearchTestObject, _ *targetSearchTestArg, b *behavior) { b.direction = 2 }, calls: []string{"team", "direction"}},
		{name: "self", setup: func(c *targetSearchTestObject, a *targetSearchTestArg, _ *behavior) { a.Source36 = c }, calls: []string{"team", "direction"}},
		{name: "cannot interact", setup: func(_ *targetSearchTestObject, _ *targetSearchTestArg, b *behavior) { b.interact = false }, calls: []string{"team", "direction", "interact"}},
		{name: "class not allowed", setup: func(_ *targetSearchTestObject, a *targetSearchTestArg, _ *behavior) {
			a.ClassAllow12 = object.ClassPlayer
		}, calls: []string{"team", "direction", "interact"}},
		{name: "class disallowed", setup: func(_ *targetSearchTestObject, a *targetSearchTestArg, _ *behavior) {
			a.ClassDisallow16 = object.ClassMonster
		}, calls: []string{"team", "direction", "interact"}},
		{name: "flags not allowed", setup: func(_ *targetSearchTestObject, a *targetSearchTestArg, _ *behavior) {
			a.FlagsAllow28 = object.FlagEnabled
		}, calls: []string{"team", "direction", "interact"}},
		{name: "flags disallowed", setup: func(_ *targetSearchTestObject, a *targetSearchTestArg, _ *behavior) {
			a.FlagsDisallow32 = object.FlagActive
		}, calls: []string{"team", "direction", "interact"}},
		{name: "zero subclass bypasses masks", setup: func(c *targetSearchTestObject, a *targetSearchTestArg, _ *behavior) {
			c.ObjSubClass = 0
			a.SubClassAllow20 = 0
			a.SubClassDisallow24 = ^object.SubClass(0)
		}, selectCandidate: true, calls: []string{"team", "direction", "interact"}},
		{name: "subclass not allowed", setup: func(_ *targetSearchTestObject, a *targetSearchTestArg, _ *behavior) {
			a.SubClassAllow20 = object.SubClass(0x20)
		}, calls: []string{"team", "direction", "interact"}},
		{name: "subclass disallowed", setup: func(_ *targetSearchTestObject, a *targetSearchTestArg, _ *behavior) {
			a.SubClassDisallow24 = object.SubClass(0x10)
		}, calls: []string{"team", "direction", "interact"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &targetSearchTestObject{PosVec: types.Ptf(10, 0)}
			candidate := targetSearchTestCandidate4E6EA0(0, 0)
			arg := targetSearchTestArg4E6EA0(source)
			b := behavior{enemy: true, direction: 1, interact: true}
			if tc.setup != nil {
				tc.setup(candidate, arg, &b)
			}
			var calls []string
			h := targetSearchTestHooks4E6EA0(nil)
			h.sameTeam = func(*targetSearchTestObject, *targetSearchTestObject) bool {
				calls = append(calls, "team")
				return b.sameTeam
			}
			h.playerStatus = func(*targetSearchTestObject) uint32 { calls = append(calls, "player"); return b.playerStatus }
			h.isEnemy = func(*targetSearchTestObject, *targetSearchTestObject) bool {
				calls = append(calls, "enemy")
				return b.enemy
			}
			h.direction = func(types.Pointf, int16, types.Pointf) uint32 { calls = append(calls, "direction"); return b.direction }
			h.canInteract = func(*targetSearchTestObject, *targetSearchTestObject, int) bool {
				calls = append(calls, "interact")
				return b.interact
			}
			best := float32(200)
			var found *targetSearchTestObject
			targetSearchCandidate4E6EF0(candidate, arg, &best, &found, h)
			if got := found == candidate; got != tc.selectCandidate {
				t.Fatalf("selected = %v, want %v", got, tc.selectCandidate)
			}
			if !reflect.DeepEqual(calls, tc.calls) {
				t.Fatalf("calls = %v, want %v", calls, tc.calls)
			}
		})
	}
}
