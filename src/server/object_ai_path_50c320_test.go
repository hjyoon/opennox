package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestAIPathReconstruction50C320UsesExecutableInset(t *testing.T) {
	if got := math.Float32bits(aiPathReconstructionInset50C320); got != 0x40133333 {
		t.Fatalf("AI path inset bits = %#x, want 0x40133333", got)
	}

	update := new(MonsterUpdateData)
	obj := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	paths := &serverAIPaths{points: make([]types.Pointf, 4)}
	end := &AIVisitNode{X0: 10, Y2: 10}
	end.Field4 = &AIVisitNode{X0: 9, Y2: 11}

	paths.Sub_50C320(obj, end, nil)
	if got := paths.Points(); len(got) != 2 {
		t.Fatalf("reconstructed point count = %d, want 2", len(got))
	} else {
		wantBits := [][2]uint32{
			{0x4363b333, 0x437f4ccd},
			{0x43718000, 0x43718000},
		}
		for i, point := range got {
			bits := [2]uint32{math.Float32bits(point.X), math.Float32bits(point.Y)}
			if bits != wantBits[i] {
				t.Fatalf("reconstructed point %d bits = %#x, want %#x", i, bits, wantBits[i])
			}
		}
	}
}

func TestAIPathReconstruction50C320ConsumesDuplicateCellSlot(t *testing.T) {
	update := new(MonsterUpdateData)
	obj := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	stale := types.Pointf{X: 123, Y: -456}
	paths := &serverAIPaths{points: []types.Pointf{{}, stale, {}, {}}}
	end := &AIVisitNode{X0: 10, Y2: 10}
	end.Field4 = &AIVisitNode{X0: 10, Y2: 10}

	paths.Sub_50C320(obj, end, nil)
	got := paths.Points()
	if len(got) != 2 {
		t.Fatalf("duplicate-cell point count = %d, want 2", len(got))
	}
	if got[0] != stale {
		t.Fatalf("duplicate-cell preserved slot = %v, want %v", got[0], stale)
	}
	wantEnd := types.Pointf{X: 241.5, Y: 241.5}
	if got[1] != wantEnd {
		t.Fatalf("duplicate-cell endpoint = %v, want %v", got[1], wantEnd)
	}
}
