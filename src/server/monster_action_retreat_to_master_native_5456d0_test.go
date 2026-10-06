package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Independent GAME.EXE 00545701..0054572F register model: retain the two
// differences and Field329+30 at precision 53/ToZero, Y-square, X-square,
// FADDP, radius-square, then C0-only FCOMPP. There is no binary32 spill.
// This reference calls no production residual/Nextafter helper.
func monsterRetreatMasterReference545727(unit, owner types.Pointf, follow float32) bool {
	if math.IsNaN(float64(follow)) || math.IsNaN(float64(unit.X)) || math.IsNaN(float64(unit.Y)) ||
		math.IsNaN(float64(owner.X)) || math.IsNaN(float64(owner.Y)) {
		return true // C0 is also set on unordered.
	}
	for _, pair := range [][2]float32{{unit.X, owner.X}, {unit.Y, owner.Y}} {
		if math.IsInf(float64(pair[0]), 0) && pair[0] == pair[1] {
			return true // same-sign infinity subtraction is unordered.
		}
	}
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	input := func(v float32) *big.Float { return chop().SetFloat64(float64(v)) }
	dx := chop().Sub(input(unit.X), input(owner.X))
	dy := chop().Sub(input(unit.Y), input(owner.Y))
	radius := chop().Add(input(follow), chop().SetInt64(30))
	distance := chop().Add(chop().Mul(dy, dy), chop().Mul(dx, dx))
	return chop().Mul(radius, radius).Cmp(distance) < 0
}

func monsterRetreatMasterNativeFixture5456D0(t *testing.T) (*Server, *Object, *MonsterUpdateData, []*Object) {
	t.Helper()
	s, unit, update := monsterHealthRetreatNativeFixture547807(t)
	*unit.HealthData = HealthData{Cur: 10, Max: 100}
	*update = MonsterUpdateData{AIStackInd: 0, Field329: 20, ResumeLevel: 0.5}
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_RETREAT_TO_MASTER), Field5: 0}
	owners := make([]*Object, 3)
	for i := range owners {
		owner, free := alloc.New(Object{})
		t.Cleanup(free)
		owner.PosVec = types.Ptf(100+float32(i)*10, 200+float32(i)*10)
		owners[i] = owner
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 {
			t.Fatalf("C-owned owner below 4 GiB: %p", owner)
		}
	}
	unit.ObjOwner = owners[0]
	return s, unit, update, owners
}

func monsterRetreatMasterAssertNative5456D0(t *testing.T, s *Server, unit *Object, update *MonsterUpdateData, public bool) {
	t.Helper()
	wantMove := monsterRetreatMasterReference545727(unit.PosVec, unit.ObjOwner.PosVec, update.Field329)
	unitBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	ownerBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit.ObjOwner), unsafe.Sizeof(*unit.ObjOwner))
	before := *update
	health := *unit.HealthData
	var events []ai.ActionType
	var items []*AIStackItem
	if public {
		if !s.MonsterActionRetreatToMaster5456D0(unit) {
			t.Fatal("public native server entry rejected retreat-to-master")
		}
		for i := 1; i <= int(update.AIStackInd); i++ {
			item := &update.AIStack[i]
			events = append(events, item.Type())
			items = append(items, item)
		}
	} else if !monsterActionRetreatToMaster5456D0(unit, monsterActionRetreatToMasterHooks5456D0{
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			events = append(events, action)
			item, free := alloc.New(AIStackItem{})
			t.Cleanup(free)
			item.Action = uint32(action)
			item.SetArgs(args...)
			items = append(items, item)
			return item
		},
		pop: func() int { t.Fatal("injured owned unit unexpectedly popped"); return 0 },
	}) {
		t.Fatal("native hook entry rejected retreat-to-master")
	}
	want := []ai.ActionType(nil)
	if wantMove {
		want = []ai.ActionType{ai.DEPENDENCY_OBJECT_FARTHER_THAN, ai.ACTION_MOVE_TO}
	}
	if !reflect.DeepEqual(events, want) {
		t.Errorf("original owner-distance branch: events=%v want=%v unit=%v owner=%v follow=%g", events, want, unit.PosVec, unit.ObjOwner.PosVec, update.Field329)
	}
	for _, item := range items {
		if item.ArgObj(2) != unit.ObjOwner {
			t.Errorf("native owner identity=%p want=%p", item.ArgObj(2), unit.ObjOwner)
		}
		if item.Type() == ai.DEPENDENCY_OBJECT_FARTHER_THAN {
			if item.ArgU32(0) != math.Float32bits(update.Field329) || item.Args[1] != 0 {
				t.Errorf("original dependency args=%#v", item.Args)
			}
		} else if item.ArgU32(0) != math.Float32bits(unit.ObjOwner.PosVec.X) || item.ArgU32(1) != math.Float32bits(unit.ObjOwner.PosVec.Y) {
			t.Errorf("raw owner position not preserved: %#v", item.Args)
		}
	}
	if !bytes.Equal(unitBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
		!bytes.Equal(ownerBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit.ObjOwner), unsafe.Sizeof(*unit.ObjOwner))) || *unit.HealthData != health {
		t.Fatal("retreat scheduling changed unit, owner or health")
	}
	if public && wantMove {
		// Original successful pushes reset these progress fields. No other
		// cached update bytes, including the original head, may change.
		before.AIStackInd = 2
		before.AIStack[1], before.AIStack[2] = update.AIStack[1], update.AIStack[2]
		before.Field2, before.Field67, before.Field74, before.Field91 = 0, 0, 0, 0
		before.Field120_1, before.Field120_2, before.Field120_3 = 0, 0, 0
		before.Field124, before.Field137 = s.Frame(), s.Frame()
	}
	if !bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update)),
		monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&before), unsafe.Sizeof(before))) {
		t.Fatal("retreat scheduling changed unrelated update bytes")
	}
	if s.Rand.Logic.Index() != 1496 || s.Rand.Other.Index() != 3905 {
		t.Fatal("retreat-to-master consumed game RNG")
	}
}

func TestMonsterRetreatToMaster5456D0IndependentCalibration(t *testing.T) {
	for _, tc := range []struct {
		unit, owner types.Pointf
		follow      float32
		move        bool
	}{
		{types.Ptf(50, 0), types.Pointf{}, 20, false},
		{types.Ptf(50, 6e-7), types.Pointf{}, 20, false},
		{types.Ptf(6e-7, 50), types.Pointf{}, 20, false},
		{types.Ptf(50, 0), types.Ptf(-6e-7, 0), 20, true},
		{types.Ptf(-50, 0), types.Ptf(6e-7, 0), 20, true},
		{types.Ptf(50, 1e-6), types.Pointf{}, 20, true},
	} {
		if got := monsterRetreatMasterReference545727(tc.unit, tc.owner, tc.follow); got != tc.move {
			t.Fatalf("independent retained distance: unit=%v owner=%v follow=%g move=%v want=%v", tc.unit, tc.owner, tc.follow, got, tc.move)
		}
	}
}

func TestMonsterRetreatToMaster5456D0NativeOriginalDistance(t *testing.T) {
	for _, tc := range []struct {
		name        string
		unit, owner types.Pointf
		follow      float32
	}{
		{"inside", types.Ptf(49, 0), types.Pointf{}, 20},
		{"equal", types.Ptf(50, 0), types.Pointf{}, 20},
		{"outside", types.Ptf(math.Nextafter32(50, 100), 0), types.Pointf{}, 20},
		{"chopped-sum", types.Ptf(50, 6e-7), types.Pointf{}, 20},
		{"chopped-sum-swapped", types.Ptf(6e-7, 50), types.Pointf{}, 20},
		{"chopped-sum-negative", types.Ptf(-50, -6e-7), types.Pointf{}, 20},
		{"retained-difference", types.Ptf(50, 0), types.Ptf(-6e-7, 0), 20},
		{"retained-difference-negative", types.Ptf(-50, 0), types.Ptf(6e-7, 0), 20},
		{"retained-difference-swapped", types.Ptf(0, 50), types.Ptf(0, -6e-7), 20},
		{"above-chop", types.Ptf(50, 1e-6), types.Pointf{}, 20},
		{"map-equal", types.Ptf(4156, 2212), types.Ptf(3300, 2212), 826},
		{"map-outside", types.Ptf(4156, math.Nextafter32(2212, 3000)), types.Ptf(3300, 2212), 826},
		{"subnormal", types.Ptf(0, math.Float32frombits(1)), types.Pointf{}, -30},
		{"largest-finite", types.Ptf(math.MaxFloat32, 0), types.Ptf(-math.MaxFloat32, 0), math.MaxFloat32},
		{"qnan-unit", types.Ptf(math.Float32frombits(0x7fc12345), 0), types.Pointf{}, 20},
		{"snan-owner", types.Pointf{}, types.Ptf(0, math.Float32frombits(0x7f812345)), 20},
		{"qnan-radius", types.Pointf{}, types.Pointf{}, math.Float32frombits(0x7fc13579)},
		{"infinite-distance", types.Ptf(float32(math.Inf(1)), 0), types.Pointf{}, 20},
		{"infinite-radius", types.Ptf(50, 0), types.Pointf{}, float32(math.Inf(1))},
		{"infinite-equality", types.Ptf(float32(math.Inf(1)), 0), types.Pointf{}, float32(math.Inf(1))},
		{"unordered-difference", types.Ptf(float32(math.Inf(1)), 0), types.Ptf(float32(math.Inf(1)), 0), 20},
	} {
		for _, public := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/public=%v", tc.name, public), func(t *testing.T) {
				s, unit, update, owners := monsterRetreatMasterNativeFixture5456D0(t)
				unit.PosVec, owners[0].PosVec, update.Field329 = tc.unit, tc.owner, tc.follow
				monsterRetreatMasterAssertNative5456D0(t, s, unit, update, public)
			})
		}
	}
}

func TestMonsterRetreatToMaster5456D0NativeIndependentFiniteMatrix(t *testing.T) {
	s, unit, update, owners := monsterRetreatMasterNativeFixture5456D0(t)
	rng := rand.New(rand.NewSource(0x5456d0)) // independent, not game RNG
	for i := 0; i < 2048; i++ {
		*update = MonsterUpdateData{AIStackInd: 0, ResumeLevel: 0.5}
		update.AIStack[0].Action = uint32(ai.ACTION_RETREAT_TO_MASTER)
		unit.PosVec = types.Ptf(float32(rng.Intn(10000)-5000), float32(rng.Intn(10000)-5000))
		owners[0].PosVec = types.Ptf(float32(rng.Intn(10000)-5000), float32(rng.Intn(10000)-5000))
		update.Field329 = float32(rng.Intn(9000))
		if i%4 == 0 {
			unit.PosVec, owners[0].PosVec, update.Field329 = types.Ptf(50, float32(rng.Float64()*1e-6)), types.Pointf{}, 20
		} else if i%4 == 1 {
			unit.PosVec, owners[0].PosVec, update.Field329 = types.Ptf(50, 0), types.Ptf(float32(-rng.Float64()*1e-6), 0), 20
		}
		monsterRetreatMasterAssertNative5456D0(t, s, unit, update, i%2 == 0)
	}
}

func TestMonsterRetreatToMaster5456D0NativePushReloads(t *testing.T) {
	for _, refused := range []int{-1, 0, 1, 2} { // both, first, second, neither
		t.Run(fmt.Sprint(refused), func(t *testing.T) {
			s, unit, cached, owners := monsterRetreatMasterNativeFixture5456D0(t)
			unit.PosVec = types.Pointf{}
			live, free := alloc.New(MonsterUpdateData{})
			t.Cleanup(free)
			*live = MonsterUpdateData{Field329: 777, ResumeLevel: 0.5}
			var items [2]*AIStackItem
			for i := range items {
				item, free := alloc.New(AIStackItem{})
				t.Cleanup(free)
				*item = AIStackItem{Args: [4]uintptr{0xabc, 0xdef, 0x123, 0x456}, Field5: 0xfedcba98}
				items[i] = item
			}
			var events []ai.ActionType
			if !monsterActionRetreatToMaster5456D0(unit, monsterActionRetreatToMasterHooks5456D0{
				push: func(action ai.ActionType, args ...any) *AIStackItem {
					i := len(events)
					events = append(events, action)
					if len(args) != 0 {
						t.Errorf("arguments were evaluated before push/cancel: action=%v count=%d", action, len(args))
					}
					if i == 0 {
						cached.Field329 = math.Float32frombits(0x7fc12345)
						unit.UpdateData, unit.ObjOwner = unsafe.Pointer(live), owners[1]
					} else {
						owners[2].PosVec = types.Ptf(math.Float32frombits(0x80000000), math.Float32frombits(0x7fc54321))
						unit.ObjOwner = owners[2]
					}
					if refused == -1 || refused == i {
						return nil
					}
					items[i].Action = uint32(action)
					items[i].SetArgs(args...)
					return items[i]
				},
				pop: func() int { t.Fatal("owned injured unit popped"); return 0 },
			}) {
				t.Fatal("native callback reload fixture rejected")
			}
			if want := []ai.ActionType{ai.DEPENDENCY_OBJECT_FARTHER_THAN, ai.ACTION_MOVE_TO}; !reflect.DeepEqual(events, want) {
				t.Fatalf("push rejection changed original continuation: events=%v", events)
			}
			for i, item := range items {
				want := AIStackItem{Args: [4]uintptr{0xabc, 0xdef, 0x123, 0x456}, Field5: 0xfedcba98}
				if refused != -1 && refused != i {
					want.Action = uint32(events[i])
					if i == 0 {
						want.Args[0], want.Args[2] = uintptr(0x7fc12345), uintptr(owners[1].CObj())
					} else {
						want.Args[0], want.Args[1], want.Args[2] = uintptr(0x80000000), uintptr(0x7fc54321), uintptr(owners[2].CObj())
					}
				}
				if *item != want {
					t.Errorf("original post-push cached/live stores: item%d=%#v want=%#v", i, *item, want)
				}
			}
			if *live != (MonsterUpdateData{Field329: 777, ResumeLevel: 0.5}) || s.Rand.Logic.Index() != 1496 || s.Rand.Other.Index() != 3905 {
				t.Fatal("callback scheduling changed live update or game RNG")
			}
		})
	}
}
