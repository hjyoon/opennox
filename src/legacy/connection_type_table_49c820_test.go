package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestConnectionTypeTable49C820UsesPackedPointerSlots(t *testing.T) {
	InitBlobData()

	const (
		tableBase = uintptr(0x587000)
		tableOff  = uintptr(164928)
	)
	wantPacked := []byte{
		0x50, 0xf4, 0x5a, 0x00,
		0x60, 0xf4, 0x5a, 0x00,
		0x70, 0xf4, 0x5a, 0x00,
		0x80, 0xf4, 0x5a, 0x00,
	}
	if got := memmap.Slice(tableBase, tableOff)[:len(wantPacked)]; unsafe.Sizeof(uintptr(0)) > 4 && !bytes.Equal(got, wantPacked) {
		t.Fatalf("packed PE32 connection-type table = %x, want %x", got, wantPacked)
	}

	want := []string{
		"general.c:T1",
		"general.c:Cable",
		"general.c:ISDN",
		"general.c:Modem",
	}
	for i, expected := range want {
		off := tableOff + 4*uintptr(i)
		ptr := *memmap.PtrPtr(tableBase, off)
		if ptr == nil {
			t.Fatalf("connection type %d pointer at offset %d is nil", i, off)
		}
		if got := GoStringP(ptr); got != expected {
			t.Fatalf("connection type %d = %q, want %q", i, got, expected)
		}
	}
}
