package legacy

import (
	"strconv"
	"testing"
)

func TestMapgenOccupiedRects521BC0PreserveNativePointers(t *testing.T) {
	got := mapgenOccupiedRectsFixture521BC0()
	if !got.legacyWrapperAdded {
		t.Fatal("legacy wrapper did not add the first occupied rectangle")
	}
	if got.headAddress != got.secondAddress {
		t.Fatalf("head = %#x, want second rectangle %#x", got.headAddress, got.secondAddress)
	}
	if got.nextAddress != got.firstAddress {
		t.Fatalf("head next = %#x, want first rectangle %#x", got.nextAddress, got.firstAddress)
	}
	if !got.intersects {
		t.Fatal("overlapping rectangle was not detected")
	}
	if !got.pointOccupied {
		t.Fatal("point inside occupied rectangle was not detected")
	}
	if !got.transientRemoved || got.remainingAddress != got.firstAddress {
		t.Fatalf("transient removal left %#x, want first rectangle %#x", got.remainingAddress, got.firstAddress)
	}
	if !got.allRemoved {
		t.Fatal("occupied rectangle list was not emptied")
	}
	if got.storedHeadToken == 0 {
		t.Fatal("room stored a zero occupied-rectangle token")
	}
	if strconv.IntSize == 64 {
		for name, address := range map[string]uintptr{
			"room":   got.roomAddress,
			"first":  got.firstAddress,
			"second": got.secondAddress,
		} {
			if address <= 1<<32 {
				t.Fatalf("%s fixture address = %#x, want address above PE32 range", name, address)
			}
		}
		if uintptr(got.storedHeadToken) == got.headAddress {
			t.Fatalf("stored head %#x unexpectedly equals native address %#x", got.storedHeadToken, got.headAddress)
		}
	}
}
