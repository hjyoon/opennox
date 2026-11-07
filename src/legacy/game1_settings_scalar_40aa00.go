package legacy

/*
#include <stdint.h>
extern uint32_t dword_5d4594_10984;
static void* game1_settings_actual_global_40e090(void) { return &dword_5d4594_10984; }
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// The original 32-byte body reads frame, start, then FPS. Both arithmetic
// operations wrap at DWORD width, and equality does not satisfy the comparison.
func sub40AA00Native() int32 {
	frame := gameFrame()
	start := *memmap.PtrUint32(0x5D4594, 3520)
	fps := gameFPS()
	if frame-start > fps*20 {
		return 1
	}
	return 0
}

func sub40AA20Native() int32 {
	return int32(*memmap.PtrUint32(0x5D4594, 3536))
}

func sub40AA30Native(value int32) int32 {
	*memmap.PtrUint32(0x5D4594, 3536) = uint32(value)
	return value
}

func sub40AA40Native() int32 {
	return int32(*memmap.PtrUint32(0x5D4594, 3476))
}

func sub40AA60Native(value int32) int32 {
	*memmap.PtrUint32(0x5D4594, 3508) = uint32(value)
	return value
}

func game1SettingsGlobal40E090() unsafe.Pointer {
	return C.game1_settings_actual_global_40e090()
}

func sub40E090Native() {
	// This is an independent C symbol, not the packed blob at PE offset 10984.
	*(*uint32)(game1SettingsGlobal40E090()) = 0
}
