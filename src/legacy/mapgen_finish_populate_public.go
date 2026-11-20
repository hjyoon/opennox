package legacy

/*
#include "GAME2_1.h"
#include "GAME3_2.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
#include "server__mapgen__generate__populate.h"

extern uint32_t dword_5d4594_1550916;
extern uint32_t dword_5d4594_2487576;
extern uint32_t dword_5d4594_2487580;
extern uint32_t dword_5d4594_2487672;

static inline void mapgenPrintNoExit(wchar2_t* text) {
    nox_xxx_printToAll_4D9FD0(0, text);
}
*/
import "C"

import (
	"math"
	"unsafe"
)

type mapgenPopulateTheme struct {
	_            [60]byte
	crystalTiles uint32
	_            [472]byte
	ambientColor [3]uint32
}

type mapgenPopulateRoom struct {
	mapgenRoomGeometry
	flags         uint32
	_             [mapGenOccupiedRectHeadOffset - unsafe.Sizeof(mapgenRoomGeometry{}) - 4]byte
	occupiedRects uint32
	decor         uint32
}

var (
	_ [536]byte = [unsafe.Offsetof(mapgenPopulateTheme{}.ambientColor)]byte{}
	_ [52]byte  = [unsafe.Offsetof(mapgenPopulateRoom{}.flags)]byte{}
	_ [368]byte = [unsafe.Offsetof(mapgenPopulateRoom{}.occupiedRects)]byte{}
	_ [372]byte = [unsafe.Offsetof(mapgenPopulateRoom{}.decor)]byte{}
)

type mapgenFinishPopulateAccess struct {
	setFlags        func()
	makeExit        func(*mapgenPopulateTheme) bool
	printNoExit     func(int32)
	maxRoomDistance func()
	firstRoom       func() *mapgenPopulateRoom
	nextRoom        func(*mapgenPopulateRoom) *mapgenPopulateRoom
	decorate        func(*mapgenPopulateTheme, *mapgenPopulateRoom)
	resolveRect     func(uint32) *mapGenOccupiedRect5227B0
	selectCrystal   func(blue bool)
	fillCrystal     func(*mapGenOccupiedRect5227B0, int32, int32)
	playerStart     func(mapgenWorldPoint)
	setAmbientColor func(*[3]uint32)
	cleanup         func()
}

func mapgenFinishPopulate(theme *mapgenPopulateTheme, access mapgenFinishPopulateAccess) {
	if theme == nil {
		return
	}
	access.setFlags()
	if !access.makeExit(theme) {
		for line := int32(848); line <= 850; line++ {
			access.printNoExit(line)
		}
	}
	access.maxRoomDistance()
	for room := access.firstRoom(); room != nil; room = access.nextRoom(room) {
		access.setFlags()
		if room.decor != 0 && room.flags&2 == 0 {
			access.decorate(theme, room)
		}
		if theme.crystalTiles == 0 {
			continue
		}
		for token := room.occupiedRects; token != 0; {
			rect := access.resolveRect(token)
			if rect == nil {
				break
			}
			access.selectCrystal(rect.transient != 0)
			cols := mapgenCrystalTileCount(rect.bounds[0], rect.bounds[2])
			rows := mapgenCrystalTileCount(rect.bounds[1], rect.bounds[3])
			access.fillCrystal(rect, cols, rows)
			token = rect.next
		}
	}
	// The original reads the root again after decor and crystal placement.
	if root := access.firstRoom(); root != nil {
		center := func(axis int) float32 {
			minimum := float64(math.Float32frombits(uint32(root.bounds[axis])))
			maximum := float64(math.Float32frombits(uint32(root.bounds[axis+2])))
			return storeFloat32TowardZero(addFloat64TowardZero(maximum, minimum) * 0.5)
		}
		access.playerStart(mapgenWorldPoint{x: center(0), y: center(1)})
	}
	access.setAmbientColor(&theme.ambientColor)
	access.cleanup()
}

func mapgenCrystalTileCount(minimum, maximum float32) int32 {
	extent := addFloat64TowardZero(float64(maximum), -float64(minimum))
	biased := addFloat64TowardZero(extent, 0.5)
	return x87TruncQwordLow(mulFloat64TowardZero(biased, mapgenWorldToCellScale))
}

func mapgenReleaseProcessRecords(slot *uint32, release func(uint32)) {
	release(*slot)
	*slot = 0
}

//export sub_4D42C0
func sub_4D42C0() C.int {
	return C.int(C.dword_5d4594_1550916)
}

//export sub_5259D0
func sub_5259D0() C.double {
	return C.double(math.Float32frombits(uint32(C.dword_5d4594_2487580)))
}

//export sub_5259E0
func sub_5259E0() C.int {
	return C.int(C.dword_5d4594_2487576)
}

//export sub_526A90
func sub_526A90() {
	slot := (*uint32)(unsafe.Pointer(&C.dword_5d4594_2487672))
	mapgenReleaseProcessRecords(slot, func(token uint32) {
		records := C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token))
		C.nox_mapgenLegacyPtrForget(records)
		nox_mapgen_release_memory(records)
	})
}

//export nox_mapgenFinishPopulateNative_5228B0
func nox_mapgenFinishPopulateNative_5228B0(theme *C.uint8_t) {
	mapgenFinishPopulate((*mapgenPopulateTheme)(unsafe.Pointer(theme)), mapgenFinishPopulateAccess{
		setFlags: func() {
			flags := uint8(0x9d)
			C.nox_xxx_mapGenSetFlags_5235F0(C.char(flags))
		},
		makeExit: func(theme *mapgenPopulateTheme) bool {
			return nox_mapgenMakeExitNative_522A40((*C.uint8_t)(unsafe.Pointer(theme))) != 0
		},
		printNoExit: func(line int32) {
			text := nox_strman_loadString_40F1D0(internCStr("NoExit"), nil,
				internCStr(`C:\NoxPost\src\Server\MapGen\Generate\populate.c`), line)
			C.mapgenPrintNoExit(text)
		},
		maxRoomDistance: func() { C.nox_mapgenMaxRoomDistanceNative_5259D0() },
		firstRoom: func() *mapgenPopulateRoom {
			return (*mapgenPopulateRoom)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(C.dword_5d4594_1550916)))
		},
		nextRoom: func(room *mapgenPopulateRoom) *mapgenPopulateRoom {
			return (*mapgenPopulateRoom)(unsafe.Pointer(C.nox_mapgenRoomSortedNextNative_525C90((*C.uint8_t)(unsafe.Pointer(room)))))
		},
		decorate: func(theme *mapgenPopulateTheme, room *mapgenPopulateRoom) {
			nox_mapgenPopulateDecorNative_522340((*C.uint8_t)(unsafe.Pointer(theme)), (*C.uint8_t)(unsafe.Pointer(room)))
		},
		resolveRect: mapGenResolveOccupiedRect,
		selectCrystal: func(blue bool) {
			name := "CrystalRed"
			if blue {
				name = "CrystalBlue"
			}
			C.nox_xxx_tileGetDefByName_51D4D0(internCStr(name))
		},
		fillCrystal: func(rect *mapGenOccupiedRect5227B0, cols, rows int32) {
			mapgenFillTileRectangle((*mapgenWorldPoint)(unsafe.Pointer(&rect.bounds[0])), cols, rows, mapgenPlaceFillTile)
		},
		playerStart: func(position mapgenWorldPoint) {
			nox_xxx_mapGenGetObjID_527940(internCStr("PlayerStart"))
			nox_xxx_mapGenPlaceObj_5279B0((*C.float2)(unsafe.Pointer(&position)))
		},
		setAmbientColor: func(color *[3]uint32) {
			C.sub_469B90((*C.int)(unsafe.Pointer(&color[0])))
		},
		cleanup: sub_526A90,
	})
}

//export nox_xxx_mapGenFinishPopulate_5228B0_mapgen_populate
func nox_xxx_mapGenFinishPopulate_5228B0_mapgen_populate(themeToken C.int) {
	theme := C.nox_mapgenLegacyPtrResolve(C.uintptr_t(uint32(themeToken)))
	nox_mapgenFinishPopulateNative_5228B0((*C.uint8_t)(theme))
}
