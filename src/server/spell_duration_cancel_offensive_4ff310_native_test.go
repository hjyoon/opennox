package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestSpellDurationCancelOffensive4FF310NativeLayout(t *testing.T) {
	wantCaster := uintptr(16)
	wantNext := uintptr(116)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantCaster = 24
		wantNext = 176
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Spell", unsafe.Offsetof(DurSpell{}.Spell), 4},
		{"Caster16", unsafe.Offsetof(DurSpell{}.Caster16), wantCaster},
		{"Next", unsafe.Offsetof(DurSpell{}.Next), wantNext},
		{"Spell width", unsafe.Sizeof(DurSpell{}.Spell), 4},
		{"Caster16 width", unsafe.Sizeof(DurSpell{}.Caster16), unsafe.Sizeof(uintptr(0))},
		{"Next width", unsafe.Sizeof(DurSpell{}.Next), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("DurSpell %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSpellDurationCancelOffensiveNative4FF310AccessorOrder(t *testing.T) {
	recordA := new(DurSpell)
	recordB := new(DurSpell)
	caster := new(Object)
	other := new(Object)
	var events []string
	deps := spellDurationCancelOffensiveNativeDeps4FF310{
		loadFirst: func() *DurSpell {
			events = append(events, "first")
			return recordA
		},
		loadCasterArg: func() *Object {
			events = append(events, "caster-arg")
			return caster
		},
		loadCaster: func(record *DurSpell) *Object {
			if record == recordA {
				events = append(events, "caster:A")
				return other
			}
			events = append(events, "caster:B")
			return caster
		},
		loadNext: func(record *DurSpell) *DurSpell {
			if record == recordA {
				events = append(events, "next:A")
				return recordB
			}
			events = append(events, "next:B")
			return nil
		},
		loadSpell: func(record *DurSpell) int32 {
			if record != recordB {
				t.Fatalf("spell record = %p, want B %p", record, recordB)
			}
			events = append(events, "spell:B")
			return math.MinInt32
		},
		loadSpellFlags: func(spellID int32) uint32 {
			if spellID != math.MinInt32 {
				t.Fatalf("spell ID = %d, want signed dword minimum", spellID)
			}
			events = append(events, "flags")
			return 0xffffff20
		},
		cancel: func(record *DurSpell) {
			if record != recordB {
				t.Fatalf("cancel record = %p, want B %p", record, recordB)
			}
			events = append(events, "cancel:B")
		},
	}

	spellDurationCancelOffensiveNative4FF310(deps)
	want := []string{
		"first", "caster-arg",
		"caster:A", "next:A",
		"caster:B", "next:B", "spell:B", "flags", "cancel:B",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestSpellDurationCancelOffensive4FF310ServerBinding(t *testing.T) {
	var srv Server
	srv.Spells.init(&srv)
	unsignedHighSpell := uint32(1) << 31
	srv.Spells.byID = map[spell.ID]*SpellDef{
		31: {
			Def: things.Spell{Flags: things.SpellOffensive},
		},
		32: {
			Def: things.Spell{Flags: things.SpellDefensive},
		},
		spell.ID(unsignedHighSpell): {
			Def: things.Spell{Flags: things.SpellOffensive},
		},
	}

	caster, freeCaster := alloc.New(Object{})
	other, freeOther := alloc.New(Object{})
	recordA, freeA := alloc.New(DurSpell{})
	recordB, freeB := alloc.New(DurSpell{})
	recordC, freeC := alloc.New(DurSpell{})
	recordD, freeD := alloc.New(DurSpell{})
	recordNil, freeNil := alloc.New(DurSpell{})
	t.Cleanup(func() {
		freeCaster()
		freeOther()
		freeA()
		freeB()
		freeC()
		freeD()
		freeNil()
	})
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"caster": unsafe.Pointer(caster),
			"other":  unsafe.Pointer(other),
			"record": unsafe.Pointer(recordC),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	recordA.Caster16, recordA.Spell, recordA.Flags88, recordA.Next = other, 31, 0x11111100, recordB
	recordB.Caster16, recordB.Spell, recordB.Flags88, recordB.Next = caster, 32, 0x22222200, recordC
	recordC.Caster16, recordC.Spell, recordC.Flags88, recordC.Next = caster, 31, 0xa1b2c300, recordD
	recordD.Caster16, recordD.Spell, recordD.Flags88 = caster, unsignedHighSpell, 0x44444400
	srv.Spells.Dur.List = recordA

	srv.Spells.Dur.SpellDurationCancelOffensive4FF310(caster)
	if recordA.Flags88 != 0x11111100 || recordB.Flags88 != 0x22222200 || recordD.Flags88 != 0x44444400 {
		t.Fatalf("nonmatching flags = %#x/%#x/%#x, want unchanged", recordA.Flags88, recordB.Flags88, recordD.Flags88)
	}
	if recordC.Flags88 != 0xa1b2c301 {
		t.Fatalf("offensive flags = %#x, want low cancellation bit 0xa1b2c301", recordC.Flags88)
	}

	recordNil.Spell = 31
	recordNil.Flags88 = 0x55667700
	srv.Spells.Dur.List = recordNil
	srv.Spells.Dur.SpellDurationCancelOffensive4FF310(nil)
	if recordNil.Flags88 != 0x55667701 {
		t.Fatalf("nil-caster flags = %#x, want low cancellation bit 0x55667701", recordNil.Flags88)
	}

	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
	runtime.KeepAlive(recordD)
	runtime.KeepAlive(recordNil)
}

func TestSpellDurationCancelOffensive4FF310NativeNilReceiverFaultsAtHead(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil duration-list receiver did not fault on its head load")
		}
	}()
	var durations *SpellsDuration
	durations.SpellDurationCancelOffensive4FF310(nil)
}
