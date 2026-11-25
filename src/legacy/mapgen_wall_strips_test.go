package legacy

import (
	"encoding/binary"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

func newMapgenWallTestState(t *testing.T) (*server.Server, *uint32) {
	t.Helper()
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
	t.Cleanup(cfixture.DisableMapgenWallFixtureModes())
	return srv, direction
}

func TestMapgenWallStripsNative(t *testing.T) {
	srv, direction := newMapgenWallTestState(t)
	if got := cfixture.MapgenWallStripFixture(0, 0, -3, false); got != -3 || *direction != 0 {
		t.Fatalf("empty horizontal strip returned %d with direction %d", got, *direction)
	}
	if got := cfixture.MapgenWallStripFixture(0, 0, 0, true); got != 1 || *direction != 1 {
		t.Fatalf("empty vertical strip returned %d with direction %d", got, *direction)
	}

	const cell = 32.526913
	if got := cfixture.MapgenWallStripFixture(cell, 0, 4, false); got != 1 {
		t.Fatalf("horizontal strip returned %d", got)
	}
	if got := cfixture.MapgenWallStripFixture(0, cell, 4, true); got != 1 {
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

func TestMapgenWallJunctionsNative(t *testing.T) {
	srv, direction := newMapgenWallTestState(t)
	neighbors := [...]image.Point{image.Pt(127, 127), image.Pt(129, 129), image.Pt(127, 129), image.Pt(129, 127)}
	center := image.Pt(128, 128)
	junctionTypes := [...]int32{-1, -1, -1, 1, -1, 10, 9, 6, -1, 7, 8, 4, 0, 3, 5, 2}
	for mask, kind := range junctionTypes {
		srv.Walls.Reset()
		*direction = 13
		for bit, pos := range neighbors {
			if mask&(1<<bit) != 0 && srv.Walls.CreateAtGrid(pos) == nil {
				t.Fatalf("could not create neighbor at %v", pos)
			}
		}
		result := cfixture.MapgenWallJunctionFixture(0, 0)
		wall := srv.Walls.GetWallAtGridRaw(center)
		if kind < 0 {
			if result != int32(mask)-3 || *direction != 13 || wall != nil {
				t.Fatalf("mask %#x: result %d, direction %d, wall %v", mask, result, *direction, wall)
			}
		} else if result != 1 || *direction != uint32(kind) || wall == nil || wall.Dir0 != byte(kind) {
			t.Fatalf("mask %#x: result %d, direction %d, wall %v, want junction type %d", mask, result, *direction, wall, kind)
		}
	}
}

func TestMapgenWallPlacementNative(t *testing.T) {
	srv, direction := newMapgenWallTestState(t)
	variant := memmap.PtrUint32(0x973F18, 35956)
	oldVariant := *variant
	t.Cleanup(func() { *variant = oldVariant })
	*variant = 3
	*direction = 0

	definition := srv.Walls.DefByInd(0)
	definition.Field749 = 5
	definition.Variations12272[0][0] = 5
	definition.Variations12272[0][2] = 5
	first := image.Pt(128, 128)
	if got := cfixture.MapgenWallPlacementFixture(0, 0); got != 1 {
		t.Fatalf("new wall returned %d", got)
	}
	wall := srv.Walls.GetWallAtGridRaw(first)
	if wall == nil || wall.Dir0 != 0 || wall.Field2 != 3 || wall.Flags4 != 0x80 {
		t.Fatalf("new wall has unexpected direction, variation or flags at %v", first)
	}
	if got := cfixture.MapgenWallPlacementFixture(16.263456, 0); got != 0 {
		t.Fatalf("odd-parity cell returned %d", got)
	}

	*direction = 1
	sub_526C80(1)
	cfixture.EnableMapgenWallFixtureRepeat()
	if got := cfixture.MapgenWallPlacementFixture(0, 0); got != 1 || wall.Dir0 != 2 || wall.Field2 != 3 {
		t.Fatalf("merged wall returned %d with direction %d and variation %d", got, wall.Dir0, wall.Field2)
	}

	definition.Field749 = 0
	definition.Variations12272[0][1] = 5
	second := image.Pt(129, 129)
	if got := cfixture.MapgenWallPlacementFixture(0, 32.526913); got != 1 {
		t.Fatalf("wall with empty repeat period returned %d", got)
	}
	if placed := srv.Walls.GetWallAtGridRaw(second); placed == nil || placed.Dir0 != 1 || placed.Field2 != 0 {
		t.Fatalf("wall with empty repeat period has unexpected direction or variation at %v", second)
	}

	*memmap.PtrUint32(0x973F18, 35948) = 255
	if got := sub_526E60(nil); got != 1 {
		t.Fatalf("disabled wall selection returned %d", got)
	}
}
