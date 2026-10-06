package opennox

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func aiNotHealthyDependencyCheck546A70(t *testing.T, s *Server, unit *server.Object, shape string, level uint32, valid, wantFault bool) {
	t.Helper()
	// Reuse the actual native stack fixture, not fake pop/reset callbacks.
	// Condition 64 reads cached ResumeLevel, never these raw Args words.
	update := aiLocationDependencyStack546A70(unit, ai.DEPENDENCY_NOT_HEALTHY, shape, level^0x80000000, 0x89abcdef, 0xffc12345)
	update.ResumeLevel = math.Float32frombits(level)
	update.Stamina = 0x87
	before := *update
	s.AI.StackChanged = false
	var fault any
	func() {
		defer func() { fault = recover() }()
		(&aiData{s: s}).nox_xxx_mobActionDependency(unit)
	}()
	if (fault != nil) != wantFault {
		t.Fatalf("NOT_HEALTHY original fault=%v wantFault=%t level=%08x native unit=%p health=%p", fault, wantFault, level, unit, unit.HealthData)
	}
	want := before
	wantPopped := !wantFault && !valid && (shape == "and" || shape == "or-false")
	if wantPopped {
		want.AIStackInd = 0
		want.Field2, want.Field67, want.Field74, want.Field91 = 0, 0, 0, 0
		want.Field120_1, want.Field120_2, want.Field120_3 = 0, 0, 0
		want.Field124, want.Field137 = 702, 702
	}
	// NaN != itself in a Go struct comparison. Check its raw DWORD first,
	// then normalize only the local comparison copies, not native memory.
	got := *update
	got.ResumeLevel, want.ResumeLevel = 0, 0
	if got != want || math.Float32bits(update.ResumeLevel) != level || s.AI.StackChanged != wantPopped {
		t.Fatalf("NOT_HEALTHY original retained comparison/pop differs: index=%d/%d changed=%t/%t level=%08x native unit=%p health=%p", update.AIStackInd, want.AIStackInd, s.AI.StackChanged, wantPopped, level, unit, unit.HealthData)
	}
}

func TestAINotHealthyDependency546A70NativeOriginalFractionAndGroup(t *testing.T) {
	s, unit, target, health := aiAliveDependencyNative546A70(t)
	unit.HealthData, target.HealthData = health, nil
	unit.ObjClass = object.ClassMonster | object.ClassFood
	unit.ObjFlags = object.FlagDead | object.FlagNoUpdate | object.FlagDestroyed
	unit.PosVec = types.Pointf{X: math.Float32frombits(0x7fc12345), Y: math.Float32frombits(0xff800001)}
	// Literal GAME.EXE 00546EC4..00546F0B: max unsigned WORD first;
	// nonzero max uses unsigned current WORD / max with no binary32 spill.
	// Zero max uses 00583068 (3f800000) without reading current HP.
	// FCOMP tests C0 only: less/unordered passes; equal/greater fails.
	for _, tc := range []struct {
		name     string
		cur, max uint16
		level    uint32
		valid    bool
	}{
		{"zero-below-one", 0, 80, 0x3f800000, true},
		{"zero-equal-positive-zero", 0, 80, 0, false},
		{"zero-equal-negative-zero", 0, 80, 0x80000000, false},
		{"zero-above-negative-one", 0, 80, 0xbf800000, false},
		{"exact-half", 40, 80, 0x3f000000, false},
		{"below-half", 39, 80, 0x3f000000, true},
		{"above-half", 41, 80, 0x3f000000, false},
		{"unspilled-troll-three-tenths", 24, 80, 0x3e99999a, true},
		{"unspilled-npc-three-tenths", 45, 150, 0x3e99999a, true},
		{"unspilled-forty-nine-fiftieths", 49, 50, 0x3f7ae148, true},
		{"unspilled-one-eightieth", 1, 80, 0x3c4ccccd, true},
		{"unspilled-one-third", 1, 3, 0x3eaaaaab, true},
		{"unspilled-two-thirds", 2, 3, 0x3f2aaaab, true},
		{"unspilled-one-tenth", 1, 10, 0x3dcccccd, true},
		{"narrowed-boundary-below-exact", 76, 80, 0x3f733333, false},
		{"exact-one", 80, 80, 0x3f800000, false},
		{"unsigned-high-current", 0x8000, 0xffff, 0x3f000000, false},
		{"unsigned-high-max-below-half", 0x7fff, 0xffff, 0x3f000000, true},
		{"unsigned-high-max-below-one", 0xfffe, 0xffff, 0x3f800000, true},
		{"unsigned-high-max-exact", 1, 0x8000, 0x38000000, false},
		{"unsigned-current-over-max-equal", 0xffff, 1, 0x477fff00, false},
		{"unsigned-current-over-max-below", 0xffff, 1, 0x47800000, true},
		{"zero-max-zero-current", 0, 0, 0x3f800000, false},
		{"zero-max-unsigned-current", 0xffff, 0, 0x3f800000, false},
		{"zero-max-below-nextafter-one", 0, 0, 0x3f800001, true},
		{"zero-max-above-previous-one", 0xffff, 0, 0x3f7fffff, false},
		{"positive-infinite-level", 0xffff, 1, 0x7f800000, true},
		{"negative-infinite-level", 0, 80, 0xff800000, false},
		{"positive-subnormal-level", 0, 80, 1, true},
		{"negative-subnormal-level", 0, 80, 0x80000001, false},
		{"positive-qnan-level", 1, 80, 0x7fc12345, true},
		{"negative-qnan-level", 1, 80, 0xffc12345, true},
		{"positive-snan-level", 1, 80, 0x7f800001, true},
		{"negative-snan-level", 1, 80, 0xff800001, true},
		{"zero-max-qnan-level", 0, 0, 0x7fc12345, true},
		{"zero-max-snan-level", 0xffff, 0, 0xff800001, true},
		{"zero-max-positive-infinite-level", 0, 0, 0x7f800000, true},
		{"zero-max-negative-infinite-level", 0xffff, 0, 0xff800000, false},
	} {
		for _, shape := range aiLocationDependencyShapes546A70 {
			t.Run(tc.name+"/"+shape, func(t *testing.T) {
				*health = server.HealthData{Cur: tc.cur, Max: tc.max, Field2: 0xabcd, Field16: 0x12345678}
				beforeHealth := *health
				beforeUnit := aiObjectDistanceInputWords546B13(unit)
				aiNotHealthyDependencyCheck546A70(t, s, unit, shape, tc.level, tc.valid, false)
				if *health != beforeHealth || unit.HealthData != health || target.HealthData != nil || aiObjectDistanceInputWords546B13(unit) != beforeUnit || unit.ObjClass != object.ClassMonster|object.ClassFood || unit.ObjFlags != object.FlagDead|object.FlagNoUpdate|object.FlagDestroyed {
					t.Fatal("NOT_HEALTHY changed native HP/identity/class/flags/position")
				}
			})
		}
	}
}

func TestAINotHealthyDependency546A70NativeNilHealthFaultPrefix(t *testing.T) {
	s, unit, _, health := aiAliveDependencyNative546A70(t)
	unit.HealthData = nil
	*health = server.HealthData{Cur: 7, Max: 80, Field2: 0xabcd, Field16: 0x12345678}
	before := *health
	for _, shape := range aiLocationDependencyShapes546A70 {
		t.Run(shape, func(t *testing.T) {
			// Missing health faults at the max read even if an OR member
			// already passed, before reading the NaN level or popping.
			aiNotHealthyDependencyCheck546A70(t, s, unit, shape, 0x7fc12345, false, true)
			if *health != before || unit.HealthData != nil {
				t.Fatal("missing-health prefix changed native HP/identity")
			}
		})
	}
}
