package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func requireSpellDurationProcessNativePointers4FEEF0(t *testing.T, values ...unsafe.Pointer) {
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

func TestSpellDurationProcess4FEEF0NativeLayouts(t *testing.T) {
	type layout struct {
		durSize uintptr
		obj12   uintptr
		caster  uintptr
		flag20  uintptr
		obj24   uintptr
		frame60 uintptr
		frame68 uintptr
		flags88 uintptr
		update  uintptr
		next    uintptr
		objFlag uintptr
	}
	wants := map[uintptr]layout{
		4: {120, 12, 16, 20, 24, 60, 68, 88, 96, 116, 16},
		8: {184, 16, 24, 32, 40, 88, 96, 120, 136, 176, 20},
	}
	ptrSize := unsafe.Sizeof(uintptr(0))
	want, ok := wants[ptrSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
	got := layout{
		unsafe.Sizeof(DurSpell{}),
		unsafe.Offsetof(DurSpell{}.Obj12),
		unsafe.Offsetof(DurSpell{}.Caster16),
		unsafe.Offsetof(DurSpell{}.Flag20),
		unsafe.Offsetof(DurSpell{}.Obj24),
		unsafe.Offsetof(DurSpell{}.Frame60),
		unsafe.Offsetof(DurSpell{}.Frame68),
		unsafe.Offsetof(DurSpell{}.Flags88),
		unsafe.Offsetof(DurSpell{}.Update),
		unsafe.Offsetof(DurSpell{}.Next),
		unsafe.Offsetof(Object{}.ObjFlags),
	}
	if got != want {
		t.Fatalf("native layout on %s/%s = %+v, want %+v", runtime.GOOS, runtime.GOARCH, got, want)
	}
}

func TestSpellDurationProcessNative4FEEF0PreservesPointersAndOrder(t *testing.T) {
	recordA := &DurSpell{Flags88: 1}
	recordB := &DurSpell{Frame60: 7, Frame68: 7}
	recordA.Next = recordB
	caster := &Object{ObjFlags: object.Flags(0x80000000)}
	obj12 := &Object{ObjFlags: object.Flags(0x8000)}
	obj24 := &Object{ObjFlags: object.FlagDestroyed}
	recordB.Caster16 = caster
	recordB.Obj12 = obj12
	recordB.Obj24 = obj24
	updateCell := new(byte)
	recordB.Update = unsafe.Pointer(updateCell)
	requireSpellDurationProcessNativePointers4FEEF0(t,
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(caster),
		unsafe.Pointer(obj12), unsafe.Pointer(obj24), unsafe.Pointer(updateCell),
	)

	var events []string
	spellDurationProcessNative4FEEF0(spellDurationProcessNativeDeps4FEEF0{
		loadFirst: func() *DurSpell {
			events = append(events, "head")
			return recordA
		},
		loadFlagsLowByte: func(record *DurSpell) byte {
			events = append(events, "flags")
			return byte(record.Flags88)
		},
		loadNext: func(record *DurSpell) *DurSpell {
			events = append(events, "next")
			return record.Next
		},
		destroy: func(record *DurSpell) {
			events = append(events, "destroy")
		},
		loadCaster: func(record *DurSpell) *Object {
			events = append(events, "caster")
			return record.Caster16
		},
		loadObjectFlags: func(object *Object) uint32 {
			events = append(events, "flags-full")
			return uint32(object.ObjFlags)
		},
		storeCaster: func(record *DurSpell, object *Object) {
			events = append(events, "store-caster")
			record.Caster16 = object
		},
		loadObj12: func(record *DurSpell) *Object {
			events = append(events, "obj12")
			return record.Obj12
		},
		loadObjectFlagsLowByte: func(object *Object) byte {
			events = append(events, "flags-low")
			return byte(object.ObjFlags)
		},
		storeObj12: func(record *DurSpell, object *Object) {
			events = append(events, "store-obj12")
			record.Obj12 = object
		},
		loadFlag20: func(record *DurSpell) uint32 {
			events = append(events, "flag20")
			return record.Flag20
		},
		loadObj24: func(record *DurSpell) *Object {
			events = append(events, "obj24")
			return record.Obj24
		},
		storeObj24: func(record *DurSpell, object *Object) {
			events = append(events, "store-obj24")
			record.Obj24 = object
		},
		loadFrame68: func(record *DurSpell) uint32 {
			events = append(events, "frame68")
			return record.Frame68
		},
		loadFrame60: func(record *DurSpell) uint32 {
			events = append(events, "frame60")
			return record.Frame60
		},
		loadCurrentFrame: func() uint32 {
			events = append(events, "current-frame")
			return 0
		},
		loadUpdate: func(record *DurSpell) unsafe.Pointer {
			events = append(events, "update")
			return record.Update
		},
		callUpdate: func(update unsafe.Pointer, record *DurSpell) int32 {
			events = append(events, "call-update")
			if update != unsafe.Pointer(updateCell) || record != recordB {
				t.Fatalf("callback args = %p/%p, want %p/%p", update, record, updateCell, recordB)
			}
			return 0
		},
		cancel: func(record *DurSpell) {
			events = append(events, "cancel")
		},
	})
	want := []string{
		"head", "flags", "next", "destroy",
		"flags", "next", "caster", "flags-full", "obj12", "flags-low", "caster",
		"obj24", "flags-low", "store-obj24", "frame68", "frame60", "update", "call-update",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if recordB.Caster16 != caster || recordB.Obj12 != obj12 || recordB.Obj24 != nil {
		t.Fatalf("record pointers = caster %p obj12 %p obj24 %p", recordB.Caster16, recordB.Obj12, recordB.Obj24)
	}
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(caster)
	runtime.KeepAlive(obj12)
	runtime.KeepAlive(obj24)
	runtime.KeepAlive(updateCell)
}

func TestSpellDurationProcess4FEEF0ServerBinding(t *testing.T) {
	srv := new(Server)
	srv.SetFrame(10)
	deadCaster := &Object{ObjFlags: object.FlagDead}
	destroyed12 := &Object{ObjFlags: object.FlagDestroyed}
	liveCaster := &Object{ObjFlags: object.Flags(0x80000000)}
	destroyed24 := &Object{ObjFlags: object.FlagDestroyed}
	recordA := &DurSpell{Flags88: 0x12345601}
	recordB := &DurSpell{
		Flags88: 0xabcdef80, Caster16: deadCaster, Obj12: destroyed12,
	}
	recordC := &DurSpell{
		Flags88: 0x76543280, Caster16: liveCaster, Obj24: destroyed24,
		Frame60: 1, Frame68: 10,
	}
	recordD := &DurSpell{
		Flags88: 0xfedcba80, Caster16: liveCaster,
		Frame60: 1, Frame68: 11,
	}
	recordA.Next = recordB
	recordB.Next = recordC
	recordC.Next = recordD
	durations := SpellsDuration{s: srv, List: recordA}
	requireSpellDurationProcessNativePointers4FEEF0(t,
		unsafe.Pointer(srv), unsafe.Pointer(deadCaster), unsafe.Pointer(destroyed12),
		unsafe.Pointer(liveCaster), unsafe.Pointer(destroyed24), unsafe.Pointer(recordA),
		unsafe.Pointer(recordB), unsafe.Pointer(recordC), unsafe.Pointer(recordD),
	)

	var destroyed []*DurSpell
	durations.SpellDurationProcess4FEEF0(SpellDurationProcessRuntime4FEEF0{
		Destroy: func(record *DurSpell) {
			destroyed = append(destroyed, record)
			durations.SpellDurationUnlink4FE900(record)
		},
		CallUpdate: func(unsafe.Pointer, *DurSpell) int32 {
			t.Fatal("nil Update unexpectedly invoked")
			return 0
		},
	})
	if !reflect.DeepEqual(destroyed, []*DurSpell{recordA}) {
		t.Fatalf("destroyed = %p, want [%p]", destroyed, recordA)
	}
	if recordB.Caster16 != nil || recordB.Obj12 != nil || recordB.Flags88 != 0xabcdef81 {
		t.Fatalf("record B = caster %p obj12 %p flags %#x, want nil/nil/0xabcdef81", recordB.Caster16, recordB.Obj12, recordB.Flags88)
	}
	if recordC.Obj24 != nil || recordC.Flags88 != 0x76543281 {
		t.Fatalf("record C = obj24 %p flags %#x, want nil/0x76543281", recordC.Obj24, recordC.Flags88)
	}
	if recordD.Flags88 != 0xfedcba80 {
		t.Fatalf("future record D flags = %#x, want unchanged", recordD.Flags88)
	}
	if durations.List != recordB || recordB.Next != recordC || recordC.Next != recordD {
		t.Fatalf("surviving list = %p -> %p -> %p, want B -> C -> D", durations.List, recordB.Next, recordC.Next)
	}
	runtime.KeepAlive(srv)
	runtime.KeepAlive(deadCaster)
	runtime.KeepAlive(destroyed12)
	runtime.KeepAlive(liveCaster)
	runtime.KeepAlive(destroyed24)
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
	runtime.KeepAlive(recordD)
}

func TestSpellDurationProcess4FEEF0ServerNilReceiverFaultsAtHead(t *testing.T) {
	var durations *SpellsDuration
	defer func() {
		if recover() == nil {
			t.Fatal("nil SpellsDuration receiver did not fault at the head load")
		}
	}()
	durations.SpellDurationProcess4FEEF0(SpellDurationProcessRuntime4FEEF0{})
}
