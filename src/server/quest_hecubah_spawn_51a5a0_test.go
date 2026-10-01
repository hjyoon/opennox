package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func questHecubahTestHooks51A5A0(f *questNecroTest51A7A0) questHecubahSpawnHooks51A5A0[*Object, *MonsterDef, *MonsterUpdateData, *HealthData, *ObjectType, types.Pointf] {
	h := questHecubahSpawnHooks51A5A0[*Object, *MonsterDef, *MonsterUpdateData, *HealthData, *ObjectType, types.Pointf]{
		questNecroSpawnHooks51A7A0: f.hooks(),
		balanceFloat:               func(key string) float32 { f.record("balance:" + key); return 999 },
	}
	h.newObject = func(name string) *Object {
		f.record("new:" + name)
		if name == "Hecubah" {
			return f.unit
		}
		if name == "RewardMarker" {
			return f.marker
		}
		panic(name)
	}
	return h
}

func TestQuestHecubahSpawn51A5A0OrderAndEveryFaultPrefix(t *testing.T) {
	want := []string{
		"new:Hecubah", "scale", "update", "definition", "quest-health", "hp:37",
		"health", "max:38", "health", "current", "health", "maximum",
		"ai:411:10000000", "ai:423:10000000", "ai:340:00000004", "ai:326:3f547ae1",
		"ai:510:00000003", "ai:410:08000000", "ai:444:20000000", "ai:388:40000000", "ai:415:40000000",
		"balance:HecubahQuestSkill", "ai:330:3f59999a", "position", "create:17:23", "new:RewardMarker",
		"stage", "reward:7", "inventory:0", "stage", "reward:7", "inventory:0",
		"stage", "reward:7", "inventory:0", "stage", "reward:7", "inventory:0", "free",
	}
	f := newQuestNecroTest51A7A0()
	questHecubahSpawn51A5A0(questHecubahTestHooks51A5A0(f))
	if !reflect.DeepEqual(f.trace, want) || f.unit.HealthData.Cur != 37 || f.unit.HealthData.Max != 38 {
		t.Fatalf("trace=%v health=%+v", f.trace, f.unit.HealthData)
	}
	for i := range want {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			f := newQuestNecroTest51A7A0()
			f.fault = i
			func() {
				defer func() {
					if recover() != "test fault" {
						t.Fatal("missing injected fault")
					}
				}()
				questHecubahSpawn51A5A0(questHecubahTestHooks51A5A0(f))
			}()
			if !reflect.DeepEqual(f.trace, want[:i+1]) {
				t.Fatalf("fault prefix=%v", f.trace)
			}
		})
	}
}

func TestQuestHecubahSpawn51A5A0LiveRewardStagesAndNullBranches(t *testing.T) {
	t.Run("allocation failure reads scale", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.unit = nil
		questHecubahSpawn51A5A0(questHecubahTestHooks51A5A0(f))
		if !reflect.DeepEqual(f.trace, []string{"new:Hecubah", "scale"}) {
			t.Fatal(f.trace)
		}
	})
	t.Run("marker failure skips stages", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.marker = nil
		questHecubahSpawn51A5A0(questHecubahTestHooks51A5A0(f))
		if f.trace[len(f.trace)-1] != "new:RewardMarker" {
			t.Fatal(f.trace)
		}
	})
	t.Run("four live stages wrap and nil rewards continue", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.stage = math.MaxUint32
		h := questHecubahTestHooks51A5A0(f)
		stages := []uint32{}
		h.stage = func() uint32 { s := f.stage; f.stage++; f.record("stage"); return s }
		h.activateReward = func(m *Object, stage uint32) *Object {
			stages = append(stages, stage)
			f.record(fmt.Sprint("reward:", stage))
			if m != f.marker {
				t.Fatal("wrong marker")
			}
			if len(stages)%2 == 0 {
				return nil
			}
			return f.reward
		}
		questHecubahSpawn51A5A0(h)
		count := 0
		for _, e := range f.trace {
			if e == "inventory:0" {
				count++
			}
		}
		if !reflect.DeepEqual(stages, []uint32{1, 2, 3, 4}) || count != 2 || f.trace[len(f.trace)-1] != "free" {
			t.Fatalf("stages=%v inventory=%d trace=%v", stages, count, f.trace)
		}
	})
	t.Run("nil reward still makes four attempts", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.reward = nil
		questHecubahSpawn51A5A0(questHecubahTestHooks51A5A0(f))
		if !reflect.DeepEqual(f.trace[len(f.trace)-9:], []string{"stage", "reward:7", "stage", "reward:7", "stage", "reward:7", "stage", "reward:7", "free"}) {
			t.Fatal(f.trace)
		}
	})
}

func TestQuestHecubahSpawn51A5A0HealthAndCachedUpdate(t *testing.T) {
	for _, scale := range []float64{-2, math.Inf(-1), math.NaN()} {
		f := newQuestNecroTest51A7A0()
		f.scale = scale
		questHecubahSpawn51A5A0(questHecubahTestHooks51A5A0(f))
		if f.unit.HealthData.Cur != 25 || f.unit.HealthData.Max != 25 {
			t.Fatalf("scale=%v health=%+v", scale, f.unit.HealthData)
		}
	}
	f := newQuestNecroTest51A7A0()
	f.update.MonsterDef = nil
	questHecubahSpawn51A5A0(questHecubahTestHooks51A5A0(f))
	if !reflect.DeepEqual(f.trace[4:9], []string{"type", "lookup:9", "type-health", "maximum", "hp:30"}) {
		t.Fatal(f.trace)
	}
	for _, base := range []uint32{0, math.MaxUint32} {
		f := newQuestNecroTest51A7A0()
		f.scale = 1
		f.update.MonsterDef.HealthQuest72 = base
		f.onEvent = func(e string) {
			if e == "balance:HecubahQuestSkill" {
				f.position = types.Pointf{X: 41, Y: 43}
				f.unit.UpdateData = unsafe.Pointer(new(MonsterUpdateData))
			}
		}
		questHecubahSpawn51A5A0(questHecubahTestHooks51A5A0(f))
		want := uint16(1)
		if base != 0 {
			want = 65535
		}
		if f.unit.HealthData.Cur != want || f.unit.HealthData.Max != want {
			t.Fatalf("signed base=%x health=%+v", base, f.unit.HealthData)
		}
		found := false
		for _, e := range f.trace {
			found = found || e == "create:41:43"
		}
		if !found {
			t.Fatal("lost late position read")
		}
	}
}
