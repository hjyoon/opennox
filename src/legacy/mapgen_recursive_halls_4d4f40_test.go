package legacy

import (
	"strconv"
	"testing"
)

func TestMapgenRecursiveHalls4D4F40PreserveNativePointers(t *testing.T) {
	InitBlobData()

	oldPlatformTicks := PlatformTicks
	PlatformTicks = func() uint64 { return 0 }
	t.Cleanup(func() { PlatformTicks = oldPlatformTicks })

	got := mapgenRecursiveHallsFixture4D4F40()
	if !got.specialGridInitialized {
		t.Fatal("special-layout occupancy grid was not initialized")
	}
	if got.specialRoomCount != 48 {
		t.Fatalf("special-layout room count = %d, want 48", got.specialRoomCount)
	}
	if got.specialRootAddress == 0 || got.specialRootToken == 0 ||
		got.specialRootResolvedAddress != got.specialRootAddress {
		t.Fatalf("special root mismatch: address=%#x token=%#x resolved=%#x",
			got.specialRootAddress, got.specialRootToken, got.specialRootResolvedAddress)
	}
	if got.trackingHeadToken == 0 || got.trackingHeadAddress == 0 ||
		!got.trackingValid || got.trackedRoomCount == 0 || got.trackedRoomCount > got.specialRoomCount {
		t.Fatalf("tracking chain invalid: token=%#x address=%#x valid=%v count=%d",
			got.trackingHeadToken, got.trackingHeadAddress, got.trackingValid, got.trackedRoomCount)
	}
	if !got.traversalPreservedRooms {
		t.Fatal("tracking-chain traversal changed the special-layout room list")
	}

	if !got.recursiveGridInitialized || got.recursiveResult != 1 {
		t.Fatalf("recursive fixture failed: grid=%v result=%d",
			got.recursiveGridInitialized, got.recursiveResult)
	}
	if got.recursiveRoomCount != 2 || got.recursiveRootLinkCount != 1 {
		t.Fatalf("recursive expansion mismatch: rooms=%d root links=%d, want 2/1",
			got.recursiveRoomCount, got.recursiveRootLinkCount)
	}
	if got.recursiveRootAddress == 0 || got.recursiveChildAddress == 0 ||
		got.recursiveLinkToken == 0 || got.recursiveLinkAddress != got.recursiveChildAddress {
		t.Fatalf("recursive link mismatch: root=%#x child=%#x token=%#x resolved=%#x",
			got.recursiveRootAddress, got.recursiveChildAddress,
			got.recursiveLinkToken, got.recursiveLinkAddress)
	}
	if got.recursiveChildType != 1 {
		t.Fatalf("recursive child type = %d, want room type 1", got.recursiveChildType)
	}
	if !got.cleanupHeadIsNull {
		t.Fatal("mapgen room list survived fixture cleanup")
	}

	if strconv.IntSize == 64 {
		addresses := []uintptr{
			got.themeAddress,
			got.specialRootAddress,
			got.trackingHeadAddress,
			got.recursiveRootAddress,
			got.recursiveChildAddress,
		}
		for i, address := range addresses {
			if address <= 1<<32 {
				t.Fatalf("fixture address[%d] = %#x, want address above PE32 range", i, address)
			}
		}
	}
}
