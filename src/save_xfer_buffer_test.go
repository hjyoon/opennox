package opennox

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

func TestXferBufferUsesNativePointer(t *testing.T) {
	const ind = 1
	xferFree446580(ind)
	t.Cleanup(func() { xferFree446580(ind) })

	want := []byte("arm64 MOTD buffer\nsecond line")
	xferSet446520(ind, want)

	ptr := legacy.GetXferBuffer446520(ind)
	if ptr == nil {
		t.Fatal("native transfer buffer is nil")
	}
	if got := *memmap.PtrPtr(0x5D4594, 826056+4*ind); got != ptr {
		t.Fatalf("Go transfer buffer = %p, native transfer buffer = %p", got, ptr)
	}
	got := unsafe.Slice((*byte)(ptr), len(want))
	if !bytes.Equal(got, want) {
		t.Fatalf("native transfer buffer = %q, want %q", got, want)
	}
	if got := memmap.Uint32(0x5D4594, 826048+4*ind); got != uint32(len(want)) {
		t.Fatalf("transfer length = %d, want %d", got, len(want))
	}
	if got := memmap.Uint32(0x5D4594, 826064+4*ind); got != 1 {
		t.Fatalf("transfer ready flag = %d, want 1", got)
	}

	xferFree446580(ind)
	if got := legacy.GetXferBuffer446520(ind); got != nil {
		t.Fatalf("native transfer buffer after free = %p, want nil", got)
	}
	if got := *memmap.PtrPtr(0x5D4594, 826056+4*ind); got != nil {
		t.Fatalf("Go transfer buffer after free = %p, want nil", got)
	}
	if got := memmap.Uint32(0x5D4594, 826048+4*ind); got != 0 {
		t.Fatalf("transfer length after free = %d, want 0", got)
	}
	if got := memmap.Uint32(0x5D4594, 826064+4*ind); got != 0 {
		t.Fatalf("transfer ready flag after free = %d, want 0", got)
	}
}

func TestXferBufferReplacementAndEmptyReset(t *testing.T) {
	const ind = 1
	xferFree446580(ind)
	t.Cleanup(func() { xferFree446580(ind) })

	xferSet446520(ind, []byte("first"))
	xferSet446520(ind, []byte("replacement"))
	ptr := legacy.GetXferBuffer446520(ind)
	if ptr == nil {
		t.Fatal("replacement transfer buffer is nil")
	}
	if got := unsafe.Slice((*byte)(ptr), len("replacement")); !bytes.Equal(got, []byte("replacement")) {
		t.Fatalf("replacement transfer buffer = %q", got)
	}

	xferSet446520(ind, nil)
	if got := legacy.GetXferBuffer446520(ind); got != nil {
		t.Fatalf("native transfer buffer after empty replacement = %p, want nil", got)
	}
	if got := memmap.Uint32(0x5D4594, 826064+4*ind); got != 0 {
		t.Fatalf("transfer ready flag after empty replacement = %d, want 0", got)
	}
}
