package opennox

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The unchanged PE32 Main -> dependency -> close MOVE_TO pipeline produces
// these literal states for the R82 Troll capture at 654/655 and 763/764.
// NPC class bits and a separate due-tick input are bounded synthetic cases.
// Original roots, stack helpers and arrival execute; buffs, mode, visibility,
// audio and RNG are declared scalar services. This is not original world
// travel, regeneration, PICKUP consumption, or a passing public food e2e.
//
// Native objects/update/health are C-owned above 4 GiB. Native visibility and
// wall rays below are real services, not replaced root/condition results.
func TestAIPassiveFoodPipeline547210NativeOriginalCancellation(t *testing.T) {
	oldGame := noxflags.GetGame()
	noxflags.UnsetGame(oldGame)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.SetGame(oldGame)
	})
	for _, npc := range []bool{false, true} {
		for _, tc := range []struct {
			name             string
			frames           []uint32
			progress, corner uint32
			hp               uint16
			path             [3]uint32
			status           object.MonsterStatus
		}{
			{"first-stall", []uint32{654, 655}, 639, 0, 76, [3]uint32{7, 6, 0}, 0x10000},
			{"second-stall", []uint32{763, 764}, 748, 655, 78, [3]uint32{}, 0x210000},
			{"periodic-due", []uint32{672}, 672, 0, 76, [3]uint32{}, 0x10000},
		} {
			t.Run(fmt.Sprintf("npc-%t/%s", npc, tc.name), func(t *testing.T) {
				s, unit, food := aiVisibleDependencyNative546A70(t)
				s.SetTickRate(30)
				update := unit.UpdateDataMonster()
				health, freeHealth := alloc.New(server.HealthData{})
				t.Cleanup(freeHealth)
				*health = server.HealthData{Cur: tc.hp, Max: 80}
				unit.HealthData = health
				unit.ObjClass, unit.ObjSubClass, unit.ObjFlags = 2097154, 66050, 16777732
				if npc {
					unit.ObjSubClass = 69778
				}
				unit.NetCode, unit.SpeedBase = 499, 2.3
				unit.PosVec = types.Ptf(3426.8457, 2309.9163)
				unit.NewPos, unit.PrevPos = unit.PosVec, unit.PosVec
				food.PosVec, food.NewPos = types.Ptf(3428, 2308), types.Ptf(3428, 2308)
				food.ObjSubClass = object.SubClass(object.FoodApple)
				food.ObjFlags |= object.FlagNoCollide
				*update = server.MonsterUpdateData{
					AIStackInd: 6, Aggression: 0.01, RetreatLevel: 0.98, ResumeLevel: 1,
					StatusFlags: tc.status, Field124: tc.progress, Field127: tc.corner,
					Field2: tc.path[0], Field67: tc.path[1], Field91: tc.path[2],
					Field125: math.Float32bits(unit.PosVec.X), Field126: math.Float32bits(unit.PosVec.Y),
				}
				if tc.progress == 639 {
					update.Field125, update.Field126 = math.Float32bits(3423.5134), math.Float32bits(2311.778)
				}
				for index, action := range []ai.ActionType{ai.DEPENDENCY_NOT_CORNERED, ai.ACTION_RETREAT,
					ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY,
					ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO} {
					item := &update.AIStack[index]
					item.Action = uint32(action)
					if index == 4 || index == 6 {
						item.SetArgs(food.PosVec, unsafe.Pointer(food))
					}
					if index == 5 {
						item.SetArgs(unsafe.Pointer(food))
					}
					if index == 1 || index == 5 || index == 6 {
						item.Field5 = 1
					}
				}
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(food), unsafe.Pointer(update), unsafe.Pointer(health)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("C-owned pipeline record below 4 GiB: %p", pointer)
					}
				}
				if !s.CanInteract(unit, food, 0) || !s.MapTraceRay(unit.PosVec, food.PosVec, server.MapTraceFlag1) {
					t.Fatal("real native visibility/ray differs from the declared original clear input")
				}
				foodCalls := 0
				var randomCalls [][3]int
				runtime := server.MonsterMainRuntime547210{
					RandomInt: func(min, max int) int {
						value := 20
						if min == 0 && max == 100 {
							value = 50
						} else if min != 15 || max != 60 {
							t.Fatalf("unexpected original RNG bounds: %d/%d", min, max)
						}
						randomCalls = append(randomCalls, [3]int{min, max, value})
						return value
					},
					TileAt:   func(types.Pointf) int { t.Fatal("unreachable dodge terrain query"); return 0 },
					TraceRay: s.MapTraceRay, TraceObstacles: s.MapTraceObstacles,
					SearchEdible:   func(*server.Object, float32) *server.Object { foodCalls++; return nil },
					PlaceInventory: func(*server.Object, *server.Object, int, int) bool { foodCalls++; return false },
					UseByNetCode:   func(*server.Object, *server.Object) int32 { foodCalls++; return 0 },
				}
				for _, frame := range tc.frames {
					s.SetFrame(frame) // Explicit captured clock input; no simulation tick is fabricated.
					check := func(phase string, wantStack []uint32, progress, corner uint32, path [3]uint32, status object.MonsterStatus) {
						t.Helper()
						var stack []uint32
						for _, item := range update.GetAIStack() {
							stack = append(stack, item.Action)
						}
						tracked := update.AIStackHead().ArgObj(2) == food
						if !reflect.DeepEqual(stack, wantStack) || update.Field124 != progress || update.Field127 != corner ||
							[3]uint32{update.Field2, update.Field67, update.Field91} != path || update.StatusFlags != status ||
							health.Cur != tc.hp || health.Max != 80 || foodCalls != 0 || s.Frame() != frame ||
							tracked != (wantStack[len(wantStack)-1] == uint32(ai.ACTION_MOVE_TO)) {
							t.Fatalf("%d/%s original state differs: stack=%v progress=%d corner=%d path=%d/%d/%d status=%x hp=%d tracked=%t food=%d",
								frame, phase, stack, update.Field124, update.Field127, update.Field2, update.Field67, update.Field91, uint32(update.StatusFlags), health.Cur, tracked, foodCalls)
						}
						row, err := json.Marshal(map[string]any{
							"npc": npc, "case": tc.name, "frame": frame, "phase": phase, "stack": stack,
							"progress": progress, "corner": corner, "path": path, "status": fmt.Sprintf("0x%x", uint32(status)),
							"hp": health.Cur, "tracked": tracked, "food_calls": foodCalls,
						})
						if err != nil {
							t.Fatal(err)
						}
						t.Logf("PASSIVE_PIPELINE_JSON %s", row)
					}
					wantStack := []uint32{68, 6, 64, 56, 48, 11, 7}
					progress, corner, path, status := tc.progress, tc.corner, tc.path, tc.status
					stalled := frame == 655 || frame == 764
					if stalled {
						wantStack = append(wantStack, 1)
						progress, corner, path, status = frame, frame, [3]uint32{}, tc.status|object.MonStatusFrustrated
					}
					if !s.MonsterMainNativeRuntime547210(unit, runtime) {
						t.Fatal("actual Main rejected the admitted original scalar input")
					}
					check("main", wantStack, progress, corner, path, status)
					(&aiData{s: s}).nox_xxx_mobActionDependency(unit)
					if stalled {
						wantStack = []uint32{0}
					}
					check("dependency", wantStack, progress, corner, path, status)
					if !stalled {
						if !s.MonsterActionMoveTo5443F0(unit,
							func(*server.Object, *types.Pointf) *server.Waypoint {
								t.Fatal("close arrival must not request a waypoint")
								return nil
							},
							func(*server.Object, *types.Pointf) { t.Fatal("close arrival must not request a detailed path") }) {
							t.Fatal("actual MOVE_TO rejected the retained food target")
						}
						check("move-to", wantStack, progress, corner, path, status)
					}
				}
				var wantRandom [][3]int
				if tc.name != "periodic-due" {
					wantRandom = [][3]int{{0, 100, 50}, {15, 60, 20}}
				}
				if !reflect.DeepEqual(randomCalls, wantRandom) {
					t.Fatalf("original RNG sequence differs: %v want %v", randomCalls, wantRandom)
				}
			})
		}
	}
}
