package legacy

import (
	"math"
	"strconv"
	"testing"
)

func TestMapgenHallRecordNative523E30(t *testing.T) {
	for _, tc := range []struct {
		kind, width, depth int32
		cells              [2]int32
	}{
		{1, 5, 6, [2]int32{0, 0}},
		{2, -3, 7, [2]int32{1, 7}},
		{3, 4, -2, [2]int32{4, -2}},
		{4, 0, -8, [2]int32{-8, 1}},
		{5, math.MaxInt32, 3, [2]int32{3, math.MaxInt32}},
	} {
		want := [9]uint32{0: uint32(tc.kind), 3: uint32(tc.cells[0]), 4: uint32(tc.cells[1])}
		for i, width := range tc.cells {
			want[7+i] = math.Float32bits(float32(float64(width) * 32.526913))
		}
		if got := mapgenHallRecordFixture(tc.kind, tc.width, tc.depth); got != want {
			t.Fatalf("hall type %d width %d depth %d: got %#x, want %#x", tc.kind, tc.width, tc.depth, got, want)
		}
	}
}

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
			if uint64(address) <= 1<<32 {
				t.Fatalf("fixture address[%d] = %#x, want address above PE32 range", i, address)
			}
		}
	}
}
