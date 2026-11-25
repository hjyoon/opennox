package legacy

/*
#include "GAME4_2.h"
#include "cgo_const.h"
#include "mapgen_theme_native.h"
*/
import "C"

import (
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const mapgenThemeWordCount = 0x45c / 4

//export nox_xxx_mapGenSpellIdByName_51E1D0
func nox_xxx_mapGenSpellIdByName_51E1D0(name *C.nox_cgo_const_char) C.int {
	full := "SPELL_" + mapgenUppercaseToken(GoStringP(unsafe.Pointer(name)))
	return C.int(nox_xxx_spellNameToN_4243F0(internCStr(full)))
}

func mapgenReadThemeSections(theme *[mapgenThemeWordCount]C.int, file *C.FILE) bool {
	buf := mapgenThemeTokenBuffer()
	text := (*C.char)(unsafe.Pointer(buf))
	data := (*C.uint8_t)(unsafe.Pointer(theme))
	for nox_xxx_mapGenReadLine_51E540(file, buf) != 0 {
		var result C.int
		switch mapgenUppercaseToken(GoString(text)) {
		case "ALGORITHM_DATA":
			result = nox_mapgenReadAlgDataNative_51EBB0(data, file)
		case "EXIT":
			result = nox_mapgenReadExitNative_51F800(data, file)
		case "PREFAB":
			result = nox_xxx_genReadPrefab_520BF0(data, file)
		case "AMBIENT_LIGHT":
			for i := 134; i <= 136; i++ {
				if nox_xxx_mapGenReadLine_51E540(file, buf) == 0 {
					return false
				}
				theme[i] = C.atoi(text)
			}
			continue
		case "SPELL_SET":
			result = nox_mapgenReadSpellSetNative_51EFB0(data, file)
		case "ARMOR_SET":
			result = nox_xxx_genReadArmorSet_51F640(data, file)
		case "WEAPON_SET":
			result = nox_xxx_genReadWeaponSet_51F030(data, file)
		case "DECOR":
			result = nox_xxx_genReadDecor_51F9F0((*C.uint32_t)(unsafe.Pointer(theme)), file)
		}
		if result == 0 {
			return false
		}
	}
	return true
}

//export nox_mapgenReadThemeNative_51E260
func nox_mapgenReadThemeNative_51E260(record *C.int, name *C.nox_cgo_const_char) C.int {
	theme := (*[mapgenThemeWordCount]C.int)(unsafe.Pointer(record))
	clear(theme[:])
	if registerMapgenPointer(unsafe.Pointer(record)) == 0 {
		return 0
	}
	*theme = [mapgenThemeWordCount]C.int{
		1: 5, 2: 2, 3: 3, 4: 1, 5: 3, 6: 40, 7: 20, 8: 10, 9: 5, 10: 20,
		mapgenThemeMergeRateIndex: 100, 13: 100, 14: 1, 16: 1157234688, 18: 100, 19: C.int(time.Now().Unix()),
	}
	*memmap.PtrUint32(0x5D4594, 2487520) = 0
	themeName := "(null)"
	if name != nil {
		themeName = GoStringP(unsafe.Pointer(name))
	}
	path, releasePath := alloc.CString("mapgen/" + themeName + ".thm")
	file := nox_binfile_open_408CC0((*C.char)(unsafe.Pointer(path)), 0)
	releasePath()
	if file == nil || nox_binfile_cryptSet_408D40(file, 1) == 0 {
		return 0
	}
	parsed := mapgenReadThemeSections(theme, file)
	nox_binfile_close_408D90(file)
	if !parsed || nox_xxx_mapgenCheckSettings_520AD0(&theme[22]) == 0 || nox_xxx_mapgenCheckSettings_520AD0(&theme[30]) == 0 {
		return 0
	}
	if theme[46] != 0 && nox_xxx_mapgenCheckSettings_520AD0(&theme[46]) == 0 {
		return 0
	}
	return 1
}

//export nox_xxx_mapGenReadTheme_51E260
func nox_xxx_mapGenReadTheme_51E260(theme *C.int, token C.int) C.int {
	name := (*C.nox_cgo_const_char)(resolveMapgenPointer(uintptr(uint32(token))))
	return nox_mapgenReadThemeNative_51E260(theme, name)
}
