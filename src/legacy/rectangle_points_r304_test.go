package legacy

import (
	"math"
	"slices"
	"testing"
	"unsafe"
)

func TestFloorBrushProducerPointsOriginalR304(t *testing.T) {
	// These bits come from GAME.EXE 5245A0 and 524610 using the gameplay x87 mode.
	rectangle := [][2]uint32{
		{0x42e906e3, 0xc209f238},
		{0x43150a55, 0xc209f238},
		{0x42e906e3, 0xbffad520},
		{0x43150a55, 0xbffad520},
	}
	column := [][2]uint32{
		{0x41d21b8f, 0x42110dc7},
		{0x41d21b8f, 0x428994ab},
		{0x41d21b8f, 0x42caa272},
	}
	cases := []struct {
		name          string
		column        bool
		columns, rows int32
		points        [][2]uint32
	}{
		{name: "rectangle-two-by-two", columns: 2, rows: 2, points: rectangle},
		{name: "column-three", column: true, rows: 3, points: column},
		{name: "rectangle-zero-width", rows: 2},
		{name: "rectangle-zero-height", columns: 2},
		{name: "rectangle-negative-width", columns: -2, rows: 2},
		{name: "rectangle-negative-height", columns: 2, rows: -2},
		{name: "column-zero", column: true},
		{name: "column-negative", column: true, rows: -3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record := struct {
				prefix uint32
				point  mapgenWorldPoint
				suffix uint32
			}{prefix: 0x12345678, point: mapgenWorldPoint{x: 100.25, y: -50.75}, suffix: 0x9abcdef0}
			if tc.column {
				record.point = mapgenWorldPoint{x: 10, y: 20}
			}
			before := record
			acknowledgement := uintptr(0x12345678)
			if unsafe.Sizeof(uintptr(0)) == 8 {
				fullWidth := uint64(0xf1212345678)
				acknowledgement = uintptr(fullWidth)
			}
			var observed [][2]uint32
			place := func(point *mapgenWorldPoint) uintptr {
				if point == &record.point {
					t.Fatal("tile placement changed from a stack position to the input record")
				}
				observed = append(observed, [2]uint32{math.Float32bits(point.x), math.Float32bits(point.y)})
				return acknowledgement
			}
			var got, want uintptr
			if tc.column {
				got = mapgenFillTileColumn(&record.point, tc.rows, place)
			} else {
				got = mapgenFillTileRectangle(&record.point, tc.columns, tc.rows, place)
				want = uintptr(int64(tc.rows))
			}
			if len(tc.points) != 0 {
				want = acknowledgement
			}
			if !slices.Equal(observed, tc.points) || got != want || record != before {
				t.Fatalf("tile fill points %08x, result %#x, record %+v, expected %08x, %#x, %+v",
					observed, got, record, tc.points, want, before)
			}
		})
	}
}
