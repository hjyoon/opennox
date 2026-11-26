package legacy

/*
#include <stdlib.h>
#include "client__io__win95__focus.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

var windowFocused = true

// SetWindowFocused delivers focus transitions through the callbacks registered
// for GAME.EXE 0042ED20 and 0042ED70.
func SetWindowFocused(focused bool) {
	if focused == windowFocused {
		return
	}
	if focused {
		windowFocused = true
		runFocusCallbacks(&worldOwnerStorage.focus_primary, &worldOwnerStorage.focus_primary_count)
	} else {
		runFocusCallbacks(&worldOwnerStorage.focus_secondary, &worldOwnerStorage.focus_secondary_count)
		windowFocused = false
	}
}

func runFocusCallbacks(records **C.obj_5D4594_754088_t, count *C.int) {
	// A callback can register another entry and move the array.
	for i := 0; *records != nil && i < int(*count); i++ {
		record := unsafe.Slice(*records, int(*count))[i]
		ccall.CallVoidInt(unsafe.Pointer(record.fnc), int(record.field_4))
	}
}

//export sub_42EDC0
func sub_42EDC0() {
	C.free(unsafe.Pointer(worldOwnerStorage.focus_primary))
	C.free(unsafe.Pointer(worldOwnerStorage.focus_secondary))
	worldOwnerStorage.focus_primary, worldOwnerStorage.focus_secondary = nil, nil
	worldOwnerStorage.focus_primary_count, worldOwnerStorage.focus_secondary_count = 0, 0
	windowFocused = true
}

//export sub_42EBB0
func sub_42EBB0(kind C.uint, callback C.nox_focus_callback, field C.int, name *C.nox_focus_name) {
	var records **C.obj_5D4594_754088_t
	var count *C.int
	switch kind {
	case 1:
		records, count = &worldOwnerStorage.focus_primary, &worldOwnerStorage.focus_primary_count
	case 2:
		records, count = &worldOwnerStorage.focus_secondary, &worldOwnerStorage.focus_secondary_count
	default:
		return
	}
	label := GoString((*C.char)(name))
	n := int(*count)
	size := C.size_t(n+1) * C.size_t(unsafe.Sizeof(**records))
	*records = (*C.obj_5D4594_754088_t)(C.realloc(unsafe.Pointer(*records), size))
	if *records == nil {
		C.abort()
		return
	}
	record := &unsafe.Slice(*records, n+1)[n]
	record.fnc = callback
	record.field_4 = field
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&record.name[0])), len(record.name))
	end := copy(raw[:len(raw)-1], label)
	raw[end] = 0
	*count += 1
}
