package legacy

import (
	"encoding/binary"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func TestMapWallGridNativeLegacyAndSparseVersions(t *testing.T) {
	srv, _ := newMapgenWallTestState(t)
	srv.Walls.DefByInd(0).Health41 = 23

	flags := noxflags.GetGame()
	noxflags.UnsetGame(noxflags.GameFlag23)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(flags)
	})
	for _, offset := range []uintptr{mapWallMinX, mapWallMinY, mapWallFlags, mapBreakableWallID, mapSecretWallID} {
		field := memmap.PtrUint32(0x5D4594, offset)
		saved := *field
		t.Cleanup(func() { *field = saved })
	}
	merge := memmap.PtrUint32(0x5D4594, mapWallFlags)
	*merge = 0
	magic := memmap.PtrUint8(0x5D4594, mapWallMagicTile)
	savedMagic := *magic
	t.Cleanup(func() { *magic = savedMagic })
	*magic = 0

	region := [8]int32{12 * 23, 8 * 23, 8 * 23, 12 * 23, 16 * 23, 12 * 23, 12 * 23, 16 * 23}
	for _, tc := range []struct {
		name    string
		version uint16
		data    []byte
		offset  *[8]int32
		variant byte
		zone    uint16
		extra   uint32
	}{
		{"dense version 1", 1, []byte{0x81}, nil, 0, 1, 0},
		{"dense version 2", 2, []byte{0x81, 0}, nil, 0, 1, 0},
		{"dense version 5", 5, []byte{0x81, 0, 3, 42, 5}, nil, 3, 42, 5},
		{"sparse version 6 with region", 6, []byte{4, 4, 0x81, 0, 3, 42, 0xff}, &region, 3, 1, 0},
		{"sparse version 7", 7, []byte{12, 12, 0x81, 0, 3, 77, 66, 0xff}, nil, 3, 77, 66},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv.Walls.Reset()
			origin := uint32(12)
			if tc.offset != nil {
				origin = 0
			}
			wire := binary.LittleEndian.AppendUint16(nil, tc.version)
			for _, word := range [...]uint32{origin, origin, 0, 0} {
				wire = binary.LittleEndian.AppendUint32(wire, word)
			}
			wire = append(wire, tc.data...)
			file := filepath.Join(t.TempDir(), "walls.mapsection")
			if err := os.WriteFile(file, wire, 0o600); err != nil {
				t.Fatal(err)
			}
			cf, err := cryptfile.OpenFile(file, cryptfile.ReadOnly, -1)
			if err != nil {
				t.Fatal(err)
			}
			defer cf.Close()
			if err := mapRWWallGrid(cf, tc.offset); err != nil {
				t.Fatal(err)
			}
			if w := srv.Walls.GetWallAtGrid(image.Pt(12, 12)); w == nil || w.Dir0 != 1 ||
				w.Tile1 != 0 || w.Field2 != tc.variant || !w.Flags4.Has(wall.Flag8) ||
				w.Health7 != 23 || w.Field8 != tc.zone || w.Field12 != tc.extra {
				t.Fatalf("wall record fields changed: %+v", w)
			}
		})
	}
}
