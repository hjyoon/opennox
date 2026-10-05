package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// 00545625/00545655 call PushAction before reading food X/Y. Each rejected
// result suppresses only its own payload, never any subsequent push.
func TestMonsterRetreatCheckEdibles5455E0FoodPushOrder(t *testing.T) {
	for _, quest := range []bool{false, true} {
		for accepted := uint32(0); accepted < 32; accepted++ {
			t.Run(fmt.Sprintf("quest=%t/accepted=%02x", quest, accepted), func(t *testing.T) {
				unit := monsterActionTestObject50A910(t)
				food := monsterActionTestObject50A910(t)
				food.PosVec = types.Ptf(11, 22)
				searchedFood := food
				other := &Object{PosVec: types.Ptf(900, 1000)}
				wantActions := []ai.ActionType{
					ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY,
					ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO,
				}
				positions := []types.Pointf{
					types.Ptf(100, 200), types.Ptf(300, 400),
					types.Ptf(math.Float32frombits(0x80000000), math.Float32frombits(0x7fc54321)),
					types.Ptf(500, 600), types.Ptf(700, 800),
				}
				initial := [4]uintptr{71, 72, 73, 74}
				var items [5]*AIStackItem
				var actions []ai.ActionType
				var payloadCount, searches int
				monsterRetreatCheckEdibles5455E0(unit, monsterActionRetreatHooks545440{
					quest: quest,
					searchFood: func(got *Object, radius float32) *Object {
						searches++
						wantRadius := float32(250)
						if quest {
							wantRadius = 640
						}
						if got != unit || radius != wantRadius || searches != 1 {
							t.Fatalf("search arguments=%p/%g calls=%d", got, radius, searches)
						}
						return searchedFood
					},
					push: func(action ai.ActionType, args ...any) *AIStackItem {
						index := len(actions)
						actions = append(actions, action)
						payloadCount += len(args)
						if index >= len(wantActions) || action != wantActions[index] {
							t.Fatalf("push order=%v", actions)
						}
						// The search result is a cached native identity, not reselected.
						searchedFood = other
						food.PosVec = positions[index]
						if accepted&(1<<index) == 0 {
							return nil
						}
						item := &AIStackItem{Action: uint32(action), Args: initial}
						item.SetArgs(args...)
						items[index] = item
						return item
					},
				})
				if searches != 1 || payloadCount != 0 || !reflect.DeepEqual(actions, wantActions) {
					t.Errorf("searches=%d push payloads=%d actions=%v", searches, payloadCount, actions)
				}
				for index, item := range items {
					if item == nil {
						continue
					}
					want := initial
					switch index {
					case 2, 4:
						want[0], want[1], want[2] = uintptr(math.Float32bits(positions[index].X)),
							uintptr(math.Float32bits(positions[index].Y)), uintptr(unsafe.Pointer(food))
					case 3:
						want[0] = uintptr(unsafe.Pointer(food))
					}
					if item.Args != want {
						t.Errorf("push %d payload=%#x want=%#x", index, item.Args, want)
					}
				}
				runtime.KeepAlive(unit)
				runtime.KeepAlive(food)
			})
		}
	}
}

// 00545697 writes only the first DWORD; 0054569E writes only the low BYTE of
// argument 2. Neither argument 1 nor the remaining bytes are normalized.
func TestMonsterRetreatCheckEdibles5455E0RoamPartialStores(t *testing.T) {
	for _, quest := range []bool{false, true} {
		for accepted := uint32(0); accepted < 16; accepted++ {
			t.Run(fmt.Sprintf("quest=%t/accepted=%02x", quest, accepted), func(t *testing.T) {
				unit := monsterActionTestObject50A910(t)
				initial := [4]uintptr{0x11223344, 0x55667788, 0x99aabbcc, 0xddeeff00}
				if unsafe.Sizeof(uintptr(0)) == 8 {
					high := uintptr(0x12345678)
					high <<= 32
					initial[1] |= high
					initial[2] |= high
				}
				wantActions := []ai.ActionType{
					ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY,
					ai.DEPENDENCY_NO_VISIBLE_FOOD, ai.ACTION_ROAM,
				}
				var actions []ai.ActionType
				var items [4]*AIStackItem
				var payloadCount int
				monsterRetreatCheckEdibles5455E0(unit, monsterActionRetreatHooks545440{
					quest: quest,
					searchFood: func(got *Object, radius float32) *Object {
						wantRadius := float32(250)
						if quest {
							wantRadius = 640
						}
						if got != unit || radius != wantRadius {
							t.Fatalf("search arguments=%p/%g", got, radius)
						}
						return nil
					},
					push: func(action ai.ActionType, args ...any) *AIStackItem {
						index := len(actions)
						actions = append(actions, action)
						payloadCount += len(args)
						if index >= len(wantActions) || action != wantActions[index] {
							t.Fatalf("push order=%v", actions)
						}
						if accepted&(1<<index) == 0 {
							return nil
						}
						item := &AIStackItem{Action: uint32(action), Args: initial}
						item.SetArgs(args...)
						items[index] = item
						return item
					},
				})
				if payloadCount != 0 || !reflect.DeepEqual(actions, wantActions) {
					t.Errorf("push payloads=%d actions=%v", payloadCount, actions)
				}
				for index, item := range items {
					if item == nil {
						continue
					}
					want := initial
					if index == 3 {
						want[0], want[2] = 0, initial[2]&^uintptr(0xff)|0x80
					}
					if item.Args != want {
						t.Errorf("push %d payload=%#x want=%#x", index, item.Args, want)
					}
				}
			})
		}
	}
}

// Push and Cancel below are the real server services, with native object
// argument slots and every zero-to-six remaining-stack-capacity boundary.
func TestMonsterRetreatCheckEdibles5455E0RealStackCapacityAndCancel(t *testing.T) {
	for _, hasFood := range []bool{false, true} {
		for slots := 0; slots <= 6; slots++ {
			t.Run(fmt.Sprintf("food=%t/slots=%d", hasFood, slots), func(t *testing.T) {
				s, unit, update := monsterRetreatSpellFixture545440(t, false)
				update.StatusFlags, update.CurrentEnemy = 0, nil
				unit.PosVec = types.Ptf(100, 200)
				food := monsterActionTestObject50A910(t)
				food.ObjClass = object.ClassFood
				food.PosVec = types.Ptf(123, 245)
				update.AIStackInd = int8(len(update.AIStack) - 1 - slots)
				initialIndex := update.AIStackInd
				update.AIStackHead().Action = uint32(ai.ACTION_RETREAT)
				update.AIStackHead().Field5 = 1
				originalHead := *update.AIStackHead()
				var cancelCalls int
				old, existed := aiActions[ai.ACTION_RETREAT]
				aiActions[ai.ACTION_RETREAT] = monsterRetreatCancelAction545440{cancel: func(got *Object) {
					if got != unit {
						t.Fatal("Cancel lost native unit identity")
					}
					cancelCalls++
					food.PosVec = types.Ptf(math.Float32frombits(0x80000000), math.Float32frombits(0x7fc54321))
				}}
				t.Cleanup(func() {
					if existed {
						aiActions[ai.ACTION_RETREAT] = old
					} else {
						delete(aiActions, ai.ACTION_RETREAT)
					}
				})
				monsterRetreatCheckEdibles5455E0(unit, monsterActionRetreatHooks545440{
					searchFood: func(owner *Object, radius float32) *Object {
						return monsterSearchEdible544A00(owner, radius, monsterSearchEdibleHooks544A00{
							eachInCircle: func(center types.Pointf, r float32, each func(*Object) bool) {
								if center != types.Ptf(100, 200) || r != 250 {
									t.Fatal("edible search lost live center/radius")
								}
								if hasFood {
									each(food)
								}
							},
							canInteract: func(owner, target *Object, flags int) bool {
								return owner == unit && target == food && flags == 0
							},
						})
					},
					push: unit.MonsterPushAction,
				})
				wantActions := []ai.ActionType{ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_NO_VISIBLE_FOOD, ai.ACTION_ROAM}
				if hasFood {
					wantActions = []ai.ActionType{ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO}
				}
				accepted := min(slots, len(wantActions))
				wantCancel := 0
				if slots > 0 {
					wantCancel = 1
				}
				if cancelCalls != wantCancel || s.AI.StackChanged != (slots > 0) ||
					update.AIStackInd != initialIndex+int8(accepted) || update.AIStack[initialIndex] != originalHead {
					t.Fatalf("real stack prefix: cancel=%d changed=%t index=%d want=%d", cancelCalls, s.AI.StackChanged, update.AIStackInd, initialIndex+int8(accepted))
				}
				for index := 0; index < accepted; index++ {
					item := &update.AIStack[int(initialIndex)+1+index]
					want := [4]uintptr{}
					if hasFood {
						switch index {
						case 2, 4:
							want = [4]uintptr{0x80000000, 0x7fc54321, uintptr(unsafe.Pointer(food)), 0}
						case 3:
							want[0] = uintptr(unsafe.Pointer(food))
						}
					} else if index == 3 {
						want[2] = 0x80
					}
					if item.Type() != wantActions[index] || item.Args != want ||
						hasFood && (index == 2 || index == 4) && item.ArgObj(2) != food ||
						hasFood && index == 3 && item.ArgObj(0) != food {
						t.Fatalf("real stack item %d = %+v want action=%v args=%#x", index, item, wantActions[index], want)
					}
				}
				runtime.KeepAlive(unit)
				runtime.KeepAlive(food)
			})
		}
	}
}
