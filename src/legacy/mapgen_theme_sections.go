package legacy

/*
#include "GAME4_2.h"
#include "cgo_const.h"
#include "mapgen_legacy_ptr.h"
#include "mapgen_theme_native.h"
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const mapgenThemeMergeRateIndex = 12

var mapgenAlgorithmIntegerFields = map[string]int{
	"MIDHALLLENGTH": 1, "HALLLENGTHVARIANCE": 2, "MIDHALLWIDTH": 3, "HALLWIDTHVARIANCE": 4,
	"HALLLIMIT": 5, "HALLBRANCHRATE": 6, "HALLROOMRATE": 7, "MIDROOMSIZE": 8,
	"ROOMVARIANCE": 9, "IRREGULARROOMRATE": 10, "EMPTYROOMRATE": 11, "MERGERATE": mapgenThemeMergeRateIndex,
	"ADJACENTPORTALRATE": 13, "RECURSIONLIMIT": 18, "SEED": 19,
}

func mapgenReadAlgorithmValue(file *C.FILE) []byte {
	value := make([]byte, 0, 60)
	put := func(index int, b byte) { value = append(value[:index], b) }
	if mapgenReadPlainToken(file, put) == 0 || sub_51E670(file) == 0 {
		return nil
	}
	return value
}

func mapgenSkeletonIndex(name string) uint32 {
	name = mapgenUppercaseToken(name)
	for i := uint32(0); ; i++ {
		skeleton := *memmap.PtrPtr(0x587000, 253244+4*uintptr(i))
		if skeleton == nil {
			return 0
		}
		if mapgenUppercaseToken(GoStringP(skeleton)) == name {
			return i
		}
	}
}

//export nox_mapgenReadAlgDataNative_51EBB0
func nox_mapgenReadAlgDataNative_51EBB0(raw *C.uint8_t, file *C.FILE) C.int {
	theme := (*[mapgenThemeWordCount]uint32)(unsafe.Pointer(raw))
	token := mapgenThemeTokenBuffer()
	for nox_xxx_mapGenReadLine_51E540(file, token) != 0 {
		if mapgenUppercaseToken(GoStringP(unsafe.Pointer(token))) == "END" {
			return 1
		}
		value := mapgenReadAlgorithmValue(file)
		if value == nil {
			return 0
		}
		cvalue := (*C.char)(unsafe.Pointer(&value[0]))
		key := mapgenUppercaseToken(GoStringP(unsafe.Pointer(token)))
		switch key {
		case "MAPSIZE":
			theme[16] = math.Float32bits(float32(C.atof(cvalue)))
		case "USEDOORS":
			theme[14] = uint32(bool2int(mapgenUppercaseToken(alloc.GoStringS(value)) == "TRUE"))
		case "DEBUG":
			theme[15] = uint32(bool2int(mapgenUppercaseToken(alloc.GoStringS(value)) == "TRUE"))
		case "SKELETON":
			theme[0] = mapgenSkeletonIndex(alloc.GoStringS(value))
		default:
			if index, ok := mapgenAlgorithmIntegerFields[key]; ok {
				theme[index] = uint32(C.atoi(cvalue))
			}
		}
	}
	return 0
}

func mapgenThemeByToken(token C.int) *C.uint8_t {
	return (*C.uint8_t)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(uint32(token))))
}

//export nox_xxx_genReadAlgData_51EBB0
func nox_xxx_genReadAlgData_51EBB0(token C.int, file *C.FILE) C.int {
	return nox_mapgenReadAlgDataNative_51EBB0(mapgenThemeByToken(token), file)
}

//export nox_mapgenReadSpellSetNative_51EFB0
func nox_mapgenReadSpellSetNative_51EFB0(raw *C.uint8_t, file *C.FILE) C.int {
	const maxSpells = 137
	count := (*uint32)(unsafe.Add(unsafe.Pointer(raw), mapGenThemeSpellCountOffset5220E0))
	list := (*[maxSpells]uint32)(unsafe.Add(unsafe.Pointer(raw), mapGenThemeSpellIDsOffset5220E0))
	token := mapgenThemeTokenBuffer()
	for nox_xxx_mapGenReadLine_51E540(file, token) != 0 {
		if mapgenUppercaseToken(GoStringP(unsafe.Pointer(token))) == "END" || *count >= maxSpells {
			return 1
		}
		if id := nox_xxx_mapGenSpellIdByName_51E1D0((*C.nox_cgo_const_char)(unsafe.Pointer(token))); id != 0 {
			list[*count] = uint32(id)
			*count++
		}
	}
	return 0
}

//export nox_xxx_genReadSpellSet_51EFB0
func nox_xxx_genReadSpellSet_51EFB0(token C.int, file *C.FILE) C.int {
	return nox_mapgenReadSpellSetNative_51EFB0(mapgenThemeByToken(token), file)
}

func copyMapgenName(dst []byte, name string) bool {
	if len(name) >= len(dst) {
		return false
	}
	alloc.StrCopy(dst, name)
	return true
}

//export nox_mapgenReadExitNative_51F800
func nox_mapgenReadExitNative_51F800(raw *C.uint8_t, file *C.FILE) C.int {
	theme := (*mapgenExitTheme)(unsafe.Pointer(raw))
	token := mapgenThemeTokenBuffer()
	read := func() bool { return nox_xxx_mapGenReadLine_51E540(file, token) != 0 }
	text := func() string { return GoStringP(unsafe.Pointer(token)) }
	for read() {
		switch mapgenUppercaseToken(text()) {
		case "END":
			return 1
		case "LINKDATA":
			if !read() || !copyMapgenName(theme.destination[:], text()) {
				return 0
			}
		case "OBJECT":
			if !read() || uint32(theme.exitCount) >= uint32(len(theme.exits)) {
				return 0
			}
			exit := &theme.exits[theme.exitCount]
			if !copyMapgenName(exit.name[:], text()) || !read() {
				return 0
			}
			switch mapgenUppercaseToken(text()) {
			case "NORTH":
				exit.side = 0
			case "SOUTH":
				exit.side = 1
			case "EAST":
				exit.side = 2
			case "WEST":
				exit.side = 3
			}
			theme.exitCount++
		default:
			return 0
		}
	}
	return 0
}

//export nox_xxx_genReadExit_51F800
func nox_xxx_genReadExit_51F800(token C.int, file *C.FILE) C.int {
	return nox_mapgenReadExitNative_51F800(mapgenThemeByToken(token), file)
}
