//go:build !server

package legacy

/*
#include "GAME1_3.h"
*/
import "C"
import "unsafe"

// Declaration-only access to the real HUD entries and original ring scalar.
func CenterMessageResetC445450() unsafe.Pointer { return unsafe.Pointer(C.sub_445450()) }
func CenterMessageDrawC445530() int             { return int(C.nox_xxx_drawMessageLines_445530()) }
func CenterMessageIndexC445530() uint32         { return uint32(clientItemStorage.center_message_index) }
func CenterMessageSetIndexC445530(index uint32) {
	clientItemStorage.center_message_index = C.uint32_t(index)
}
