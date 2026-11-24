package legacy

/*
#include "GAME4.h"
#include "GAME4_2.h"
#include "mapgen_decor_records.h"
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
)

type mapGenDecorAreaFlags struct {
	_     [60]byte
	flags uint32
}

type mapGenDecorObjectPlacement struct {
	allocate func() (*mapGenOccupiedRect5227B0, uint32)
	release  func(*mapGenOccupiedRect5227B0)
	width    func(int32) float64
	height   func(int32) float64
	record   func(int32) *mapGenDecorAreaFlags
	roll     func(int32, int32) int32
	inside   func(*mapGenOccupiedRect5227B0) bool
	overlaps func(*mapGenOccupiedRect5227B0) bool
	loadArea func(int32)
	foreach  func(uint32)
	moveArea func(*types.Pointf)
}

func mapGenDecorGridPosition(cells int32, roomEdge float32) float32 {
	delta := mulFloat64TowardZero(float64(cells), mapgenCellSize)
	return storeFloat32TowardZero(addFloat64TowardZero(delta, float64(roomEdge)))
}

func mapGenDecorOppositeEdge(edge, extent float32) float32 {
	// An x87 subtraction propagates the NaN operand with its original sign.
	if !math.IsNaN(float64(extent)) {
		extent = -extent
	}
	return storeFloat32TowardZero(addFloat64TowardZero(float64(edge), float64(extent)))
}

func mapGenPlaceDecorObject521CB0(room *mapgenRoomGeometry, head *uint32, foreach *C.uint32_t,
	index int32, rt mapGenDecorObjectPlacement) (placed bool) {
	rect, token := rt.allocate()
	if rect == nil {
		return false
	}
	defer func() {
		if !placed {
			rt.release(rect)
		}
	}()
	width := storeFloat32TowardZero(rt.width(index))
	height := storeFloat32TowardZero(rt.height(index))
	widthCells := x87TruncQwordLow(float64(width) * mapgenWorldToCellScale)
	heightCells := x87TruncQwordLow(float64(height) * mapgenWorldToCellScale)
	horizontal := int32(room.cellDimensions[0]) - widthCells
	vertical := int32(room.cellDimensions[1]) - heightCells
	if horizontal < 0 || vertical < 0 {
		return false
	}
	record := rt.record(index)
	if record == nil {
		return false
	}
	edge := func(index int) float32 { return math.Float32frombits(uint32(room.bounds[index])) }
	x := rt.roll(0, horizontal)
	point := types.Pointf{X: mapGenDecorGridPosition(x, edge(0))}
	y := rt.roll(0, vertical)
	point.Y = mapGenDecorGridPosition(y, edge(1))
	flags := record.flags
	if flags&1 != 0 {
		point.Y = edge(1)
	} else if flags&2 != 0 {
		point.Y = mapGenDecorOppositeEdge(edge(3), height)
	}
	if flags&4 != 0 {
		point.X = mapGenDecorOppositeEdge(edge(2), width)
	} else if flags&8 != 0 {
		point.X = edge(0)
	}
	if flags&16 != 0 {
		// Centering rereads the room dimensions after the two RNG calls.
		point.X = mapGenDecorGridPosition((int32(room.cellDimensions[0])-widthCells)/2, edge(0))
		point.Y = mapGenDecorGridPosition((int32(room.cellDimensions[1])-heightCells)/2, edge(1))
	}
	rect.transient = 1
	rect.bounds = [4]float32{
		point.X, point.Y,
		storeFloat32TowardZero(addFloat64TowardZero(float64(point.X), float64(width))),
		storeFloat32TowardZero(addFloat64TowardZero(float64(point.Y), float64(height))),
	}
	rect.objectIndex = index
	if !rt.inside(rect) || rt.overlaps(rect) {
		return false
	}
	rt.loadArea(index)
	rt.foreach(uint32(*foreach))
	rt.moveArea(&point)
	rect.next = *head
	*head = token
	return true
}

//export nox_mapgenPlaceDecorObjectNative_521CB0
func nox_mapgenPlaceDecorObjectNative_521CB0(theme, room *C.uint8_t, entry *C.nox_mapgen_decor_entry_legacy, index C.int) C.int {
	if theme == nil || room == nil || entry == nil {
		return 0
	}
	return C.int(bool2int(mapGenPlaceDecorObject521CB0(
		(*mapgenRoomGeometry)(unsafe.Pointer(room)), mapGenOccupiedRectHead(room), &entry.foreach_token, int32(index),
		mapGenDecorObjectPlacement{
			allocate: mapGenAllocateOccupiedRect,
			release:  mapGenReleaseOccupiedRect,
			width:    func(index int32) float64 { return float64(C.sub_502E70(C.int(index))) },
			height:   func(index int32) float64 { return float64(C.sub_502EA0(C.int(index))) },
			record: func(index int32) *mapGenDecorAreaFlags {
				return (*mapGenDecorAreaFlags)(unsafe.Pointer(C.sub_526AA0(C.int(index))))
			},
			roll: func(low, high int32) int32 {
				return int32(nox_xxx_mapGenRandFunc_526AC0(C.int(low), C.int(high)))
			},
			inside: func(rect *mapGenOccupiedRect5227B0) bool {
				return sub_521EB0(room, (*C.nox_mapgen_occupied_rect_legacy)(unsafe.Pointer(rect))) != 0
			},
			overlaps: func(rect *mapGenOccupiedRect5227B0) bool {
				return nox_mapgenOccupiedRectsIntersectNative_521F10(
					room, (*C.nox_mapgen_occupied_rect_legacy)(unsafe.Pointer(rect))) != nil
			},
			loadArea: func(index int32) { C.sub_502D70(C.int(index)) },
			foreach:  func(token uint32) { nox_mapgenApplyForeachNative_521C60(theme, C.uintptr_t(token)) },
			moveArea: func(point *types.Pointf) { C.sub_503B30((*C.float2)(unsafe.Pointer(point))) },
		},
	)))
}
