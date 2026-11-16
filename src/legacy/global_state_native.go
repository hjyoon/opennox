package legacy

/*
#include "defs.h"
#include <string.h>
extern uintptr_t dword_8531A0_2576, dword_5d4594_1062480;
extern nox_window *dword_5d4594_1049508, *dword_5d4594_1305680, *dword_5d4594_1305684;
extern uint32_t dword_5d4594_1062488, dword_5d4594_1319056, dword_5d4594_1320964;
extern uint32_t dword_5d4594_1098628, dword_5d4594_1321800;
extern uint32_t nox_xxx_waypointCounterMB_587000_154948, dword_5d4594_2487244;
extern uint32_t nox_xxx_minimap_587000_149232;
extern nox_window_yyy nox_windows_arr_1093036[7];
void* nox_mapgen_last_waypoint_slot_native_51D0E0(void);
// Address bridges only. Existing C callers and Go entries share these owners.
static void* global_state_slot_C(unsigned i) {
 switch(i) {
 case 0: case 1: return &dword_8531A0_2576;
 case 2: return &dword_5d4594_1062480;
 case 3: return &dword_5d4594_1049508;
 case 4: return &dword_5d4594_1062488;
 case 5: return &dword_5d4594_1319056;
 case 6: return &dword_5d4594_1320964;
 case 7: return &dword_5d4594_1098628;
 case 8: return &dword_5d4594_1305680;
 case 9: return &dword_5d4594_1305684;
 case 10: return &dword_5d4594_1321800;
 case 11: return &nox_xxx_waypointCounterMB_587000_154948;
 case 12: return &dword_5d4594_2487244;
 case 13: case 14: return &nox_xxx_minimap_587000_149232;
 case 15: return &nox_windows_arr_1093036[0].field_1;
 case 16: return &nox_windows_arr_1093036[0].field_2;
 case 17: return nox_mapgen_last_waypoint_slot_native_51D0E0();
 default: return NULL;
 }
}
static void global_state_pointer_store_C(void* slot, void* value) { memcpy(slot, &value, sizeof(value)); }
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func globalStateNativeSlot(index int) unsafe.Pointer {
	return C.global_state_slot_C(C.uint(index))
}

func globalStateNativeWord(index int) *uint32 {
	return (*uint32)(globalStateNativeSlot(index))
}

func globalStateNativePointer(index int) unsafe.Pointer {
	return *(*unsafe.Pointer)(globalStateNativeSlot(index))
}

func globalStateNativeStorePointer(slot, value unsafe.Pointer) {
	// These slots belong to C and may contain foreign pointer bits. Copy them
	// without applying Go's deletion barrier to the previous contents.
	C.global_state_pointer_store_C(slot, value)
}

func globalStateNativeValue(index int) int32 {
	switch index {
	case 0:
		player := (*server.Player)(globalStateNativePointer(0))
		// The stock instruction compares the complete active word, including its
		// three reserved bytes. Locate that word through the native Player field.
		if player != nil && *(*uint32)(unsafe.Pointer(&player.Active)) == 1 && player.Field3680&3 != 0 {
			return 1
		}
		return 0
	case 1:
		if player := (*server.Player)(globalStateNativePointer(1)); player != nil {
			return int32(int8(player.Level))
		}
		return 1
	case 2:
		if cell := (*C.nox_inventory_cell_t)(globalStateNativePointer(2)); cell != nil {
			return int32(cell.field_4)
		}
		return 0
	case 3, 8, 9:
		if globalStateNativePointer(index) != nil {
			return 1
		}
		return 0
	case 7:
		if *globalStateNativeWord(index) == 2 {
			return 1
		}
		return 0
	case 4, 5, 6, 15, 16:
		return int32(*globalStateNativeWord(index))
	}
	panic("invalid global value entry")
}

func globalStateNativeMutate(index int) {
	word := globalStateNativeWord(index)
	switch index {
	case 10:
		*word = 0
	case 11:
		*word = ^uint32(0)
	case 12:
		// Keep the existing full-width waypoint projection in sync with its
		// original legacy word; the original executable had one pointer word.
		globalStateNativeStorePointer(globalStateNativeSlot(17), nil)
		*word = 0
	case 13:
		value := *word - 10
		if int32(value) < 500 {
			value = 500
		}
		*word = value
	case 14:
		value := *word + 10
		if int32(value) > 4000 {
			value = 4000
		}
		*word = value
	default:
		panic("invalid global mutation entry")
	}
}
