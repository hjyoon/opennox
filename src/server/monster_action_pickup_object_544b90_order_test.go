package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// GAME.EXE 00544BB4..00544BD9 retains x87 53-bit differences, Y-square,
// X-square and sum until FCOM. C0 alone admits unordered as well as less.
// Literal input bits and admission results do not use the implementation.
func TestMonsterActionPickupObject544B90RetainedArithmetic(t *testing.T) {
	for _, tc := range []struct {
		name               string
		unitX, unitY, x, y uint32
		admit              bool
	}{
		{"rounded-boundary-1", 0, 0, 0x4295ffff, 0x3cf5c28f, true},
		{"rounded-boundary-2", 0, 0, 0x4295fff1, 0x3e051eb8, true},
		{"rounded-boundary-3", 0, 0, 0x4295ffdd, 0x3e4ccccd, true},
		{"rounded-boundary-4", 0, 0, 0x4295ffac, 0x3e9eb852, true},
		{"rounded-boundary-5", 0, 0, 0x4295ff7b, 0x3ec7ae14, true},
		{"rounded-boundary-6", 0, 0, 0x4295ff74, 0x3ecccccd, true},
		{"rounded-boundary-7", 0, 0, 0x4295ff4f, 0x3ee66666, true},
		{"rounded-boundary-8", 0, 0, 0x4295ff2e, 0x3efae148, true},
		{"unspilled-subtraction", 0x35800000, 0, 0x42960000, 0, true},
		{"exact-75", 0, 0, 0x42960000, 0, false},
		{"exact-45-60", 0, 0, 0x42340000, 0x42700000, false},
		{"just-outside", 0, 0, 0x42960001, 0, false},
		{"NaN-X", 0, 0, 0x7fc54321, 0, true},
		{"NaN-Y", 0, 0, 0, 0x7fc54321, true},
		{"positive-infinity", 0, 0, 0x7f800000, 0, false},
		{"negative-infinity", 0, 0, 0xff800000, 0, false},
		{"infinity-minus-infinity", 0x7f800000, 0, 0x7f800000, 0, true},
		{"signed-zero", 0x80000000, 0, 0, 0x80000000, true},
		{"subnormal", 0, 0, 1, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := monsterActionTestObject50A910(t)
			target.PosVec = types.Ptf(math.Float32frombits(tc.x), math.Float32frombits(tc.y))
			unit, _ := pickupObjectMonster544B90(target)
			unit.PosVec = types.Ptf(math.Float32frombits(tc.unitX), math.Float32frombits(tc.unitY))
			var events []string
			got := monsterActionPickupObject544B90(unit, monsterActionPickupObjectHooks544B90{
				canInteract: func(owner, item *Object, flags int) bool {
					if owner != unit || item != target || flags != 0 {
						t.Fatal("interaction identity/flags changed")
					}
					events = append(events, "interact")
					return true
				},
				placeInventory: func(owner, item *Object, a, b int) bool {
					if owner != unit || item != target || a != 1 || b != 1 {
						t.Fatal("placement identity/flags changed")
					}
					events = append(events, "place")
					return false
				},
				useByNetCode: func(*Object, *Object) int32 { t.Fatal("ordinary target used"); return 0 },
				pop:          func() int { events = append(events, "pop"); return -129 },
			})
			want := []string{"pop"}
			if tc.admit {
				want = []string{"interact", "place", "pop"}
			}
			if got != -129 || !reflect.DeepEqual(events, want) {
				t.Fatalf("result=%d events=%v, want -129/%v", got, events, want)
			}
			runtime.KeepAlive(target)
		})
	}
}

// 00544BEB reloads the entry-cached slot AFTER visibility, before placement.
// 00544BF9 reloads that SAME slot again after placement, ignoring its result.
func TestMonsterActionPickupObject544B90CallbackReloadOrder(t *testing.T) {
	for _, mode := range []string{"new-target", "nil-target-to-placement", "new-update", "new-head-index"} {
		t.Run(mode, func(t *testing.T) {
			original, placed, used := monsterActionTestObject50A910(t), monsterActionTestObject50A910(t), monsterActionTestObject50A910(t)
			original.PosVec = types.Ptf(20, 20)
			placed.PosVec = types.Ptf(10000, 20000) // No second range/visibility check exists.
			used.ObjSubClass = 0x10010
			unit, cachedHead := pickupObjectMonster544B90(original)
			update := unit.UpdateDataMonster()
			var events []string
			wantPlaced := placed
			if mode == "nil-target-to-placement" {
				wantPlaced = nil
			}
			got := monsterActionPickupObject544B90(unit, monsterActionPickupObjectHooks544B90{
				canInteract: func(owner, item *Object, flags int) bool {
					if owner != unit || item != original || flags != 0 {
						t.Fatal("visibility must see the original identity")
					}
					events = append(events, "interact")
					cachedHead.SetArgs(wantPlaced)
					switch mode {
					case "new-update":
						unit.UpdateData = unsafe.Pointer(&MonsterUpdateData{AIStackInd: 0})
					case "new-head-index":
						update.AIStackInd = 1
						update.AIStack[1].Action = uint32(ai.ACTION_PICKUP_OBJECT)
						update.AIStack[1].SetArgs(original)
					}
					unit.PosVec = types.Ptf(-10000, -20000)
					return true
				},
				placeInventory: func(owner, item *Object, a, b int) bool {
					if owner != unit || item != wantPlaced || a != 1 || b != 1 {
						t.Errorf("placement=%p/%p/%d/%d, want %p/%p/1/1", owner, item, a, b, unit, wantPlaced)
					}
					events = append(events, "place")
					cachedHead.SetArgs(used)
					return false
				},
				useByNetCode: func(owner, item *Object) int32 {
					if owner != unit || item != used {
						t.Fatal("use must reload the same cached slot after placement")
					}
					events = append(events, "use")
					return -7
				},
				pop: func() int { events = append(events, "pop"); return 37 },
			})
			if got != 37 || !reflect.DeepEqual(events, []string{"interact", "place", "use", "pop"}) {
				t.Fatalf("result=%d events=%v", got, events)
			}
			runtime.KeepAlive(original)
			runtime.KeepAlive(placed)
			runtime.KeepAlive(used)
			runtime.KeepAlive(update)
		})
	}
}

// 00544BFF tests only bit 0x10 of the live low byte, without a Food class gate.
func TestMonsterActionPickupObject544B90PostPlacementSubclassByte(t *testing.T) {
	for _, tc := range []struct {
		name string
		bits uint32
		use  bool
	}{
		{"zero", 0, false}, {"health", 0x10, true}, {"mushroom", 0x80, false},
		{"upper-health-only", 0x1000, false}, {"upper-with-health", 0x10010, true},
		{"all-bits", 0xffffffff, true}, {"all-except-health", 0xffffffef, false},
		{"potion-without-health", 0x04, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := monsterActionTestObject50A910(t)
			target.ObjClass, target.ObjSubClass, target.PosVec = 0, 0x10, types.Ptf(20, 20)
			unit, _ := pickupObjectMonster544B90(target)
			var events []string
			monsterActionPickupObject544B90(unit, monsterActionPickupObjectHooks544B90{
				canInteract: func(*Object, *Object, int) bool { events = append(events, "interact"); return true },
				placeInventory: func(*Object, *Object, int, int) bool {
					events = append(events, "place")
					target.ObjSubClass = object.SubClass(tc.bits)
					return false
				},
				useByNetCode: func(owner, item *Object) int32 {
					if owner != unit || item != target {
						t.Fatal("use identity changed")
					}
					events = append(events, "use")
					return 0
				},
				pop: func() int { events = append(events, "pop"); return 0 },
			})
			want := []string{"interact", "place", "pop"}
			if tc.use {
				want = []string{"interact", "place", "use", "pop"}
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events=%v want=%v", events, want)
			}
			runtime.KeepAlive(target)
		})
	}
}

// A callback can invalidate the cached slot or a bound service. Preserve the
// original completed-call prefix; never convert these faults into a Pop.
func TestMonsterActionPickupObject544B90CallbackFaultPrefix(t *testing.T) {
	for _, mode := range []string{"nil-after-visibility", "nil-after-placement", "nil-visibility-service", "nil-placement-service", "nil-use-service"} {
		t.Run(mode, func(t *testing.T) {
			target := monsterActionTestObject50A910(t)
			target.PosVec, target.ObjSubClass = types.Ptf(20, 20), 0x10
			unit, head := pickupObjectMonster544B90(target)
			var events []string
			hooks := monsterActionPickupObjectHooks544B90{
				canInteract: func(*Object, *Object, int) bool {
					events = append(events, "interact")
					if mode == "nil-after-visibility" {
						head.SetArgs((*Object)(nil))
					}
					return true
				},
				placeInventory: func(_ *Object, item *Object, _, _ int) bool {
					events = append(events, "place")
					if mode == "nil-after-visibility" && item != nil {
						t.Errorf("placement must receive nil, got %p", item)
					}
					if mode == "nil-after-placement" {
						head.SetArgs((*Object)(nil))
					}
					return true
				},
				useByNetCode: func(*Object, *Object) int32 { events = append(events, "use"); return 0 },
				pop:          func() int { events = append(events, "pop"); return 0 },
			}
			want := []string{"interact", "place"}
			switch mode {
			case "nil-visibility-service":
				hooks.canInteract = nil
				want = nil
			case "nil-placement-service":
				hooks.placeInventory = nil
				want = []string{"interact"}
			case "nil-use-service":
				hooks.useByNetCode = nil
			}
			fault := func() (caught any) {
				defer func() { caught = recover() }()
				monsterActionPickupObject544B90(unit, hooks)
				return nil
			}()
			if fault == nil || !reflect.DeepEqual(events, want) {
				t.Fatalf("fault=%v completed=%v want=%v", fault, events, want)
			}
			runtime.KeepAlive(target)
		})
	}
}
