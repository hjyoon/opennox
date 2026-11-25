package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

const floorMapDenseVersion = 3

func readDenseFloorMap(cf *cryptfile.CryptFile, grid floorMapGrid, region *[8]int32) error {
	header, err := readFloorGridHeader(cf)
	if err != nil {
		return err
	}
	x0, y0 := header[0], header[1]
	if region == nil {
		grid.clearFlags()
	} else {
		bounds := mapWallRegionBounds(region)
		x0, y0 = bounds[0]/wall.GridStep, bounds[1]/wall.GridStep
	}
	for y := y0; y < y0+header[3]; y++ {
		for x := x0; x < x0+header[2]; x++ {
			tileFlags, err := cf.ReadU8()
			if err != nil {
				return err
			}
			if region != nil {
				if tileFlags != 0 {
					if err := readFloorRegionTile(cf, grid, x, y, true); err != nil {
						return err
					}
				}
				continue
			}
			tile := grid.cell(x, y)
			if tile == nil {
				return fmt.Errorf("floor tile (%d, %d) is outside the grid", x, y)
			}
			tile.field_0 = C.uint8_t(tileFlags)
			for _, layer := range [...]byte{1, 2} {
				if tileFlags&layer != 0 {
					if err := floorGridLayerIO(cf, tile, layer, true); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func writeDenseFloorMap(cf *cryptfile.CryptFile, grid floorMapGrid, region *[8]int32) error {
	x0, y0, x1, y1 := grid.writeBounds(region)
	if region == nil && (x1 < 0 || y1 < 0) {
		return nil
	}
	if err := writeFloorGridHeader(cf, x0, y0, x1, y1); err != nil {
		return err
	}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if region != nil {
				tile, layer := floorRegionTileForWrite(grid, region, x, y)
				if err := cf.WriteU8(byte(bool2int(tile != nil))); err != nil {
					return err
				}
				if tile != nil {
					if err := floorGridLayerIO(cf, tile, layer, true); err != nil {
						return err
					}
				}
				continue
			}
			tile := grid.cell(x, y)
			if err := cf.WriteU8(byte(tile.field_0)); err != nil {
				return err
			}
			for _, layer := range [...]byte{1, 2} {
				if byte(tile.field_0)&layer != 0 {
					if err := floorGridLayerIO(cf, tile, layer, true); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func floorMapV3(cf *cryptfile.CryptFile, grid floorMapGrid, region *[8]int32) error {
	if cf.ReadOnly() {
		return readDenseFloorMap(cf, grid, region)
	}
	return writeDenseFloorMap(cf, grid, region)
}

//export nox_xxx_tile_422C10
func nox_xxx_tile_422C10(version C.int, region unsafe.Pointer) C.int {
	cf := cryptfile.Global()
	if cf == nil || int16(version) < floorMapDenseVersion {
		return 0
	}
	return C.int(bool2int(floorMapV3(cf, liveFloorMapGrid(), (*[8]int32)(region)) == nil))
}
