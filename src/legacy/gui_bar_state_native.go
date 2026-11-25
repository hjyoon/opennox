package legacy

/*
#include "defs.h"
extern nox_window_yyy nox_windows_arr_1093036[7];
extern uint32_t dword_5d4594_1090280, dword_5d4594_1090284;
extern uint32_t dword_5d4594_1096260, dword_5d4594_1096264;
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func guiBarStateNativeWord(index int) *uint32 {
	switch uint32(index) {
	case 0:
		return (*uint32)(&C.dword_5d4594_1090280)
	case 1:
		return (*uint32)(&C.dword_5d4594_1090284)
	case 2:
		return g2ServerWord(1091964)
	case 3:
		return g2ServerWord(1092992)
	case 4:
		return (*uint32)(&C.nox_windows_arr_1093036[0].field_1)
	case 5:
		return (*uint32)(&C.nox_windows_arr_1093036[0].field_2)
	case 6:
		return (*uint32)(&C.nox_windows_arr_1093036[0].color_1)
	case 7:
		return (*uint32)(&C.nox_windows_arr_1093036[0].color_2)
	case 8:
		return (*uint32)(&C.nox_windows_arr_1093036[1].field_1)
	case 9:
		return (*uint32)(&C.nox_windows_arr_1093036[1].field_2)
	case 10:
		return (*uint32)(&C.dword_5d4594_1096260)
	case 11:
		return (*uint32)(&C.dword_5d4594_1096264)
	default:
		return nil
	}
}

func guiBarStateNative(index int, value, maximum int32) int32 {
	switch index {
	case 0:
		*guiBarStateNativeWord(11) = uint32(value)
		first, second := *guiBarStateNativeWord(1), *guiBarStateNativeWord(0)
		result := second
		if value != 0 {
			first, second = *guiBarStateNativeWord(2), *guiBarStateNativeWord(3)
			result = first
		}
		*guiBarStateNativeWord(6), *guiBarStateNativeWord(7) = first, second
		return int32(result)
	case 1, 3:
		if player := (*server.Player)(globalStateNativePointer(0)); player != nil {
			if index == 1 {
				player.Info().SetField2247(uint32(maximum))
			} else {
				player.Info().SetField2243(uint32(maximum))
			}
		}
		current := 4
		if index == 3 {
			current = 8
		}
		*guiBarStateNativeWord(current + 1) = uint32(maximum)
		*guiBarStateNativeWord(current) = uint32(value)
		*guiBarStateNativeWord(10) = 32
		return value
	case 2:
		*guiBarStateNativeWord(4) = uint32(value)
		return value
	case 4:
		*guiBarStateNativeWord(8) = uint32(value)
		return value
	case 5:
		return int32(*guiBarStateNativeWord(8))
	}
	panic("invalid GUI bar state entry")
}
