package legacy

/*
#include "defs.h"
extern uint32_t dword_5d4594_1200832;
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func questNoticeNamePending48EA70() int {
	return int(C.dword_5d4594_1200832)
}

// Populate the genuine native guide record; the real 427240 lookup performs
// its ordinary ID/type gate and returns this C-owned UTF-16 title.
func questNoticeGuideFixture48EA70(title string) func() {
	old := C.nox_guide_entries_native[1]
	C.nox_guide_entries_native[1].thing_type = 1
	C.nox_guide_entries_native[1].name = (*C.wchar2_t)(unsafe.Pointer(alloc.InternCString16(title)))
	return func() { C.nox_guide_entries_native[1] = old }
}
