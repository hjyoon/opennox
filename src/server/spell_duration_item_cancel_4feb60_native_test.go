package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func requireItemCancelDurSpellsNativePointers4FEB60(t *testing.T, values ...unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for i, value := range values {
		if value == nil || uintptr(value) <= math.MaxUint32 {
			t.Fatalf("pointer %d = %p, want native address above 4 GiB", i, value)
		}
	}
}

func TestItemCancelDurSpells4FEB60NativeObjectLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantSubclass := uintptr(12)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantSubclass = 16
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubclass},
		{"ObjClass width", unsafe.Sizeof(Object{}.ObjClass), 4},
		{"ObjSubClass width", unsafe.Sizeof(Object{}.ObjSubClass), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestItemCancelDurSpellsNative4FEB60PreservesPointersAndReloadsSubclass(t *testing.T) {
	owner := new(Object)
	item := &Object{
		ObjClass:    object.Class(itemCancelDurSpellsClass4FEB60),
		ObjSubClass: object.SubClass(itemCancelDurSpellsSpell43Bit4FEB60),
	}
	requireItemCancelDurSpellsNativePointers4FEB60(t, unsafe.Pointer(owner), unsafe.Pointer(item))

	var spells []int32
	var owners []*Object
	itemCancelDurSpellsNative4FEB60(owner, item, itemCancelDurSpellsNativeDeps4FEB60{
		cancel: func(spellID int32, gotOwner *Object) {
			spells = append(spells, spellID)
			owners = append(owners, gotOwner)
			if spellID == itemCancelDurSpellsSpell43ID4FEB60 {
				item.ObjSubClass |= object.SubClass(itemCancelDurSpellsSpell59Bit4FEB60)
			}
		},
	})

	if want := []int32{43, 59}; !reflect.DeepEqual(spells, want) {
		t.Fatalf("spells = %v, want %v", spells, want)
	}
	if len(owners) != 2 || owners[0] != owner || owners[1] != owner {
		t.Fatalf("owners = %p, want [%p %p]", owners, owner, owner)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}

func TestItemCancelDurSpells4FEB60ServerBinding(t *testing.T) {
	owner := new(Object)
	item := &Object{
		ObjClass: object.Class(itemCancelDurSpellsClass4FEB60),
		ObjSubClass: object.SubClass(itemCancelDurSpellsSpell43Bit4FEB60 |
			itemCancelDurSpellsSpell59Bit4FEB60),
	}
	record43 := &DurSpell{Spell: uint32(itemCancelDurSpellsSpell43ID4FEB60), Caster16: owner, Flags88: 0x12345620}
	record59 := &DurSpell{Spell: uint32(itemCancelDurSpellsSpell59ID4FEB60), Caster16: owner, Flags88: 0x89abcdee}
	record43.Next = record59
	durations := SpellsDuration{List: record43}
	requireItemCancelDurSpellsNativePointers4FEB60(t,
		unsafe.Pointer(owner), unsafe.Pointer(item), unsafe.Pointer(record43), unsafe.Pointer(record59),
	)

	durations.ItemCancelDurSpells4FEB60(owner, item)

	if record43.Flags88 != 0x12345621 || record59.Flags88 != 0x89abcdef {
		t.Fatalf("flags = %#x/%#x, want 0x12345621/0x89abcdef", record43.Flags88, record59.Flags88)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}

func TestItemCancelDurSpells4FEB60ServerBindingForwardsNilOwner(t *testing.T) {
	item := &Object{
		ObjClass:    object.Class(itemCancelDurSpellsClass4FEB60),
		ObjSubClass: object.SubClass(itemCancelDurSpellsSpell43Bit4FEB60),
	}
	record := &DurSpell{Spell: uint32(itemCancelDurSpellsSpell43ID4FEB60), Flags88: 0x76543210}
	durations := SpellsDuration{List: record}

	durations.ItemCancelDurSpells4FEB60(nil, item)

	if record.Flags88 != 0x76543211 {
		t.Fatalf("flags = %#x, want 0x76543211", record.Flags88)
	}
}

func TestItemCancelDurSpells4FEB60NativeDoesNotGuardNilItem(t *testing.T) {
	var durations SpellsDuration
	defer func() {
		if recover() == nil {
			t.Fatal("nil item did not fault at the native ObjClass load")
		}
	}()
	durations.ItemCancelDurSpells4FEB60(new(Object), nil)
}
