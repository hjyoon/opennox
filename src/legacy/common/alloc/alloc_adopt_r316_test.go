package alloc

import (
	"testing"
	"unsafe"
)

func TestAdoptCAllocation316KeepsExistingOwner(t *testing.T) {
	p, free := New(uint64(0))
	defer free()
	*p = 0x123456789abcdef0
	AdoptCAllocation(unsafe.Pointer(p), unsafe.Sizeof(*p))
	AdoptCAllocation(unsafe.Pointer(p), unsafe.Sizeof(*p))
	if *p != 0x123456789abcdef0 {
		t.Fatal("adoption changed the allocation contents")
	}
	allocMu.Lock()
	size, ok := allocs[unsafe.Pointer(p)]
	allocMu.Unlock()
	if !ok || size != unsafe.Sizeof(*p) {
		t.Fatal("adoption lost the existing owner or size")
	}
}

func TestAdoptCAllocation316RejectsInvalidOwnership(t *testing.T) {
	p, free := New(uint64(0))
	defer free()
	for _, tc := range []struct {
		name       string
		ptr        unsafe.Pointer
		size       uintptr
		panicValue string
	}{
		{"nil", nil, 8, "invalid C allocation"},
		{"zero_size", unsafe.Pointer(p), 0, "invalid C allocation"},
		{"conflicting_size", unsafe.Pointer(p), 1, "conflicting C allocation size"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tc.panicValue {
					t.Fatalf("panic = %v, want %q", got, tc.panicValue)
				}
			}()
			AdoptCAllocation(tc.ptr, tc.size)
		})
	}
	allocMu.Lock()
	size, ok := allocs[unsafe.Pointer(p)]
	allocMu.Unlock()
	if !ok || size != unsafe.Sizeof(*p) {
		t.Fatal("rejected adoption changed existing ownership")
	}
}
