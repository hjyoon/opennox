package server

import (
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSpellMoonglowCreate531A00RejectsMissingOrDestroyedTarget(t *testing.T) {
	for _, target := range []*Object{nil, {ObjFlags: object.Flags(0x8000)}, {ObjFlags: object.Flags(0x20)}} {
		calls := 0
		record := &DurSpell{Target48: target}
		runtime := SpellMoonglowRuntime531A00{
			EnchantmentDuration: func() float32 { calls++; return 60 },
		}
		if got := SpellMoonglowCreate531A00(record, runtime); got != 1 {
			t.Fatalf("target %p: create = %d, want 1", target, got)
		}
		if calls != 1 {
			t.Fatalf("target %p: duration calls = %d, want 1", target, calls)
		}
	}
}

func TestSpellMoonglowCreate531A00NonPlayerAppliesLight(t *testing.T) {
	target := &Object{ObjClass: object.ClassMonster}
	record := &DurSpell{Target48: target, Level: 0x102, Field72: 0x55aa}
	calls := 0
	runtime := SpellMoonglowRuntime531A00{
		EnchantmentDuration: func() float32 { return 60.9 },
		ApplyBuff: func(got *Object, buff EnchantID, duration int16, power int8) {
			calls++
			if got != target || buff != ENCHANT_LIGHT || duration != 60 || power != 2 {
				t.Errorf("apply = %p/%d/%d/%d", got, buff, duration, power)
			}
		},
	}
	if got := SpellMoonglowCreate531A00(record, runtime); got != 1 {
		t.Fatalf("create = %d, want 1", got)
	}
	if calls != 1 || record.Field72 != 0x55aa {
		t.Fatalf("light calls = %d, PE32 pointer field = %#x", calls, record.Field72)
	}
}

func TestSpellMoonglowCreateAndDestroy531A00PlayerVisual(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status uint32
		point  types.Pointf
	}{
		{"default", 0, types.Pointf{X: 2944, Y: 2944}},
		{"cursor", 0x10, types.Pointf{X: -123, Y: 456}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			player := &Player{Field3680: tc.status, CursorVec: image.Point{X: -123, Y: 456}}
			update := &PlayerUpdateData{Player: player}
			target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
			visual := &Object{}
			record := &DurSpell{Target48: target, Level: 0x103, Field72: 0x55aa, Pos: types.Pointf{X: 1.725}}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
				t.Fatal("expected target pointer above 4 GiB")
			}
			var stored *Object
			var events []string
			runtime := SpellMoonglowRuntime531A00{
				EnchantmentDuration: func() float32 { events = append(events, "duration"); return 70.9 },
				NewObject: func(id string) *Object {
					events = append(events, "new")
					if id != "Moonglow" {
						t.Errorf("type ID = %q", id)
					}
					return visual
				},
				CreateAt: func(gotVisual, gotOwner *Object, point types.Pointf) {
					events = append(events, "create")
					if gotVisual != visual || gotOwner != target || point != tc.point {
						t.Errorf("create = %p/%p/%v", gotVisual, gotOwner, point)
					}
				},
				ApplyBuff: func(got *Object, buff EnchantID, duration int16, power int8) {
					events = append(events, "apply")
					if got != target || buff != ENCHANT_MOONGLOW || duration != 70 || power != 3 {
						t.Errorf("apply = %p/%d/%d/%d", got, buff, duration, power)
					}
				},
				BuffOff: func(got *Object, buff EnchantID) {
					events = append(events, "off")
					if got != target || buff != ENCHANT_MOONGLOW {
						t.Errorf("buff-off = %p/%d", got, buff)
					}
				},
				DelayedDelete: func(got *Object) {
					events = append(events, "delete")
					if got != visual {
						t.Errorf("deleted visual = %p", got)
					}
				},
				LoadVisual: func(*DurSpell) *Object { return stored },
				StoreVisual: func(got *DurSpell, newVisual *Object) {
					events = append(events, "store")
					if got != record {
						t.Errorf("stored record = %p", got)
					}
					stored = newVisual
				},
			}
			if got := SpellMoonglowCreate531A00(record, runtime); got != 0 {
				t.Fatalf("create = %d, want 0", got)
			}
			if stored != visual || record.Field72 != 0x55aa || record.Pos.X != 1.725 {
				t.Fatalf("stored visual = %p, PE32 pointer field = %#x, position = %v", stored, record.Field72, record.Pos)
			}
			SpellMoonglowDestroy531AF0(record, runtime)
			if stored != nil || record.Field72 != 0x55aa {
				t.Fatalf("visual after destroy = %p, PE32 pointer field = %#x", stored, record.Field72)
			}
			want := []string{"duration", "new", "store", "create", "apply", "delete", "store", "off"}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestSpellMoonglowCreate531A00MissingVisualAndDestroyOtherTargets(t *testing.T) {
	target := &Object{ObjClass: object.ClassPlayer}
	record := &DurSpell{Target48: target}
	storeCalls := 0
	runtime := SpellMoonglowRuntime531A00{
		EnchantmentDuration: func() float32 { return 60 },
		NewObject:           func(string) *Object { return nil },
		StoreVisual: func(*DurSpell, *Object) {
			storeCalls++
		},
	}
	if got := SpellMoonglowCreate531A00(record, runtime); got != 1 || storeCalls != 1 {
		t.Fatalf("missing visual: result = %d, stores = %d", got, storeCalls)
	}

	record.Target48 = nil
	SpellMoonglowDestroy531AF0(record, SpellMoonglowRuntime531A00{})
	record.Target48 = &Object{ObjClass: object.ClassMonster}
	buffCalls := 0
	SpellMoonglowDestroy531AF0(record, SpellMoonglowRuntime531A00{
		BuffOff: func(got *Object, buff EnchantID) {
			buffCalls++
			if got != record.Target48 || buff != ENCHANT_LIGHT {
				t.Errorf("buff-off = %p/%d", got, buff)
			}
		},
	})
	if buffCalls != 1 {
		t.Fatalf("non-player buff-off calls = %d, want 1", buffCalls)
	}
}
