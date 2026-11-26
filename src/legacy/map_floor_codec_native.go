package legacy

/*
#include "GAME1_1.h"
#include "world_native_storage.h"

*/
import "C"

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

const floorMapVersion = 4

type floorMapGrid []*C.obj_5D4594_2650668_t

func liveFloorMapGrid() floorMapGrid {
	if worldOwnerStorage.tile_coordinates == nil {
		return nil
	}
	return unsafe.Slice(worldOwnerStorage.tile_coordinates, int(C.NOX_WORLD_TILE_CAPACITY))
}

func (grid floorMapGrid) cell(x, y int32) *C.obj_5D4594_2650668_t {
	size := int32(len(grid))
	if x < 0 || y < 0 || x >= size || y >= size || grid[x] == nil {
		return nil
	}
	return &unsafe.Slice(grid[x], size)[y]
}

func (grid floorMapGrid) bounds() (x0, y0, x1, y1 int32) {
	x0, y0, x1, y1 = int32(len(grid)), int32(len(grid)), -1, -1
	for x, row := range grid {
		for y, tile := range unsafe.Slice(row, len(grid)) {
			if tile.field_0 != 0 {
				x0, x1 = min(x0, int32(x)), max(x1, int32(x))
				y0, y1 = min(y0, int32(y)), max(y1, int32(y))
			}
		}
	}
	return
}

func (grid floorMapGrid) clearFlags() {
	for _, raw := range grid {
		column := unsafe.Slice(raw, len(grid))
		for y := range column {
			column[y].field_0 = 0
		}
	}
}

func (grid floorMapGrid) writeBounds(region *[8]int32) (x0, y0, x1, y1 int32) {
	if region == nil {
		return grid.bounds()
	}
	bounds := mapWallRegionBounds(region)
	return bounds[0] / wall.GridStep, bounds[1] / wall.GridStep, bounds[2] / wall.GridStep, bounds[3] / wall.GridStep
}

func floorGridLayerIO(cf *cryptfile.CryptFile, tile *C.obj_5D4594_2650668_t, layer byte, dense bool) error {
	if cf.ReadOnly() {
		tile.field_0 |= C.uint8_t(layer)
	}
	if layer == 1 {
		return floorTileLayerIO(cf, (*[4]uint32)(unsafe.Pointer(&tile.field_1)), &tile.field_5, dense)
	}
	return floorTileLayerIO(cf, (*[4]uint32)(unsafe.Pointer(&tile.field_6)), &tile.field_10, dense)
}

func floorRegionLayer(x, y int32) (column, row int32, layer byte) {
	if y&1 != 0 {
		return (wall.GridStep*x - wall.GridStep) / (2 * wall.GridStep),
			(wall.GridStep*y + wall.GridStep) / (2 * wall.GridStep), 1
	}
	return wall.GridStep * x / (2 * wall.GridStep), wall.GridStep * y / (2 * wall.GridStep), 2
}

func readFloorRegionTile(cf *cryptfile.CryptFile, grid floorMapGrid, x, y int32, dense bool) error {
	column, row, layer := floorRegionLayer(x, y)
	if noxflags.HasGame(noxflags.GameFlag23) {
		entry := nox_xxx_tileAllocTileInCoordList_5040A0(C.int(column), C.int(row), C.float(math.Float32frombits(uint32(layer))))
		if entry == nil || entry.layer == nil {
			return fmt.Errorf("allocate floor layer %d at (%d, %d)", layer, column, row)
		}
		return floorTileLayerIO(cf, (*[4]uint32)(unsafe.Pointer(&entry.layer.field_0)), &entry.layer.subtiles, dense)
	}
	tile := grid.cell(column, row)
	if tile == nil {
		return fmt.Errorf("floor region tile (%d, %d) is outside the grid", column, row)
	}
	return floorGridLayerIO(cf, tile, layer, dense)
}

func readFloorGridHeader(cf *cryptfile.CryptFile) ([4]int32, error) {
	var header [4]int32
	for i := range header {
		raw, err := cf.ReadU32()
		if err != nil {
			return header, err
		}
		header[i] = int32(raw)
	}
	return header, nil
}

func writeFloorGridHeader(cf *cryptfile.CryptFile, x0, y0, x1, y1 int32) error {
	for _, value := range [...]int32{x0, y0, x1 - x0 + 1, y1 - y0 + 1} {
		if err := cf.WriteU32(uint32(value)); err != nil {
			return err
		}
	}
	return nil
}

func readFloorMapTiles(cf *cryptfile.CryptFile, grid floorMapGrid, region *[8]int32) error {
	header, err := readFloorGridHeader(cf)
	if err != nil {
		return err
	}
	var dx, dy int32
	if region != nil {
		bounds := mapWallRegionBounds(region)
		dx, dy = bounds[0]/wall.GridStep-header[0], bounds[1]/wall.GridStep-header[1]
	} else {
		grid.clearFlags()
	}

	for {
		position, err := cf.ReadU16()
		if err != nil || position == 0xffff {
			return err
		}
		if region != nil {
			x, y := dx+int32(position>>8), dy+int32(byte(position))
			if err := readFloorRegionTile(cf, grid, x, y, false); err != nil {
				return err
			}
			continue
		}
		upper, lower := position&0x8000 != 0, position&0x80 != 0
		if !upper && !lower {
			continue
		}
		x, y := int32(position>>8&0x7f), int32(position&0x7f)
		tile := grid.cell(x, y)
		if tile == nil {
			return fmt.Errorf("floor tile (%d, %d) is outside the grid", x, y)
		}
		if upper {
			if err := floorGridLayerIO(cf, tile, 1, false); err != nil {
				return err
			}
		}
		if lower {
			if err := floorGridLayerIO(cf, tile, 2, false); err != nil {
				return err
			}
		}
	}
}

func floorRegionTileForWrite(grid floorMapGrid, region *[8]int32, x, y int32) (*C.obj_5D4594_2650668_t, byte) {
	if (x+y)&1 != 0 {
		return nil, 0
	}
	column, row, layer := floorRegionLayer(x, y)
	tile := grid.cell(column, row)
	if tile == nil || byte(tile.field_0)&layer == 0 ||
		!mapWallRegionContains(wall.GridStep*x+11, wall.GridStep*y+34, region) {
		return nil, 0
	}
	return tile, layer
}

func writeFloorRegionTile(cf *cryptfile.CryptFile, grid floorMapGrid, region *[8]int32, x, y int32) error {
	tile, layer := floorRegionTileForWrite(grid, region, x, y)
	if tile == nil {
		return nil
	}
	if err := cf.WriteU16(uint16(byte(x))<<8 | uint16(y)); err != nil {
		return err
	}
	return floorGridLayerIO(cf, tile, layer, false)
}

func writeFloorWholeTile(cf *cryptfile.CryptFile, grid floorMapGrid, x, y int32) error {
	tile := grid.cell(x, y)
	if tile == nil || tile.field_0&3 == 0 {
		return nil
	}
	flags := tile.field_0
	encoded := uint16(x)<<8 | uint16(byte(y))
	if flags&1 != 0 {
		encoded |= 0x8000
	}
	if flags&2 != 0 {
		encoded |= 0x80
	}
	if err := cf.WriteU16(encoded); err != nil {
		return err
	}
	if flags&1 != 0 {
		if err := floorGridLayerIO(cf, tile, 1, false); err != nil {
			return err
		}
	}
	if flags&2 != 0 {
		return floorGridLayerIO(cf, tile, 2, false)
	}
	return nil
}

func writeFloorMapTiles(cf *cryptfile.CryptFile, grid floorMapGrid, region *[8]int32) error {
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
				if err := writeFloorRegionTile(cf, grid, region, x, y); err != nil {
					return err
				}
			} else if err := writeFloorWholeTile(cf, grid, x, y); err != nil {
				return err
			}
		}
	}
	return cf.WriteU16(0xffff)
}

func mapRWFloorGrid(cf *cryptfile.CryptFile, region *[8]int32) error {
	if cf == nil {
		cf = cryptfile.Global()
		if cf == nil {
			return fmt.Errorf("floor map requires an open crypt file")
		}
	}
	version, err := cf.ReadWriteU16(floorMapVersion)
	if err != nil {
		return err
	}
	if version < floorMapDenseVersion || version > floorMapVersion {
		return fmt.Errorf("unsupported floor map version %d", version)
	}
	grid := liveFloorMapGrid()
	if version == floorMapDenseVersion {
		return floorMapV3(cf, grid, region)
	}
	if cf.ReadOnly() {
		return readFloorMapTiles(cf, grid, region)
	}
	return writeFloorMapTiles(cf, grid, region)
}

//export nox_server_mapRWFloorMap_422230
func nox_server_mapRWFloorMap_422230(region unsafe.Pointer) C.int {
	return C.int(bool2int(mapRWFloorGrid(nil, (*[8]int32)(region)) == nil))
}
