package server

import (
	"math"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSpellManaBombCreate530F90PlayerStateAndNativeCharge(t *testing.T) {
	caster := &Object{
		ObjClass: object.ClassPlayer,
		Mass:     17.5,
		VelVec:   types.Ptf(1, 2), ForceVec: types.Ptf(3, 4), Pos24: types.Ptf(5, 6),
	}
	charge := &Object{}
	record := &DurSpell{Caster16: caster, Level: 3, Pos: types.Ptf(100, 200), Field76: ^uintptr(0), Field84: 9}
	var gotCharge *Object
	var buffs []EnchantID
	runtime := SpellManaBombRuntime530F90{
		BalanceLevel: func(key string, level uint32) float32 {
			if key != "ManaBombInitPower" || level != 2 {
				t.Fatalf("balance level = %s/%d", key, level)
			}
			return 10.5 // x87 round-to-even
		},
		TickRate: func() uint32 { return 60 },
		ApplyBuff: func(got *Object, buff EnchantID, duration int16, power int8) {
			if got != caster || duration != 600 || power != 5 {
				t.Fatalf("buff = %p/%d/%d/%d", got, buff, duration, power)
			}
			buffs = append(buffs, buff)
		},
		NewObject: func(id string) *Object {
			if id != "ManaBombCharge" {
				t.Fatalf("object ID = %q", id)
			}
			return charge
		},
		CreateAt: func(got, owner *Object, pos types.Pointf) {
			if got != charge || owner != nil || pos != record.Pos {
				t.Fatalf("charge = %p/%p/%v", got, owner, pos)
			}
		},
		StoreCharge: func(got *DurSpell, visual *Object) {
			if got != record {
				t.Fatalf("record = %p", got)
			}
			gotCharge = visual
		},
	}
	if got := SpellManaBombCreate530F90(record, runtime); got != 0 {
		t.Fatalf("create = %d", got)
	}
	if gotCharge != charge || !reflect.DeepEqual(buffs, []EnchantID{5, 14, 29}) ||
		record.Field72 != 10 || record.Field76 != 0 || record.Field80 != math.Float32bits(17.5) || record.Field84 != 0 ||
		math.Float32bits(caster.Mass) != 1203982323 || caster.VelVec != (types.Pointf{}) ||
		caster.ForceVec != (types.Pointf{}) || caster.Pos24 != (types.Pointf{}) {
		t.Fatalf("create state = charge %p, buffs %v, record %+v, caster %+v", gotCharge, buffs, record, caster)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(charge)) <= uintptr(^uint32(0)) {
		t.Fatal("expected charge pointer above 4 GiB")
	}
}

func TestSpellManaBombCreate530F90GlyphDeadline(t *testing.T) {
	record := &DurSpell{Flag20: 1, Level: 1, Field76: ^uintptr(0), Field84: 7}
	runtime := SpellManaBombRuntime530F90{
		BalanceLevel: func(key string, level uint32) float32 { return 2.5 },
		Balance: func(key string) float32 {
			if key != "ManaBombGlyphDuration" {
				t.Fatalf("key = %q", key)
			}
			return 12.5
		},
		Frame:     func() uint32 { return 100 },
		NewObject: func(string) *Object { return nil },
	}
	if got := SpellManaBombCreate530F90(record, runtime); got != 0 ||
		record.Field72 != 2 || record.Frame68 != 112 || record.Field76 != 0 || record.Field84 != 0 {
		t.Fatalf("glyph = %d/%+v", got, record)
	}
}

func TestSpellManaBombUpdate5310C0PlayerChargeAndExplosion(t *testing.T) {
	update := &PlayerUpdateData{ManaCur: 20}
	caster := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), PosVec: types.Ptf(21, 34)}
	charge := &Object{}
	record := &DurSpell{Caster16: caster, Level: 2, Frame68: 100, Field72: 10, Pos: types.Ptf(99, 88)}
	frame := uint32(10)
	stored := charge
	var events []string
	runtime := SpellManaBombRuntime530F90{
		Frame: func() uint32 { return frame },
		LoadCharge: func(got *DurSpell) *Object {
			if got != record {
				t.Fatal("wrong record")
			}
			return stored
		},
		StoreCharge: func(_ *DurSpell, got *Object) { stored = got; events = append(events, "store") },
		DelayedDelete: func(got *Object) {
			if got != charge {
				t.Fatal("wrong charge")
			}
			events = append(events, "delete")
		},
		BalanceLevel: func(key string, level uint32) float32 {
			if key != "ManaBombDeltaPower" || level != 1 {
				t.Fatalf("level balance = %s/%d", key, level)
			}
			return 3.5
		},
		ManaSub: func(got *Object, amount int32) {
			if got != caster || amount != 2 {
				t.Fatalf("mana sub = %p/%d", got, amount)
			}
			events = append(events, "mana")
		},
		Balance: func(key string) float32 {
			switch key {
			case "ManaBombOutRadius":
				return 60
			case "ManaBombInRadius":
				return 20
			case "ManaBombShakeMag":
				return 4.5
			default:
				t.Fatalf("balance = %q", key)
				return 0
			}
		},
		DamageAround: func(pos types.Pointf, outer, inner float32, damage int, got *Object) {
			if pos != caster.PosVec || outer != 60 || inner != 20 || damage != 18 || got != caster {
				t.Fatalf("damage = %v/%v/%v/%d/%p", pos, outer, inner, damage, got)
			}
			events = append(events, "damage")
		},
		Earthquake: func(pos types.Pointf, magnitude int) {
			if pos != caster.PosVec || magnitude != 4 {
				t.Fatalf("earthquake = %v/%d", pos, magnitude)
			}
			events = append(events, "earthquake")
		},
		PointFX: func(effect netmsg.Op, pos types.Pointf) {
			if pos != caster.PosVec {
				t.Fatalf("FX pos = %v", pos)
			}
			switch effect {
			case 129:
				events = append(events, "fx129")
			case 154:
				events = append(events, "fx154")
			default:
				t.Fatalf("FX = %d", effect)
			}
		},
		Audio: func(got *Object) {
			if got != caster {
				t.Fatal("wrong audio object")
			}
			events = append(events, "audio")
		},
	}
	if got := SpellManaBombUpdate5310C0(record, runtime); got != 0 || record.Field72 != 14 || stored != charge {
		t.Fatalf("charging = %d/%d/%p", got, record.Field72, stored)
	}
	update.ManaCur = 14
	frame++
	if got := SpellManaBombUpdate5310C0(record, runtime); got != 0 || record.Field72 != 18 || stored != nil {
		t.Fatalf("charge removed = %d/%d/%p", got, record.Field72, stored)
	}
	update.ManaCur = 0
	frame++
	if got := SpellManaBombUpdate5310C0(record, runtime); got != 1 || record.Field84 != 1 {
		t.Fatalf("explosion = %d/%d", got, record.Field84)
	}
	want := []string{"mana", "delete", "store", "mana", "damage", "earthquake", "fx129", "fx154", "audio"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpellManaBombUpdate5310C0GlyphAndOrphan(t *testing.T) {
	record := &DurSpell{Flag20: 1, Frame68: 20, Pos: types.Ptf(40, 50), Field72: 7}
	charge := &Object{}
	stored := charge
	frame := uint32(11)
	var deleted, damaged int
	runtime := SpellManaBombRuntime530F90{
		Frame:         func() uint32 { return frame },
		LoadCharge:    func(*DurSpell) *Object { return stored },
		StoreCharge:   func(_ *DurSpell, value *Object) { stored = value },
		DelayedDelete: func(*Object) { deleted++ },
		BalanceLevel:  func(string, uint32) float32 { return 1 },
		Balance:       func(string) float32 { return 1 },
		DamageAround: func(pos types.Pointf, _, _ float32, damage int, caster *Object) {
			if pos != record.Pos || damage != 8 || caster != nil {
				t.Fatalf("glyph damage = %v/%d/%p", pos, damage, caster)
			}
			damaged++
		},
		Earthquake: func(types.Pointf, int) {},
		PointFX:    func(netmsg.Op, types.Pointf) {},
		Audio:      func(*Object) {},
	}
	if got := SpellManaBombUpdate5310C0(record, runtime); got != 0 || deleted != 1 || stored != nil || record.Field72 != 8 {
		t.Fatalf("glyph charge = %d/%d/%p/%d", got, deleted, stored, record.Field72)
	}
	frame = 19
	if got := SpellManaBombUpdate5310C0(record, runtime); got != 1 || damaged != 1 {
		t.Fatalf("glyph explosion = %d/%d", got, damaged)
	}
	record.Flag20 = 0
	if got := SpellManaBombUpdate5310C0(record, SpellManaBombRuntime530F90{}); got != 1 {
		t.Fatalf("orphan = %d", got)
	}
}

func TestSpellManaBombDestroy531290RestoresPlayerAndCancels(t *testing.T) {
	caster := &Object{ObjClass: object.ClassPlayer, Mass: 99}
	charge := &Object{}
	record := &DurSpell{Caster16: caster, Pos: types.Ptf(1, 2), Field80: math.Float32bits(17.5)}
	stored := charge
	var events []string
	runtime := SpellManaBombRuntime530F90{
		LoadCharge:  func(*DurSpell) *Object { return stored },
		StoreCharge: func(_ *DurSpell, got *Object) { stored = got; events = append(events, "store") },
		DelayedDelete: func(got *Object) {
			if got != charge {
				t.Fatal("wrong charge")
			}
			events = append(events, "delete")
		},
		BuffOff: func(got *Object, buff EnchantID) {
			if got != caster {
				t.Fatal("wrong caster")
			}
			events = append(events, strconv.Itoa(int(buff)))
		},
		PointFX: func(effect netmsg.Op, pos types.Pointf) {
			if effect != 163 || pos != record.Pos {
				t.Fatalf("FX = %d/%v", effect, pos)
			}
			events = append(events, "cancel")
		},
	}
	SpellManaBombDestroy531290(record, runtime)
	if stored != nil || caster.Mass != 17.5 || !reflect.DeepEqual(events, []string{"delete", "store", "5", "14", "29", "cancel"}) {
		t.Fatalf("destroy = %p/%v/%v", stored, caster.Mass, events)
	}
	events = nil
	record.Flag20 = 1
	record.Field84 = 1
	SpellManaBombDestroy531290(record, runtime)
	if !reflect.DeepEqual(events, []string{"store"}) {
		t.Fatalf("exploded glyph destroy = %v", events)
	}
}
