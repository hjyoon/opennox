package legacy

/*
#include <stddef.h>
#include <stdint.h>

extern uint32_t dword_5d4594_3484;
extern int nox_server_gameSettingsUpdated;
extern int sub_4D7EA0(void);

_Static_assert(sizeof(uint32_t) == 4, "settings flags DWORD ABI");
_Static_assert(sizeof(int) == 4, "settings dirty int ABI");
_Static_assert(sizeof(dword_5d4594_3484) == 4, "actual subflags global width");
_Static_assert(sizeof(nox_server_gameSettingsUpdated) == 4, "actual dirty global width");
_Static_assert(_Generic(&dword_5d4594_3484, uint32_t*: 1, default: 0), "actual subflags pointer type");
_Static_assert(_Generic(&nox_server_gameSettingsUpdated, int*: 1, default: 0), "actual dirty pointer type");

static void* setting_flags_subflags_cell_409e40(void) { return &dword_5d4594_3484; }
static void* setting_flags_dirty_cell_409e40(void) { return &nox_server_gameSettingsUpdated; }
static size_t setting_flags_subflags_width_409e40(void) { return sizeof(dword_5d4594_3484); }
static size_t setting_flags_dirty_width_409e40(void) { return sizeof(nox_server_gameSettingsUpdated); }
static size_t setting_flags_subflags_align_409e40(void) { return _Alignof(uint32_t); }
static size_t setting_flags_dirty_align_409e40(void) { return _Alignof(int); }
static int setting_flags_refresh_409e40(void) { return sub_4D7EA0(); }
*/
import "C"

import "unsafe"

// These are addresses of the actual independent C globals. No PE-address
// memmap cell, Go copy, adjacent layout, allocation or integer-pointer recovery.
func settingFlagsNativeCell409E40(field int) unsafe.Pointer {
	switch field {
	case 0:
		return C.setting_flags_subflags_cell_409e40()
	case 1:
		return C.setting_flags_dirty_cell_409e40()
	default:
		panic("unknown setting flags field")
	}
}

func settingFlagsNativeCellWidth409E40(field int) uintptr {
	switch field {
	case 0:
		return uintptr(C.setting_flags_subflags_width_409e40())
	case 1:
		return uintptr(C.setting_flags_dirty_width_409e40())
	default:
		panic("unknown setting flags field")
	}
}

func settingFlagsNativeCellAlign409E40(field int) uintptr {
	switch field {
	case 0:
		return uintptr(C.setting_flags_subflags_align_409e40())
	case 1:
		return uintptr(C.setting_flags_dirty_align_409e40())
	default:
		panic("unknown setting flags field")
	}
}

func settingFlagsNativeRefresh409E40() int32 {
	return int32(C.setting_flags_refresh_409e40())
}
