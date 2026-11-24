package legacy

import (
	"encoding/binary"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMapgenWallStripsNative(t *testing.T) {
	InitBlobData()
	srv := newMapgenTestWallGrid(t)
	// An empty wall definition has a 40-byte aligned header and zero sprites per direction.
	size := 40 + 8*int(server.WallDirLeftHalfArrowUp-server.WallDirUp+1)
	data, _ := alloc.Make([]byte{}, size+4)
	binary.LittleEndian.PutUint32(data[size:], 0x454e4420)
	file := binfile.NewMemFile(unsafe.Pointer(&data[0]), len(data))
	t.Cleanup(file.Free)
	if err := srv.Walls.ReadWall(file); err != nil {
		t.Fatal(err)
	}

	tile, direction := memmap.PtrUint32(0x973F18, 35948), memmap.PtrUint32(0x973F18, 35952)
	previousTile, previousDirection := *tile, *direction
	t.Cleanup(func() { *tile, *direction = previousTile, previousDirection })
	*tile = 0
	if got := mapgenWallStripFixture(0, 0, -3, false); got != -3 || *direction != 0 {
		t.Fatalf("empty horizontal strip returned %d with direction %d", got, *direction)
	}
	if got := mapgenWallStripFixture(0, 0, 0, true); got != 1 || *direction != 1 {
		t.Fatalf("empty vertical strip returned %d with direction %d", got, *direction)
	}

	const cell = 32.526913
	if got := mapgenWallStripFixture(cell, 0, 4, false); got != 1 {
		t.Fatalf("horizontal strip returned %d", got)
	}
	if got := mapgenWallStripFixture(0, cell, 4, true); got != 1 {
		t.Fatalf("vertical strip returned %d", got)
	}
	want := map[image.Point]byte{
		image.Pt(129, 127): 0, image.Pt(130, 126): 0, image.Pt(131, 125): 0, image.Pt(132, 124): 0,
		image.Pt(129, 129): 1, image.Pt(130, 130): 1, image.Pt(131, 131): 1, image.Pt(132, 132): 1,
	}
	srv.Walls.EachWallRaw(func(wall *server.Wall) bool {
		pos := wall.GridPos()
		dir, present := want[pos]
		if !present || wall.Dir0 != dir || wall.Tile1 != 0 || wall.Flags4 != 0x80 {
			t.Errorf("unexpected wall at %v with direction %d, tile %d, flags %#x", pos, wall.Dir0, wall.Tile1, wall.Flags4)
		}
		delete(want, pos)
		return true
	})
	if len(want) != 0 {
		t.Fatalf("wall strips missed coordinates %v", want)
	}
}
