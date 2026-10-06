package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// GAME.EXE 00544A9E..00544ACB: two retained 53-bit differences,
// Y-square, X-square, sum, C0-only comparison and a winning FSTP DWORD.
// CRT 004031F2 and gameplay 0043E2C1 select precision 53 / ToZero.
// This reference uses neither the production residual/Nextafter helpers nor
// nearest-even arithmetic. Inputs in these tests are finite binary32 values.
func monsterEdibleReferenceSquare544A9E(unit, food types.Pointf) *big.Float {
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	input := func(v float32) *big.Float { return chop().SetFloat64(float64(v)) }
	dx := chop().Sub(input(food.X), input(unit.X))
	dy := chop().Sub(input(food.Y), input(unit.Y))
	return chop().Add(chop().Mul(dy, dy), chop().Mul(dx, dx))
}

func monsterEdibleReferenceSpill544AC5(distance *big.Float) float32 {
	if distance.Sign() == 0 {
		return 0
	}
	// Subnormals have a fixed 2^-149 quantum, not 24 significant bits.
	// A winning finite distance is below the original 10,000,000 cache,
	// so overflow is unreachable and every rounded result fits binary32.
	precision := min(24, distance.MantExp(nil)+149)
	if precision <= 0 {
		return 0
	}
	spill := new(big.Float).SetPrec(uint(precision)).SetMode(big.ToZero).Set(distance)
	result, accuracy := spill.Float32()
	if accuracy != big.Exact {
		panic("edible reference spill must be exactly binary32")
	}
	return result
}

func monsterEdibleReferenceWinner544A00(unit types.Pointf, foods []types.Pointf) int {
	best := new(big.Float).SetPrec(53).SetInt64(10000000)
	winner := -1
	for i, food := range foods {
		distance := monsterEdibleReferenceSquare544A9E(unit, food)
		if distance.Cmp(best) < 0 {
			best.SetFloat64(float64(monsterEdibleReferenceSpill544AC5(distance)))
			winner = i
		}
	}
	return winner
}

func TestMonsterSearchEdible544A00IndependentChopCalibration(t *testing.T) {
	for _, tc := range []struct {
		food types.Pointf
		bits uint32
	}{
		{types.Ptf(1, 0.0003), 0x3f800000},
		{types.Ptf(-1, -0.0003), 0x3f800000},
		{types.Ptf(0.0003, 1), 0x3f800000},
		{types.Ptf(3, 4), 0x41c80000},
		{types.Ptf(0, math.Float32frombits(1)), 0},
		{types.Ptf(0, 0x1p-74), 2},
		{types.Ptf(0, 0x1p-75), 0},
	} {
		if got := math.Float32bits(monsterEdibleReferenceSpill544AC5(monsterEdibleReferenceSquare544A9E(types.Pointf{}, tc.food))); got != tc.bits {
			t.Fatalf("independent FSTP for %v: %08x want %08x", tc.food, got, tc.bits)
		}
	}
	if got := monsterEdibleReferenceWinner544A00(types.Pointf{}, []types.Pointf{types.Ptf(1, 0.0003), types.Ptf(1, 0.0003)}); got != 0 {
		t.Fatalf("original cached rounded tie selected %d, want first", got)
	}
}

func monsterEdibleNativeFoods544A00(t *testing.T, count int) []*Object {
	t.Helper()
	foods := make([]*Object, count)
	for i := range foods {
		food, free := alloc.New(Object{})
		t.Cleanup(free)
		*food = Object{ObjClass: object.ClassFood, ObjSubClass: object.SubClass(object.FoodApple),
			ObjFlags: object.FlagActive | object.FlagEnabled}
		foods[i] = food
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(food)) <= math.MaxUint32 {
			t.Fatalf("C-owned food below 4 GiB: %p", food)
		}
	}
	return foods
}

func monsterEdibleAssertNativeWinner544A00(t *testing.T, unit *Object, foods []*Object, positions []types.Pointf) {
	t.Helper()
	wantIndex := monsterEdibleReferenceWinner544A00(unit.PosVec, positions)
	var want *Object
	if wantIndex >= 0 {
		want = foods[wantIndex]
	}
	unitBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	updateBefore := monsterMoveSelectionBytes50D3B0(unit.UpdateData, unsafe.Sizeof(MonsterUpdateData{}))
	foodBefore := make([][]byte, len(positions))
	for i, pos := range positions {
		foods[i].PosVec = pos
		foodBefore[i] = monsterMoveSelectionBytes50D3B0(unsafe.Pointer(foods[i]), unsafe.Sizeof(*foods[i]))
	}
	visits, calls := 0, 0
	got := monsterSearchEdible544A00(unit, 250, monsterSearchEdibleHooks544A00{
		eachInCircle: func(pos types.Pointf, radius float32, each func(*Object) bool) {
			if pos != unit.PosVec || radius != 250 {
				t.Fatal("search changed the enumeration center or radius")
			}
			for i := range positions {
				visits++
				if !each(foods[i]) {
					t.Fatal("search stopped before the original last candidate")
				}
			}
		},
		canInteract: func(owner, candidate *Object, flags int) bool {
			if owner != unit || candidate != foods[calls] || flags != 0 {
				t.Fatal("original candidate/interaction order changed")
			}
			calls++
			return true
		},
		online: true,
	})
	if got != want || calls != len(positions) || visits != len(positions) {
		t.Errorf("food winner=%p want=%p (original index %d) unit=%v foods=%v visits/calls=%d/%d",
			got, want, wantIndex, unit.PosVec, positions, visits, calls)
	}
	if !bytes.Equal(unitBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
		!bytes.Equal(updateBefore, monsterMoveSelectionBytes50D3B0(unit.UpdateData, unsafe.Sizeof(MonsterUpdateData{}))) {
		t.Fatal("read-only food search changed unit/update state")
	}
	for i := range positions {
		if !bytes.Equal(foodBefore[i], monsterMoveSelectionBytes50D3B0(unsafe.Pointer(foods[i]), unsafe.Sizeof(*foods[i]))) {
			t.Fatalf("read-only search changed candidate %d", i)
		}
	}
}

func TestMonsterSearchEdible544A00NativeChopWinner(t *testing.T) {
	for _, tc := range []struct {
		name  string
		unit  types.Pointf
		foods []types.Pointf
	}{
		{"rounded-tie", types.Pointf{}, []types.Pointf{types.Ptf(1, 0.0003), types.Ptf(1, 0.0003)}},
		{"rounded-cache-not-geometric-nearest", types.Pointf{}, []types.Pointf{types.Ptf(1, 0.0003), types.Ptf(1, 0.0001)}},
		{"opposite-signs", types.Pointf{}, []types.Pointf{types.Ptf(-1, -0.0003), types.Ptf(1, 0.00034)}},
		{"swapped-coordinates", types.Pointf{}, []types.Pointf{types.Ptf(0.0003, 1), types.Ptf(0.00034, 1)}},
		{"map-rounded-tie", types.Ptf(1000, 1000), []types.Pointf{types.Ptf(1001, 1000.0003), types.Ptf(1001, 1000.0003)}},
		{"map-rounded-cache", types.Ptf(1000, 1000), []types.Pointf{types.Ptf(1001, 1000.0003), types.Ptf(1001, 1000.0002)}},
		{"strictly-better-still-replaces", types.Ptf(1000, 1000), []types.Pointf{types.Ptf(1001, 1000.0003), types.Ptf(1000.5, 1000)}},
		{"original-limit-tie", types.Pointf{}, []types.Pointf{types.Ptf(1000, 3000), types.Ptf(3000, 1000)}},
		{"retained-subtraction", types.Ptf(0x1p-15, 0), []types.Pointf{types.Ptf(1000, 3000)}},
		{"subnormal-spill", types.Pointf{}, []types.Pointf{types.Ptf(0, 0x1p-74), types.Ptf(0, 0x1p-75)}},
		{"finite-largest", types.Ptf(-math.MaxFloat32, math.MaxFloat32), []types.Pointf{types.Ptf(math.MaxFloat32, -math.MaxFloat32)}},
	} {
		for _, npc := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/NPC-%t", tc.name, npc), func(t *testing.T) {
				s, unit, _ := monsterHealthRetreatNativeFixture547807(t)
				unit.PosVec = tc.unit
				if npc {
					unit.ObjSubClass = object.SubClass(object.MonsterNPC)
				}
				monsterEdibleAssertNativeWinner544A00(t, unit, monsterEdibleNativeFoods544A00(t, len(tc.foods)), tc.foods)
				if s.Rand.Logic.Index() != 1496 || s.Rand.Other.Index() != 3905 {
					t.Fatal("food search consumed game RNG")
				}
			})
		}
	}
}

func TestMonsterSearchEdible544A00NativeFiniteSelectionMatrix(t *testing.T) {
	s, unit, _ := monsterHealthRetreatNativeFixture547807(t)
	foods := monsterEdibleNativeFoods544A00(t, 4)
	// A private deterministic generator, never the game RNG or its seed.
	rng := rand.New(rand.NewSource(54400))
	for i := 0; i < 1024; i++ {
		unit.PosVec = types.Ptf(float32(rng.Intn(5000)), float32(rng.Intn(5000)))
		first := types.Ptf(unit.PosVec.X+float32(rng.Intn(128)-64), unit.PosVec.Y+float32(rng.Intn(128)-64))
		first.Y = math.Nextafter32(first.Y, float32(math.Inf(1)))
		second := first
		if i%2 == 0 {
			second.X = math.Nextafter32(second.X, unit.PosVec.X)
		}
		positions := []types.Pointf{first, second, first, unit.PosVec.Add(types.Ptf(75, 100))}
		monsterEdibleAssertNativeWinner544A00(t, unit, foods, positions)
		if t.Failed() {
			t.Fatalf("finite matrix first failing index %d", i)
		}
	}
	if s.Rand.Logic.Index() != 1496 || s.Rand.Other.Index() != 3905 {
		t.Fatal("selection matrix consumed either game RNG")
	}
}
