package legacy

import (
	"math"
	"strconv"
	"testing"
)

func TestMapgenRoomRanking5259F0PreservesNativePointers(t *testing.T) {
	oldPlatformTicks := PlatformTicks
	PlatformTicks = func() uint64 { return 0 }
	t.Cleanup(func() { PlatformTicks = oldPlatformTicks })

	got := mapgenRoomRankingFixture5259F0()
	if !got.gridInitialized || got.roomsAdded != 5 || got.linksAdded != 4 {
		t.Fatalf("fixture setup failed: grid=%v rooms=%d links=%d", got.gridInitialized, got.roomsAdded, got.linksAdded)
	}
	if got.farthestAddress != got.roomAddresses[4] ||
		got.nativeFarthestAddress != got.roomAddresses[4] ||
		got.legacyFarthestAddress != got.roomAddresses[4] {
		t.Fatalf("farthest mismatch: return=%#x native=%#x legacy=%#x want=%#x",
			got.farthestAddress, got.nativeFarthestAddress, got.legacyFarthestAddress, got.roomAddresses[4])
	}
	if got.sortedHeadAddress != got.roomAddresses[0] || got.legacySortedHeadAddress != got.roomAddresses[0] {
		t.Fatalf("sorted head mismatch: native=%#x legacy=%#x want=%#x",
			got.sortedHeadAddress, got.legacySortedHeadAddress, got.roomAddresses[0])
	}

	wantBuckets := [5]uint32{1, 2, 4, 8, 32}
	for i := range got.roomAddresses {
		if got.sortedAddresses[i] != got.roomAddresses[i] {
			t.Fatalf("sorted[%d] = %#x, want room %#x", i, got.sortedAddresses[i], got.roomAddresses[i])
		}
		wantPrev := uintptr(0)
		wantNext := uintptr(0)
		if i > 0 {
			wantPrev = got.roomAddresses[i-1]
		}
		if i+1 < len(got.roomAddresses) {
			wantNext = got.roomAddresses[i+1]
		}
		if got.sortedPrevAddresses[i] != wantPrev || got.sortedNextAddresses[i] != wantNext {
			t.Fatalf("sorted links[%d] = prev %#x next %#x, want %#x/%#x",
				i, got.sortedPrevAddresses[i], got.sortedNextAddresses[i], wantPrev, wantNext)
		}
		if (i > 0 && got.sortedPrevTokens[i] == 0) || (i+1 < len(got.roomAddresses) && got.sortedNextTokens[i] == 0) {
			t.Fatalf("sorted token[%d] unexpectedly zero: prev=%#x next=%#x",
				i, got.sortedPrevTokens[i], got.sortedNextTokens[i])
		}
		if got.buckets[i] != wantBuckets[i] {
			t.Fatalf("bucket[%d] = %d, want %d", i, got.buckets[i], wantBuckets[i])
		}
		if got.visits[i] != 2 {
			t.Fatalf("visit[%d] = %d, want 2", i, got.visits[i])
		}
		if i > 0 && !(got.distances[i] > got.distances[i-1]) {
			t.Fatalf("distances not strictly ascending: %v", got.distances)
		}
		wantNormalized := float32(i) / 4
		if math.Abs(float64(got.normalized[i]-wantNormalized)) > 1e-5 {
			t.Fatalf("normalized[%d] = %v, want %v", i, got.normalized[i], wantNormalized)
		}
	}
	if got.flags[4]&1 == 0 {
		t.Fatalf("farthest room flags = %#x, want bit 0 set", got.flags[4])
	}
	if got.maxDistance != got.distances[4] {
		t.Fatalf("max distance = %v, want final sorted distance %v", got.maxDistance, got.distances[4])
	}
	if got.farthestAfterCleanup != 0 || got.sortedHeadAfterCleanup != 0 || got.maxDistanceAfterCleanup != 0 {
		t.Fatalf("ranking state survived cleanup: farthest=%#x head=%#x max=%v",
			got.farthestAfterCleanup, got.sortedHeadAfterCleanup, got.maxDistanceAfterCleanup)
	}

	if strconv.IntSize == 64 {
		if got.themeAddress <= 1<<32 {
			t.Fatalf("theme fixture address = %#x, want address above PE32 range", got.themeAddress)
		}
		for i, address := range got.roomAddresses {
			if address <= 1<<32 {
				t.Fatalf("room[%d] fixture address = %#x, want address above PE32 range", i, address)
			}
			if got.roomTokens[i] == 0 {
				t.Fatalf("room[%d] token is zero", i)
			}
		}
	}
}
