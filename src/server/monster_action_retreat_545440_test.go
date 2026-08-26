package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func retreatMonsterTestObject545440(t *testing.T) *Object {
	t.Helper()
	unit := passiveMonsterTestObject547210(t)
	unit.PosVec = types.Ptf(100, 200)
	unit.HealthData = &HealthData{Cur: 10, Max: 100}
	unit.UpdateDataMonster().ResumeLevel = 0.5
	return unit
}

func retreatHooks545440(t *testing.T, unit *Object, events *[]ai.ActionType) monsterActionRetreatHooks545440 {
	t.Helper()
	return monsterActionRetreatHooks545440{
		frame:    func() uint32 { return 1000 },
		tickRate: func() uint32 { return 30 },
		random: func(minimum, maximum int) int {
			if minimum != 120 || maximum != 180 {
				t.Fatalf("random bounds = %d..%d, want 120..180", minimum, maximum)
			}
			return 151
		},
		castRelated: func(*Object) bool { return false },
		searchFood:  func(*Object, float32) *Object { return nil },
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			*events = append(*events, action)
			item := &AIStackItem{Action: uint32(action)}
			item.SetArgs(args...)
			return item
		},
		pop: func() int {
			*events = append(*events, ai.ACTION_INVALID)
			return 0
		},
	}
}

func TestMonsterCanResumeAttack545520(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	if monsterCanResumeAttack545520(unit) {
		t.Fatal("10 percent health resumed at a 50 percent threshold")
	}
	unit.HealthData.Cur = 50
	if !monsterCanResumeAttack545520(unit) {
		t.Fatal("threshold equality did not resume")
	}
	unit.HealthData.Max = 0
	unit.UpdateDataMonster().ResumeLevel = 1
	if !monsterCanResumeAttack545520(unit) {
		t.Fatal("zero maximum did not use the original 1.0 ratio")
	}
}

func TestMonsterActionRetreat545440RecoveredPops(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	unit.HealthData.Cur = 75
	var events []ai.ActionType
	hooks := retreatHooks545440(t, unit, &events)
	monsterActionRetreat545440(unit, hooks)
	if len(events) != 1 || events[0] != ai.ACTION_INVALID {
		t.Fatalf("events = %v, want pop", events)
	}
}

func TestMonsterActionRetreat545440EnemyFlees(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	enemy := &Object{PosVec: types.Ptf(300, 400)}
	unit.UpdateDataMonster().CurrentEnemy = enemy
	var events []ai.ActionType
	var pushed []*AIStackItem
	hooks := retreatHooks545440(t, unit, &events)
	hooks.push = func(action ai.ActionType, args ...any) *AIStackItem {
		events = append(events, action)
		item := &AIStackItem{Action: uint32(action)}
		item.SetArgs(args...)
		pushed = append(pushed, item)
		return item
	}
	monsterActionRetreat545440(unit, hooks)
	if len(events) != 2 || events[0] != ai.DEPENDENCY_TIME || events[1] != ai.ACTION_FLEE {
		t.Fatalf("events = %v", events)
	}
	if pushed[0].ArgU32(0) != 1151 || pushed[1].ArgPos(0) != enemy.PosVec || pushed[1].ArgU32(2) != 0 {
		t.Fatalf("flee args = %#v %#v", pushed[0].Args, pushed[1].Args)
	}
}

func TestMonsterActionRetreat545440FoodStack(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	food := &Object{PosVec: types.Ptf(123, 456)}
	var events []ai.ActionType
	var pushed []*AIStackItem
	hooks := retreatHooks545440(t, unit, &events)
	hooks.quest = true
	hooks.searchFood = func(got *Object, radius float32) *Object {
		if got != unit || radius != 640 {
			t.Fatalf("food search = %p/%g, want unit/640", got, radius)
		}
		return food
	}
	hooks.push = func(action ai.ActionType, args ...any) *AIStackItem {
		events = append(events, action)
		item := &AIStackItem{Action: uint32(action)}
		item.SetArgs(args...)
		pushed = append(pushed, item)
		return item
	}
	monsterActionRetreat545440(unit, hooks)
	want := []ai.ActionType{ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO}
	if len(events) != len(want) {
		t.Fatalf("events = %v", events)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
	if pushed[2].ArgPos(0) != food.PosVec || pushed[2].ArgObj(2) != food || pushed[3].ArgObj(0) != food || pushed[4].ArgObj(2) != food {
		t.Fatal("food pointer or position was not preserved in the action stack")
	}
}

func TestMonsterActionRetreat545440RoamStack(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	var events []ai.ActionType
	var roam *AIStackItem
	hooks := retreatHooks545440(t, unit, &events)
	hooks.push = func(action ai.ActionType, args ...any) *AIStackItem {
		events = append(events, action)
		item := &AIStackItem{Action: uint32(action)}
		item.SetArgs(args...)
		if action == ai.ACTION_ROAM {
			roam = item
		}
		return item
	}
	monsterActionRetreat545440(unit, hooks)
	want := []ai.ActionType{ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_NO_VISIBLE_FOOD, ai.ACTION_ROAM}
	for i := range want {
		if len(events) != len(want) || events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
	if roam == nil || roam.ArgU32(0) != 0 || roam.ArgU32(1) != 0 || roam.ArgU32(2) != uint32(0xffffff80) {
		t.Fatalf("roam args = %#v", roam)
	}
}

func TestMonsterSearchEdible544A00FiltersAndChoosesNearest(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	unit.ObjSubClass = object.SubClass(object.MonsterNPC)
	unit.Poison540 = 0
	blocked := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodSimple), PosVec: types.Ptf(101, 200)}
	jug := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodJug), PosVec: types.Ptf(102, 200)}
	mushroom := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodMushroom), PosVec: types.Ptf(103, 200)}
	ordinaryPotion := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodPotion | object.FoodManaPotion), PosVec: types.Ptf(104, 200)}
	healthPotion := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodPotion | object.FoodHealthPotion), PosVec: types.Ptf(120, 200)}
	appleFar := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodApple), PosVec: types.Ptf(140, 200)}
	appleNear := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodApple), PosVec: types.Ptf(130, 200)}
	items := []*Object{blocked, jug, mushroom, ordinaryPotion, healthPotion, appleFar, appleNear}
	hooks := monsterSearchEdibleHooks544A00{
		eachInCircle: func(pos types.Pointf, radius float32, fn func(*Object) bool) {
			if pos != unit.PosVec || radius != 640 {
				t.Fatalf("circle = %v/%g", pos, radius)
			}
			for _, item := range items {
				fn(item)
			}
		},
		canInteract: func(_ *Object, candidate *Object, flags int) bool {
			return flags == 0 && candidate != blocked
		},
		online: true,
	}
	if got := monsterSearchEdible544A00(unit, 640, hooks); got != healthPotion {
		t.Fatalf("nearest edible = %p, want health potion %p", got, healthPotion)
	}
	hooks.online = false
	if got := monsterSearchEdible544A00(unit, 640, hooks); got != appleNear {
		t.Fatalf("offline nearest edible = %p, want apple %p", got, appleNear)
	}
	unit.Poison540 = 1
	if got := monsterSearchEdible544A00(unit, 640, hooks); got != mushroom {
		t.Fatalf("poisoned nearest edible = %p, want mushroom %p", got, mushroom)
	}
}
