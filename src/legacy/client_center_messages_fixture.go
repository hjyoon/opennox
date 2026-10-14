//go:build !server

package legacy

/*
#include "GAME1_3.h"
extern uint32_t dword_5d4594_825736;
*/
import "C"
import "unsafe"

// Declaration-only access to the real HUD entries and original ring scalar.
func CenterMessageResetC445450() unsafe.Pointer { return unsafe.Pointer(C.sub_445450()) }
func CenterMessageDrawC445530() int             { return int(C.nox_xxx_drawMessageLines_445530()) }
func CenterMessageIndexC445530() uint32         { return uint32(C.dword_5d4594_825736) }
func CenterMessageSetIndexC445530(index uint32) { C.dword_5d4594_825736 = C.uint32_t(index) }
