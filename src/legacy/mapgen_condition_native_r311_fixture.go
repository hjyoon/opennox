package legacy

/*
#include "GAME4_2.h"
#include "common__binfile.h"
#include "mapgen_legacy_ptr.h"
#include "mapgen_theme_native.h"

typedef struct nox_r311_condition_result {
	int result, guards_preserved, high_token, count;
	uint32_t value;
	uintptr_t file_address;
	char tokens[16][256];
} nox_r311_condition_result;

static nox_r311_condition_result nox_r311_read_condition(char* path, int mode, int legacy) {
	nox_r311_condition_result out = {0};
	uint32_t words[3] = {UINT32_C(0x12345678), UINT32_C(0xA5A5A5A5), UINT32_C(0x87654321)};
	FILE* file = nox_binfile_open_408CC0(path, 0);
	if (!file) return out;
	out.file_address = (uintptr_t)file;
	if (!nox_binfile_cryptSet_408D40(file, 1)) {
		nox_binfile_close_408D90(file);
		return out;
	}
	uint32_t token = 0;
	void* collision = NULL;
	if (legacy) {
#if UINTPTR_MAX > UINT32_MAX
		uint32_t low = (uint32_t)(uintptr_t)file;
		collision = (void*)(uintptr_t)low;
		uint32_t occupied = nox_mapgenLegacyPtrRegister(collision);
		token = nox_mapgenLegacyPtrRegister(file);
		out.high_token = occupied == low && token >= UINT32_C(0xF0000000);
#else
		token = nox_mapgenLegacyPtrRegister(file);
		out.high_token = token != 0;
#endif
	}
	if (mode == 0) {
		out.result = legacy ? sub_51E800((int)token, &words[1]) : nox_mapgenReadConditionNative_51E800(file, &words[1]);
	} else if (mode == 1) {
		out.result = legacy ? sub_51EAF0((int)token, &words[1]) : nox_mapgenReadComparisonNative_51EAF0(file, &words[1]);
	} else {
		uint8_t* buffer = getMemAt(0x5D4594, 2487264);
		while (out.count < 16 && nox_xxx_mapGenReadLine_51E540(file, buffer)) {
			memcpy(out.tokens[out.count++], buffer, strlen((char*)buffer) + 1);
		}
		out.result = out.count != 0;
	}
	if (legacy) {
		nox_mapgenLegacyPtrForget(file);
		nox_mapgenLegacyPtrForget(collision);
	}
	out.value = words[1];
	out.guards_preserved = words[0] == UINT32_C(0x12345678) && words[2] == UINT32_C(0x87654321);
	nox_binfile_close_408D90(file);
	return out;
}
*/
import "C"

import "unsafe"

type mapgenConditionResultR311 struct {
	result, guardsPreserved, highToken bool
	value                              uint32
	fileAddress                        uintptr
	tokens                             []string
}

func mapgenConditionFixtureR311(path string, mode int, legacy bool) mapgenConditionResultR311 {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	var flag C.int
	if legacy {
		flag = 1
	}
	out := C.nox_r311_read_condition(cpath, C.int(mode), flag)
	result := mapgenConditionResultR311{
		result: out.result != 0, guardsPreserved: out.guards_preserved != 0,
		highToken: out.high_token != 0, value: uint32(out.value), fileAddress: uintptr(out.file_address),
	}
	for i := 0; i < int(out.count); i++ {
		result.tokens = append(result.tokens, C.GoString(&out.tokens[i][0]))
	}
	return result
}
