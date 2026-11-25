package legacy

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestFloorMapV3DenseToV4SparseNativeWire(t *testing.T) {
	const size = 8
	grid := make(floorMapGrid, size)
	for x := range grid {
		column, free := alloc.Make(unsafe.Slice(grid[x], 0), size)
		grid[x] = &column[0]
		t.Cleanup(free)
	}
	tile := grid.cell(2, 3)
	t.Cleanup(func() {
		nox_xxx_tileFreeTile_422200(&tile.field_5)
		nox_xxx_tileFreeTile_422200(&tile.field_10)
	})
	stale := grid.cell(7, 7)
	stale.field_0, stale.field_0_1, stale.field_0_2 = 3, 0xa5, 0xc6d7

	header := binary.LittleEndian.AppendUint32(nil, 1)
	for _, field := range []uint32{2, 4, 3} {
		header = binary.LittleEndian.AppendUint32(header, field)
	}
	layers := [...][][4]int32{
		{{0xa1, -1234, 0xb2, 0xc3}, {0xd1, -42, 0xe1, 0xf1}, {0xd2, -84, 0xe2, 0xf2}},
		{{0xa2, -2222, 0xb3, 0xc4}, {0xd3, -126, 0xe3, 0xf3}},
	}
	appendLayer := func(wire []byte, rows [][4]int32, dense bool) []byte {
		for i, row := range rows {
			if dense {
				for _, value := range row {
					wire = binary.LittleEndian.AppendUint32(wire, uint32(value))
				}
			} else {
				wire = append(wire, byte(row[0]))
				wire = binary.LittleEndian.AppendUint16(wire, uint16(row[1]))
				wire = append(wire, byte(row[2]), byte(row[3]))
			}
			if i == 0 {
				wire = append(wire, byte(len(rows)-1))
			}
		}
		return wire
	}

	denseWant := bytes.Clone(header)
	for _, flag := range []byte{0x80, 0, 0, 0, 0, 7, 0, 0, 0, 0, 0, 4} {
		denseWant = append(denseWant, flag)
		if flag == 7 {
			for _, rows := range layers {
				denseWant = appendLayer(denseWant, rows, true)
			}
		}
	}
	denseInput := filepath.Join(t.TempDir(), "dense-in")
	if err := os.WriteFile(denseInput, denseWant, 0o600); err != nil {
		t.Fatal(err)
	}
	input, err := cryptfile.OpenFile(denseInput, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if err := readDenseFloorMap(input, grid, nil); err != nil {
		t.Fatal(err)
	}
	if stale.field_0 != 0 || stale.field_0_1 != 0xa5 || stale.field_0_2 != 0xc6d7 {
		t.Fatalf("v3 flag clear changed adjacent native fields: %+v", *stale)
	}
	if tile.field_0 != 7 || tile.field_1 != 0xa1 || tile.field_2 != -1234 ||
		tile.field_6 != 0xa2 || tile.field_7 != -2222 ||
		tile.field_5 == nil || tile.field_5.field_1 != -42 ||
		tile.field_5.next == nil || tile.field_5.next.field_1 != -84 || tile.field_5.next.next != nil ||
		tile.field_10 == nil || tile.field_10.field_1 != -126 || tile.field_10.next != nil {
		t.Fatalf("v3 fields or subtile links differ: %+v", *tile)
	}

	sparseWant := binary.LittleEndian.AppendUint16(bytes.Clone(header), 0x8283)
	for _, rows := range layers {
		sparseWant = appendLayer(sparseWant, rows, false)
	}
	sparseWant = binary.LittleEndian.AppendUint16(sparseWant, 0xffff)
	for _, tc := range []struct {
		name string
		want []byte
		save func(*cryptfile.CryptFile, floorMapGrid, *[8]int32) error
	}{
		{"dense-v3", denseWant, writeDenseFloorMap},
		{"sparse-v4", sparseWant, writeFloorMapTiles},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "floor")
			output, err := cryptfile.OpenFile(file, cryptfile.WriteOnly, -1)
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			if err := tc.save(output, grid, nil); err != nil {
				t.Fatal(err)
			}
			if err := output.Close(); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("floor wire=%x, want %x", got, tc.want)
			}
		})
	}
}
