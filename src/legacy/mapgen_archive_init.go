package legacy

/*
#include "GAME3_2.h"
#include "cgo_const.h"

extern uint32_t dword_5d4594_3835368;
extern uint32_t dword_5d4594_3835372;
extern uint32_t dword_5d4594_3835396;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const mapgenWorkspacePathSize = 2048

func mapgenSetWorkspacePath(offset uintptr, directory, suffix string) *C.char {
	raw := memmap.PtrOff(0x973F18, offset)
	path := unsafe.Slice((*byte)(raw), mapgenWorkspacePathSize)
	at := copy(path, directory)
	at += copy(path[at:], suffix)
	path[at] = 0
	return (*C.char)(raw)
}

func mapgenResetWorkspaceSettings() {
	*memmap.PtrUint64(0x5D4594, 1549772) = uint64(nox_platform_get_ticks())
	clear(unsafe.Slice((*byte)(memmap.PtrOff(0x973F18, 35912)), 0x48))
	legacyWords.dword_5d4594_3835348 = 0
	floorEditState.brush_mode = 0
	floorEditState.edge_tile = 255
	floorEditState.edge_variation = 0
	legacyWords.dword_5d4594_3835364 = 1
	C.dword_5d4594_3835368 = 1
	C.dword_5d4594_3835372 = 1
	legacyAddressableWords.dword_5d4594_3835388 = 0
	legacyAddressableWords.dword_5d4594_3835392 = 1
	C.dword_5d4594_3835396 = ^C.uint32_t(0)
	*memmap.PtrUint8(0x973F18, 35972) = 2
	sub_51D0E0()
}

func mapgenInitializeAreaArchive(directory string) C.int {
	const archive = "\\AreaMap.lib"
	if len(directory) >= mapgenWorkspacePathSize-len(archive) {
		return 0
	}
	mapgenResetWorkspaceSettings()
	mapgenSetWorkspacePath(42152, directory, "")
	oldObjects := mapgenSetWorkspacePath(36008, directory, "\\oldObj.obj")
	blendObjects := mapgenSetWorkspacePath(38056, directory, "\\blend.obj")
	nox_fs_remove(oldObjects)
	nox_fs_remove(blendObjects)

	nox_xxx_mapReset_5028E0()
	archivePath, release := alloc.CString(directory + archive)
	defer release()
	sub_502A50((*C.char)(unsafe.Pointer(archivePath)))
	sub_502AB0((*C.char)(unsafe.Pointer(archivePath)))
	result := sub_502B10()
	legacyWords.dword_5d4594_3835312 = 0
	*memmap.PtrUint32(0x973F18, 35880) = 0
	*memmap.PtrUint32(0x5D4594, 1599580) = 0
	return result
}

//export sub_4D39F0
func sub_4D39F0(directory *C.nox_cgo_const_char) C.int {
	if directory == nil {
		mapgenResetWorkspaceSettings()
		for i, path := range [...]uintptr{42152, 40104, 36008, 38056} {
			*memmap.PtrUint8(0x973F18, path) = *memmap.PtrUint8(0x5D4594, uintptr(1549780+4*i))
		}
		nox_xxx_mapReset_5028E0()
		return 0
	}
	return mapgenInitializeAreaArchive(GoStringP(unsafe.Pointer(directory)))
}
