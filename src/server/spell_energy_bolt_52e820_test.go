package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSpellEnergyBoltCreate52E820NativePointers(t *testing.T) {
	caster := &Object{ObjClass: object.ClassMonster}
	record := &DurSpell{Caster16: caster, Pos: types.Ptf(7, 9)}
	var events []string
	rt := SpellEnergyBoltRuntime52E820{
		CancelSpell: func(id int32, got *Object) {
			if id != 43 || got != caster {
				t.Fatalf("cancel = %d/%p", id, got)
			}
			events = append(events, "cancel")
		},
		PointFX: func(code uint8, pos types.Pointf) {
			if code != 130 || pos != record.Pos {
				t.Fatalf("effect = %d/%v", code, pos)
			}
			events = append(events, "effect")
		},
	}
	if got := SpellEnergyBoltCreate52E820(record, rt); got != 0 || !reflect.DeepEqual(events, []string{"cancel", "effect"}) {
		t.Fatalf("create = %d, events %v", got, events)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(caster))>>32 == 0 {
		t.Skip("allocator did not place the object above the PE32 address range")
	}
}

func TestSpellEnergyBoltUpdate52E850RayTransitions(t *testing.T) {
	caster := &Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(0, 0)}
	a := &Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(3, 0)}
	b := &Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(4, 0)}
	record := &DurSpell{Caster16: caster, Level: 2, Frame60: 1}
	var previous *Object
	var found *Object
	var events []string
	rt := SpellEnergyBoltRuntime52E820{
		Frame:    func() uint32 { return 2 },
		TickRate: func() uint32 { return 30 },
		Balance: func(key string) float32 {
			switch key {
			case "LightningRange":
				return 20
			case "LightningSearchTime":
				return 5
			}
			t.Fatalf("unexpected balance key %q", key)
			return 0
		},
		BalanceLevel: func(key string, level uint32) float32 {
			if key != "EnergyBoltDamage" || level != 1 {
				t.Fatalf("damage lookup = %q/%d", key, level)
			}
			return 1.75
		},
		ObjectsInCircle: func(_ types.Pointf, _ float32, visit func(*Object) bool) {
			if found != nil {
				visit(found)
			}
		},
		CanInteract:   func(*Object, *Object) bool { return true },
		IsEnemy:       func(*Object, *Object) bool { return true },
		InFront:       func(*Object, *Object) bool { return true },
		PositionDelta: func(*Object, *types.Pointf) int32 { return 0 },
		StartRay:      func(*DurSpell) { events = append(events, "start") },
		StopRay: func(_ *DurSpell, target *Object) {
			if target != previous {
				t.Fatalf("stop target = %p, previous %p", target, previous)
			}
			events = append(events, "stop")
		},
		Damage: func(target, source *Object, amount int32) {
			if source != caster || target != found && target != record.Target48 {
				t.Fatalf("damage target/source = %p/%p", target, source)
			}
			events = append(events, "damage")
		},
		Audio:          func(uint16, *Object) {},
		PointFX:        func(uint8, types.Pointf) {},
		LoadRayTarget:  func(*DurSpell) *Object { return previous },
		StoreRayTarget: func(_ *DurSpell, target *Object) { previous = target },
	}
	found = a
	if got := SpellEnergyBoltUpdate52E850(record, rt); got != 0 || previous != a || record.Target48 != a ||
		record.Frame68 != 7 || record.Field72 != int32(math.Float32bits(-0.25)) ||
		!reflect.DeepEqual(events, []string{"start", "damage"}) {
		t.Fatalf("first update = %d, previous %p, target %p, frame %d, fraction %#x, events %v", got, previous, record.Target48, record.Frame68, uint32(record.Field72), events)
	}
	events = nil
	record.Target48 = nil
	found = b
	if got := SpellEnergyBoltUpdate52E850(record, rt); got != 0 || previous != b ||
		!reflect.DeepEqual(events, []string{"stop", "start", "damage"}) {
		t.Fatalf("retarget = %d, previous %p, events %v", got, previous, events)
	}
	events = nil
	record.Target48 = nil
	found = nil
	if got := SpellEnergyBoltUpdate52E850(record, rt); got != 0 || previous != nil ||
		!reflect.DeepEqual(events, []string{"stop"}) {
		t.Fatalf("lost target = %d, previous %p, events %v", got, previous, events)
	}
}

func TestSpellEnergyBoltUpdate52E850Glyph(t *testing.T) {
	near := &Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(1, 0)}
	far := &Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(4, 0)}
	record := &DurSpell{Flag20: 1, Pos: types.Ptf(0, 0)}
	var gotTarget *Object
	var gotDamage int32
	rt := SpellEnergyBoltRuntime52E820{
		Balance: func(key string) float32 {
			if key == "LightningRange" {
				return 10
			}
			if key == "EnergyBoltGlyphDamage" {
				return 3.5
			}
			t.Fatalf("unexpected balance key %q", key)
			return 0
		},
		ObjectsInCircle: func(_ types.Pointf, _ float32, visit func(*Object) bool) {
			visit(far)
			visit(near)
		},
		Damage:    func(target, _ *Object, amount int32) { gotTarget, gotDamage = target, amount },
		Audio:     func(uint16, *Object) {},
		CastSound: func() uint16 { return 77 },
		PointFX:   func(uint8, types.Pointf) {},
	}
	if got := SpellEnergyBoltUpdate52E850(record, rt); got != 1 || gotTarget != near || gotDamage != 4 {
		t.Fatalf("glyph = %d, target %p, damage %d", got, gotTarget, gotDamage)
	}
}
