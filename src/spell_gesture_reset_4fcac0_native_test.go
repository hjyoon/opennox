package opennox

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellGestureResetNative4FCAC0Layout(t *testing.T) {
	// Eleven preceding pointer words widen, and two intervening fields acquire
	// an additional four-byte alignment pad on 64-bit hosts.
	shift := uintptr(13) * (unsafe.Sizeof(uintptr(0)) - 4)
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"PlayerUpdateData.Field47_0", unsafe.Offsetof(server.PlayerUpdateData{}.Field47_0), 188 + shift},
		{"PlayerUpdateData.TrapSpells", unsafe.Offsetof(server.PlayerUpdateData{}.TrapSpells), 192 + shift},
		{"PlayerUpdateData.TrapSpellsCnt", unsafe.Offsetof(server.PlayerUpdateData{}.TrapSpellsCnt), 212 + shift},
		{"PlayerUpdateData.SpellCastStart", unsafe.Offsetof(server.PlayerUpdateData{}.SpellCastStart), 216 + shift},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSpellGestureResetNative4FCAC0BindsPointersFieldsAndCallbacks(t *testing.T) {
	allocatorClass := new(alloc.Class)
	allocator := alloc.ClassT[server.MagicEntityClass]{Class: allocatorClass}
	unit1 := new(server.Object)
	unit2 := new(server.Object)
	update1 := &server.PlayerUpdateData{
		Field47_0:      0xa1,
		TrapSpells:     [5]uint32{1, 2, 3, 4, 5},
		TrapSpellsCnt:  0xa1b2c3d4,
		SpellCastStart: 0x11223344,
	}
	update2 := &server.PlayerUpdateData{
		Field47_0:      0xb2,
		TrapSpells:     [5]uint32{6, 7, 8, 9, 10},
		TrapSpellsCnt:  0x55667788,
		SpellCastStart: 0xaabbccdd,
	}
	caster := new(server.Object)
	updates := map[*server.Object]*server.PlayerUpdateData{unit1: update1, unit2: update2}
	var events []string
	var casterGlobal *server.Object

	assertReset := func(name string, update *server.PlayerUpdateData, wantCount uint32) {
		t.Helper()
		if update.Field47_0 != 0 || update.SpellCastStart != 0 || update.TrapSpells != [5]uint32{} || update.TrapSpellsCnt != wantCount {
			t.Fatalf("%s state = field47 %#x, start %#x, traps %#v, count %#x", name, update.Field47_0, update.SpellCastStart, update.TrapSpells, update.TrapSpellsCnt)
		}
	}

	got := spellGestureResetNative4FCAC0(-7, 9, spellGestureResetNativeDeps4FCAC0{
		resetDurations: func(value int32) {
			if value != -7 {
				t.Fatalf("duration argument = %d, want -7", value)
			}
			events = append(events, "reset")
		},
		loadMagicClass: func() alloc.ClassT[server.MagicEntityClass] {
			events = append(events, "load-magic")
			return allocator
		},
		freeAllMagicObjects: func(value alloc.ClassT[server.MagicEntityClass]) {
			if value.Class != allocatorClass {
				t.Fatalf("allocator pointer = %p, want %p", value.Class, allocatorClass)
			}
			events = append(events, "free-magic")
		},
		clearMagicEntityHead: func() { events = append(events, "clear-head") },
		firstPlayerUnit: func() *server.Object {
			events = append(events, "first-unit")
			return unit1
		},
		loadPlayerUpdate: func(unit *server.Object) *server.PlayerUpdateData {
			if unit != unit1 && unit != unit2 {
				t.Fatalf("unexpected unit pointer %p", unit)
			}
			events = append(events, "load-update")
			return updates[unit]
		},
		nextPlayerUnit: func(unit *server.Object) *server.Object {
			events = append(events, "next-unit")
			switch unit {
			case unit1:
				assertReset("unit1", update1, 0xa1b2c300)
				return unit2
			case unit2:
				assertReset("unit2", update2, 0x55667700)
				return nil
			default:
				t.Fatalf("unexpected next-unit pointer %p", unit)
				return nil
			}
		},
		newObjectByTypeID: func(name string) *server.Object {
			if name != "ImaginaryCaster" {
				t.Fatalf("object type = %q", name)
			}
			events = append(events, "new-object")
			return caster
		},
		storeImaginaryCaster: func(value *server.Object) {
			if value != caster {
				t.Fatalf("caster pointer = %p, want %p", value, caster)
			}
			events = append(events, "store-caster")
			casterGlobal = value
		},
		createObjectAt: func(object, owner *server.Object, pos types.Pointf) {
			if object != caster || owner != nil || casterGlobal != caster {
				t.Fatalf("create pointers/global = (%p, %p, %p), want (%p, nil, %p)", object, owner, casterGlobal, caster, caster)
			}
			if math.Float32bits(pos.X) != math.Float32bits(2944) || math.Float32bits(pos.Y) != math.Float32bits(2944) {
				t.Fatalf("create position bits = (%08x, %08x)", math.Float32bits(pos.X), math.Float32bits(pos.Y))
			}
			events = append(events, "create-object")
		},
	})

	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	wantEvents := []string{
		"reset", "load-magic", "free-magic", "clear-head", "first-unit",
		"load-update", "next-unit", "load-update", "next-unit",
		"new-object", "store-caster", "create-object",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit1": uintptr(unsafe.Pointer(unit1)), "unit2": uintptr(unsafe.Pointer(unit2)),
			"update1": uintptr(unsafe.Pointer(update1)), "update2": uintptr(unsafe.Pointer(update2)),
			"caster": uintptr(unsafe.Pointer(caster)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}
	runtime.KeepAlive(allocatorClass)
	runtime.KeepAlive(unit1)
	runtime.KeepAlive(unit2)
	runtime.KeepAlive(update1)
	runtime.KeepAlive(update2)
	runtime.KeepAlive(caster)
}

func TestSpellGestureResetNative4FCAC0ForwardsNilAllocatorAndCaster(t *testing.T) {
	storedCaster := new(server.Object)
	freed := false
	got := spellGestureResetNative4FCAC0(0, 1, spellGestureResetNativeDeps4FCAC0{
		resetDurations: func(int32) {},
		loadMagicClass: func() alloc.ClassT[server.MagicEntityClass] { return alloc.ClassT[server.MagicEntityClass]{} },
		freeAllMagicObjects: func(value alloc.ClassT[server.MagicEntityClass]) {
			freed = true
			if value.Class != nil {
				t.Fatalf("allocator pointer = %p, want nil", value.Class)
			}
		},
		clearMagicEntityHead: func() {},
		firstPlayerUnit:      func() *server.Object { return nil },
		newObjectByTypeID:    func(string) *server.Object { return nil },
		storeImaginaryCaster: func(value *server.Object) { storedCaster = value },
		createObjectAt: func(*server.Object, *server.Object, types.Pointf) {
			t.Fatal("nil caster reached create")
		},
	})

	if got != 0 || !freed || storedCaster != nil {
		t.Fatalf("result/freed/caster = (%d, %t, %p), want (0, true, nil)", got, freed, storedCaster)
	}
}
