package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"os"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/ifs"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	mapgenInfoOffset       = 2408
	mapgenInfoSize         = 0x5b8
	mapgenObjectHeadOffset = 1550924
	mapgenMaxAttempts      = 100
)

func mapgenGenerateWithRetries(step func() int) bool {
	for range mapgenMaxAttempts {
		switch step() {
		case 0:
			return false
		case 1:
			return true
		}
	}
	return false
}

func mapgenGenerationStep() int { return int(C.nox_xxx_mapGenStep_4D44E0()) }

func mapgenFileAbsent(name string) bool {
	_, err := ifs.Stat(name)
	return os.IsNotExist(err)
}

func mapgenMoveScratchObjects(allowMissing bool) bool {
	root := datapath.Data()
	temporary := root + `\nc.obj`
	blend := root + `\blend.obj`
	if (!allowMissing || !mapgenFileAbsent(temporary)) && ifs.Remove(temporary) != nil {
		return false
	}
	if allowMissing && mapgenFileAbsent(blend) {
		return true
	}
	return ifs.Rename(blend, temporary) == nil
}

func mapgenResetHeader() []byte {
	info := memmap.Slice(0x973F18, mapgenInfoOffset)[:mapgenInfoSize]
	clear(info)
	return info
}

//export nox_xxx_mapGenStart_4D4320
func nox_xxx_mapGenStart_4D4320() C.int {
	nox_xxx_mapSwitchLevel_4D12E0(1)
	noxflags.SetGame(noxflags.GameFlag23)
	previous := memmap.PtrPtr(0x5D4594, mapgenObjectHeadOffset)
	*previous = unsafe.Pointer(nox_get_and_zero_server_objects_4DA3C0())
	info := mapgenResetHeader()
	if !mapgenGenerateWithRetries(mapgenGenerationStep) || !mapgenMoveScratchObjects(true) {
		return 0
	}
	for obj := GetServer().S().Objs.First(); obj != nil; obj = obj.Next() {
		obj.ScriptIDVal = 0
	}
	nox_xxx_mapGenMakeInfo_4D5DB0(unsafe.Pointer(&info[0]))
	name := GoString(nox_xxx_getRandMapName_4D4310())
	directory := datapath.Data() + `\Maps\$` + name
	_ = ifs.Mkdir(directory)
	if !Nox_xxx_mapSaveMap_51E010(directory+`\$`+name+".map", 1) {
		return 0
	}
	nox_xxx_mapSwitchLevel_4D12E0(1)
	nox_set_server_objects_4DA3E0((*nox_object_t)(*previous))
	*previous = nil
	noxflags.UnsetGame(noxflags.GameFlag23)
	return 1
}

//export nox_xxx_mapGenStartAlt_4D5F30
func nox_xxx_mapGenStartAlt_4D5F30() C.int {
	noxflags.SetGame(noxflags.GameFlag23)
	info := mapgenResetHeader()
	if !mapgenGenerateWithRetries(mapgenGenerationStep) || !mapgenMoveScratchObjects(false) {
		return 0
	}
	nox_xxx_mapGenMakeInfo_4D5DB0(unsafe.Pointer(&info[0]))
	noxflags.UnsetGame(noxflags.GameFlag23)
	return 1
}
