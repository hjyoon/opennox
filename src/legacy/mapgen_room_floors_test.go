package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func TestMapgenRoomFloorInterior(t *testing.T) {
	allocate := func() (unsafe.Pointer, uint32) {
		raw, token := mapgenAllocateDecorRecord(mapgenRoomRecordSize)
		if raw == nil || token == 0 {
			t.Fatal("could not allocate room floor fixture")
		}
		t.Cleanup(func() { mapgenReleaseDecorRecord(raw) })
		return raw, uint32(token)
	}
	raw, _ := allocate()
	ordinary, roomToken := allocate()
	hall, hallToken := allocate()
	*(*uint32)(ordinary) = 1
	*(*uint32)(hall) = 2
	room := (*mapgenRoomGeometry)(raw)
	links := (*mapgenRoomLinkRecord)(raw)
	*(*[2]int32)(unsafe.Pointer(&room.cellDimensions)) = [2]int32{6, 9}
	origin := [2]float32{-234.25, 111.5}
	*(*[2]float32)(unsafe.Pointer(&room.position)) = origin

	const cellSize = 32.526913
	check := func(wantOrigin [2]float32, wantCells [2]int32, first, last int) {
		t.Helper()
		point, cells := mapgenRoomFloorInterior(room)
		gotOrigin := [2]uint32{math.Float32bits(float32(point.field_0)), math.Float32bits(float32(point.field_4))}
		wantBits := [2]uint32{math.Float32bits(wantOrigin[0]), math.Float32bits(wantOrigin[1])}
		gotCells := [2]int32{int32(cells[0]), int32(cells[1])}
		if gotOrigin != wantBits || gotCells != wantCells {
			t.Fatalf("type %d, neighbors %d/%d: origin %x, cells %v, want %x and %v",
				links.kind, first, last, gotOrigin, gotCells, wantBits, wantCells)
		}
	}
	check([2]float32{}, [2]int32{}, 0, 0)
	links.kind = 1
	check([2]float32{origin[0] + cellSize, origin[1] + cellSize}, [2]int32{4, 7}, 0, 0)

	// The first and last neighbors are indexed as missing, room, and hall.
	tokens := [3]uint32{0, roomToken, hallToken}
	lengthChange := [3][3]int32{{-1, -2, 0}, {-1, -2, 0}, {0, -1, 1}}
	nearShift := [3]float32{cellSize, cellSize, 0}
	farShift := [3]float32{0, cellSize, -cellSize}
	for _, tc := range []struct {
		kind               uint32
		first, last, axis  int
		startsAtLowerBound bool
	}{
		{2, 0, 1, 1, true},
		{3, 1, 0, 1, false},
		{4, 2, 3, 0, false},
		{5, 3, 2, 0, true},
	} {
		links.kind = tc.kind
		links.links = [4][8]uint32{}
		for first, firstToken := range tokens {
			for last, lastToken := range tokens {
				links.links[tc.first][0] = firstToken
				links.links[tc.last][0] = lastToken
				point, cells := origin, [2]int32{6, 9}
				cross := 1 - tc.axis
				point[cross] += cellSize
				cells[cross] -= 2
				cells[tc.axis] += lengthChange[first][last]
				if tc.startsAtLowerBound {
					point[tc.axis] += nearShift[first]
				} else {
					point[tc.axis] += farShift[last]
				}
				check(point, cells, first, last)
			}
		}
	}
}
