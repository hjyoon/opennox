package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func playerQuestSpawnTestPlayerUnit4E8210(gate *Object, field76 uint32, questExit *Object) (*Object, *PlayerUpdateData) {
	update := &PlayerUpdateData{Field76: field76, SoulGate: gate, QuestExit: questExit}
	return &Object{UpdateData: unsafe.Pointer(update)}, update
}

func TestPlayerQuestSpawnNative4E8210UsesNamedFields(t *testing.T) {
	data1 := &SoulGateCollideData{LastUsedFrame: 3}
	data2 := &SoulGateCollideData{LastUsedFrame: 0x80000000}
	gate1 := &Object{PosVec: types.Pointf{X: 1, Y: 2}, CollideData: unsafe.Pointer(data1)}
	gate2 := &Object{PosVec: types.Pointf{
		X: math.Float32frombits(0x80000000),
		Y: math.Float32frombits(0x7fa12345),
	}, CollideData: unsafe.Pointer(data2)}
	unit1, _ := playerQuestSpawnTestPlayerUnit4E8210(gate1, 0, nil)
	unit2, _ := playerQuestSpawnTestPlayerUnit4E8210(gate2, 0, nil)
	questExit := &Object{}
	joining, joiningUpdate := playerQuestSpawnTestPlayerUnit4E8210(gate1, 0x11223344, questExit)
	next := map[*Object]*Object{unit1: unit2}
	want := types.Pointf{X: math.Float32frombits(0x80000001), Y: math.Float32frombits(0x7fc12345)}
	got, ok := playerQuestSpawnNative4E8210(joining, playerQuestSpawnNativeDeps4E8210{
		firstUnit: func() *Object { return unit1 },
		nextUnit:  func(unit *Object) *Object { return next[unit] },
		randomReachablePos: func(radius float32, pos types.Pointf) types.Pointf {
			if math.Float32bits(radius) != 0x42700000 {
				t.Fatalf("radius bits = %#x", math.Float32bits(radius))
			}
			if math.Float32bits(pos.X) != math.Float32bits(gate2.PosVec.X) || math.Float32bits(pos.Y) != math.Float32bits(gate2.PosVec.Y) {
				t.Fatalf("source position bits = (%#x, %#x), want (%#x, %#x)",
					math.Float32bits(pos.X), math.Float32bits(pos.Y), math.Float32bits(gate2.PosVec.X), math.Float32bits(gate2.PosVec.Y))
			}
			return want
		},
	})
	if !ok || math.Float32bits(got.X) != math.Float32bits(want.X) || math.Float32bits(got.Y) != math.Float32bits(want.Y) {
		t.Fatalf("result = (%#v, %v), want (%#v, true)", got, ok, want)
	}
	if joiningUpdate.SoulGate != gate2 || joiningUpdate.Field76 != 0x11223344 || joiningUpdate.QuestExit != questExit {
		t.Fatalf("joining update = %#v, want selected gate with adjacent fields preserved", joiningUpdate)
	}
	if data1.LastUsedFrame != 3 || data2.LastUsedFrame != 0x80000000 {
		t.Fatalf("SoulGate frames changed: %#v %#v", data1, data2)
	}
}

func TestPlayerQuestSpawnNative4E8210FailureSkipsJoiningUpdate(t *testing.T) {
	got, ok := playerQuestSpawnNative4E8210(nil, playerQuestSpawnNativeDeps4E8210{
		firstUnit: func() *Object { return nil },
		nextUnit: func(*Object) *Object {
			t.Fatal("empty traversal called next")
			return nil
		},
		randomReachablePos: func(float32, types.Pointf) types.Pointf {
			t.Fatal("empty traversal searched a point")
			return types.Pointf{}
		},
	})
	if ok || got != (types.Pointf{}) {
		t.Fatalf("result = (%#v, %v), want zero/false", got, ok)
	}
}

func TestPlayerQuestSpawnNative4E8210NilCollideDataFaults(t *testing.T) {
	unit, _ := playerQuestSpawnTestPlayerUnit4E8210(&Object{}, 0, nil)
	defer func() {
		if recover() == nil {
			t.Fatal("nil native SoulGate collide data did not fault")
		}
	}()
	playerQuestSpawnNative4E8210(unit, playerQuestSpawnNativeDeps4E8210{
		firstUnit:          func() *Object { return unit },
		nextUnit:           func(*Object) *Object { return nil },
		randomReachablePos: func(float32, types.Pointf) types.Pointf { return types.Pointf{} },
	})
}

func TestPlayerQuestSpawnNative4E8210DefersJoiningFaultUntilAfterTraversal(t *testing.T) {
	data := &SoulGateCollideData{LastUsedFrame: 1}
	gate := &Object{CollideData: unsafe.Pointer(data)}
	unit, _ := playerQuestSpawnTestPlayerUnit4E8210(gate, 0, nil)
	nextCalls := 0
	randomCalls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil joining player did not fault after selecting a SoulGate")
		}
		if nextCalls != 1 || randomCalls != 0 {
			t.Fatalf("next/random calls = (%d, %d), want (1, 0)", nextCalls, randomCalls)
		}
	}()
	playerQuestSpawnNative4E8210(nil, playerQuestSpawnNativeDeps4E8210{
		firstUnit: func() *Object { return unit },
		nextUnit: func(*Object) *Object {
			nextCalls++
			return nil
		},
		randomReachablePos: func(float32, types.Pointf) types.Pointf {
			randomCalls++
			return types.Pointf{}
		},
	})
}

func TestServerSub4E8210EmptyPlayerListDoesNotInspectJoiningPlayer(t *testing.T) {
	s := &Server{}
	got, ok := s.Sub4E8210(nil)
	if ok || got != (types.Pointf{}) {
		t.Fatalf("result = (%#v, %v), want zero/false", got, ok)
	}
}

func TestPlayerQuestSpawnNativeLayout4E8210(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantSoulGate := uintptr(308)
	wantPos := uintptr(56)
	wantCollideData := uintptr(700)
	if ptrSize == 8 {
		wantSoulGate = 376
		wantPos = 60
		wantCollideData = 776
	}
	if got := unsafe.Offsetof(PlayerUpdateData{}.SoulGate); got != wantSoulGate {
		t.Fatalf("PlayerUpdateData.SoulGate offset = %d, want %d", got, wantSoulGate)
	}
	if got := unsafe.Offsetof(Object{}.PosVec); got != wantPos {
		t.Fatalf("Object.PosVec offset = %d, want %d", got, wantPos)
	}
	if got := unsafe.Offsetof(Object{}.CollideData); got != wantCollideData {
		t.Fatalf("Object.CollideData offset = %d, want %d", got, wantCollideData)
	}
	if got := unsafe.Sizeof(SoulGateCollideData{}); got != 4 {
		t.Fatalf("SoulGateCollideData size = %d, want 4", got)
	}
	if got := unsafe.Offsetof(SoulGateCollideData{}.LastUsedFrame); got != 0 {
		t.Fatalf("SoulGateCollideData.LastUsedFrame offset = %d, want 0", got)
	}
}
