package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
)

func TestSpellFirewalkUpdate52ED40NativeWidthAndMovement(t *testing.T) {
	s := &Server{}
	s.Rand.Logic = prand.New(0)
	target := &Object{PosVec: types.Ptf(30, 40)}
	target.Shape.Circle.R = 2
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
		t.Fatal("expected target above 4 GiB")
	}
	record := &DurSpell{Target48: target, Frame60: 100, Frame64: 100, Level: 4}
	record.Pos.X = math.Float32frombits(0x3fdccccc) // PE32's false target pointer.
	var flames []types.Pointf
	runtime := SpellFirewalkRuntime52ED40{SpawnFlame: func(kind int, position types.Pointf) {
		if kind < 0 || kind > 2 {
			t.Fatalf("flame kind = %d", kind)
		}
		flames = append(flames, position)
	}}
	if got := s.SpellFirewalkUpdate52ED40(record, runtime); got != 0 {
		t.Fatalf("first update = %d", got)
	}
	if record.Frame64 != 101 || math.Float32frombits(uint32(record.Field72)) != 30 ||
		math.Float32frombits(uint32(record.Field76)) != 40 || record.Field80 != math.Float32bits(30) ||
		record.Field84 != math.Float32bits(40) || len(flames) != 0 {
		t.Fatalf("first tick state = frame %d, previous (%g,%g), anchor (%g,%g), flames %v",
			record.Frame64, math.Float32frombits(uint32(record.Field72)), math.Float32frombits(uint32(record.Field76)),
			math.Float32frombits(record.Field80), math.Float32frombits(record.Field84), flames)
	}
	target.PosVec = types.Ptf(70, 80)
	if got := s.SpellFirewalkUpdate52ED40(record, runtime); got != 0 {
		t.Fatalf("second update = %d", got)
	}
	target.PosVec = types.Ptf(110, 120)
	if got := s.SpellFirewalkUpdate52ED40(record, runtime); got != 0 {
		t.Fatalf("third update = %d", got)
	}
	if len(flames) != 4 || flames[0] != (types.Ptf(30, 40)) || flames[1] != (types.Ptf(30, 40)) ||
		flames[2] != (types.Ptf(70, 80)) || flames[3] != (types.Ptf(50, 60)) {
		t.Fatalf("flame positions = %v", flames)
	}
	if math.Float32frombits(uint32(record.Field72)) != 110 || math.Float32frombits(uint32(record.Field76)) != 120 ||
		math.Float32frombits(record.Field80) != 70 || math.Float32frombits(record.Field84) != 80 {
		t.Fatalf("final position state = previous (%g,%g), anchor (%g,%g)",
			math.Float32frombits(uint32(record.Field72)), math.Float32frombits(uint32(record.Field76)),
			math.Float32frombits(record.Field80), math.Float32frombits(record.Field84))
	}
}

func TestSpellFirewalkUpdate52ED40StopsWithoutTargetOrWithDestroyedTarget(t *testing.T) {
	s := &Server{}
	record := &DurSpell{Frame60: 7, Frame64: 7}
	if got := s.SpellFirewalkUpdate52ED40(record, SpellFirewalkRuntime52ED40{}); got != 1 {
		t.Fatalf("nil target update = %d", got)
	}
	record.Target48 = &Object{ObjFlags: object.Flags(0x8020)}
	if got := s.SpellFirewalkUpdate52ED40(record, SpellFirewalkRuntime52ED40{}); got != 1 {
		t.Fatalf("destroyed target update = %d", got)
	}
	if record.Frame64 != 7 {
		t.Fatalf("stopped callback changed frame to %d", record.Frame64)
	}
}

func TestSpellFirewalkUpdate52ED40RadiusThreshold(t *testing.T) {
	s := &Server{}
	target := &Object{PosVec: types.Ptf(12, 16)}
	target.Shape.Circle.R = 5
	record := &DurSpell{Target48: target, Frame60: 1, Frame64: 2,
		Field72: int32(math.Float32bits(0)), Field76: uintptr(math.Float32bits(0))}
	if got := s.SpellFirewalkUpdate52ED40(record, SpellFirewalkRuntime52ED40{}); got != 0 {
		t.Fatalf("at threshold update = %d", got)
	}
	if record.Field72 != 0 || record.Field76 != 0 {
		t.Fatal("below-threshold update advanced the flame position")
	}
}
