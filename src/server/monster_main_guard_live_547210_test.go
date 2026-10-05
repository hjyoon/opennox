package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterMainGuard547210UsesLivePredicatesButCachedStimulusRecord(t *testing.T) {
	for _, tc := range []struct {
		name                string
		liveAggression      float32
		liveGuard, liveHunt bool
		wantQuery           bool
	}{
		{"replacement-guard", 1, true, false, true},
		{"replacement-passive", 0, true, false, false},
		{"replacement-hunt", 1, true, true, false},
		{"replacement-no-guard", 1, false, false, false},
		{"replacement-head-guard", 1, true, false, false},
		{"replacement-unordered-aggression", float32(math.NaN()), true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, cached := monsterScriptHitFixture515A30(t)
			s.SetFrame(208)
			cached.Aggression = 1
			cached.CurrentEnemy = &Object{}
			if !tc.wantQuery {
				cached.AIStackInd = 1
				cached.AIStack[0].Action = uint32(ai.ACTION_GUARD)
				cached.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			}
			live := new(MonsterUpdateData)
			live.Aggression = tc.liveAggression
			live.AIStack[0].Action = uint32(ai.ACTION_WAIT)
			if tc.liveGuard {
				live.AIStackInd++
				live.AIStack[live.AIStackInd].Action = uint32(ai.ACTION_GUARD)
			}
			if tc.liveHunt {
				live.AIStackInd++
				live.AIStack[live.AIStackInd].Action = uint32(ai.ACTION_HUNT)
			}
			live.AIStackInd++
			live.AIStack[live.AIStackInd].Action = uint32(ai.ACTION_WAIT)
			if tc.name == "replacement-head-guard" {
				live.AIStack[live.AIStackInd].Action = uint32(ai.ACTION_GUARD)
			}
			unit.UpdateData = unsafe.Pointer(live)
			before := *live
			calls := 0
			s.monsterMainGuardEnemyStimulus547210(unit, cached, MonsterMainRuntime547210{
				EnemyAggro: func(got *Object, radius float32) *Object {
					calls++
					if got != unit || radius != 100 {
						t.Fatal("GUARD query arguments changed")
					}
					s.SetFrame(209)
					return &Object{ObjClass: object.ClassPlayer}
				},
			}, ai.ACTION_WAIT)
			if (calls == 1) != tc.wantQuery || calls > 1 {
				t.Fatalf("query count = %d, want query %t", calls, tc.wantQuery)
			}
			after := *live
			unchangedAggression := math.Float32bits(after.Aggression) == math.Float32bits(before.Aggression)
			after.Aggression, before.Aggression = 0, 0 // NaN is unequal to itself.
			if !unchangedAggression || after != before {
				t.Fatal("GUARD stimulus wrote the replacement update record")
			}
			if tc.wantQuery && (!cached.StatusFlags.Has(object.MonStatusInjured) ||
				unit.Obj130 != cached.CurrentEnemy || unit.Field131 != 11 || unit.Frame134 != 209) {
				t.Fatal("GUARD did not retain the cached record and post-query frame")
			}
		})
	}
}
