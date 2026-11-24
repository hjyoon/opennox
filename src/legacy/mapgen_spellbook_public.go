package legacy

/*
#include "GAME4_2.h"
#include "cgo_const.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

const (
	mapGenThemeSpellCountOffset5220E0 = 1096
	mapGenThemeSpellIDsOffset5220E0   = 548
	mapGenThemeSpellStride5220E0      = 4
)

type mapGenSpellbookRuntime5220E0 struct {
	allocate func() *server.Object
	release  func(*server.Object)
	random   func(int32, int32) int32
	spellID  func() int
	finish   func(*server.Object, byte)
}

func mapGenMakeSpellbook5220E0(theme unsafe.Pointer, name func() string, runtime mapGenSpellbookRuntime5220E0) *server.Object {
	count := (*uint32)(unsafe.Add(theme, mapGenThemeSpellCountOffset5220E0))
	if *count == 0 {
		return nil
	}
	book := runtime.allocate()
	if book == nil {
		return nil
	}

	var id byte
	if name() == "*" {
		// The allocator can replace the count. The original reads it again
		// before calling the signed-range map RNG.
		index := runtime.random(0, int32(*count-1))
		id = *(*byte)(unsafe.Add(theme, mapGenThemeSpellIDsOffset5220E0+int(index)*mapGenThemeSpellStride5220E0))
	} else {
		id = byte(runtime.spellID())
		if id == 0 {
			runtime.release(book)
			return nil
		}
	}
	runtime.finish(book, id)
	return book
}

func mapGenFinishSpellbook527DB0(unit *server.Object, id byte, transfersSpellReward func(*server.Object) bool) bool {
	if unit == nil || !transfersSpellReward(unit) || unit.UseData.Ptr == nil {
		return false
	}
	*(*byte)(unit.UseData.Ptr) = id
	return true
}

//export nox_xxx_mapGenMakeSpellbook_5220E0
func nox_xxx_mapGenMakeSpellbook_5220E0(theme *C.uint8_t, name *C.nox_mapgen_spell_name_t) *nox_object_t {
	namePtr := (*C.char)(unsafe.Pointer(name))
	binding := GetServer().S()
	book := mapGenMakeSpellbook5220E0(unsafe.Pointer(theme), func() string { return GoString(namePtr) }, mapGenSpellbookRuntime5220E0{
		allocate: func() *server.Object { return binding.NewObjectByTypeID("SpellBook") },
		release: func(unit *server.Object) {
			binding.Objs.FreeObject(unit)
		},
		random: func(start, end int32) int32 {
			return int32(nox_xxx_mapGenRandFunc_526AC0(C.int(start), C.int(end)))
		},
		spellID: func() int {
			return int(nox_xxx_mapGenSpellIdByName_51E1D0((*C.nox_cgo_const_char)(namePtr)))
		},
		finish: func(unit *server.Object, id byte) {
			nox_xxx_mapGenFinishSpellbook_527DB0(asObjectC(unit), C.char(id))
		},
	})
	return asObjectC(book)
}

//export nox_xxx_mapGenFinishSpellbook_527DB0
func nox_xxx_mapGenFinishSpellbook_527DB0(unit *nox_object_t, id C.char) C.int {
	finish := mapGenFinishSpellbook527DB0(asObjectS(unit), byte(id), func(unit *server.Object) bool {
		name := nox_xxx_getUnitName_4E39D0(asObjectC(unit))
		return nox_objectTypeGetXfer(name) == Get_nox_xxx_XFerSpellReward_4F5F30()
	})
	return C.int(bool2int(finish))
}
