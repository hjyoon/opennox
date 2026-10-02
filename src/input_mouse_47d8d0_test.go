package opennox

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

func TestOptionsMouseInit47D8D0PublishesButtonCapability(t *testing.T) {
	// The original stores one capability BYTE at 006F7A3C and a present DWORD
	// at 006F7A28. Neither write may touch the packed globals between and
	// after them. Start after the separately extracted noxInputMap region.
	region := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 1193108)), 28)
	saved := bytes.Clone(region)
	t.Cleanup(func() { copy(region, saved) })
	for previous := range 256 {
		for i := range region {
			region[i] = byte(11*i + previous)
		}
		region[20] = byte(previous)
		want := bytes.Clone(region)
		binary.NativeEndian.PutUint32(want[0:4], 1)
		want[20] = 3 // Native input accepts Left, Right and Middle, not wheel.
		inputInitMouse()
		if got := legacy.InputMouseButtonCount47DBC0(); got != 3 {
			t.Fatalf("previous=%d: C capability BYTE=%d, want 3 supported buttons", previous, got)
		}
		if !bytes.Equal(region, want) {
			t.Fatalf("previous=%d: packed globals=%x, want %x", previous, region, want)
		}
	}
}

func TestOptionsMouseButtonCount47DBC0IsByte(t *testing.T) {
	region := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 1193127)), 5)
	saved := bytes.Clone(region)
	t.Cleanup(func() { copy(region, saved) })
	for value := range 256 {
		copy(region, []byte{0xe1, byte(value), 0xff, 0x80, 0x7f})
		before := bytes.Clone(region)
		if got := legacy.InputMouseButtonCount47DBC0(); int(got) != value {
			t.Fatalf("BYTE=%d: C getter=%d", value, got)
		}
		if !bytes.Equal(region, before) {
			t.Fatalf("getter changed packed globals: %x, want %x", region, before)
		}
	}
}
