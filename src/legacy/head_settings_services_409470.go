package legacy

/*
#include "GAME1.h"
#include "GAME3_2.h"
#include "common__binfile.h"
#include <ctype.h>
#include <stdint.h>
static int head_settings_flags_409470(uint32_t mask) { return nox_common_gameFlags_check_40A5C0(mask) ? 1 : 0; }
static int head_settings_notify_409470(int value) { return nox_xxx_netNotifyRate_4D7F10(value); }
static int head_settings_compare_409470(void* a, void* b) { return nox_strcmpi((const char*)a, (const char*)b); }
static int head_settings_space_409470(uint16_t ch) { return isspace(ch); }
*/
import "C"
import "unsafe"

// Serial fixture hook. Each call reads the current observer afresh; a nested
// service may install its own observer and restore it before returning.
type headSettingsObserver409470 struct {
	Flags  func(uint32) bool
	Notify func(int32) int32
}

var headSettingsObserved409470 *headSettingsObserver409470

func headSettingsObserve409470(o *headSettingsObserver409470) func() {
	previous := headSettingsObserved409470
	headSettingsObserved409470 = o
	return func() {
		if headSettingsObserved409470 != o {
			panic("head settings observer restore order")
		}
		headSettingsObserved409470 = previous
	}
}
func headSettingsFlags409470(mask uint32) bool {
	if o := headSettingsObserved409470; o != nil && o.Flags != nil {
		return o.Flags(mask)
	}
	return C.head_settings_flags_409470(C.uint32_t(mask)) != 0
}
func headSettingsNotify409470(v int32) int32 {
	if o := headSettingsObserved409470; o != nil && o.Notify != nil {
		return o.Notify(v)
	}
	return int32(C.head_settings_notify_409470(C.int(v)))
}
func headSettingsCompare409470(a, b unsafe.Pointer) int32 {
	return int32(C.head_settings_compare_409470(a, b))
}
func headSettingsSpace409470(v uint16) bool { return C.head_settings_space_409470(C.uint16_t(v)) != 0 }
func headSettingsRead409470(file, out unsafe.Pointer) {
	nox_binfile_fread_408E40((*C.char)(out), 1, 1, (*C.FILE)(file))
}
func headSettingsLastErr409470(file unsafe.Pointer) int32 {
	return nox_binfile_lastErr_409370((*C.FILE)(file))
}
func headSettingsSkipLine409470(file unsafe.Pointer) { nox_binfile_skipLine_409520((*C.FILE)(file)) }

// Reference-only services use the same controlled service boundary while the
// private C bodies themselves are retained byte-for-byte apart from names.
//
//export head_settings_reference_flags_409470
func head_settings_reference_flags_409470(mask C.uint32_t) C.int {
	if headSettingsFlags409470(uint32(mask)) {
		return 1
	}
	return 0
}

//export head_settings_reference_notify_409470
func head_settings_reference_notify_409470(v C.int) C.int {
	return C.int(headSettingsNotify409470(int32(v)))
}
