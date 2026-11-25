package legacy

import (
	"bytes"
	"encoding/binary"
	"image"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func TestMapTaggedWallsNativeLegacyRegionAndWireOrder(t *testing.T) {
	InitBlobData()
	srv := newMapgenTestWallGrid(t)
	t.Cleanup(srv.Walls.ClearBreakable)
	savedFlags := noxflags.GetGame()
	noxflags.UnsetGame(noxflags.GameFlag23)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(savedFlags)
	})

	for _, offset := range []uintptr{mapWallOffsetX, mapWallOffsetY, mapWindowWalls.count, mapBreakableWalls.count, mapBreakableWallID, mapSecretWalls.count, mapSecretWallID} {
		field := memmap.PtrUint32(0x5D4594, offset)
		old := *field
		t.Cleanup(func() { *field = old })
	}
	SetMapWallOffset(5, 8)
	nox_xxx_wallBreakableCounterClear_429520()
	nox_xxx_wallSecretCounterClear_4297B0()
	secretHead := (*unsafe.Pointer)(secretWallNativeHeadCell410550())
	savedHead := *secretHead
	t.Cleanup(func() { *secretHead = savedHead })

	pos := image.Pt(12, 12)
	center := srv.Walls.CreateAtGrid(pos)
	outside := srv.Walls.CreateAtGrid(image.Pt(32, 32))
	if center == nil || outside == nil {
		t.Fatal("create map wall fixtures")
	}
	// The left corner is three cells right of the source map origin.
	region := [8]int32{12 * 23, 8 * 23, 8 * 23, 12 * 23, 16 * 23, 12 * 23, 12 * 23, 16 * 23}
	const openWait = 0x10203040
	encode := func(version uint16, x, y int32) []byte {
		b := binary.LittleEndian.AppendUint16(nil, version)
		b = binary.LittleEndian.AppendUint16(b, 1)
		b = binary.LittleEndian.AppendUint32(b, uint32(x))
		return binary.LittleEndian.AppendUint32(b, uint32(y))
	}
	read := func(section mapTaggedWallSection, b []byte) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "input.mapsection")
		if err := os.WriteFile(path, b, 0o600); err != nil {
			t.Fatal(err)
		}
		cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
		if err != nil {
			t.Fatal(err)
		}
		defer cf.Close()
		if err := mapRWTaggedWalls(cf, unsafe.Pointer(&region), section); err != nil {
			t.Fatal(err)
		}
	}
	checkWritten := func(section mapTaggedWallSection, want []byte) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "output.mapsection")
		cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
		if err != nil {
			t.Fatal(err)
		}
		if err := mapRWTaggedWalls(cf, unsafe.Pointer(&region), section); err != nil {
			_ = cf.Close()
			t.Fatal(err)
		}
		if err := cf.Close(); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s bytes = %x, want %x", section.flag, got, want)
		}
	}

	center.Field2 = 7
	outside.Flags4 |= wall.FlagWindow | wall.FlagBreakable | wall.FlagSecret
	read(mapWindowWalls, encode(1, 9, 12))
	if !center.Flags4.Has(wall.FlagWindow) || center.Field2 != 0 {
		t.Fatal("version 1 window did not clear the mapped wall variation")
	}
	checkWritten(mapWindowWalls, encode(2, 12, 12))

	read(mapBreakableWalls, encode(1, 9, 12))
	breakable := srv.Walls.FirstBreakable()
	if !center.Flags4.Has(wall.FlagBreakable) || center.Field10 != 0 ||
		breakable == nil || breakable.Wall != center || memmap.Uint32(0x5D4594, mapBreakableWallID) != 1 {
		t.Fatal("breakable counter or server list did not track the mapped wall")
	}
	checkWritten(mapBreakableWalls, encode(1, 12, 12))

	secretV1 := binary.LittleEndian.AppendUint32(encode(1, 9, 12), openWait)
	read(mapSecretWalls, append(secretV1, 8))
	secret := center.Secret()
	if secret == nil {
		t.Fatal("secret wall was not attached")
	}
	t.Cleanup(func() {
		sub_4107A0(unsafe.Pointer(secret))
		center.SetData(nil)
	})
	want := server.SecretWall{Next: (*server.SecretWall)(savedHead), X: 12, Y: 12, Wall: center, OpenWait: openWait,
		Flags: 8, State: 3, OpenDelay: 23, PlayerBits: ^uint32(0)}
	if *secret != want || *secretHead != unsafe.Pointer(secret) || center.Field10 != 0 ||
		memmap.Uint32(0x5D4594, mapSecretWallID) != 1 {
		t.Fatalf("secret wall = %+v, want %+v", *secret, want)
	}
	secretV2 := binary.LittleEndian.AppendUint32(encode(2, 12, 12), openWait)
	secretV2 = append(secretV2, 8, 3, 23)
	secretV2 = binary.LittleEndian.AppendUint32(secretV2, 0)
	secretV2 = binary.LittleEndian.AppendUint32(secretV2, ^uint32(0))
	checkWritten(mapSecretWalls, secretV2)
}
