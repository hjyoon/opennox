package legacy

/*
#include "GAME1_1.h"
#include "GAME1_2.h"

*/
import "C"

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const (
	mapWallMinX      = 741360
	mapWallMinY      = 741368
	mapWallMagicTile = 741372
)

type mapWallFrame struct {
	x, y          int32
	width, height int32
}

func writeMapWallFrame(region *[8]int32) mapWallFrame {
	minX, minY := memmap.PtrUint32(0x5D4594, mapWallMinX), memmap.PtrUint32(0x5D4594, mapWallMinY)
	maxX := (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_741356))
	maxY := (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_741364))
	if region != nil {
		bounds := mapWallRegionBounds(region)
		*minX, *minY = uint32(bounds[0]/wall.GridStep), uint32(bounds[1]/wall.GridStep)
		*maxX, *maxY = uint32(bounds[2]/wall.GridStep), uint32(bounds[3]/wall.GridStep)
	} else {
		*minX, *minY = server.WallGridSize, server.WallGridSize
		*maxX, *maxY = 0, 0
		GetServer().S().Walls.EachWallXxx(func(w *server.Wall) bool {
			wallMapBounds42A0F0(w, minX, maxX, minY, maxY)
			return true
		})
	}
	return mapWallFrame{int32(*minX), int32(*minY), int32(*maxX - *minX + 1), int32(*maxY - *minY + 1)}
}

func mapWallFrameIO(cf *cryptfile.CryptFile, region *[8]int32) (frame mapWallFrame, dx, dy int32, err error) {
	if !cf.ReadOnly() {
		frame = writeMapWallFrame(region)
	}
	for _, field := range [...]*int32{&frame.x, &frame.y, &frame.width, &frame.height} {
		raw, err := cf.ReadWriteU32(uint32(*field))
		if err != nil {
			return frame, 0, 0, err
		}
		*field = int32(raw)
	}
	if cf.ReadOnly() {
		if region != nil {
			bounds := mapWallRegionBounds(region)
			x, y := bounds[0]/wall.GridStep, bounds[1]/wall.GridStep
			dx, dy = x-frame.x, y-frame.y
			frame.x, frame.y = x, y
		}
		*memmap.PtrInt32(0x5D4594, mapWallMinX) = frame.x
		*memmap.PtrInt32(0x5D4594, mapWallMinY) = frame.y
	}
	return frame, dx, dy, nil
}

func mapWallForRead(x, y int32, direction byte, merge, temporary bool) *server.Wall {
	var w *server.Wall
	if temporary {
		node := sub_504290(C.int8_t(x), C.int8_t(y))
		if node != nil {
			w = asWallP(unsafe.Pointer(node.wall))
		}
	} else {
		pos := image.Pt(int(x), int(y))
		grid := &GetServer().S().Walls
		w = grid.GetWallAtGrid(pos)
		if w == nil {
			w = grid.CreateAtGrid(pos)
		} else if merge {
			direction = WallMergeDirections(w.Dir0, direction)
		}
	}
	if w != nil {
		w.Dir0 = direction
	}
	return w
}

func wallMapV2Variation(w *server.Wall) bool {
	w.Field2 = 0
	if w.Dir0 < 2 {
		w.Field2 = w.X5 % 3
	}
	variant := (int32(w.Flags4) >> 2) & 2
	if nox_xxx_getWallSprite_46A3B0(int32(w.Tile1), int32(w.Dir0), int32(w.Field2), variant) == nil {
		w.Field2 = 0
		return false
	}
	return true
}

func readMapWall(cf *cryptfile.CryptFile, x, y int32, packed byte, version int16, merge, temporary, sparse bool) (err error) {
	w := mapWallForRead(x, y, packed&0x7f, merge, temporary)
	if w == nil {
		return fmt.Errorf("create map wall at (%d, %d)", x, y)
	}
	if packed&0x80 != 0 {
		w.Flags4 |= wall.Flag8
	}
	zone := byte(1)
	omitLoadedZone := sparse && version == 6
	if version < 2 {
		w.Tile1 = 0
		w.Health7 = nox_xxx_mapWallGetHpByTile_410E20(0)
		w.Field8 = w.Field8&0xff00 | uint16(zone)
		return nil
	}
	if w.Tile1, err = cf.ReadU8(); err != nil {
		return err
	}
	if version >= 3 {
		if w.Field2, err = cf.ReadU8(); err != nil {
			return err
		}
	} else {
		wallMapV2Variation(w)
	}
	if merge && w.Field2 >= nox_xxx_mapWallMaxVariation_410DD0(int32(w.Tile1), int32(w.Dir0), 0) {
		w.Field2 = 0
	}
	w.Health7 = nox_xxx_mapWallGetHpByTile_410E20(int32(w.Tile1))
	if version >= 4 {
		value, err := cf.ReadU8()
		if err != nil {
			return err
		}
		if !omitLoadedZone {
			zone = value
		}
	}
	w.Field8 = w.Field8&0xff00 | uint16(zone)
	var extra byte
	if version >= 5 && !omitLoadedZone {
		if extra, err = cf.ReadU8(); err != nil {
			return err
		}
	}
	w.Field12 = uint32(extra)
	return nil
}

func mapWallZone(w *server.Wall, fallback byte) byte {
	point := C.int2{field_0: wall.GridStep*C.int(w.X5) + 11, field_4: wall.GridStep*C.int(w.Y6) + 11}
	poly := Nox_xxx_polygonIsPlayerInPolygon_4217B0(unsafe.Pointer(&point), 0)
	if poly == nil {
		poly = (*Nox_player_polygon_check_data)(unsafe.Pointer(sub_421990(&point, 10, 0)))
	}
	if poly == nil {
		return fallback
	}
	return byte(poly.Field_0[32] >> 16)
}

func writeMapWall(cf *cryptfile.CryptFile, w *server.Wall, sparse bool) error {
	extra, fallback := byte(w.Field12), byte(1)
	if sparse {
		if err := cf.WriteU8(w.X5); err != nil {
			return err
		}
		if err := cf.WriteU8(w.Y6); err != nil {
			return err
		}
		fallback = 100
		if noxflags.HasGame(noxflags.GameFlag22) {
			extra = 0
		}
	}
	// Each legacy byte is transferred separately because the map CRC follows
	// the original transfer boundaries. The direction is stored as a byte.
	for _, value := range [...]byte{w.Dir0, w.Tile1, w.Field2} {
		if err := cf.WriteU8(value); err != nil {
			return err
		}
	}
	if err := cf.WriteU8(mapWallZone(w, fallback)); err != nil {
		return err
	}
	return cf.WriteU8(extra)
}

func mapWallGridIO(cf *cryptfile.CryptFile, region *[8]int32, version int16, sparse bool) error {
	frame, dx, dy, err := mapWallFrameIO(cf, region)
	if err != nil {
		return err
	}
	if cf.ReadOnly() {
		merge := nox_xxx_wallGet_426A30()&1 != 0
		temporary := noxflags.HasGame(noxflags.GameFlag23)
		if sparse {
			for {
				x, err := cf.ReadU8()
				if err != nil || x == 0xff {
					return err
				}
				y, err := cf.ReadU8()
				if err != nil {
					return err
				}
				dir, err := cf.ReadU8()
				if err != nil {
					return err
				}
				x, y = byte(int32(x)+dx), byte(int32(y)+dy)
				if err := readMapWall(cf, int32(x), int32(y), dir, version, merge, temporary, sparse); err != nil {
					return err
				}
			}
		}
		if frame.width > server.WallGridSize || frame.height > server.WallGridSize {
			return fmt.Errorf("wall grid dimensions %d by %d exceed the map", frame.width, frame.height)
		}
		for y := int32(0); y <= frame.height; y++ {
			for x := int32(0); x <= frame.width; x++ {
				dir, err := cf.ReadU8()
				if err != nil {
					return err
				}
				if dir != 0xff {
					if err := readMapWall(cf, frame.x+x, frame.y+y, dir, version, merge, temporary, sparse); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	grid := &GetServer().S().Walls
	magic := memmap.Uint8(0x5D4594, mapWallMagicTile)
	for y := frame.y; y <= frame.y+frame.height; y++ {
		for x := frame.x; x <= frame.x+frame.width; x++ {
			w := grid.GetWallAtGrid(image.Pt(int(x), int(y)))
			if w != nil && region != nil && !mapWallRegionContains(wall.GridStep*x, wall.GridStep*y, region) {
				w = nil
			}
			if sparse {
				if w == nil || w.Tile1 == magic {
					continue
				}
			} else if w == nil || w.Dir0 == 0xff {
				if err := cf.WriteU8(0xff); err != nil {
					return err
				}
				continue
			}
			if err := writeMapWall(cf, w, sparse); err != nil {
				return err
			}
		}
	}
	if sparse {
		return cf.WriteU8(0xff)
	}
	return nil
}

func mapRWWallGrid(cf *cryptfile.CryptFile, region *[8]int32) error {
	if cf == nil {
		cf = cryptfile.Global()
		if cf == nil {
			return fmt.Errorf("wall map requires an open crypt file")
		}
	}
	magic := memmap.PtrUint8(0x5D4594, mapWallMagicTile)
	if *magic == 0 {
		*magic = byte(GetServer().S().Walls.DefIndByName("MagicWallSystemUseOnly"))
	}
	nox_xxx_wallSecretCounterClear_4297B0()
	nox_xxx_wallBreakableCounterClear_429520()
	version, err := cf.ReadWriteU16(7)
	if err != nil {
		return err
	}
	if int16(version) > 7 {
		return fmt.Errorf("unsupported wall map version %d", version)
	}
	return mapWallGridIO(cf, region, int16(version), int16(version) >= 6)
}

//export nox_server_mapRWWallMap_429B20
func nox_server_mapRWWallMap_429B20(region *C.uint32_t) C.int {
	return C.int(bool2int(mapRWWallGrid(nil, (*[8]int32)(unsafe.Pointer(region))) == nil))
}

//export sub_42A150
func sub_42A150(version C.short, region *C.uint32_t) C.int {
	cf := cryptfile.Global()
	if cf == nil {
		return 0
	}
	return C.int(bool2int(mapWallGridIO(cf, (*[8]int32)(unsafe.Pointer(region)), int16(version), false) == nil))
}

//export sub_42A650
func sub_42A650(raw *C.uchar) C.int {
	return C.int(bool2int(wallMapV2Variation(asWallP(unsafe.Pointer(raw)))))
}
