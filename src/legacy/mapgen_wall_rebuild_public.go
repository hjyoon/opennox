package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const mapgenPreservedWallFlags = 0x8c

type mapgenWallRebuildAccess struct {
	floorLayers func(x, y int32) (first, second int32)
	direction   func(mask int) uint32
	style       func() uint32
	lookup      func(image.Point) *server.Wall
	create      func(image.Point) *server.Wall
	remove      func(image.Point)
	release     func(unsafe.Pointer)
	variations  func(tile byte) byte
}

func mapgenWallRebuildOrigin(grid int32) int32 {
	grid--
	if grid <= 0 {
		return 0
	}
	return grid & 0xfffe
}

func mapgenRebuildWallsNearTile(x, y int32, access mapgenWallRebuildAccess) int32 {
	startX, startY := mapgenWallRebuildOrigin(x), mapgenWallRebuildOrigin(y)
	for gy := startY; gy <= startY+4; gy++ {
		for gx := startX; gx <= startX+4; gx++ {
			wall := access.lookup(image.Pt(int(gx), int(gy)))
			if wall == nil || wall.Flags4&mapgenPreservedWallFlags != 0 {
				continue
			}
			if wall.Data != nil {
				access.release(wall.Data)
				wall.SetData(nil)
			}
			access.remove(wall.GridPos())
		}
	}
	for gy := startY; gy <= startY+4; gy += 2 {
		for gx := startX; gx <= startX+4; gx += 2 {
			x, y := gx/2, gy/2
			if x >= 128 || y >= 128 {
				continue
			}
			mask := mapgenEvenWallMask(x, y, access.floorLayers)
			access.rebuildWall(image.Pt(int(gx), int(gy)), access.direction(mask))
			mask = mapgenOddWallMask(x, y, access.floorLayers)
			access.rebuildWall(image.Pt(int(gx)-1, int(gy)-1), access.direction(mask))
		}
	}
	return startX + 6
}

func mapgenEvenWallMask(x, y int32, layers func(int32, int32) (int32, int32)) int {
	first, second := layers(x, y)
	_, upper := layers(x, y-1)
	left, _ := layers(x-1, y)
	mask := 0
	if left != 255 {
		mask |= 1
	}
	if first != 255 {
		mask |= 2
	}
	if upper != 255 {
		mask |= 4
	}
	if second != 255 {
		mask |= 8
	}
	return mask
}

func mapgenOddWallMask(x, y int32, layers func(int32, int32) (int32, int32)) int {
	_, upper := layers(x, y-1)
	left, _ := layers(x-1, y)
	diagonalFirst, diagonalSecond := layers(x-1, y-1)
	mask := 0
	if diagonalSecond != 255 {
		mask |= 1
	}
	if upper != 255 {
		mask |= 2
	}
	if diagonalFirst != 255 {
		mask |= 4
	}
	if left != 255 {
		mask |= 8
	}
	return mask
}

func (access mapgenWallRebuildAccess) rebuildWall(grid image.Point, direction uint32) {
	if direction == 255 {
		return
	}
	wall := access.lookup(grid)
	if wall == nil {
		wall = access.create(grid)
		if wall == nil {
			return
		}
		wall.Tile1 = byte(access.style())
	} else if wall.Flags4&mapgenPreservedWallFlags != 0 ||
		uint32(wall.Dir0) == direction || uint32(wall.Tile1) == access.style() {
		return
	}
	wall.Field2 = 0
	wall.Dir0 = byte(direction)
	if wall.Dir0 <= 1 {
		if count := access.variations(wall.Tile1); count != 0 {
			wall.Field2 = wall.X5 % count
		}
	}
}

func mapgenUpdateWallsAroundPoint(x, y float32, rebuild func(x, y int32)) {
	sx := x87TruncQwordLow(mulFloat64TowardZero(float64(x), mapgenWallCoordinateScale))
	sy := x87TruncQwordLow(mulFloat64TowardZero(float64(y), mapgenWallCoordinateScale))
	for _, point := range [3][2]int32{{sx, sy}, {sx - 1, sy}, {sx, sy - 1}} {
		if px, py := point[0], point[1]; px > 0 && px < 255 && py > 0 && py < 255 {
			rebuild(px, py)
		}
	}
}

func mapgenFloorWallLayers(x, y int32) (int32, int32) {
	if x < 0 || y < 0 {
		return 255, 255
	}
	head := *(*unsafe.Pointer)(worldTileHeadCell411160())
	if head == nil {
		return 255, 255
	}
	row := worldTileRow411160(head, x)
	if row == nil {
		return 255, 255
	}
	cells := unsafe.Slice((*worldTileCell411160)(row), worldTileCapacity411160())
	return cells[y].Value1, cells[y].Value6
}

func mapgenWallRebuildProductionAccess() mapgenWallRebuildAccess {
	walls := &GetServer().S().Walls
	return mapgenWallRebuildAccess{
		floorLayers: mapgenFloorWallLayers,
		direction:   func(mask int) uint32 { return *memmap.PtrUint32(0x587000, 255052+4*uintptr(mask)) },
		style:       func() uint32 { return *memmap.PtrUint32(0x973F18, 35948) },
		lookup:      walls.GetWallAtGrid,
		create:      walls.CreateAtGrid,
		remove:      walls.DeleteAtGrid,
		release:     func(data unsafe.Pointer) { sub_4107A0(data) },
		variations: func(tile byte) byte {
			if def := walls.DefByInd(int(tile)); def != nil {
				return def.Field749
			}
			return 0
		},
	}
}

//export sub_527380
func sub_527380(point *C.float2) C.int {
	if point == nil {
		return 0
	}
	access := mapgenWallRebuildProductionAccess()
	mapgenUpdateWallsAroundPoint(float32(point.field_0), float32(point.field_4), func(x, y int32) {
		mapgenRebuildWallsNearTile(x, y, access)
	})
	return 1
}

//export sub_527450
func sub_527450(point *C.uint32_t) C.int {
	if point == nil {
		return 0
	}
	grid := unsafe.Slice(point, 2)
	return C.int(mapgenRebuildWallsNearTile(int32(grid[0]), int32(grid[1]), mapgenWallRebuildProductionAccess()))
}
