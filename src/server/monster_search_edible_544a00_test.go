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
)

// GAME.EXE 00544A9E..00544ACB keeps the distance unspilled until FCOM,
// tests only C0 (less or unordered), then spills the winning distance to F32.
// The expected identities below are independent of the Go distance helper.
func TestMonsterSearchEdible544A00Arithmetic(t *testing.T) {
	for _, mode := range []string{
		"nearest", "exact-tie", "best-rounds-up", "rounded-tie-replaces",
		"best-rounds-down", "unspilled-subtraction", "exact-limit",
		"beyond-limit", "NaN-last", "NaN-middle", "NaN-then-infinity",
		"infinite", "infinity-minus-infinity", "signed-zero", "subnormal",
	} {
		t.Run(mode, func(t *testing.T) {
			unit := &Object{}
			first := &Object{ObjClass: object.ClassFood, PosVec: types.Ptf(30, 40)}
			second := &Object{ObjClass: object.ClassFood, PosVec: types.Ptf(10, 20)}
			items, want := []*Object{first, second}, second
			switch mode {
			case "exact-tie":
				second.PosVec, want = types.Ptf(40, 30), first
			case "best-rounds-up":
				first.PosVec, second.PosVec = types.Ptf(1, 0.0003), types.Ptf(1, 0.00034)
			case "rounded-tie-replaces":
				first.PosVec, second.PosVec = types.Ptf(1, 0.0003), types.Ptf(1, 0.0003)
			case "best-rounds-down":
				first.PosVec, second.PosVec, want = types.Ptf(1, 0.0002), types.Ptf(1, 0.0001), first
			case "unspilled-subtraction":
				unit.PosVec.X = 0x1p-15
				first.PosVec, items, want = types.Ptf(1000, 3000), []*Object{first}, first
			case "exact-limit":
				first.PosVec, second.PosVec, want = types.Ptf(1000, 3000), types.Ptf(3000, 1000), nil
			case "beyond-limit":
				first.PosVec, second.PosVec, want = types.Ptf(1001, 3000), types.Ptf(3001, 1000), nil
			case "NaN-last":
				second.PosVec.X = math.Float32frombits(0x7fc54321)
			case "NaN-middle":
				first.PosVec.X = math.Float32frombits(0x7fc54321)
			case "NaN-then-infinity":
				first.PosVec.X, second.PosVec.X = float32(math.NaN()), float32(math.Inf(1))
			case "infinite":
				first.PosVec.X, second.PosVec.X, want = float32(math.Inf(1)), float32(math.Inf(-1)), nil
			case "infinity-minus-infinity":
				unit.PosVec.X, first.PosVec.X = float32(math.Inf(1)), float32(math.Inf(1))
				items, want = []*Object{first}, first
			case "signed-zero":
				first.PosVec, second.PosVec, want = types.Ptf(math.Float32frombits(0x80000000), 0), types.Ptf(0, 0), first
			case "subnormal":
				first.PosVec, second.PosVec, want = types.Ptf(0, math.SmallestNonzeroFloat32), types.Ptf(0, 2*math.SmallestNonzeroFloat32), first
			}
			interactions := 0
			got := monsterSearchEdible544A00(unit, 5000, monsterSearchEdibleHooks544A00{
				eachInCircle: func(pos types.Pointf, radius float32, each func(*Object) bool) {
					if math.Float32bits(pos.X) != math.Float32bits(unit.PosVec.X) ||
						math.Float32bits(pos.Y) != math.Float32bits(unit.PosVec.Y) || radius != 5000 {
						t.Fatal("enumeration arguments changed")
					}
					for _, item := range items {
						if !each(item) {
							t.Fatal("search stopped before all candidates")
						}
					}
				},
				canInteract: func(owner, item *Object, flags int) bool {
					if owner != unit || flags != 0 {
						t.Fatal("interaction arguments changed")
					}
					interactions++
					return true
				},
			})
			if got != want || interactions != len(items) {
				t.Fatalf("winner=%p want=%p; interactions=%d want=%d", got, want, interactions, len(items))
			}
		})
	}
}

func TestMonsterSearchEdible544A00FoodByteFilters(t *testing.T) {
	for _, poisoned := range []bool{false, true} {
		for _, online := range []bool{false, true} {
			for _, npc := range []bool{false, true} {
				t.Run(fmt.Sprintf("poisoned=%t/online=%t/NPC=%t", poisoned, online, npc), func(t *testing.T) {
					unit := &Object{}
					if poisoned {
						unit.Poison540 = 255
					}
					if npc {
						unit.ObjSubClass = 0x10
					}
					for bits := uint32(0); bits < 256; bits++ {
						candidate := &Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(bits), PosVec: types.Ptf(1, 2)}
						calls := 0
						got := monsterSearchEdible544A00(unit, 75, monsterSearchEdibleHooks544A00{
							eachInCircle: func(_ types.Pointf, _ float32, each func(*Object) bool) { each(candidate) },
							canInteract: func(owner, item *Object, flags int) bool {
								if owner != unit || item != candidate || flags != 0 {
									t.Fatal("filter callback arguments changed")
								}
								calls++
								return true
							},
							online: online,
						})
						eligible := bits&4 == 0 && (bits&0x80 == 0 || poisoned) &&
							(bits&8 == 0 || bits&0x10 != 0 && online && npc)
						if eligible && (got != candidate || calls != 1) || !eligible && (got != nil || calls != 0) {
							t.Fatalf("food byte=%02x winner=%p calls=%d eligible=%t", bits, got, calls, eligible)
						}
					}
				})
			}
		}
	}
}

func TestMonsterSearchEdible544A00LivePositionsAfterInteraction(t *testing.T) {
	for _, allow := range []bool{false, true} {
		t.Run(fmt.Sprintf("allow=%t", allow), func(t *testing.T) {
			unit := &Object{PosVec: types.Ptf(100, 200)}
			notFood := &Object{ObjClass: object.ClassWeapon}
			first := &Object{ObjClass: object.ClassFood, PosVec: types.Ptf(130, 200)}
			second := &Object{ObjClass: object.ClassFood, PosVec: types.Ptf(50000, 50000)}
			var calls []*Object
			got := monsterSearchEdible544A00(unit, 250, monsterSearchEdibleHooks544A00{
				eachInCircle: func(pos types.Pointf, radius float32, each func(*Object) bool) {
					if pos != types.Ptf(100, 200) || radius != 250 {
						t.Fatal("initial enumeration center changed")
					}
					for _, item := range []*Object{nil, notFood, first, second} {
						if !each(item) {
							t.Fatal("search stopped early")
						}
					}
				},
				canInteract: func(owner, item *Object, flags int) bool {
					calls = append(calls, item)
					if owner != unit || flags != 0 {
						t.Fatal("interaction arguments changed")
					}
					if item == second {
						// Eligibility is cached before this call; positions are not.
						second.ObjClass, second.ObjSubClass = 0, object.SubClass(object.FoodJug)
						unit.PosVec, second.PosVec = types.Ptf(300, 400), types.Ptf(301, 402)
						return allow
					}
					return true
				},
			})
			want := first
			if allow {
				want = second
			}
			if got != want || !reflect.DeepEqual(calls, []*Object{first, second}) {
				t.Fatalf("live selection: winner=%p want=%p calls=%v", got, want, calls)
			}
		})
	}
}

func TestMonsterSearchEdible544A00NativeIdentityAndSearchReset(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	food := monsterActionTestObject50A910(t)
	food.ObjClass = object.ClassFood
	food.PosVec = types.Ptf(1, 2)
	var candidates []*Object
	hooks := monsterSearchEdibleHooks544A00{
		eachInCircle: func(_ types.Pointf, _ float32, each func(*Object) bool) {
			for _, item := range candidates {
				each(item)
			}
		},
		canInteract: func(*Object, *Object, int) bool { return true },
	}
	candidates = []*Object{food}
	if got := monsterSearchEdible544A00(unit, 250, hooks); got != food {
		t.Fatalf("native food identity=%p want=%p (address %#x)", got, food, uintptr(unsafe.Pointer(food)))
	}
	candidates = nil
	if got := monsterSearchEdible544A00(unit, 250, hooks); got != nil {
		t.Fatalf("empty subsequent search retained previous native pointer %p", got)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(food)
}
