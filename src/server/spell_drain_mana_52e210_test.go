package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func drainManaTestPlayer52E210(pos types.Pointf, mana, max uint16) (*Object, *PlayerUpdateData) {
	update := &PlayerUpdateData{ManaCur: mana, ManaMax: max, Player: &Player{}}
	return &Object{ObjClass: object.ClassPlayer, PosVec: pos, UpdateData: unsafe.Pointer(update)}, update
}

func drainManaTestObelisk52E210(pos types.Pointf, mana int32) (*Object, *ObeliskUpdateData) {
	update := &ObeliskUpdateData{Mana: mana}
	return &Object{ObjClass: object.ClassImmobile, ObjSubClass: object.SubClass(object.OtherVisibleObelisk), PosVec: pos, UpdateData: unsafe.Pointer(update)}, update
}

func TestSpellDrainManaClosest52E610WeightedAndBlocked(t *testing.T) {
	caster, _ := drainManaTestPlayer52E210(types.Ptf(0, 0), 2, 10)
	obelisk, _ := drainManaTestObelisk52E210(types.Ptf(4, 0), 5)
	player, _ := drainManaTestPlayer52E210(types.Ptf(5, 0), 6, 10)
	blocked := false
	rt := SpellDrainManaRuntime52E210{
		Balance: func(key string) float32 {
			if key != "ManaDrainRange" {
				t.Fatalf("balance key %q", key)
			}
			return 10
		},
		ObjectsInCircle: func(_ types.Pointf, _ float32, visit func(*Object) bool) {
			visit(obelisk)
			visit(player)
		},
		IsEnemy:  func(source, target *Object) bool { return source == caster && target == player },
		SameTeam: func(*Object, *Object) bool { return false },
		TraceRay: func(_ types.Pointf, to types.Pointf, flags MapTraceFlags) bool {
			if flags != 5 {
				t.Fatalf("trace flags %d", flags)
			}
			return !blocked || to != player.PosVec
		},
	}
	if got := spellDrainManaClosest52E610(caster.PosVec, caster, rt); got != player {
		t.Fatalf("weighted target = %p, want player %p", got, player)
	}
	blocked = true
	if got := spellDrainManaClosest52E610(caster.PosVec, caster, rt); got != obelisk {
		t.Fatalf("blocked target = %p, want obelisk %p", got, obelisk)
	}
}

func TestSpellDrainManaUpdate52E210RayAndTransfer(t *testing.T) {
	caster, mana := drainManaTestPlayer52E210(types.Ptf(2, 3), 3, 10)
	target, supply := drainManaTestObelisk52E210(types.Ptf(5, 3), 7)
	record := &DurSpell{Caster16: caster, Pos: caster.PosVec, Level: 2}
	var previous *Object
	var available = true
	var events []string
	rt := SpellDrainManaRuntime52E210{
		Frame:    func() uint32 { return 15 },
		TickRate: func() uint32 { return 30 },
		Balance: func(key string) float32 {
			if key != "ManaDrainRange" {
				t.Fatalf("balance key %q", key)
			}
			return 20
		},
		BalanceLevel: func(key string, level uint32) float32 {
			if key != "ManaDrainCoeff" || level != 1 {
				t.Fatalf("coefficient %q/%d", key, level)
			}
			return 2.5
		},
		ObjectsInCircle: func(_ types.Pointf, _ float32, visit func(*Object) bool) {
			if available {
				visit(target)
			}
		},
		IsEnemy:   func(*Object, *Object) bool { return true },
		SameTeam:  func(*Object, *Object) bool { return false },
		TraceRay:  func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
		QuestMode: func() bool { return false },
		AddMana: func(got *Object, amount int16) {
			if got != caster {
				t.Fatalf("mana receiver = %p", got)
			}
			mana.ManaCur += uint16(amount)
			events = append(events, "add")
		},
		StartRay: func(got *DurSpell) {
			if got != record || got.Target48 != target {
				t.Fatal("ray starts with wrong target")
			}
			events = append(events, "start")
		},
		StopRay: func(got *DurSpell, old *Object) {
			if got != record || old != target {
				t.Fatal("ray stops with wrong target")
			}
			events = append(events, "stop")
		},
		Audio: func(id uint16, _ *Object) {
			if id != 230 && id != 229 {
				t.Fatalf("audio %d", id)
			}
			events = append(events, "audio")
		},
		LoadRayTarget:  func(*DurSpell) *Object { return previous },
		StoreRayTarget: func(_ *DurSpell, obj *Object) { previous = obj },
	}
	if got := SpellDrainManaUpdate52E210(record, rt); got != 0 || previous != target ||
		mana.ManaCur != 5 || supply.Mana != 5 || record.Field72 != int32(math.Float32bits(0.5)) ||
		!reflect.DeepEqual(events, []string{"start", "add", "audio", "audio"}) {
		t.Fatalf("update = %d, previous %p, mana %d, supply %d, fraction %#x, events %v", got, previous, mana.ManaCur, supply.Mana, uint32(record.Field72), events)
	}
	available = false
	events = nil
	if got := SpellDrainManaUpdate52E210(record, rt); got != 1 || previous != nil || record.Target48 != nil ||
		!reflect.DeepEqual(events, []string{"stop"}) {
		t.Fatalf("no target = %d, previous %p, target %p, events %v", got, previous, record.Target48, events)
	}
}

func TestSpellDrainManaUpdate52E210GlyphDrainsPlayer(t *testing.T) {
	target, mana := drainManaTestPlayer52E210(types.Ptf(2, 2), 20, 30)
	record := &DurSpell{Flag20: 1, Pos: types.Ptf(0, 0)}
	var amount int32
	rt := SpellDrainManaRuntime52E210{
		Balance:         func(string) float32 { return 10 },
		ObjectsInCircle: func(_ types.Pointf, _ float32, visit func(*Object) bool) { visit(target) },
		SameTeam:        func(*Object, *Object) bool { return false },
		TraceRay:        func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
		SubMana: func(got *Object, value int32) {
			if got != target {
				t.Fatalf("glyph target = %p", got)
			}
			amount = value
			mana.ManaCur = 0
		},
	}
	if got := SpellDrainManaUpdate52E210(record, rt); got != 1 || amount != 50 || mana.ManaCur != 0 {
		t.Fatalf("glyph = %d, amount %d, mana %d", got, amount, mana.ManaCur)
	}
}

func TestSpellDrainManaTransfer52E450QuestObelisk(t *testing.T) {
	caster, mana := drainManaTestPlayer52E210(types.Ptf(0, 0), 1, 10)
	target, supply := drainManaTestObelisk52E210(types.Ptf(2, 0), 3)
	rt := SpellDrainManaRuntime52E210{
		QuestMode:      func() bool { return true },
		QuestManaScale: func(*Object) float32 { return 1.5 },
		AddMana:        func(_ *Object, amount int16) { mana.ManaCur += uint16(amount) },
	}
	if !spellDrainManaTransfer52E450(caster, target, 2, rt) || mana.ManaCur != 4 || supply.Mana != 3 {
		t.Fatalf("quest transfer = mana %d, supply %d", mana.ManaCur, supply.Mana)
	}
}
