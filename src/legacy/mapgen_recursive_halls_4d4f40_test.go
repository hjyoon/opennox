package legacy

import (
	"math"
	"strconv"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestMapgenHallExpansionKindKeepsRandomDrawOrder(t *testing.T) {
	preserveLegacyCRTRand315(t)
	branchRate, roomRate := memmap.PtrInt32(0x5D4594, 1549820), memmap.PtrInt32(0x5D4594, 1549824)
	oldBranch, oldRoom := *branchRate, *roomRate
	t.Cleanup(func() { *branchRate, *roomRate = oldBranch, oldRoom })

	for _, tc := range []struct {
		name           string
		kind           uint32
		width, height  int32
		branch, room   int32
		seed           uint32
		want, nextDraw int32
	}{
		{"square", 2, 5, 5, 100, 100, 1, 1, 41},
		{"negative width as unsigned", 2, -1, 5, 100, 100, 1, 1, 41},
		{"vertical branch", 2, 3, 9, 1, 0, 1, 32, 6334},
		{"horizontal room", 4, 9, 3, 0, 57, 1, 1, 6334},
		{"straight", 2, 3, 9, 0, 56, 1, 2, 26500},
		{"elbow", 2, 3, 9, 0, 89, 2, 4, 17795},
	} {
		t.Run(tc.name, func(t *testing.T) {
			*branchRate, *roomRate = tc.branch, tc.room
			kind := mapgenHallExpansionKindFixture(tc.seed, tc.kind, tc.width, tc.height)
			if next := nox_platform_rand(); kind != tc.want || next != tc.nextDraw {
				t.Fatalf("expansion kind and next draw = %d, %d, want %d, %d", kind, next, tc.want, tc.nextDraw)
			}
		})
	}
}

func TestMapgenRoomExpansionUsesUnblockedDirections(t *testing.T) {
	InitBlobData()
	preserveLegacyCRTRand315(t)
	oldTicks := PlatformTicks
	PlatformTicks = func() uint64 { return 0 }
	t.Cleanup(func() { PlatformTicks = oldTicks })
	result, links := mapgenRoomExpansionFixture()
	if result != 1 || links[0] != 1 || links[1] == 0 || links[2] == 0 || links[3] == 0 {
		t.Fatalf("expansion returned %d with direction counts %v, want the original north link and new links on the other three sides", result, links)
	}
}

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

func TestMapgenHallConnectionPointsNative(t *testing.T) {
	for _, tc := range []struct {
		kind  int32
		x, y  [3]float32
		edges [2]float32
	}{
		{2, [3]float32{30, 30, 30}, [3]float32{20, 80, 23}, [2]float32{20, 80}},
		{3, [3]float32{30, 30, 30}, [3]float32{80, 20, 77}, [2]float32{}},
		{4, [3]float32{50, 10, 46}, [3]float32{50, 50, 50}, [2]float32{50, 10}},
		{5, [3]float32{10, 50, 14}, [3]float32{50, 50, 50}, [2]float32{10, 50}},
	} {
		for _, linked := range []bool{false, true} {
			got, words, token := mapgenHallPointsFixture(tc.kind, linked)
			want := [3][2]float32{{tc.x[0], tc.y[0]}, {tc.x[1], tc.y[1]}, {tc.x[2], tc.y[2]}}
			if linked {
				want[2] = [2]float32{30, 50}
			}
			returns := [2]uint32{math.Float32bits(tc.edges[0]), math.Float32bits(tc.edges[1])}
			if tc.kind == 3 {
				returns = [2]uint32{token, token}
			}
			if token == 0 || got != want || words != returns {
				t.Fatalf("hall type %d linked=%t: points %v, words %x, want %v and %x", tc.kind, linked, got, words, want, returns)
			}
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

func TestMapgenRoomExpansionUsesThemeMergeRate(t *testing.T) {
	InitBlobData()
	preserveLegacyCRTRand315(t)
	for _, tc := range []struct {
		rate, wantLength int32
		wantLinks        uint8
	}{
		{-1, 0, 0},
		{0, 0, 0},
		{100, 4, 1},
	} {
		joined, length, links := mapgenHallMergeRateFixture(tc.rate)
		if joined != (tc.wantLinks != 0) || length != tc.wantLength || links != tc.wantLinks {
			t.Errorf("merge rate %d: joined %v, hall length %d and target links %d, want length %d and links %d",
				tc.rate, joined, length, links, tc.wantLength, tc.wantLinks)
		}
	}
}
