package legacy

import (
	"math"
	"strconv"
	"testing"
)

func TestMapgenRoomConnections522CA0PreserveNativePointers(t *testing.T) {
	oldPlatformTicks := PlatformTicks
	PlatformTicks = func() uint64 { return 0 }
	t.Cleanup(func() { PlatformTicks = oldPlatformTicks })

	got := mapgenRoomConnectionsFixture522CA0()
	if !got.gridInitialized || !got.roomAdded {
		t.Fatalf("fixture setup failed: grid=%v room=%v", got.gridInitialized, got.roomAdded)
	}
	if got.roomAtCellAddress != got.roomAddress {
		t.Fatalf("room-at-cell = %#x, want %#x", got.roomAtCellAddress, got.roomAddress)
	}
	if got.addFirst != 1 || got.addDuplicate != 1 || got.addLegacyDuplicate != 1 || got.roomPointCount != 1 {
		t.Fatalf("point insertion mismatch: first=%d duplicate=%d legacy=%d count=%d",
			got.addFirst, got.addDuplicate, got.addLegacyDuplicate, got.roomPointCount)
	}
	assertPoint522CA0(t, "stored room point", got.roomPoint, [2]float32{17.25, -31.5})
	if !got.nativeLinesReturnedNull || !got.nativePointsReturnedNull ||
		!got.legacyLinesReturnedNull || !got.legacyPointsReturnedNull {
		t.Fatalf("connection pass return mismatch: native=%v/%v legacy=%v/%v",
			got.nativeLinesReturnedNull, got.nativePointsReturnedNull,
			got.legacyLinesReturnedNull, got.legacyPointsReturnedNull)
	}

	if !got.hallLinksAdded {
		t.Fatal("hall links were not added")
	}
	for i := range got.neighborAddresses {
		if got.linkedNeighborAddresses[i] != got.neighborAddresses[i] {
			t.Fatalf("linked neighbor[%d] = %#x, want %#x",
				i, got.linkedNeighborAddresses[i], got.neighborAddresses[i])
		}
	}
	for i, typ := range []int{2, 3, 4, 5} {
		bounds := got.hallBounds[i]
		midpoint := [2]float32{(bounds[0] + bounds[2]) * 0.5, (bounds[1] + bounds[3]) * 0.5}
		assertPoint522CA0(t, "center", got.centerPoints[i], midpoint)
		var near, far [2]float32
		switch typ {
		case 2:
			near = [2]float32{midpoint[0], bounds[1]}
			far = [2]float32{midpoint[0], bounds[3]}
		case 3:
			near = [2]float32{midpoint[0], bounds[3]}
			far = [2]float32{midpoint[0], bounds[1]}
		case 4:
			near = [2]float32{bounds[2], midpoint[1]}
			far = [2]float32{bounds[0], midpoint[1]}
		case 5:
			near = [2]float32{bounds[0], midpoint[1]}
			far = [2]float32{bounds[2], midpoint[1]}
		}
		assertPoint522CA0(t, "near", got.nearPoints[i], near)
		assertPoint522CA0(t, "far", got.farPoints[i], far)
		if got.hallResolvedAddresses[i] != got.hallAddresses[i] {
			t.Fatalf("resolved hall[%d] = %#x, want %#x",
				i, got.hallResolvedAddresses[i], got.hallAddresses[i])
		}
	}
	assertPoint522CA0(t, "legacy near", got.legacyNearPoint, got.nearPoints[0])
	assertPoint522CA0(t, "legacy far", got.legacyFarPoint, got.farPoints[0])
	assertPoint522CA0(t, "legacy center", got.legacyCenterPoint, got.centerPoints[0])
	if got.legacyCenterReturnAddress != got.hallAddresses[0] {
		t.Fatalf("legacy center return = %#x, want %#x",
			got.legacyCenterReturnAddress, got.hallAddresses[0])
	}
	if !got.adjustRejected {
		t.Fatal("non-overlapping hall adjustment was not rejected")
	}

	if !got.waypointNativeLink || !got.waypointDuplicateRejected || !got.waypointLegacyLink ||
		!got.waypointSelfRejected || !got.waypointCapacityRejected {
		t.Fatalf("waypoint link result mismatch: native=%v duplicate=%v legacy=%v self=%v capacity=%v",
			got.waypointNativeLink, got.waypointDuplicateRejected, got.waypointLegacyLink,
			got.waypointSelfRejected, got.waypointCapacityRejected)
	}
	if got.waypointCounts != [2]uint8{1, 1} {
		t.Fatalf("waypoint counts = %v, want [1 1]", got.waypointCounts)
	}
	if got.waypointLinkAddresses != [2]uintptr{got.waypointAddresses[1], got.waypointAddresses[0]} {
		t.Fatalf("waypoint links = %#x, want %#x",
			got.waypointLinkAddresses, [2]uintptr{got.waypointAddresses[1], got.waypointAddresses[0]})
	}
	if got.waypointKinds != [2]uint8{7, 9} {
		t.Fatalf("waypoint kinds = %v, want [7 9]", got.waypointKinds)
	}
	if !got.cleanupHeadIsNull {
		t.Fatal("mapgen room list survived fixture cleanup")
	}

	if strconv.IntSize == 64 {
		addresses := []uintptr{got.themeAddress, got.roomAddress, got.hallAddresses[0],
			got.hallAddresses[1], got.hallAddresses[2], got.hallAddresses[3],
			got.neighborAddresses[0], got.neighborAddresses[1],
			got.waypointAddresses[0], got.waypointAddresses[1]}
		for i, address := range addresses {
			if address <= 1<<32 {
				t.Fatalf("fixture address[%d] = %#x, want address above PE32 range", i, address)
			}
		}
		tokens := []uint32{got.themeToken, got.roomToken, got.hallTokens[0], got.hallTokens[1],
			got.hallTokens[2], got.hallTokens[3], got.waypointTokens[0], got.waypointTokens[1]}
		for i, token := range tokens {
			if token == 0 {
				t.Fatalf("fixture token[%d] is zero", i)
			}
		}
	}
}

func assertPoint522CA0(t *testing.T, name string, got, want [2]float32) {
	t.Helper()
	const epsilon = 1e-4
	if math.Abs(float64(got[0]-want[0])) > epsilon || math.Abs(float64(got[1]-want[1])) > epsilon {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}
