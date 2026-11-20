package legacy

/*
#include "GAME2.h"
#include "GAME5_2.h"
#include "cgo_const.h"

extern uint32_t dword_5d4594_2650652;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const (
	ruleSourceMap      = 1
	ruleSourceUser     = 2
	ruleSourceInternet = 4
)

//export sub_57A1E0
func sub_57A1E0(data *C.int, filename *C.nox_cgo_const_char, notes *C.int, sources C.char, mode C.short) C.char {
	return readMapRuleDefaults57A1E0((*server.Settings2)(unsafe.Pointer(data)), (*C.char)(filename),
		notes, uint8(sources), uint16(mode))
}

func readMapRuleDefaults57A1E0(settings *server.Settings2, filename *C.char, notes *C.int, sources uint8, mode uint16) C.char {
	if notes != nil {
		C.sub_57ADF0(notes)
	}
	for i := range settings.Field24.Vals {
		settings.Field24.Vals[i] = 0xffffffff
	}
	settings.Field44 = [4]byte{0xff, 0xff, 0xff, 0xff}
	settings.Field48 = 0xffffffff

	selection := C.int(noxflags.GameFlag(mode) & noxflags.GameModeMask)
	if sources&(ruleSourceMap|ruleSourceUser) != 0 {
		mapName := alloc.GoStringS(settings.Field0[:8])
		prefix := "maps\\" + mapName + "\\"
		userLoaded := false
		if sources&ruleSourceUser != 0 {
			name := "user.rul"
			if filename != nil {
				name = GoString(filename)
			}
			userLoaded = readLegacyMapRuleFile57A1E0(prefix+name, settings, notes, selection)
		}
		if sources&ruleSourceMap != 0 && !userLoaded {
			readLegacyMapRuleFile57A1E0(prefix+mapName+".rul", settings, notes, selection)
		}
	}
	if sources&ruleSourceInternet != 0 && C.dword_5d4594_2650652 != 0 {
		readLegacyMapRuleFile57A1E0("internet.rul", settings, notes, selection)
	}
	if noxflags.GameFlag(mode)&noxflags.GameModeFlagBall != 0 {
		return C.char(sub_453FA0(unsafe.Pointer(&settings.Field24), C.int(spell.SPELL_WALL), 0))
	}
	return C.char(int8(mode))
}

func readLegacyMapRuleFile57A1E0(path string, settings *server.Settings2, notes *C.int, mode C.int) bool {
	name := CString(path)
	defer StrFree(name)
	return C.sub_57A3F0(name, (*C.char)(unsafe.Pointer(settings)), notes, mode) != 0
}
