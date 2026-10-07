package server

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Compose the real native health-retreat, indexed edible search, action-stack
// refresh and MOVE_TO arrival services. GAME.EXE 00544500 keeps a nonnil
// tracked target even after 0050D5A0 reports arrival. The independent periodic
// food tail skips aggression 0.01 through 00534440/00547B88.
//
// These are isolated close-food inputs, including the previously observed
// stopping coordinates. They are not full game ticks, travel, consumption,
// stock-world reproduction, or execution of the original Windows program.
func TestMonsterRetreatFoodArrival5443F0VisiblePassiveContract(t *testing.T) {
	for _, npc := range []bool{false, true} {
		for _, foodKind := range []struct {
			name string
			kind object.FoodClass
		}{
			{"apple", object.FoodApple},
			{"simple-food", object.FoodSimple},
		} {
			for _, positions := range []struct {
				name     string
				from, to types.Pointf
			}{
				{"captured-stop", types.Ptf(3426.8457, 2309.9163), types.Ptf(3428, 2308)},
				{"reflected-stop", types.Ptf(3429.1543, 2306.0837), types.Ptf(3428, 2308)},
				{"coincident", types.Ptf(3428, 2308), types.Ptf(3428, 2308)},
				{"inside-biased-eight", types.Ptf(3420.02, 2308), types.Ptf(3428, 2308)},
			} {
				for _, frame := range []uint32{656, 672} {
					t.Run(fmt.Sprintf("npc-%t/%s/%s/frame-%d", npc, foodKind.name, positions.name, frame), func(t *testing.T) {
						s, unit, update := moveToNativeRetryFixture5443F0(t, ai.ACTION_IDLE)
						health, freeHealth := alloc.New(HealthData{})
						t.Cleanup(freeHealth)
						*health = HealthData{Cur: 76, Max: 80} // Initial injury, not a healing result.
						food, freeFood := alloc.New(Object{})
						t.Cleanup(freeFood)
						s.Map.Init()
						s.Walls.byPos = make([]*Wall, wallsPerBucket*WallGridSize)
						t.Cleanup(s.Map.Free)
						s.SetFrame(frame) // Initial fixture input, never advanced below.
						unit.HealthData = health
						unit.PosVec, unit.NewPos, unit.ObjFlags = positions.from, positions.from, object.FlagActive
						if npc {
							unit.ObjSubClass = object.SubClass(object.MonsterNPC)
						}
						update.Aggression, update.RetreatLevel, update.ResumeLevel = 0.01, 0.98, 1
						*food = Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(foodKind.kind),
							ObjFlags: object.FlagActive | object.FlagNoCollide, PosVec: positions.to, NewPos: positions.to}
						// Explicit synthetic shapes make both records searchable;
						// ShapeKindNone has no collision bounds in the real index.
						// These radii are fixture inputs, not stock type claims.
						unit.Shape.Kind, food.Shape.Kind = ShapeKindCircle, ShapeKindCircle
						unit.Shape.Circle.R, unit.Shape.Circle.R2 = 12, 144
						food.Shape.Circle.R, food.Shape.Circle.R2 = 4, 16
						for _, pointer := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(health), unsafe.Pointer(food)} {
							if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
								t.Fatalf("C-owned retreat-food record below 4 GiB: %p", pointer)
							}
						}
						s.Map.AddObjectToIndex(unit)
						s.Map.AddObjectToIndex(food)
						logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
						if !s.CanInteract(unit, food, 0) || s.MonsterSearchEdible544A00(unit, 250) != food {
							t.Fatal("actual indexed search did not select visible food")
						}
						if !s.monsterMainHealthRetreat547210(unit, update, nil, MonsterMainRuntime547210{}) {
							t.Fatal("actual health branch did not schedule retreat")
						}
						s.MonsterActionRetreat545440(unit)
						wantActions := []ai.ActionType{ai.DEPENDENCY_NOT_CORNERED, ai.ACTION_RETREAT,
							ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY,
							ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO}
						if int(update.AIStackInd)+1 != len(wantActions) {
							t.Fatalf("actual retreat stack=%+v", update.GetAIStack())
						}
						for index, action := range wantActions {
							if update.AIStack[index].Type() != action {
								t.Fatalf("stack[%d]=%s want=%s", index, update.AIStack[index].Type(), action)
							}
						}
						for _, index := range []int{4, 6} {
							if update.AIStack[index].ArgPos(0) != positions.to || update.AIStack[index].ArgObj(2) != food {
								t.Fatalf("stack[%d] lost full tracked food identity", index)
							}
						}
						if update.AIStack[5].ArgObj(0) != food {
							t.Fatal("PICKUP_OBJECT lost full food identity")
						}
						beforeUnit, beforeUpdate, beforeHealth, beforeFood := *unit, *update, *health, *food
						stackChanged := s.AI.StackChanged
						for pass := 0; pass < 3; pass++ {
							if s.MonsterActionRefresh50A910(unit) != 0 {
								t.Fatal("actual tracked-target refresh failed")
							}
							if !s.MonsterActionMoveTo5443F0(unit,
								func(*Object, *types.Pointf) *Waypoint { t.Fatal("arrival must not find a waypoint"); return nil },
								func(*Object, *types.Pointf) { t.Fatal("arrival must not build a detailed path") }) {
								t.Fatal("actual close-food MOVE_TO was rejected")
							}
							if !s.monsterMainEatNearbyFood547210(unit, update, MonsterMainRuntime547210{
								SearchEdible:   func(*Object, float32) *Object { t.Fatal("passive periodic food search must be skipped"); return nil },
								PlaceInventory: func(*Object, *Object, int, int) bool { t.Fatal("passive food must not be placed"); return false },
								UseByNetCode:   func(*Object, *Object) int32 { t.Fatal("passive food must not be used"); return 0 },
							}) {
								t.Fatal("passive periodic food tail was not handled")
							}
							if *update != beforeUpdate || *health != beforeHealth || s.AI.StackChanged != stackChanged {
								t.Fatalf("pass %d popped MOVE_TO, changed the stack/path or fabricated healing", pass)
							}
							// Spatial queries alone can change visitation Field62.
							for _, pair := range [][2]*Object{{&beforeUnit, unit}, {&beforeFood, food}} {
								actual := *pair[1]
								actual.Field62 = pair[0].Field62
								if !bytes.Equal(moveToRawSnapshot5443F0(unsafe.Pointer(&actual), unsafe.Sizeof(actual)),
									moveToRawSnapshot5443F0(unsafe.Pointer(pair[0]), unsafe.Sizeof(*pair[0]))) {
									t.Fatal("arrival changed an object outside its visitation token")
								}
							}
							if s.Frame() != frame || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
								t.Fatal("retreat-food contract changed the frame or consumed RNG")
							}
						}
					})
				}
			}
		}
	}
}
