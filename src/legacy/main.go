package legacy

/*
#include "GAME1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_2.h"
#include "GAME3_2.h"
#include "GAME5.h"
#include "common__net_list.h"
#include "client__system__parsecmd.h"
#include "common__object__armrlook.h"
#include "common__object__weaplook.h"
#include "common__log.h"

extern unsigned int dword_5d4594_2650652;
extern void* dword_587000_81128;
extern unsigned int dword_587000_93156;

extern nox_gui_animation* nox_wnd_xxx_1309740;
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

var (
	Nox_exit                                              func(exitCode int)
	Nox_xxx_gameGetScreenBoundaries_43BEB0_get_video_mode func(w, h, d *int)
	Sub_4AA9C0                                            func() int
)

//export nox_exit
func nox_exit(exitCode_cgo int32) { exitCode := int(exitCode_cgo); Nox_exit(exitCode) }

//export nox_xxx_gameGetScreenBoundaries_43BEB0_get_video_mode
func nox_xxx_gameGetScreenBoundaries_43BEB0_get_video_mode(w_cgo, h_cgo, d_cgo *int32) {
	w, w_cgo_finish := cgoABIIntPtr(w_cgo)
	defer w_cgo_finish()
	h, h_cgo_finish := cgoABIIntPtr(h_cgo)
	defer h_cgo_finish()
	d, d_cgo_finish := cgoABIIntPtr(d_cgo)
	defer d_cgo_finish()
	Nox_xxx_gameGetScreenBoundaries_43BEB0_get_video_mode(w, h, d)
}

//export sub_4AA9C0
func sub_4AA9C0() int32 { return int32(Sub_4AA9C0()) }

func Nox_xxx_loadLook_415D50() {
	loadLookStrings(371256, 35500, "C:\\NoxPost\\src\\common\\Object\\ArmrLook.c", 380)
}

func Nox_xxx_loadModifyers_4158C0() {
	loadLookStrings(371248, 33396, "C:\\NoxPost\\src\\common\\Object\\WeapLook.c", 200)
}

func loadLookStrings(flagOff, tableOff uintptr, source string, line int32) {
	flag := memmap.PtrUint32(0x5D4594, flagOff)
	if *flag != 0 {
		return
	}
	csource := internCStr(source)
	for off := tableOff; ; off += 12 {
		name := *memmap.PtrPtr(0x587000, off)
		if name == nil {
			break
		}
		text := nox_strman_loadString_40F1D0((*C.char)(name), nil, csource, line)
		*memmap.PtrPtr(0x587000, off-4) = unsafe.Pointer(text)
	}
	*flag = 1
}

func Sub_4D11A0() {
	C.sub_4D11A0()
}

func Sub_431370() int {
	return int(C.sub_431370())
}

func Nox_xxx_tileAlloc_410F60_init() int {
	return int(C.nox_xxx_tileAlloc_410F60_init())
}

func Nox_xxx_initSinCosTables_414C90() {
	C.nox_xxx_initSinCosTables_414C90()
}

func Nox_xxx_loadMapCycle_4D0A30() {
	C.nox_xxx_loadMapCycle_4D0A30()
}

func Nox_xxx_mapSelectFirst_4D0E00() {
	C.nox_xxx_mapSelectFirst_4D0E00()
}

func Sub_4134D0() {
	C.sub_4134D0()
}

func Sub_413920() {
	C.sub_413920()
}

func Sub_431380() {
	C.sub_431380()
}

func Nox_xxx_tileFree_410FC0_free() {
	C.nox_xxx_tileFree_410FC0_free()
}

func Sub_42EDC0() {
	C.sub_42EDC0()
}

func Sub_4D11D0() {
	C.sub_4D11D0()
}

func Sub_4D0DA0() {
	C.sub_4D0DA0()
}

func Nox_common_maplist_free_4D0970() {
	C.nox_common_maplist_free_4D0970()
}

func Sub_451970() {
	C.sub_451970()
}

func Sub_431270() {
	C.sub_431270()
}

func Sub_4875F0() {
	C.sub_4875F0()
}

func Sub_4870A0() {
	C.sub_4870A0()
}

func Sub_431290() {
	C.sub_431290()
}

func Nox_xxx_servSetPlrLimit_409F80(v int) {
	C.nox_xxx_servSetPlrLimit_409F80(C.int(v))
}

func Nox_xxx_guiChatShowHide_445730(v bool) {
	C.nox_xxx_guiChatShowHide_445730(C.int(bool2int(v)))
}
