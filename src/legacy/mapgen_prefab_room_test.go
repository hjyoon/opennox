package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestMapgenPrefabStartConsumesWrittenPositions(t *testing.T) {
	InitBlobData()
	raw, _ := mapgenAllocateDecorRecord(mapgenRoomRecordSize)
	if raw == nil {
		t.Fatal("could not allocate prefab room fixture")
	}
	t.Cleanup(func() { mapgenReleaseDecorRecord(raw) })
	room := (*mapgenRoomGeometry)(raw)
	*(*[2]float32)(unsafe.Pointer(&room.dimensions)) = [2]float32{60, 80}

	slots := mapgenPrefabStartCells()
	var previousSlots [10]uint32
	for i, slot := range slots {
		previousSlots[i] = *slot
	}
	order := (*[5]uint32)(memmap.PtrOff(0x5D4594, 2487588))
	cursor := memmap.PtrUint32(0x5D4594, 2487608)
	previousOrder, previousCursor := *order, *cursor
	t.Cleanup(func() {
		*order, *cursor = previousOrder, previousCursor
		for i, slot := range slots {
			*slot = previousSlots[i]
		}
	})

	mapgenStorePrefabStartPositions([][2]float32{
		{90, 140}, {123, -456}, {999, -275}, {-950, 470}, {-72, -91},
	})
	*order = [5]uint32{1, 2, 3, 4, 0}
	*cursor = 0
	for step, want := range [5][2]int32{{2, -15}, {29, -9}, {-30, 13}, {-3, -4}, {1, 3}} {
		world, y := mapgenPrefabRoomStartFixture(raw)
		const cell = 32.526913
		expected := [2]float32{float32(float64(want[0]) * cell), float32(float64(want[1]) * cell)}
		if world != expected || y != int64(want[1]) || *cursor != uint32(step+1) {
			t.Fatalf("step %d: origin %v, Y cell %d, cursor %d, want %v, %d, %d",
				step, world, y, *cursor, expected, want[1], step+1)
		}
	}
}
