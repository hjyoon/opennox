package legacy

/*
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
#include "mapgen_theme_native.h"
*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/opennox/libs/player"
	"github.com/opennox/opennox/v1/common/memmap"
)

func mapgenThemeTokenBuffer() *C.uint8_t {
	return (*C.uint8_t)(memmap.PtrOff(0x5D4594, 2487264))
}

func mapgenUppercaseToken(value string) string {
	buf := []byte(value)
	for i, ch := range buf {
		if ch >= 'a' && ch <= 'z' {
			buf[i] = ch - ('a' - 'A')
		}
	}
	return string(buf)
}

func mapgenSkipConditional(file *C.FILE, stopAtElse bool) C.int {
	scratch := mapgenThemeTokenBuffer()
	depth := 0
	for sub_51E570(file, scratch) != 0 {
		word := GoStringP(unsafe.Pointer(scratch))
		switch mapgenUppercaseToken(word) {
		case "IF":
			depth++
		case "ENDIF":
			if depth == 0 {
				return 1
			}
			depth--
		case "ELSE":
			if stopAtElse && depth == 0 {
				return 1
			}
		}
	}
	return 0
}

//export sub_51E720
func sub_51E720(file *C.FILE) C.int { return mapgenSkipConditional(file, false) }

//export sub_51E780
func sub_51E780(file *C.FILE) C.int { return mapgenSkipConditional(file, true) }

//export sub_51E670
func sub_51E670(file *C.FILE) C.int {
	scratch := mapgenThemeTokenBuffer()
	switch mapgenUppercaseToken(GoStringP(unsafe.Pointer(scratch))) {
	case "IF":
		var passed C.uint32_t
		if nox_mapgenReadConditionNative_51E800(file, &passed) == 0 {
			return 0
		}
		if passed == 0 && sub_51E780(file) == 0 {
			return 0
		}
	case "ELSE":
		if sub_51E720(file) == 0 {
			return 0
		}
	case "ENDIF":
	default:
		return 1
	}
	return nox_xxx_mapGenReadLine_51E540(file, scratch)
}

func mapgenHasPlayerClass(class player.Class) bool {
	players := &GetServer().S().Players
	for p := players.First(); p != nil; p = players.Next(p) {
		if p.PlayerClass() == class {
			return true
		}
	}
	return false
}

func mapgenReadThreshold(file *C.FILE, left C.int, output *C.uint32_t) C.int {
	var op C.uint32_t
	scratch := mapgenThemeTokenBuffer()
	if nox_mapgenReadComparisonNative_51EAF0(file, &op) == 0 || nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 {
		return 0
	}
	right := C.atoi((*C.char)(unsafe.Pointer(scratch)))
	var matches bool
	switch op {
	case 0:
		matches = left < right
	case 1:
		matches = left > right
	case 2:
		matches = left == right
	case 3:
		matches = left != right
	}
	*output = C.uint32_t(bool2int(matches))
	return 1
}

//export nox_mapgenReadConditionNative_51E800
func nox_mapgenReadConditionNative_51E800(file *C.FILE, output *C.uint32_t) C.int {
	scratch := mapgenThemeTokenBuffer()
	if nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 {
		return 0
	}
	token := GoStringP(unsafe.Pointer(scratch))
	var passed bool
	switch mapgenUppercaseToken(token) {
	case "WARRIOR":
		passed = mapgenHasPlayerClass(player.Warrior)
	case "WIZARD":
		passed = mapgenHasPlayerClass(player.Wizard)
	case "CONJURER":
		passed = mapgenHasPlayerClass(player.Conjurer)
	case "EXPERIENCE_LEVEL":
		return mapgenReadThreshold(file, 3, output)
	case "NUMPLAYERS":
		return mapgenReadThreshold(file, C.int(nox_common_playerInfoCount_416F40()), output)
	default:
		if !strings.ContainsRune(token, '%') {
			return 0
		}
		chance := C.atoi((*C.char)(unsafe.Pointer(scratch)))
		passed = nox_xxx_mapGenRandFunc_526AC0(1, 100) <= chance
	}
	*output = C.uint32_t(bool2int(passed))
	return 1
}

//export sub_51E800
func sub_51E800(fileToken C.int, output *C.uint32_t) C.int {
	file := (*C.FILE)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(uint32(fileToken))))
	return nox_mapgenReadConditionNative_51E800(file, output)
}

//export nox_mapgenReadComparisonNative_51EAF0
func nox_mapgenReadComparisonNative_51EAF0(file *C.FILE, output *C.uint32_t) C.int {
	scratch := mapgenThemeTokenBuffer()
	if nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 {
		return 0
	}
	switch GoStringP(unsafe.Pointer(scratch)) {
	case "<":
		*output = 0
	case ">":
		*output = 1
	case "==":
		*output = 2
	case "!=":
		*output = 3
	default:
		return 0
	}
	return 1
}

//export sub_51EAF0
func sub_51EAF0(fileToken C.int, output *C.uint32_t) C.int {
	file := (*C.FILE)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(uint32(fileToken))))
	return nox_mapgenReadComparisonNative_51EAF0(file, output)
}
