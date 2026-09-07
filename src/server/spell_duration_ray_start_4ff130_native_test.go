package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestDurationRayStartNative4FF130Layout(t *testing.T) {
	wantDurSize := uintptr(120)
	wantCaster := uintptr(16)
	wantTarget := uintptr(48)
	wantSub108 := uintptr(108)
	wantNext := uintptr(116)
	wantDirection := uintptr(124)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantDurSize = 184
		wantCaster = 24
		wantTarget = 72
		wantSub108 = 160
		wantNext = 176
		wantDirection = 128
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"DurSpell size", unsafe.Sizeof(DurSpell{}), wantDurSize},
		{"DurSpell.Spell", unsafe.Offsetof(DurSpell{}.Spell), 4},
		{"DurSpell.Level", unsafe.Offsetof(DurSpell{}.Level), 8},
		{"DurSpell.Caster16", unsafe.Offsetof(DurSpell{}.Caster16), wantCaster},
		{"DurSpell.Target48", unsafe.Offsetof(DurSpell{}.Target48), wantTarget},
		{"DurSpell.Sub108", unsafe.Offsetof(DurSpell{}.Sub108), wantSub108},
		{"DurSpell.Next", unsafe.Offsetof(DurSpell{}.Next), wantNext},
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantDirection},
		{"Caster16 width", unsafe.Sizeof(DurSpell{}.Caster16), unsafe.Sizeof(uintptr(0))},
		{"Target48 width", unsafe.Sizeof(DurSpell{}.Target48), unsafe.Sizeof(uintptr(0))},
		{"Sub108 width", unsafe.Sizeof(DurSpell{}.Sub108), unsafe.Sizeof(uintptr(0))},
		{"Next width", unsafe.Sizeof(DurSpell{}.Next), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDurationRayStartNative4FF130PreservesHighPointersAndLiveReloads(t *testing.T) {
	caster := &Object{Direction1: 0xcd}
	caster2 := &Object{Direction1: 0xee}
	caster3 := &Object{Direction1: 0xff}
	target := &Object{}
	target2 := &Object{}
	record := &DurSpell{Spell: 59, Caster16: caster, Target48: target}

	var pin runtime.Pinner
	for _, pointer := range []any{caster, caster2, caster3, target, target2, record} {
		pin.Pin(pointer)
	}
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"caster":  uintptr(unsafe.Pointer(caster)),
			"caster2": uintptr(unsafe.Pointer(caster2)),
			"caster3": uintptr(unsafe.Pointer(caster3)),
			"target":  uintptr(unsafe.Pointer(target)),
			"target2": uintptr(unsafe.Pointer(target2)),
			"record":  uintptr(unsafe.Pointer(record)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	var gotPacket [7]byte
	var gotRecipient int32
	var gotRelated *Object
	var gotRemove int32
	var unmarked []*Object
	durationRayStartNative4FF130(record, durationRayStartNativeDeps4FF130{
		unitCode: func(object *Object) uint32 {
			switch object {
			case target:
				record.Caster16 = caster2
				return 0x12345678
			case caster2:
				return 0xabcdef01
			default:
				t.Fatalf("unit code object = %p", object)
				return 0
			}
		},
		sendPacket: func(recipient int32, packet [7]byte, related *Object, remove int32) int32 {
			gotRecipient, gotPacket, gotRelated, gotRemove = recipient, packet, related, remove
			record.Caster16 = caster3
			record.Target48 = target2
			return math.MinInt32
		},
		unmarkMinimap: func(object *Object, flags uint32) {
			if flags != 2 {
				t.Fatalf("unmark flags = %d, want 2", flags)
			}
			unmarked = append(unmarked, object)
		},
	})

	if gotRecipient != 255 || gotRelated != nil || gotRemove != 1 {
		t.Fatalf("send metadata = (%d,%p,%d), want (255,nil,1)", gotRecipient, gotRelated, gotRemove)
	}
	wantPacket := [7]byte{0x9e, 1, 0xcd, 0x01, 0xef, 0x78, 0x56}
	if gotPacket != wantPacket {
		t.Fatalf("packet = %v, want %v", gotPacket, wantPacket)
	}
	if want := []*Object{caster3, target2}; !reflect.DeepEqual(unmarked, want) {
		t.Fatalf("unmarked = %p, want %p", unmarked, want)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(caster2)
	runtime.KeepAlive(caster3)
	runtime.KeepAlive(target)
	runtime.KeepAlive(target2)
	runtime.KeepAlive(record)
}

func TestDurationRayStartNative4FF130ChainUsesLiveLinks(t *testing.T) {
	caster := &Object{}
	caster2 := &Object{}
	targetA := &Object{}
	targetB := &Object{}
	targetC := &Object{}
	childB := &DurSpell{Spell: 9, Level: 2, Caster16: caster2, Target48: targetB}
	childC := &DurSpell{Spell: 24, Level: 3, Caster16: caster2, Target48: targetC}
	childA := &DurSpell{Spell: 7, Level: 1, Caster16: caster, Target48: targetA, Next: childB}
	parent := &DurSpell{Spell: 43, Caster16: caster, Sub108: childA}

	codes := map[*Object]uint32{
		caster:  0x1001,
		caster2: 0x1002,
		targetA: 0x2001,
		targetB: 0x2002,
		targetC: 0x2003,
	}
	var packets [][7]byte
	durationRayStartNative4FF130(parent, durationRayStartNativeDeps4FF130{
		unitCode: func(object *Object) uint32 {
			return codes[object]
		},
		sendPacket: func(_ int32, packet [7]byte, _ *Object, _ int32) int32 {
			packets = append(packets, packet)
			if len(packets) == 1 {
				childA.Next = childC
			}
			return 0
		},
		unmarkMinimap: func(*Object, uint32) {},
	})
	want := [][7]byte{
		{0x9e, 3, 1, 0x01, 0x10, 0x01, 0x20},
		{0x9e, 4, 3, 0x02, 0x10, 0x03, 0x20},
	}
	if !reflect.DeepEqual(packets, want) {
		t.Fatalf("packets = %v, want %v", packets, want)
	}
}
