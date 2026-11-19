package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func importantRates4E4E50() *C.nox_important_rate_controls_t {
	return memmap.PtrT[C.nox_important_rate_controls_t](0x5D4594, 1565124)
}

func importantCounters4E4ED0() *C.nox_important_player_counters_t {
	return memmap.PtrT[C.nox_important_player_counters_t](0x5D4594, 1565524)
}

//export sub_4E4DE0
func sub_4E4DE0() C.int {
	if pool := C.nox_server_getImportantAllocClass_4E4DE0(); pool != nil {
		nox_free_alloc_class(pool)
	}
	C.nox_server_setImportantAllocClass_4E4DE0(nil)
	sub_4E4ED0()
	rates := importantRates4E4E50()
	clear(rates[:])
	C.nox_server_setImportantFirst_4E4F80(nil)
	C.nox_server_setImportantLast_4E4F80(nil)
	return initializeImportantRates4E4EF0(rates)
}

//export sub_4E4E50
func sub_4E4E50(index C.int) C.int {
	divisor := uint32(1)
	if Get_dword_5d4594_2650652() == 1 {
		divisor = uint32(headSettingsRate409470())
	}

	rate := &importantRates4E4E50()[index]
	interval := uint32(rate.resend_interval)
	if interval > 2 {
		frames := gameFPS() / divisor
		rate.lower_threshold = C.uint32_t(frames * (interval - 1) * uint32(rate.resends_per_update))
	} else {
		rate.lower_threshold = 0
	}
	frames := gameFPS() / divisor
	rate.threshold = C.uint32_t(frames * interval * uint32(rate.resends_per_update))
	return C.int(rate.threshold)
}

//export sub_4E4ED0
func sub_4E4ED0() C.int {
	clear(importantCounters4E4ED0()[:])
	return 0
}

//export sub_4E4EF0
func sub_4E4EF0() C.int {
	rates := importantRates4E4E50()
	clear(rates[:])
	return initializeImportantRates4E4EF0(rates)
}

func initializeImportantRates4E4EF0(rates *C.nox_important_rate_controls_t) C.int {
	var result C.int
	for index := range rates {
		result = nox_xxx_playerResetImportantCtr_4E4F40(C.int(index))
	}
	return result
}

//export sub_4E4F30
func sub_4E4F30(index C.int) C.int {
	importantCounters4E4ED0()[index] = 0
	return index
}

//export nox_xxx_playerResetImportantCtr_4E4F40
func nox_xxx_playerResetImportantCtr_4E4F40(index C.int) C.int {
	rate := &importantRates4E4E50()[index]
	rate.resend_interval = 2
	rate.resends_per_update = 1
	rate.update_rate = C.uint8_t(headSettingsRate409470())
	return sub_4E4E50(index)
}

//export sub_4E4F80
func sub_4E4F80() C.int {
	for packet := C.nox_server_getImportantFirst_4E4F80(); packet != nil; {
		kind := packet.payload[0]
		next := C.nox_server_getImportantNext_4E4F80(packet)
		if kind >= 0x31 && kind <= 0x33 {
			sub_4E4FC0(packet)
		}
		packet = next
	}
	return 0
}

//export sub_4E4FC0
func sub_4E4FC0(packet *C.nox_important_packet_t) {
	prev := C.nox_server_getImportantPrev_4E4F80(packet)
	next := C.nox_server_getImportantNext_4E4F80(packet)
	if prev != nil {
		C.nox_server_setImportantNext_4E4F80(prev, next)
	} else {
		C.nox_server_setImportantFirst_4E4F80(next)
	}
	if next != nil {
		C.nox_server_setImportantPrev_4E4F80(next, prev)
	} else {
		C.nox_server_setImportantLast_4E4F80(prev)
	}
	nox_alloc_class_free_obj_first(C.nox_server_getImportantAllocClass_4E4DE0(), unsafe.Pointer(packet))
}
