package legacy

/*
#include "world_native_storage.h"
#define dword_5d4594_3484 (nox_world_word_storage()->server_subflags)
#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(uint32_t) == 4, "settings flags DWORD ABI");
_Static_assert(sizeof(int) == 4, "settings dirty int ABI");
_Static_assert(sizeof(dword_5d4594_3484) == 4, "actual subflags global width");
_Static_assert(_Generic(&dword_5d4594_3484, uint32_t*: 1, default: 0), "actual subflags pointer type");

static size_t setting_flags_subflags_width_409e40(void) { return sizeof(dword_5d4594_3484); }
static size_t setting_flags_subflags_align_409e40(void) { return _Alignof(uint32_t); }
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// The native ABI tests snapshot 16 bytes on either side of the dirty word.
var settingsDirtyStorage, _ = alloc.NewHigh(struct {
	before [16]byte
	dirty  C.int
	after  [16]byte
}{})
var settingsDirtyWord = &settingsDirtyStorage.dirty

//export nox_settings_dirty_cell_409E40
func nox_settings_dirty_cell_409E40() *C.int { return settingsDirtyWord }

// Subflags and the settings dirty word use separate native DWORD cells.
func settingFlagsNativeCell409E40(field int) unsafe.Pointer {
	switch field {
	case 0:
		return unsafe.Pointer(&worldWordStorage.server_subflags)
	case 1:
		return unsafe.Pointer(settingsDirtyWord)
	default:
		panic("unknown setting flags field")
	}
}

func settingFlagsNativeCellWidth409E40(field int) uintptr {
	switch field {
	case 0:
		return uintptr(C.setting_flags_subflags_width_409e40())
	case 1:
		return unsafe.Sizeof(*settingsDirtyWord)
	default:
		panic("unknown setting flags field")
	}
}

func settingFlagsNativeCellAlign409E40(field int) uintptr {
	switch field {
	case 0:
		return uintptr(C.setting_flags_subflags_align_409e40())
	case 1:
		return unsafe.Alignof(*settingsDirtyWord)
	default:
		panic("unknown setting flags field")
	}
}

func settingFlagsNativeRefresh409E40() int32 {
	return int32(sub_4D7EA0())
}
