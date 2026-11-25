package legacy

/*
#include <stdlib.h>
#include "client__io__win95__focus.h"

extern obj_5D4594_754088_t *ptr_5D4594_754088, *ptr_5D4594_754092;
extern int ptr_5D4594_754088_cnt, ptr_5D4594_754092_cnt;
*/
import "C"

import "unsafe"

//export sub_42EBB0
func sub_42EBB0(kind C.uint, callback C.nox_focus_callback, field C.int, name *C.nox_focus_name) {
	var records **C.obj_5D4594_754088_t
	var count *C.int
	switch kind {
	case 1:
		records, count = &C.ptr_5D4594_754088, &C.ptr_5D4594_754088_cnt
	case 2:
		records, count = &C.ptr_5D4594_754092, &C.ptr_5D4594_754092_cnt
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
