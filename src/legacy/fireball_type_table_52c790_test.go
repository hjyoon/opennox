package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestFireballTypeTable52C790UsesPackedPointerSlots(t *testing.T) {
	InitBlobData()
	want := []string{
		"Fireball",
		"StrongFireball",
		"TitanFireball",
		"TitanFireball",
		"TitanFireball",
	}
	for level, expected := range want {
		off := uintptr(258864 + 4*(level+1))
		ptr := *memmap.PtrPtr(0x587000, off)
		if ptr == nil {
			t.Fatalf("level %d type pointer at offset %d is nil", level+1, off)
		}
		if got := GoStringP(ptr); got != expected {
			t.Fatalf("level %d type = %q, want %q", level+1, got, expected)
		}
	}
}
