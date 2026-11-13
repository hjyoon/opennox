package alloc

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"
)

func requireFixtureAddress(t *testing.T, p unsafe.Pointer) {
	t.Helper()
	if p == nil {
		t.Fatal("nil foreign allocation")
	}
	if runtime.GOOS == "windows" && unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= uintptr(^uint32(0)) {
		t.Fatalf("explicit high fixture allocation is below 4 GiB: %p", p)
	}
}

func TestHighAllocationOwnershipAndRealloc(t *testing.T) {
	ordinary, freeOrdinary := Malloc(23)
	defer freeOrdinary()
	if allocReleases[ordinary] != nil {
		t.Fatal("ordinary C allocator unexpectedly changed")
	}
	t.Logf("ordinary C address=%p native pointer width=%d", ordinary, unsafe.Sizeof(uintptr(0)))
	p, _ := MallocHigh(23)
	requireFixtureAddress(t, p)
	if !bytes.Equal(unsafe.Slice((*byte)(p), 23), make([]byte, 23)) {
		t.Fatal("new foreign fixture memory was not zeroed")
	}
	want := []byte("foreign ownership bytes")
	copy(unsafe.Slice((*byte)(p), 23), want)
	// Re-adopting an existing owner must preserve its custom release method.
	AdoptCAllocation(p, 23)
	p = Realloc(p, 79)
	requireFixtureAddress(t, p)
	if !bytes.Equal(unsafe.Slice((*byte)(p), len(want)), want) {
		t.Fatal("foreign realloc growth lost the original bytes")
	}
	p = Realloc(p, 11)
	requireFixtureAddress(t, p)
	if !bytes.Equal(unsafe.Slice((*byte)(p), 11), want[:11]) {
		t.Fatal("foreign realloc shrink lost the retained bytes")
	}
	FreePtr(p)
	if _, ok := allocs[p]; ok || allocReleases[p] != nil {
		t.Fatal("freed foreign ownership remained registered")
	}
}

func TestHighFixtureTypedFramesAndClassOwnership(t *testing.T) {
	src := []uint64{0x1122334455667788, 0xffeeddccbbaa0099}
	s, free := MakeHigh(src, 3)
	defer free()
	requireFixtureAddress(t, unsafe.Pointer(&s[0]))
	if s[0] != src[0] || s[1] != src[1] || s[2] != 0 {
		t.Fatal("typed fixture input copy or zero tail changed")
	}
	value, freeValue := NewHigh(uint64(0))
	defer freeValue()
	requireFixtureAddress(t, unsafe.Pointer(value))
	if *value != 0 {
		t.Fatal("typed fixture was not zeroed")
	}
	ordinaryText, freeOrdinaryText := CString16("ordinary")
	defer freeOrdinaryText()
	if allocReleases[unsafe.Pointer(ordinaryText)] != nil {
		t.Fatal("ordinary UTF16 allocation inherited the fixture policy")
	}
	text, freeText := CString16High("native 한글 😀")
	requireFixtureAddress(t, unsafe.Pointer(text))
	if GoString16(text) != "native 한글 😀" || GoString16(ordinaryText) != "ordinary" {
		t.Fatal("UTF16 fixture lost text or modified another allocation")
	}
	freeText()
	if _, ok := allocs[unsafe.Pointer(text)]; ok {
		t.Fatal("UTF16 fixture ownership remained registered after release")
	}
	pool := NewClassHigh("high-fixture-ownership", 32, 1)
	pool.dynamic.enabled = true
	a, b := pool.NewObject(), pool.NewObject()
	requireFixtureAddress(t, a)
	requireFixtureAddress(t, b)
	if a == b {
		t.Fatal("distinct live class records alias")
	}
	pool.Free()
}

func TestHighAllocationRejectsInvalidExtents(t *testing.T) {
	for name, call := range map[string]func(){
		"zero":     func() { MallocHigh(0) },
		"negative": func() { CallocHigh(-1, 4) },
		"overflow": func() { CallocHigh(2, ^uintptr(0)) },
		"empty":    func() { MakeHigh([]uint32(nil), 0) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid fixture extent did not fail")
				}
			}()
			call()
		})
	}
}
