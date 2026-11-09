package legacy

/*
#include "defs.h"
#include "noxstring.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func titleDefaultCompare415960(query, title unsafe.Pointer) int32 {
	v := int32(C._nox_wcsicmp((*C.wchar2_t)(query), (*C.wchar2_t)(title)))
	runtime.KeepAlive(query)
	runtime.KeepAlive(title)
	return v
}

func titleDefaultLoadString415960(key unsafe.Pointer, source string, line int32) unsafe.Pointer {
	p := nox_strman_loadString_40F1D0((*C.char)(key), nil, internCStr(source), line)
	runtime.KeepAlive(key)
	return unsafe.Pointer(p)
}
