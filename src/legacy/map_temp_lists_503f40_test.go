package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestMapTempLists503F40NativePointersAndCleanup(t *testing.T) {
	got := mapTempListsFixture503F40()
	if !got.wallLinksOK || !got.wallLookupOK || !got.wallValuesOK || !got.wallCleanupOK ||
		!got.wallPayloadPreservedOK {
		t.Fatalf("temporary wall list failed: %+v", got)
	}
	if !got.waypointLinksOK || !got.waypointValuesOK || !got.waypointCleanupOK ||
		!got.waypointPayloadPreservedOK {
		t.Fatalf("temporary waypoint list failed: %+v", got)
	}
	if !got.tileValuesOK || !got.tileCleanupOK {
		t.Fatalf("temporary tile list failed: %+v", got)
	}
	wantPointerSize := uint32(unsafe.Sizeof(uintptr(0)))
	if got.pointerSize != wantPointerSize {
		t.Fatalf("C pointer size = %d, want %d", got.pointerSize, wantPointerSize)
	}
	if wantPointerSize == 8 {
		if got.wallSize != 56 || got.wallNodeSize != 24 || got.waypointSize != 800 ||
			got.waypointNodeSize != 24 || got.tileEntrySize != 40 || got.tileLayerSize != 24 {
			t.Fatalf("native structure sizes = wall %d/%d, waypoint %d/%d, tile %d/%d",
				got.wallSize, got.wallNodeSize, got.waypointSize, got.waypointNodeSize,
				got.tileEntrySize, got.tileLayerSize)
		}
		if runtime.GOOS != "windows" {
			for i, address := range got.addresses {
				if address <= math.MaxUint32 {
					t.Fatalf("fixture address %d = %#x, want above 4 GiB", i, address)
				}
			}
		}
	}
}
