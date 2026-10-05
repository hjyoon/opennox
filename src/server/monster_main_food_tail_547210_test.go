package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterMainFoodTail547210OriginalGates(t *testing.T) {
	for _, mode := range []string{"active", "aggression-equality", "passive", "NaN-aggression", "no-health", "no-max", "full-health", "above-max", "not-due", "frustrated", "flee", "retreat", "master"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, cached := monsterMainHealthRetreatFixture547210(t)
			s.SetFrame(160)
			want := true
			switch mode {
			case "aggression-equality":
				cached.Aggression = monsterMainPassiveAggressionLimit547210
			case "passive":
				cached.Aggression = math.Nextafter32(monsterMainPassiveAggressionLimit547210, 0)
				want = false
			case "NaN-aggression":
				cached.Aggression = float32(math.NaN())
				want = false
			case "no-health":
				unit.HealthData = nil
				want = false
			case "no-max":
				unit.HealthData.Max = 0
				want = false
			case "full-health":
				unit.HealthData.Cur = unit.HealthData.Max
				want = false
			case "above-max":
				unit.HealthData.Cur = unit.HealthData.Max + 1
				want = false
			case "not-due":
				s.SetFrame(161)
				want = false
			case "frustrated":
				cached.StatusFlags |= object.MonStatusFrustrated
			case "flee", "retreat", "master":
				cached.AIStack[0].Action = uint32(map[string]ai.ActionType{"flee": ai.ACTION_FLEE, "retreat": ai.ACTION_RETREAT, "master": ai.ACTION_RETREAT_TO_MASTER}[mode])
			}
			searches := 0
			if !s.monsterMainEatNearbyFood547210(unit, cached, MonsterMainRuntime547210{
				SearchEdible: func(got *Object, radius float32) *Object {
					if got != unit || radius != 75 {
						t.Fatal("food search arguments changed")
					}
					searches++
					return nil
				},
				PlaceInventory: func(*Object, *Object, int, int) bool { t.Fatal("nil food placed"); return false },
			}) || (searches == 1) != want {
				t.Fatalf("original food gates: searches=%d, want search=%t", searches, want)
			}
		})
	}
}

func TestMonsterMainFoodTail547210LiveAggression(t *testing.T) {
	for _, active := range []bool{false, true} {
		t.Run(map[bool]string{false: "cached-active/live-passive", true: "cached-passive/live-active"}[active], func(t *testing.T) {
			s, unit, cached := monsterMainHealthRetreatFixture547210(t)
			s.SetFrame(160)
			live := &MonsterUpdateData{Aggression: 0}
			if active {
				cached.Aggression = 0
				live.Aggression = 0.83
			}
			unit.UpdateData = unsafe.Pointer(live)
			searches := 0
			if !s.monsterMainEatNearbyFood547210(unit, cached, MonsterMainRuntime547210{
				SearchEdible:   func(*Object, float32) *Object { searches++; return nil },
				PlaceInventory: func(*Object, *Object, int, int) bool { t.Fatal("nil food placed"); return false },
			}) || (searches == 1) != active {
				t.Fatal("534440 must reload live aggression")
			}
		})
	}
}

func TestMonsterMainFoodTail547210PostInventorySubclassAndNoRecheck(t *testing.T) {
	for _, tc := range []struct {
		name          string
		before, after object.SubClass
		use           bool
	}{
		{"apple-becomes-mushroom", object.SubClass(object.FoodApple), object.SubClass(object.FoodMushroom), true},
		{"mushroom-becomes-apple", object.SubClass(object.FoodMushroom), object.SubClass(object.FoodApple), false},
		{"health-potion-with-upper-bits", 0, 0x10010, true},
		{"upper-bits-only", 0x80, 0x10000, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, cached := monsterMainHealthRetreatFixture547210(t)
			s.SetFrame(160)
			food := &Object{ObjSubClass: tc.before}
			events := []string{}
			if !s.monsterMainEatNearbyFood547210(unit, cached, MonsterMainRuntime547210{
				SearchEdible: func(got *Object, radius float32) *Object {
					if got != unit || radius != 75 {
						t.Fatal("food search arguments")
					}
					events = append(events, "search")
					unit.UpdateData = unsafe.Pointer(&MonsterUpdateData{Aggression: 0})
					unit.HealthData.Max = 0
					s.SetFrame(161)
					return food
				},
				PlaceInventory: func(owner, item *Object, a, b int) bool {
					if owner != unit || item != food || a != 1 || b != 1 {
						t.Fatal("food placement arguments")
					}
					events = append(events, "place")
					food.ObjSubClass = tc.after
					return false // original does not gate consumption on placement result
				},
				UseByNetCode: func(owner, item *Object) int32 {
					if owner != unit || item != food {
						t.Fatal("food use arguments")
					}
					events = append(events, "use")
					return 0
				},
			}) {
				t.Fatal("food tail rejected")
			}
			want := []string{"search", "place"}
			if tc.use {
				want = append(want, "use")
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("callbacks=%v, want %v", events, want)
			}
		})
	}
}
