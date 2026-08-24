package memmap

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestPtrPtrPreservesPackedPE32Neighbors(t *testing.T) {
	const base = uintptr(0x12340000)
	data := []byte{
		0, 0, 0, 0,
		0x78, 0x56, 0x34, 0x12,
		0xef, 0xcd, 0xab, 0x90,
	}
	RegisterBlobData(base, "pointer_test", data)
	ResetPointerSlots()
	wantBytes := append([]byte(nil), data...)

	target := new(byte)
	want := unsafe.Pointer(target)
	*PtrPtr(base, 0) = want
	if got := *PtrPtr(base, 0); got != want {
		t.Fatalf("PtrPtr value = %p, want %p", got, want)
	}
	if ptrSize > 4 && !bytes.Equal(data, wantBytes) {
		t.Fatalf("native pointer write changed packed PE32 bytes: got %x, want %x", data, wantBytes)
	}
}

func TestResetPointerSlots(t *testing.T) {
	if ptrSize == 4 {
		t.Skip("32-bit builds store pointers directly in PE32 blobs")
	}
	const base = uintptr(0x12350000)
	RegisterBlobData(base, "pointer_reset_test", make([]byte, 8))
	*PtrPtr(base, 0) = unsafe.Pointer(new(byte))
	ResetPointerSlots()
	if got := *PtrPtr(base, 0); got != nil {
		t.Fatalf("pointer slot survived reset: %p", got)
	}
}
