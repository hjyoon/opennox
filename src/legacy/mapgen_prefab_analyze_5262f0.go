package legacy

/*
#include <stdint.h>

#include "GAME4.h"
#include "GAME4_2.h"
#include "map_object_list_5048a0.h"

extern uint32_t dword_5d4594_2487656;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	mapgenPrefabNorth5262F0 = iota
	mapgenPrefabSouth5262F0
	mapgenPrefabEast5262F0
	mapgenPrefabWest5262F0
	mapgenPrefabDirections5262F0
)

type mapgenPrefabFloatPoint5262F0 struct {
	x float32
	y float32
}

type mapgenPrefabIntPoint5262F0 struct {
	x int32
	y int32
}

type mapgenPrefabExit5262F0 struct {
	x       int32
	y       int32
	count   int32
	present int32
}

type mapgenPrefabAnalyzeDeps5262F0[T comparable] struct {
	lookup      func() int32
	storeIndex  func(int32)
	loadIndex   func() int32
	load        func(int32)
	width       func(int32) float32
	storeWidth  func(float32)
	height      func(int32) float32
	storeHeight func(float32)

	firstObject func() T
	nextObject  func(T) T
	position    func(T) mapgenPrefabFloatPoint5262F0
	round       func(mapgenPrefabFloatPoint5262F0) mapgenPrefabIntPoint5262F0
	extent      func(T) uint32
	markerID    [mapgenPrefabDirections5262F0]func() uint32

	loadExit  func(int) mapgenPrefabExit5262F0
	storeExit func(int, mapgenPrefabExit5262F0)
	capacity  func() (int32, int32)
	loadMax   func() int32
	storeMax  func(int32)
}

func mapgenPrefabIncrement5262F0(value int32) int32 {
	return int32(uint32(value) + 1)
}

func mapgenPrefabStoreMarker5262F0[T comparable](
	direction int,
	point mapgenPrefabIntPoint5262F0,
	deps mapgenPrefabAnalyzeDeps5262F0[T],
) {
	exit := deps.loadExit(direction)
	if exit.present == 0 {
		exit = mapgenPrefabExit5262F0{x: point.x, y: point.y, count: 1, present: 1}
		deps.storeExit(direction, exit)
		return
	}

	if direction == mapgenPrefabNorth5262F0 || direction == mapgenPrefabSouth5262F0 {
		if point.x < exit.x {
			exit.x = point.x
		}
		exit.y = point.y
	} else {
		if point.y < exit.y {
			exit.y = point.y
		}
		exit.x = point.x
	}
	exit.count = mapgenPrefabIncrement5262F0(exit.count)
	deps.storeExit(direction, exit)
}

// mapgenAnalyzePrefabWithDeps5262F0 preserves GAME.EXE 005262F0's write
// order and signed 32-bit comparisons. The original reads only the low word
// of each temporary object's PE32 extent before comparing it with the four
// 32-bit marker IDs; the native object field is deliberately narrowed here.
func mapgenAnalyzePrefabWithDeps5262F0[T comparable](deps mapgenPrefabAnalyzeDeps5262F0[T]) bool {
	index := deps.lookup()
	deps.storeIndex(index)
	if index == -1 {
		return false
	}

	deps.load(index)
	width := deps.width(deps.loadIndex())
	heightIndex := deps.loadIndex()
	deps.storeWidth(width)
	deps.storeHeight(deps.height(heightIndex))

	var zero T
	for object := deps.firstObject(); object != zero; object = deps.nextObject(object) {
		position := deps.position(object)
		point := deps.round(mapgenPrefabFloatPoint5262F0{
			x: position.x - float32(1),
			y: position.y - float32(1),
		})
		extent := uint32(uint16(deps.extent(object)))

		direction := -1
		if extent == deps.markerID[mapgenPrefabNorth5262F0]() {
			direction = mapgenPrefabNorth5262F0
		} else if extent == deps.markerID[mapgenPrefabSouth5262F0]() {
			direction = mapgenPrefabSouth5262F0
		} else if extent == deps.markerID[mapgenPrefabWest5262F0]() {
			direction = mapgenPrefabWest5262F0
		} else if extent == deps.markerID[mapgenPrefabEast5262F0]() {
			direction = mapgenPrefabEast5262F0
		}
		if direction >= 0 {
			mapgenPrefabStoreMarker5262F0(direction, point, deps)
		}
	}

	capacityA, capacityB := deps.capacity()
	capacity := int32(uint32(capacityA) + uint32(capacityB))
	groups := int32(0)
	for direction := 0; direction < mapgenPrefabDirections5262F0; direction++ {
		exit := deps.loadExit(direction)
		if exit.present == 0 {
			continue
		}
		if exit.count > deps.loadMax() {
			deps.storeMax(exit.count)
		}
		groups++
	}

	if deps.loadMax() > capacity || groups > 2 {
		return false
	}
	vertical := deps.loadExit(mapgenPrefabNorth5262F0).present != 0
	if !vertical {
		vertical = deps.loadExit(mapgenPrefabSouth5262F0).present != 0
	}
	if !vertical {
		return true
	}
	horizontal := deps.loadExit(mapgenPrefabEast5262F0).present != 0
	if !horizontal {
		horizontal = deps.loadExit(mapgenPrefabWest5262F0).present != 0
	}
	return !horizontal
}

func mapgenPrefabReadInt325262F0(base unsafe.Pointer, offset uintptr) int32 {
	return *(*int32)(unsafe.Add(base, offset))
}

func mapgenPrefabWriteInt325262F0(base unsafe.Pointer, offset uintptr, value int32) {
	*(*int32)(unsafe.Add(base, offset)) = value
}

func mapgenPrefabWriteFloat325262F0(base unsafe.Pointer, offset uintptr, value float32) {
	*(*float32)(unsafe.Add(base, offset)) = value
}

func mapgenPrefabRuntimeDeps5262F0(
	theme unsafe.Pointer,
	prefab unsafe.Pointer,
) mapgenPrefabAnalyzeDeps5262F0[unsafe.Pointer] {
	exitOffset := func(direction int) uintptr {
		return uintptr(80 + direction*16)
	}
	return mapgenPrefabAnalyzeDeps5262F0[unsafe.Pointer]{
		lookup: func() int32 {
			return int32(C.sub_5029A0((*C.char)(prefab)))
		},
		storeIndex: func(index int32) {
			mapgenPrefabWriteInt325262F0(prefab, 68, index)
		},
		loadIndex: func() int32 {
			return mapgenPrefabReadInt325262F0(prefab, 68)
		},
		load: func(index int32) {
			C.sub_502D70(C.int(index))
		},
		width: func(index int32) float32 {
			return float32(C.sub_502E70(C.int(index)))
		},
		storeWidth: func(width float32) {
			mapgenPrefabWriteFloat325262F0(prefab, 60, width)
		},
		height: func(index int32) float32 {
			return float32(C.sub_502EA0(C.int(index)))
		},
		storeHeight: func(height float32) {
			mapgenPrefabWriteFloat325262F0(prefab, 64, height)
		},
		firstObject: func() unsafe.Pointer {
			return unsafe.Pointer(C.sub_504980())
		},
		nextObject: func(object unsafe.Pointer) unsafe.Pointer {
			return unsafe.Pointer(C.sub_5049C0((*C.nox_object_t)(object)))
		},
		position: func(object unsafe.Pointer) mapgenPrefabFloatPoint5262F0 {
			var position C.float2
			C.sub_503EC0((*C.nox_object_t)(object), &position)
			return mapgenPrefabFloatPoint5262F0{
				x: float32(position.field_0),
				y: float32(position.field_4),
			}
		},
		round: func(point mapgenPrefabFloatPoint5262F0) mapgenPrefabIntPoint5262F0 {
			input := C.float2{field_0: C.float(point.x), field_4: C.float(point.y)}
			var output [2]C.uint32_t
			C.nox_xxx_mapGenRoundFloatToPtr_520DF0(&input, &output[0])
			return mapgenPrefabIntPoint5262F0{x: int32(output[0]), y: int32(output[1])}
		},
		extent: func(object unsafe.Pointer) uint32 {
			return uint32((*C.nox_object_t)(object).extent)
		},
		markerID: [mapgenPrefabDirections5262F0]func() uint32{
			func() uint32 { return uint32(C.dword_5d4594_2487656) },
			func() uint32 { return *memmap.PtrUint32(0x5D4594, 2487660) },
			func() uint32 { return *memmap.PtrUint32(0x5D4594, 2487664) },
			func() uint32 { return *memmap.PtrUint32(0x5D4594, 2487668) },
		},
		loadExit: func(direction int) mapgenPrefabExit5262F0 {
			offset := exitOffset(direction)
			return mapgenPrefabExit5262F0{
				x:       mapgenPrefabReadInt325262F0(prefab, offset),
				y:       mapgenPrefabReadInt325262F0(prefab, offset+4),
				count:   mapgenPrefabReadInt325262F0(prefab, offset+8),
				present: mapgenPrefabReadInt325262F0(prefab, offset+12),
			}
		},
		storeExit: func(direction int, exit mapgenPrefabExit5262F0) {
			offset := exitOffset(direction)
			mapgenPrefabWriteInt325262F0(prefab, offset, exit.x)
			mapgenPrefabWriteInt325262F0(prefab, offset+4, exit.y)
			mapgenPrefabWriteInt325262F0(prefab, offset+8, exit.count)
			mapgenPrefabWriteInt325262F0(prefab, offset+12, exit.present)
		},
		capacity: func() (int32, int32) {
			return mapgenPrefabReadInt325262F0(theme, 12), mapgenPrefabReadInt325262F0(theme, 16)
		},
		loadMax: func() int32 {
			return mapgenPrefabReadInt325262F0(prefab, 144)
		},
		storeMax: func(maximum int32) {
			mapgenPrefabWriteInt325262F0(prefab, 144, maximum)
		},
	}
}

var mapgenPrefabAnalyzeEntry5262F0 = func(theme, prefab unsafe.Pointer) bool {
	return mapgenAnalyzePrefabWithDeps5262F0(mapgenPrefabRuntimeDeps5262F0(theme, prefab))
}

//export nox_mapgenAnalyzePrefabNative_5262F0
func nox_mapgenAnalyzePrefabNative_5262F0(theme, prefab *C.uint8_t) C.int {
	if theme == nil || prefab == nil {
		return 0
	}
	return C.int(bool2int(mapgenPrefabAnalyzeEntry5262F0(unsafe.Pointer(theme), unsafe.Pointer(prefab))))
}
