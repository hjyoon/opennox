package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestSpellDurationFindActiveTarget4FF2D0NativeLayout(t *testing.T) {
	wantTarget := uintptr(48)
	wantFlags := uintptr(88)
	wantNext := uintptr(116)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantTarget = 72
		wantFlags = 120
		wantNext = 176
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Spell", unsafe.Offsetof(DurSpell{}.Spell), 4},
		{"Target48", unsafe.Offsetof(DurSpell{}.Target48), wantTarget},
		{"Flags88", unsafe.Offsetof(DurSpell{}.Flags88), wantFlags},
		{"Next", unsafe.Offsetof(DurSpell{}.Next), wantNext},
		{"Spell width", unsafe.Sizeof(DurSpell{}.Spell), 4},
		{"Flags88 width", unsafe.Sizeof(DurSpell{}.Flags88), 4},
		{"Target48 width", unsafe.Sizeof(DurSpell{}.Target48), unsafe.Sizeof(uintptr(0))},
		{"Next width", unsafe.Sizeof(DurSpell{}.Next), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("DurSpell %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSpellDurationFindActiveTarget4FF2D0NativeBindsLiveListAndFields(t *testing.T) {
	target, freeTarget := alloc.New(Object{})
	other, freeOther := alloc.New(Object{})
	recordA, freeA := alloc.New(DurSpell{})
	recordB, freeB := alloc.New(DurSpell{})
	recordC, freeC := alloc.New(DurSpell{})
	recordD, freeD := alloc.New(DurSpell{})
	t.Cleanup(freeTarget)
	t.Cleanup(freeOther)
	t.Cleanup(freeA)
	t.Cleanup(freeB)
	t.Cleanup(freeC)
	t.Cleanup(freeD)

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"target":  unsafe.Pointer(target),
			"recordA": unsafe.Pointer(recordA),
			"recordC": unsafe.Pointer(recordC),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	recordA.Flags88, recordA.Spell, recordA.Target48, recordA.Next = 0x12345601, 0x80000000, target, recordB
	recordB.Flags88, recordB.Spell, recordB.Target48, recordB.Next = 0xabcdef00, math.MaxInt32, target, recordC
	recordC.Flags88, recordC.Spell, recordC.Target48, recordC.Next = 0xfedcba00, 0x80000000, target, recordD
	recordD.Flags88, recordD.Spell, recordD.Target48 = 0x76543200, 0x80000000, other
	durations := new(SpellsDuration)
	durations.List = recordA

	got := durations.SpellDurationFindActiveTarget4FF2D0(math.MinInt32, target)
	if got != recordC {
		t.Fatalf("result = %p, want first active exact match %p", got, recordC)
	}

	recordC.Target48 = other
	if got := durations.SpellDurationFindActiveTarget4FF2D0(math.MinInt32, target); got != nil {
		t.Fatalf("result after live target replacement = %p, want nil", got)
	}

	durations.List = recordD
	recordD.Target48 = target
	if got := durations.SpellDurationFindActiveTarget4FF2D0(math.MinInt32, target); got != recordD {
		t.Fatalf("result after live head replacement = %p, want %p", got, recordD)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(other)
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
	runtime.KeepAlive(recordD)
}

func TestSpellDurationFindActiveTargetNative4FF2D0AccessorOrder(t *testing.T) {
	recordA := new(DurSpell)
	recordB := new(DurSpell)
	target := new(Object)
	var events []string
	deps := spellDurationFindActiveTargetNativeDeps4FF2D0{
		loadFirst: func() *DurSpell {
			events = append(events, "first")
			return recordA
		},
		loadFlagsLow: func(record *DurSpell) byte {
			if record == recordA {
				events = append(events, "flags:A")
				return 1
			}
			events = append(events, "flags:B")
			return 0
		},
		loadSpell: func(record *DurSpell) int32 {
			events = append(events, "spell:B")
			return -1
		},
		loadTarget: func(record *DurSpell) *Object {
			events = append(events, "target:B")
			return target
		},
		loadNext: func(record *DurSpell) *DurSpell {
			events = append(events, "next:A")
			return recordB
		},
	}

	got := spellDurationFindActiveTargetNative4FF2D0(-1, target, deps)

	want := []string{"first", "flags:A", "next:A", "flags:B", "spell:B", "target:B"}
	if got != recordB || !reflect.DeepEqual(events, want) {
		t.Fatalf("result/events = %p/%q, want %p/%q", got, events, recordB, want)
	}
}

func TestSpellDurationFindActiveTarget4FF2D0NativeNilReceiverFaultsAtHead(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil duration-list receiver did not fault on its head load")
		}
	}()
	var durations *SpellsDuration
	durations.SpellDurationFindActiveTarget4FF2D0(51, nil)
}
