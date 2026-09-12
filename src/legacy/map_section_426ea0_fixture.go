package legacy

/*
#include <stdint.h>
#include <stdlib.h>

#include "GAME1_1.h"

typedef struct nox_test_map_section_426EA0_result {
	int result;
	uint32_t error;
	uintptr_t context;
} nox_test_map_section_426EA0_result;

static nox_test_map_section_426EA0_result nox_test_map_section_426EA0(const char* name) {
	nox_test_map_section_426EA0_result out = {0};
	uint8_t context = 0x5a;
	out.error = 0xffffffffu;
	out.context = (uintptr_t)&context;
	out.result = nox_xxx_mapReadSection_426EA0(&context, (char*)name, &out.error);
	return out;
}
*/
import "C"

import "unsafe"

type mapSectionViaCResult426EA0 struct {
	result  int
	error   uint32
	context uintptr
}

func mapSectionViaC426EA0(name string) mapSectionViaCResult426EA0 {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	v := C.nox_test_map_section_426EA0(cname)
	return mapSectionViaCResult426EA0{
		result:  int(v.result),
		error:   uint32(v.error),
		context: uintptr(v.context),
	}
}
