package legacy

import (
	"strconv"
	"testing"
)

func TestMapgenRoomGrid521100PreservesNativePointers(t *testing.T) {
	got := mapgenRoomGridFixture521100()
	if !got.gridInitialized || !got.firstAdded || !got.secondAdded || !got.secondRemoved {
		t.Fatalf("lifecycle failed: grid=%v first=%v second=%v removed=%v", got.gridInitialized, got.firstAdded, got.secondAdded, got.secondRemoved)
	}
	if got.collisionAddress != got.firstAddress {
		t.Fatalf("collision room = %#x, want first room %#x", got.collisionAddress, got.firstAddress)
	}
	if got.firstHeadAddress != got.firstAddress {
		t.Fatalf("first head = %#x, want %#x", got.firstHeadAddress, got.firstAddress)
	}
	if got.secondHeadAddress != got.secondAddress {
		t.Fatalf("second head = %#x, want %#x", got.secondHeadAddress, got.secondAddress)
	}
	if got.secondNextAddress != got.firstAddress {
		t.Fatalf("second next = %#x, want first room %#x", got.secondNextAddress, got.firstAddress)
	}
	if got.removedHead != got.firstAddress {
		t.Fatalf("head after removal = %#x, want first room %#x", got.removedHead, got.firstAddress)
	}
	if got.vacatedAddress != 0 {
		t.Fatalf("vacated grid still resolves room %#x", got.vacatedAddress)
	}
	if strconv.IntSize == 64 {
		for name, address := range map[string]uintptr{
			"theme":  got.themeAddress,
			"first":  got.firstAddress,
			"second": got.secondAddress,
		} {
			if address <= 1<<32 {
				t.Fatalf("%s fixture address = %#x, want address above PE32 range", name, address)
			}
		}
	}
}
