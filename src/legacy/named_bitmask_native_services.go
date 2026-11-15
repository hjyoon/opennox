package legacy

/*
#include "defs.h"
#include "noxstring.h"
#include <string.h>
static int named_mask_prefix(const char* input) { return strncmp(input,"NULL",4); }
static char* named_mask_token(char* input) { return strtok(input,"+"); }
*/
import "C"
import "unsafe"

// Serial observation of the three declared libc boundaries in the independent
// original reference. The kernel owns all pointer traversal and mask updates.
type namedBitmaskObserver423930 struct {
	Compare func(unsafe.Pointer, unsafe.Pointer) int32
	Prefix  func(unsafe.Pointer) int32
	Token   func(unsafe.Pointer) unsafe.Pointer
}

var namedBitmaskObserved423930 *namedBitmaskObserver423930

func namedBitmaskObserve423930(o *namedBitmaskObserver423930) func() {
	previous := namedBitmaskObserved423930
	namedBitmaskObserved423930 = o
	return func() {
		if namedBitmaskObserved423930 != o {
			panic("named-bitmask observer restore order")
		}
		namedBitmaskObserved423930 = previous
	}
}

func namedBitmaskCompare4239C0(a, b unsafe.Pointer) int32 {
	if o := namedBitmaskObserved423930; o != nil && o.Compare != nil {
		return o.Compare(a, b)
	}
	return int32(C.nox_strcmpi((*C.char)(a), (*C.char)(b)))
}

func namedBitmaskPrefix423930(input unsafe.Pointer) int32 {
	if o := namedBitmaskObserved423930; o != nil && o.Prefix != nil {
		return o.Prefix(input)
	}
	return int32(C.named_mask_prefix((*C.char)(input)))
}

func namedBitmaskToken423930(input unsafe.Pointer) unsafe.Pointer {
	if o := namedBitmaskObserved423930; o != nil && o.Token != nil {
		return o.Token(input)
	}
	return unsafe.Pointer(C.named_mask_token((*C.char)(input)))
}
