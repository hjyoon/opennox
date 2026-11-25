package legacy

/*
#include <stdlib.h>
#include "GAME1_2.h"
#include "map_temp_lists_503f40.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"image"
	"unsafe"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type mapTaggedWallSection struct {
	flag    wall.Flags
	version uint16
	count   uintptr
}

var (
	mapWindowWalls    = mapTaggedWallSection{wall.FlagWindow, 2, 741336}
	mapBreakableWalls = mapTaggedWallSection{wall.FlagBreakable, 1, 741340}
	mapSecretWalls    = mapTaggedWallSection{wall.FlagSecret, 2, 741348}
)

const (
	mapBreakableWallID = 741344
	mapSecretWallID    = 741352
)

func mapTaggedWallOffset(region unsafe.Pointer) (int32, int32) {
	if region == nil {
		return 0, 0
	}
	bounds := mapWallRegionBounds((*[8]int32)(region))
	return bounds[0]/wall.GridStep - memmap.Int32(0x5D4594, mapWallOffsetX),
		bounds[1]/wall.GridStep - memmap.Int32(0x5D4594, mapWallOffsetY)
}

func mapTaggedWallAt(pos image.Point, temporary bool) *server.Wall {
	if temporary {
		node := nox_xxx_cliWallGet_5042F0(C.int32_t(pos.X), C.int32_t(pos.Y))
		if node != nil {
			return asWallP(unsafe.Pointer(node.wall))
		}
		return nil
	}
	return GetServer().S().Walls.GetWallAtGrid(pos)
}

func mapTaggedWallPosition(cf *cryptfile.CryptFile, record *server.SecretWall) error {
	var pos [8]byte
	binary.LittleEndian.PutUint32(pos[0:4], uint32(record.X))
	binary.LittleEndian.PutUint32(pos[4:8], uint32(record.Y))
	if err := mapSectionTransfer(cf, pos[:]); err != nil {
		return err
	}
	record.X = int32(binary.LittleEndian.Uint32(pos[0:4]))
	record.Y = int32(binary.LittleEndian.Uint32(pos[4:8]))
	return nil
}

func mapTaggedWallSecretData(cf *cryptfile.CryptFile, version int16, record *server.SecretWall) (err error) {
	if record.OpenWait, err = cf.ReadWriteU32(record.OpenWait); err != nil {
		return err
	}
	if record.Flags, err = cf.ReadWriteU8(record.Flags); err != nil || version < 2 {
		return err
	}
	if record.State, err = cf.ReadWriteU8(record.State); err != nil {
		return err
	}
	if record.OpenDelay, err = cf.ReadWriteU8(record.OpenDelay); err != nil {
		return err
	}
	if record.LastOpen, err = cf.ReadWriteU32(record.LastOpen); err != nil {
		return err
	}
	record.PlayerBits, err = cf.ReadWriteU32(record.PlayerBits)
	return err
}

func applyMapTaggedWall(w *server.Wall, section mapTaggedWallSection, version int16, record server.SecretWall, temporary bool) error {
	switch section.flag {
	case wall.FlagWindow:
		w.Flags4 |= wall.FlagWindow
		if version < 2 {
			w.Field2 = 0
		}
	case wall.FlagBreakable:
		w.Flags4 |= wall.FlagBreakable
		next := memmap.PtrUint32(0x5D4594, mapBreakableWallID)
		w.Field10 = uint16(*next)
		*next++
		if !temporary {
			nox_xxx_wallBreackableListAdd_410840(w.C())
		}
	case wall.FlagSecret:
		if record.State == 0 {
			if record.Flags&8 != 0 {
				record.State, record.OpenDelay, record.PlayerBits = 3, 23, ^uint32(0)
			} else {
				record.State, record.OpenDelay, record.PlayerBits = 1, 0, 0
			}
		}
		p := C.calloc(1, C.size_t(unsafe.Sizeof(record)))
		if p == nil {
			return fmt.Errorf("allocate secret wall at %v", w.GridPos())
		}
		secret := (*server.SecretWall)(p)
		record.Wall = w
		*secret = record
		next := memmap.PtrUint32(0x5D4594, mapSecretWallID)
		w.AttachSecret(secret, uint16(*next))
		*next++
		if !temporary {
			secretWallBlockNative410760(secret)
		}
	}
	return nil
}

func readMapTaggedWalls(cf *cryptfile.CryptFile, region unsafe.Pointer, section mapTaggedWallSection, version int16) error {
	count, err := cf.ReadU16()
	if err != nil {
		return err
	}
	*memmap.PtrUint16(0x5D4594, section.count) = count
	if count == 0 {
		return nil
	}
	// The original processes the first record for any nonzero count and
	// compares the remaining records against the signed 16-bit count.
	records := make([]server.SecretWall, max(1, int(int16(count))))
	dx, dy := mapTaggedWallOffset(region)
	for i := range records {
		record := &records[i]
		if err := mapTaggedWallPosition(cf, record); err != nil {
			return err
		}
		if section.flag == wall.FlagSecret {
			if err := mapTaggedWallSecretData(cf, version, record); err != nil {
				return err
			}
		}
		record.X, record.Y = record.X+dx, record.Y+dy
	}
	temporary := noxflags.HasGame(noxflags.GameFlag23)
	for _, record := range records {
		if w := mapTaggedWallAt(image.Pt(int(record.X), int(record.Y)), temporary); w != nil {
			if err := applyMapTaggedWall(w, section, version, record, temporary); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeMapTaggedWalls(cf *cryptfile.CryptFile, region unsafe.Pointer, section mapTaggedWallSection) error {
	var walls []*server.Wall
	GetServer().S().Walls.EachWallXxx(func(w *server.Wall) bool {
		if !w.Flags4.Has(section.flag) {
			return true
		}
		if region != nil && !mapWallRegionContains(wall.GridStep*int32(w.X5), wall.GridStep*int32(w.Y6), (*[8]int32)(region)) {
			return true
		}
		walls = append(walls, w)
		return true
	})
	for _, w := range walls {
		if section.flag == wall.FlagSecret && w.Secret() == nil {
			return fmt.Errorf("missing secret wall data at %v", w.GridPos())
		}
	}
	count := uint16(len(walls))
	*memmap.PtrUint16(0x5D4594, section.count) = count
	if err := cf.WriteU16(count); err != nil {
		return err
	}
	for _, w := range walls {
		var record server.SecretWall
		if section.flag == wall.FlagSecret {
			record = *w.Secret()
		}
		record.X, record.Y = int32(w.X5), int32(w.Y6)
		if err := mapTaggedWallPosition(cf, &record); err != nil {
			return err
		}
		if section.flag == wall.FlagSecret {
			if err := mapTaggedWallSecretData(cf, int16(section.version), &record); err != nil {
				return err
			}
		}
	}
	return nil
}

func mapRWTaggedWalls(cf *cryptfile.CryptFile, region unsafe.Pointer, section mapTaggedWallSection) error {
	if cf == nil {
		cf = cryptfile.Global()
		if cf == nil {
			return fmt.Errorf("wall map section requires an open crypt file")
		}
	}
	version, err := cf.ReadWriteU16(section.version)
	if err != nil {
		return err
	}
	if int16(version) > int16(section.version) {
		return fmt.Errorf("unsupported %s map section version %d", section.flag, version)
	}
	if cf.ReadOnly() {
		return readMapTaggedWalls(cf, region, section, int16(version))
	}
	return writeMapTaggedWalls(cf, region, section)
}

//export nox_server_mapRWWindowWalls_4292C0
func nox_server_mapRWWindowWalls_4292C0(region *C.uint32_t) C.int {
	return C.int(bool2int(mapRWTaggedWalls(nil, unsafe.Pointer(region), mapWindowWalls) == nil))
}

//export nox_server_mapRWDestructableWalls_429530
func nox_server_mapRWDestructableWalls_429530(region *C.uint32_t) C.int {
	return C.int(bool2int(mapRWTaggedWalls(nil, unsafe.Pointer(region), mapBreakableWalls) == nil))
}

//export nox_server_mapRWSecretWalls_4297C0
func nox_server_mapRWSecretWalls_4297C0(region *C.uint32_t) C.int {
	return C.int(bool2int(mapRWTaggedWalls(nil, unsafe.Pointer(region), mapSecretWalls) == nil))
}

//export nox_xxx_wallBreakableCounterClear_429520
func nox_xxx_wallBreakableCounterClear_429520() {
	*memmap.PtrUint32(0x5D4594, mapBreakableWallID) = 0
}

//export nox_xxx_wallSecretCounterClear_4297B0
func nox_xxx_wallSecretCounterClear_4297B0() {
	*memmap.PtrUint32(0x5D4594, mapSecretWallID) = 0
}
