package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterMainThreatFlee547210OriginalGatesAndBlinkFallbacks(t *testing.T) {
	const reject, flee, blink = 0, 1, 2
	tests := []struct {
		name     string
		setup    func(*Server, *Object, *MonsterUpdateData)
		distance float64
		want     int
	}{
		{"blink", nil, 20, blink},
		{"blink-at-unsigned-deadline", func(s *Server, _ *Object, u *MonsterUpdateData) { s.SetFrame(0x80000001); u.Field371 = 0x80000001 }, 20, blink},
		{"unsigned-deadline-still-future", func(s *Server, _ *Object, u *MonsterUpdateData) { s.SetFrame(0x7fffffff); u.Field371 = 0x80000000 }, 20, flee},
		{"no-cast-ability", func(_ *Server, _ *Object, u *MonsterUpdateData) { u.StatusFlags = 0 }, 20, flee},
		{"empty-blink-slot", func(_ *Server, _ *Object, u *MonsterUpdateData) { u.Field376 = 0 }, 20, flee},
		{"anti-magic-falls-back", func(_ *Server, o *Object, _ *MonsterUpdateData) { o.Buffs = 1 << ENCHANT_ANTI_MAGIC }, 20, flee},
		{"half-range-equality", nil, 32.5, flee},
		{"half-range-float32-spill", nil, 32.4999999, flee},
		{"half-range-next-binary32-below", nil, float64(math.Nextafter32(32.5, 0)), blink},
		{"outer-retains-unrounded-distance", nil, 64.9999999, flee},
		{"outer-boundary", nil, 65, reject},
		{"outer-above-boundary", nil, 65.0000001, reject},
		{"outer-infinity", nil, math.Inf(1), reject},
		{"unordered-distance-falls-back", nil, math.NaN(), flee},
		{"unordered-range", func(_ *Server, _ *Object, u *MonsterUpdateData) { u.FleeRange = float32(math.NaN()) }, 20, reject},
		{"zero-range", func(_ *Server, _ *Object, u *MonsterUpdateData) { u.FleeRange = 0; u.StatusFlags = 0 }, -1, reject},
		{"negative-zero-range", func(_ *Server, _ *Object, u *MonsterUpdateData) {
			u.FleeRange = math.Float32frombits(0x80000000)
			u.StatusFlags = 0
		}, -1, reject},
		{"negative-range-is-not-positivity-gated", func(_ *Server, _ *Object, u *MonsterUpdateData) { u.FleeRange = -1; u.StatusFlags = 0 }, -2, flee},
		{"passive", func(_ *Server, _ *Object, u *MonsterUpdateData) { u.Aggression = 0 }, 20, reject},
		{"passive-equality", func(_ *Server, _ *Object, u *MonsterUpdateData) {
			u.Aggression = monsterMainPassiveAggressionLimit547210
		}, 20, blink},
		{"unordered-aggression", func(_ *Server, _ *Object, u *MonsterUpdateData) { u.Aggression = float32(math.NaN()) }, 20, reject},
		{"immobile", func(_ *Server, o *Object, _ *MonsterUpdateData) { o.SpeedBase = 0 }, 20, reject},
		{"movement-equality", func(_ *Server, o *Object, _ *MonsterUpdateData) { o.SpeedBase = float32(0.0099999998) }, 20, blink},
		{"unordered-speed", func(_ *Server, o *Object, _ *MonsterUpdateData) { o.SpeedBase = float32(math.NaN()) }, 20, reject},
		{"cast-object-head", func(_ *Server, _ *Object, u *MonsterUpdateData) {
			u.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
		}, 20, reject},
		{"cast-location-head", func(_ *Server, _ *Object, u *MonsterUpdateData) {
			u.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_LOCATION)
		}, 20, reject},
		{"cast-duration-head", func(_ *Server, _ *Object, u *MonsterUpdateData) {
			u.AIStack[0].Action = uint32(ai.ACTION_CAST_DURATION_SPELL)
		}, 20, reject},
		{"cast-below-head-does-not-block", func(_ *Server, _ *Object, u *MonsterUpdateData) {
			u.AIStackInd = 1
			u.AIStack[0].Action = uint32(ai.ACTION_CAST_DURATION_SPELL)
			u.AIStack[1].Action = uint32(ai.ACTION_WAIT)
		}, 20, blink},
		{"confused", func(_ *Server, o *Object, _ *MonsterUpdateData) { o.Buffs = 1 << ENCHANT_CONFUSED }, 20, reject},
		{"move-attempt-recent", func(s *Server, _ *Object, u *MonsterUpdateData) { u.Field127 = s.Frame() - 1 }, 20, reject},
		{"move-attempt-expiry", func(s *Server, _ *Object, u *MonsterUpdateData) { u.Field127 = s.Frame() - 3*s.TickRate() }, 20, blink},
		{"no-enemy", func(_ *Server, _ *Object, u *MonsterUpdateData) { u.CurrentEnemy = nil }, 20, reject},
		{"existing-flee-blocks-only-fallback", func(_ *Server, _ *Object, u *MonsterUpdateData) {
			u.AIStack[0].Action = uint32(ai.ACTION_FLEE)
			u.Field376 = 0
		}, 20, reject},
		{"enabled-not-a-branch-gate", func(_ *Server, o *Object, _ *MonsterUpdateData) { o.ObjFlags &^= object.FlagEnabled }, 20, blink},
		{"destroyed-not-a-branch-gate", func(_ *Server, o *Object, _ *MonsterUpdateData) { o.ObjFlags |= object.FlagDestroyed }, 20, blink},
		{"other-buff-not-a-branch-gate", func(_ *Server, o *Object, _ *MonsterUpdateData) { o.Buffs = 1 << ENCHANT_SLOWED }, 20, blink},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, update, _ := newMonsterMainFleeTest547210(t)
			update.StatusFlags, update.Field376 = object.MonStatusCanCastSpells, 0x80000000
			update.Field371, update.Field370_0, update.Field370_2 = s.Frame(), 5, 5
			if tc.setup != nil {
				tc.setup(s, unit, update)
			}
			stack, index, deadline := update.AIStack, update.AIStackInd, update.Field371
			casts, rolls, distanceCalls := 0, 0, 0
			handled := s.monsterMainThreatFlee547210(unit, update, nil, MonsterMainRuntime547210{
				Distance: func(got, target *Object) float64 {
					distanceCalls++
					if got != unit || target != update.CurrentEnemy {
						t.Fatal("distance lost native objects")
					}
					return tc.distance
				},
				CastSpell: func(id int32, got *Object, arg *SpellAcceptArg) {
					casts++
					if id != 4 || got != unit || arg.Obj != unit || arg.Pos != unit.PosVec {
						t.Fatal("Blink arguments changed")
					}
				},
				RandomInt: func(min, max int) int {
					rolls++
					if tc.want == blink {
						if casts != 1 || min != 5 || max != 5 {
							t.Fatal("Blink RNG changed")
						}
						return 5
					}
					if casts != 0 || min != 0 || max != 1 {
						t.Fatal("FLEE RNG changed")
					}
					return 0
				},
			})
			if handled != (tc.want != reject) {
				t.Fatalf("handled = %t, want disposition %d", handled, tc.want)
			}
			switch tc.want {
			case reject:
				if casts != 0 || rolls != 0 || update.AIStack != stack || update.AIStackInd != index || update.Field371 != deadline {
					t.Fatal("rejected branch changed observable state")
				}
			case blink:
				if casts != 1 || rolls != 1 || distanceCalls != 1 || update.AIStack != stack || update.AIStackInd != index || update.Field371 != s.Frame()+5 {
					t.Fatal("immediate Blink/cooldown contract changed")
				}
			case flee:
				if casts != 0 || rolls != 1 || distanceCalls != 1 || update.AIStackInd != index+4 || update.AIStackHead().Type() != ai.ACTION_FLEE || update.Field371 != deadline {
					t.Fatal("ordinary FLEE fallback contract changed")
				}
			}
		})
	}
}

func TestMonsterMainThreatFlee547210ChargesCachedCooldownAfterCastAndRandom(t *testing.T) {
	s, unit, update, _ := newMonsterMainFleeTest547210(t)
	update.StatusFlags, update.Field376 = object.MonStatusCanCastSpells, 1
	update.Field370_0, update.Field370_2 = 1, 2
	replacement := new(MonsterUpdateData)
	replacement.Field371 = 0x12345678
	calls := 0
	if !s.monsterMainThreatFlee547210(unit, update, nil, MonsterMainRuntime547210{
		CastSpell: func(_ int32, got *Object, _ *SpellAcceptArg) {
			calls++
			if got != unit {
				t.Fatal("cast target changed")
			}
			update.Field370_0, update.Field370_2 = 65000, 65535
			s.SetFrame(400)
			unit.UpdateData = unsafe.Pointer(replacement)
		},
		RandomInt: func(min, max int) int {
			if calls != 1 || min != 65000 || max != 65535 {
				t.Fatal("cooldown bounds were cached before casting")
			}
			s.SetFrame(0xfffffffc)
			return 9
		},
	}) || calls != 1 || update.Field371 != 5 || replacement.Field371 != 0x12345678 {
		t.Fatal("Blink lost cached update, unsigned bounds or post-RNG wrapping frame")
	}
}

func TestMonsterMainThreatFlee547210ReadsAfterPushAndRetainsEntrySoundPointer(t *testing.T) {
	s, unit, update, _ := newMonsterMainFleeTest547210(t)
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT), Field5: 1}
	originalSound, replacementSound := [13]uint32{12: 100}, [13]uint32{12: 200}
	update.SoundSet122 = unsafe.Pointer(&originalSound[0])
	oldAction, existed := aiActions[ai.ACTION_WAIT]
	t.Cleanup(func() {
		if existed {
			aiActions[ai.ACTION_WAIT] = oldAction
		} else {
			delete(aiActions, ai.ACTION_WAIT)
		}
	})
	cancelCalls := 0
	enemy := &Object{PosVec: types.Ptf(math.Float32frombits(0x80000000), math.Float32frombits(0x7fc12345))}
	aiActions[ai.ACTION_WAIT] = monsterMainFearCancel547210{cancel: func(got *Object) {
		cancelCalls++
		unit.Direction1 = 0x8000
		update.FleeRange, update.CurrentEnemy = 95, enemy
		update.SoundSet122 = unsafe.Pointer(&replacementSound[0])
	}}
	audioCalls := 0
	if !s.monsterMainThreatFlee547210(unit, update, unsafe.Pointer(&originalSound[0]), MonsterMainRuntime547210{
		RandomInt: func(min, max int) int {
			if min != 0 || max != 1 || cancelCalls != 1 {
				t.Fatal("ordinary flee RNG ordering changed")
			}
			originalSound[12] = 300
			return 1
		},
		AudioEvent: func(id uint32, got *Object) {
			audioCalls++
			if id != 300 || got != unit {
				t.Fatal("sound did not use entry pointer and post-RNG value")
			}
		},
	}) || cancelCalls != 1 || audioCalls != 1 {
		t.Fatal("ordinary flee failed")
	}
	if update.AIStack[1].Args[0] != 0xffff8080 || update.AIStack[3].ArgF32(0) != 125 ||
		update.AIStack[4].Args != [4]uintptr{0x80000000, 0x7fc12345, 0, 0} {
		t.Fatalf("post-push direction/range/enemy bits changed: %+v", update.AIStack[:5])
	}
}

func TestMonsterMainThreatFlee547210PartialAndRejectedPushesStillRandomize(t *testing.T) {
	for _, initial := range []int8{20, 21, 22, 23} {
		s, unit, update, _ := newMonsterMainFleeTest547210(t)
		update.AIStackInd = initial
		for i := range update.AIStack {
			update.AIStack[i] = AIStackItem{Action: uint32(ai.ACTION_WAIT)}
		}
		before := update.AIStack
		rolls, sounds := 0, 0
		soundSet := [13]uint32{12: 123}
		if !s.monsterMainThreatFlee547210(unit, update, unsafe.Pointer(&soundSet[0]), MonsterMainRuntime547210{
			RandomInt: func(int, int) int { rolls++; return -1 },
			AudioEvent: func(id uint32, got *Object) {
				sounds++
				if id != 123 || got != unit {
					t.Fatal("rejected push sound changed")
				}
			},
		}) || update.AIStackInd != 23 || rolls != 1 || sounds != 1 {
			t.Fatalf("stack %d did not preserve partial-push/sound return", initial)
		}
		for i := 0; i <= int(initial); i++ {
			if update.AIStack[i] != before[i] {
				t.Fatal("partial pushes changed an earlier stack item")
			}
		}
		want := []ai.ActionType{ai.ACTION_SET_ANGLE, ai.DEPENDENCY_NOT_CORNERED, ai.DEPENDENCY_ENEMY_CLOSER_THAN, ai.ACTION_FLEE}
		for i := int(initial) + 1; i < len(update.AIStack); i++ {
			if update.AIStack[i].Type() != want[i-int(initial)-1] {
				t.Fatal("partial pushes changed action order")
			}
		}
	}
}

func TestMonsterMainThreatFlee547210UsesCollisionSurfaces(t *testing.T) {
	for _, shape := range []ShapeKind{ShapeKindCircle, ShapeKindBox} {
		s, unit, update, enemy := newMonsterMainFleeTest547210(t)
		enemy.PosVec = types.Ptf(200, 100) // centers exceed range, surfaces do not
		unit.Shape.Kind, unit.Shape.Circle.R = shape, 40
		unit.Shape.Box = ShapeBox{W: 80, H: 10}
		if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{RandomInt: func(int, int) int { return 0 }}) ||
			update.AIStackInd != 4 || update.AIStackHead().Type() != ai.ACTION_FLEE {
			t.Fatal("MainAI still compared center-distance instead of collision surfaces")
		}
	}
}

func TestMonsterMainThreatFlee547210MissingImmediateServiceDoesNotScheduleFlee(t *testing.T) {
	s, unit, update, _ := newMonsterMainFleeTest547210(t)
	update.StatusFlags, update.Field376 = object.MonStatusCanCastSpells, 1
	stack, deadline := update.AIStack, update.Field371
	if s.monsterMainThreatFlee547210(unit, update, nil, MonsterMainRuntime547210{RandomInt: func(int, int) int { return 5 }}) ||
		update.AIStack != stack || update.Field371 != deadline {
		t.Fatal("unavailable immediate Blink was replaced with a different action")
	}
}
