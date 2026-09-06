package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func confusedTestObject545140(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_CONFUSED)
	update.MonsterDef = &MonsterDef{}
	unit.Direction1 = 37
	unit.PosVec = types.Ptf(123.456, -78.25)
	return unit
}

func confusedHooks545140(t *testing.T) monsterActionConfusedHooks545140 {
	t.Helper()
	return monsterActionConfusedHooks545140{
		random: func(int, int) int {
			t.Fatal("unexpected RNG call")
			return 0
		},
		randomWalk: func(*Object) bool {
			t.Fatal("unexpected random-walk call")
			return false
		},
		canMelee: func(*Object) bool {
			t.Fatal("unexpected melee capability call")
			return false
		},
		canShoot: func(*Object) bool {
			t.Fatal("unexpected missile capability call")
			return false
		},
		push: func(*Object, ai.ActionType) *AIStackItem {
			t.Fatal("unexpected action push")
			return nil
		},
	}
}

func TestMonsterActionConfused545140RandomWalkBoundary(t *testing.T) {
	unit := confusedTestObject545140(t)
	hooks := confusedHooks545140(t)
	events := make([]string, 0, 2)
	hooks.random = func(min, max int) int {
		if min != 0 || max != 100 {
			t.Fatalf("random bounds = %d..%d, want 0..100", min, max)
		}
		events = append(events, "random")
		return 15
	}
	hooks.randomWalk = func(got *Object) bool {
		if got != unit {
			t.Fatalf("random-walk unit = %p, want %p", got, unit)
		}
		events = append(events, "walk")
		return false
	}
	if !monsterActionConfused545140(unit, hooks) {
		t.Fatal("confused action was not handled")
	}
	if want := []string{"random", "walk"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestMonsterActionConfused545140AttackSelectionOrder(t *testing.T) {
	tests := []struct {
		name       string
		melee      bool
		shoot      bool
		secondRand int
		wantAction ai.ActionType
		wantEvents []string
	}{
		{name: "melee-only", melee: true, wantAction: ai.ACTION_MELEE_ATTACK, wantEvents: []string{"random", "melee", "shoot", "push:16"}},
		{name: "both-melee", melee: true, shoot: true, secondRand: 49, wantAction: ai.ACTION_MELEE_ATTACK, wantEvents: []string{"random", "melee", "shoot", "random", "push:16"}},
		{name: "both-missile-boundary", melee: true, shoot: true, secondRand: 50, wantAction: ai.ACTION_MISSILE_ATTACK, wantEvents: []string{"random", "melee", "shoot", "random", "push:17"}},
		{name: "missile-only", shoot: true, wantAction: ai.ACTION_MISSILE_ATTACK, wantEvents: []string{"random", "melee", "shoot", "push:17"}},
		{name: "unarmed", wantAction: ai.ACTION_INVALID, wantEvents: []string{"random", "melee", "shoot"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := confusedTestObject545140(t)
			item := &AIStackItem{Args: [4]uintptr{0x111, 0x222, ^uintptr(0), 0x444}}
			events := make([]string, 0, 5)
			randomCalls := 0
			hooks := confusedHooks545140(t)
			hooks.random = func(min, max int) int {
				events = append(events, "random")
				randomCalls++
				if min != 0 || max != 100 {
					t.Fatalf("random call %d bounds = %d..%d", randomCalls, min, max)
				}
				if randomCalls == 1 {
					return 14
				}
				return tc.secondRand
			}
			hooks.canMelee = func(got *Object) bool {
				if got != unit {
					t.Fatalf("melee unit = %p, want %p", got, unit)
				}
				events = append(events, "melee")
				return tc.melee
			}
			hooks.canShoot = func(got *Object) bool {
				if got != unit {
					t.Fatalf("shoot unit = %p, want %p", got, unit)
				}
				events = append(events, "shoot")
				return tc.shoot
			}
			hooks.push = func(got *Object, action ai.ActionType) *AIStackItem {
				if got != unit {
					t.Fatalf("push unit = %p, want %p", got, unit)
				}
				events = append(events, fmt.Sprintf("push:%d", action))
				if action != tc.wantAction {
					t.Fatalf("pushed action = %s, want %s", action, tc.wantAction)
				}
				return item
			}

			if !monsterActionConfused545140(unit, hooks) {
				t.Fatal("confused action was not handled")
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}

			if tc.wantAction == ai.ACTION_MISSILE_ATTACK {
				cosine, _ := SinCosDir(byte(unit.Direction1))
				_, sine := SinCosDir(byte(unit.Direction1))
				wantX := float32(float64(cosine)*10 + float64(unit.PosVec.X))
				wantY := float32(float64(sine)*10 + float64(unit.PosVec.Y))
				want := [4]uintptr{
					uintptr(math.Float32bits(wantX)),
					uintptr(math.Float32bits(wantY)),
					0,
					0x444,
				}
				if item.Args != want {
					t.Fatalf("missile args = %#v, want %#v", item.Args, want)
				}
			} else if item.Args != [4]uintptr{0x111, 0x222, ^uintptr(0), 0x444} {
				t.Fatalf("non-missile args changed: %#v", item.Args)
			}
		})
	}
}

func TestMonsterActionConfused545140MissilePushFailure(t *testing.T) {
	unit := confusedTestObject545140(t)
	hooks := confusedHooks545140(t)
	hooks.random = func(int, int) int { return 14 }
	hooks.canMelee = func(*Object) bool { return false }
	hooks.canShoot = func(*Object) bool { return true }
	hooks.push = func(got *Object, action ai.ActionType) *AIStackItem {
		if got != unit || action != ai.ACTION_MISSILE_ATTACK {
			t.Fatalf("push = (%p, %s)", got, action)
		}
		return nil
	}
	before := *unit
	if !monsterActionConfused545140(unit, hooks) {
		t.Fatal("failed missile push was not handled")
	}
	if *unit != before {
		t.Fatal("failed missile push changed unit state")
	}
}

func TestMonsterActionConfused545140UsesRandomWalkBody(t *testing.T) {
	unit := confusedTestObject545140(t)
	unit.SpeedCur = 2.5
	unit.Direction1 = 250
	unit.Direction2 = 1
	hooks := confusedHooks545140(t)
	hooks.random = func(int, int) int { return 15 }
	hooks.randomWalk = func(got *Object) bool {
		return monsterActionRandomWalkMove545020(got, monsterActionRandomWalkHooks545020{
			random: func(min, max int) int {
				if min != -20 || max != 20 {
					t.Fatalf("walk RNG bounds = %d..%d", min, max)
				}
				return 20
			},
		})
	}
	if !monsterActionConfused545140(unit, hooks) {
		t.Fatal("confused random walk was not handled")
	}
	if unit.Direction1 != 14 || unit.Direction2 != 14 {
		t.Fatalf("directions = %d/%d, want 14", unit.Direction1, unit.Direction2)
	}
	cosine, sine := SinCosDir(14)
	if math.Float32bits(unit.ForceVec.X) != math.Float32bits(float32(2.5*float64(cosine))) ||
		math.Float32bits(unit.ForceVec.Y) != math.Float32bits(float32(2.5*float64(sine))) {
		t.Fatalf("force = %v", unit.ForceVec)
	}
}

func TestMonsterActionConfused545140RejectsInvalidAdmission(t *testing.T) {
	validHooks := confusedHooks545140(t)
	validHooks.random = func(int, int) int { return 15 }
	validHooks.randomWalk = func(*Object) bool { return true }
	validHooks.canMelee = func(*Object) bool { return false }
	validHooks.canShoot = func(*Object) bool { return false }
	validHooks.push = func(*Object, ai.ActionType) *AIStackItem { return nil }

	for _, unit := range []*Object{
		nil,
		{},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(new(MonsterUpdateData))},
	} {
		if monsterActionConfused545140(unit, validHooks) {
			t.Fatalf("invalid unit %#v was handled", unit)
		}
	}

	unit := confusedTestObject545140(t)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_GUARD)
	if monsterActionConfused545140(unit, validHooks) {
		t.Fatal("wrong stack head was handled")
	}

	unit = confusedTestObject545140(t)
	missing := []monsterActionConfusedHooks545140{
		{randomWalk: validHooks.randomWalk, canMelee: validHooks.canMelee, canShoot: validHooks.canShoot, push: validHooks.push},
		{random: validHooks.random, canMelee: validHooks.canMelee, canShoot: validHooks.canShoot, push: validHooks.push},
		{random: validHooks.random, randomWalk: validHooks.randomWalk, canShoot: validHooks.canShoot, push: validHooks.push},
		{random: validHooks.random, randomWalk: validHooks.randomWalk, canMelee: validHooks.canMelee, push: validHooks.push},
		{random: validHooks.random, randomWalk: validHooks.randomWalk, canMelee: validHooks.canMelee, canShoot: validHooks.canShoot},
	}
	for i, hooks := range missing {
		if monsterActionConfused545140(unit, hooks) {
			t.Fatalf("missing hook case %d was handled", i)
		}
	}
}

func TestMonsterActionConfused545140PreservesNativeObjectPointer(t *testing.T) {
	unit := confusedTestObject545140(t)
	want := uintptr(unsafe.Pointer(unit))
	if unsafe.Sizeof(uintptr(0)) == 8 && want <= uintptr(^uint32(0)) {
		t.Fatalf("unit pointer = %#x, want value above PE32 range", want)
	}
	hooks := confusedHooks545140(t)
	hooks.random = func(int, int) int { return 14 }
	hooks.canMelee = func(got *Object) bool {
		if uintptr(unsafe.Pointer(got)) != want {
			t.Fatalf("melee unit = %#x, want %#x", uintptr(unsafe.Pointer(got)), want)
		}
		return false
	}
	hooks.canShoot = func(got *Object) bool {
		if uintptr(unsafe.Pointer(got)) != want {
			t.Fatalf("shoot unit = %#x, want %#x", uintptr(unsafe.Pointer(got)), want)
		}
		return true
	}
	hooks.push = func(got *Object, action ai.ActionType) *AIStackItem {
		if uintptr(unsafe.Pointer(got)) != want || action != ai.ACTION_MISSILE_ATTACK {
			t.Fatalf("push = (%#x, %s), want (%#x, %s)", uintptr(unsafe.Pointer(got)), action, want, ai.ACTION_MISSILE_ATTACK)
		}
		return nil
	}
	if !monsterActionConfused545140(unit, hooks) {
		t.Fatal("native-width unit was not handled")
	}
}
