package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func requireNativeSpellDurationFreePointer4FE980(t *testing.T, name string, value unsafe.Pointer) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s pointer is nil", name)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(value) <= math.MaxUint32 {
		t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, value)
	}
}

func TestSpellDurationFreeRecursiveNative4FE980PreservesPointersAndOrder(t *testing.T) {
	root := new(DurSpell)
	sub108 := new(DurSpell)
	sub104 := new(DurSpell)
	allocatorA := new(alloc.Class)
	allocatorB := new(alloc.Class)
	root.Sub108 = sub108
	root.Sub104 = sub104
	for name, value := range map[string]unsafe.Pointer{
		"root":        unsafe.Pointer(root),
		"Sub108":      unsafe.Pointer(sub108),
		"Sub104":      unsafe.Pointer(sub104),
		"allocator A": unsafe.Pointer(allocatorA),
		"allocator B": unsafe.Pointer(allocatorB),
	} {
		requireNativeSpellDurationFreePointer4FE980(t, name, value)
	}

	liveAllocator := alloc.ClassT[DurSpell]{Class: allocatorA}
	var events []string
	spellDurationFreeRecursiveNative4FE980(root, spellDurationFreeRecursiveNativeDeps4FE980{
		loadSub108: func(value *DurSpell) *DurSpell {
			events = append(events, "sub108")
			return value.Sub108
		},
		loadSub104: func(value *DurSpell) *DurSpell {
			events = append(events, "sub104")
			return value.Sub104
		},
		loadNext: func(value *DurSpell) *DurSpell {
			events = append(events, "next")
			return value.Next
		},
		loadAllocator: func() alloc.ClassT[DurSpell] {
			events = append(events, "allocator")
			return liveAllocator
		},
		freeObjectFirst: func(allocator alloc.ClassT[DurSpell], value *DurSpell) {
			switch value {
			case sub108:
				events = append(events, "free-sub108")
				if allocator.Class != allocatorA {
					t.Fatalf("Sub108 allocator = %p, want %p", allocator.Class, allocatorA)
				}
				liveAllocator.Class = allocatorB
			case sub104:
				events = append(events, "free-sub104")
				if allocator.Class != allocatorB {
					t.Fatalf("Sub104 allocator = %p, want live %p", allocator.Class, allocatorB)
				}
			case root:
				events = append(events, "free-root")
				if allocator.Class != allocatorB {
					t.Fatalf("root allocator = %p, want live %p", allocator.Class, allocatorB)
				}
			default:
				t.Fatalf("unexpected record %p", value)
			}
		},
	})

	want := []string{
		"sub108", "next", "sub108", "sub104", "allocator", "free-sub108",
		"sub104", "next", "sub108", "sub104", "allocator", "free-sub104",
		"allocator", "free-root",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	runtime.KeepAlive(root)
	runtime.KeepAlive(sub108)
	runtime.KeepAlive(sub104)
	runtime.KeepAlive(allocatorA)
	runtime.KeepAlive(allocatorB)
}

func TestSpellDurationFreeRecursive4FE980NativeAllocatorReuse(t *testing.T) {
	spells := SpellsDuration{alloc: alloc.NewClassT("DurationFree4FE980", DurSpell{}, 5)}
	t.Cleanup(spells.alloc.Free)

	root := spells.alloc.NewObject()
	sub108A := spells.alloc.NewObject()
	sub108B := spells.alloc.NewObject()
	grandchild := spells.alloc.NewObject()
	sub104 := spells.alloc.NewObject()
	for name, record := range map[string]*DurSpell{
		"root":       root,
		"Sub108 A":   sub108A,
		"Sub108 B":   sub108B,
		"grandchild": grandchild,
		"Sub104":     sub104,
	} {
		requireNativeSpellDurationFreePointer4FE980(t, name, unsafe.Pointer(record))
	}
	root.Sub108 = sub108A
	root.Sub104 = sub104
	sub108A.Next = sub108B
	sub108A.Sub108 = grandchild

	spells.SpellDurationFreeRecursive4FE980(root)

	wantReuse := []*DurSpell{root, sub104, sub108B, sub108A, grandchild}
	for i, want := range wantReuse {
		got := spells.alloc.NewObject()
		if got != want {
			t.Fatalf("reuse %d = %p, want depth-first FreeObjectFirst order %p", i, got, want)
		}
		if *got != (DurSpell{}) {
			t.Fatalf("reuse %d record = %#v, want zeroed DurSpell", i, *got)
		}
	}
	if got, want := unsafe.Offsetof(DurSpell{}.Sub104), uintptr(152); unsafe.Sizeof(uintptr(0)) == 8 && got != want {
		t.Fatalf("64-bit Sub104 offset = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(DurSpell{}.Sub108), uintptr(160); unsafe.Sizeof(uintptr(0)) == 8 && got != want {
		t.Fatalf("64-bit Sub108 offset = %d, want %d", got, want)
	}
}

func TestSpellDurationFreeRecursive4FE980NativeDoesNotGuardNilRecord(t *testing.T) {
	var spells SpellsDuration
	defer func() {
		if recover() == nil {
			t.Fatal("nil duration record did not fault")
		}
	}()
	spells.SpellDurationFreeRecursive4FE980(nil)
}
