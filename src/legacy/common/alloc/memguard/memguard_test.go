package memguard

import (
	"testing"
	"unsafe"
)

func TestNewPage(t *testing.T) {
	size := PageSize() + 1
	b, free := New(size)
	defer free()

	if len(b) != size {
		t.Fatalf("length = %d, want %d", len(b), size)
	}
	addr := uintptr(unsafe.Pointer(unsafe.SliceData(b)))
	if addr%uintptr(PageSize()) != 0 {
		t.Fatalf("address %#x is not aligned to page size %d", addr, PageSize())
	}
}
