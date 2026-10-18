package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
#include "mapgen_theme_native.h"

typedef struct nox_r310_theme_storage {
	uint32_t before[8];
	uint32_t words[279];
	uint32_t after[8];
} nox_r310_theme_storage;

typedef struct nox_r310_theme_result {
	int result;
	int guards_preserved;
	int cleanup_cleared;
	int high_name_token;
	uintptr_t theme_address;
	uintptr_t name_address;
	uintptr_t room_address;
	uintptr_t hall_address;
	uint32_t words[279];
	char room_name[60];
	char hall_name[60];
	char exit_name[60];
	char link_data[60];
} nox_r310_theme_result;

static nox_r310_theme_result nox_r310_read_theme(const char* name, int legacy) {
	nox_r310_theme_result out = {0};
	nox_r310_theme_storage* value = malloc(sizeof(*value));
	if (!value) return out;
	memset(value, 0xA5, sizeof(*value));
	out.theme_address = (uintptr_t)value->words;
	out.name_address = (uintptr_t)name;
	if (legacy) {
#if UINTPTR_MAX > UINT32_MAX
		uint32_t low = (uint32_t)(uintptr_t)name;
		void* collision = (void*)(uintptr_t)low;
		uint32_t occupied = nox_mapgenLegacyPtrRegister(collision);
		uint32_t token = nox_mapgenLegacyPtrRegister((void*)name);
		out.high_name_token = occupied == low && token >= UINT32_C(0xF0000000);
		if (out.high_name_token) {
			out.result = nox_xxx_mapGenReadTheme_51E260((int*)value->words, (int)token);
		}
		nox_mapgenLegacyPtrForget((void*)name);
		nox_mapgenLegacyPtrForget(collision);
#else
		out.result = nox_xxx_mapGenReadTheme_51E260((int*)value->words, (int)(uintptr_t)name);
#endif
	} else {
		out.result = nox_mapgenReadThemeNative_51E260((int*)value->words, name);
	}
	memcpy(out.words, value->words, sizeof(out.words));
	void* room = nox_mapgenLegacyPtrResolve(value->words[22]);
	void* hall = nox_mapgenLegacyPtrResolve(value->words[30]);
	out.room_address = (uintptr_t)room;
	out.hall_address = (uintptr_t)hall;
	if (room) memcpy(out.room_name, room, sizeof(out.room_name));
	if (hall) memcpy(out.hall_name, hall, sizeof(out.hall_name));
	memcpy(out.exit_name, (uint8_t*)value->words + 216, sizeof(out.exit_name));
	memcpy(out.link_data, (uint8_t*)value->words + 476, sizeof(out.link_data));
	sub_520D50(value->words);
	out.cleanup_cleared = value->words[22] == 0 && value->words[23] == 0 &&
		value->words[30] == 0 && value->words[31] == 0 && value->words[38] == 0 &&
		value->words[39] == 0 && value->words[46] == 0 && value->words[47] == 0;
	out.guards_preserved = 1;
	for (unsigned i = 0; i < 8; ++i) {
		if (value->before[i] != UINT32_C(0xA5A5A5A5) || value->after[i] != UINT32_C(0xA5A5A5A5)) {
			out.guards_preserved = 0;
		}
	}
	nox_mapgenLegacyPtrForget(value->words);
	free(value);
	return out;
}
*/
import "C"

import "unsafe"

type mapgenThemeResultR310 struct {
	result, guardsPreserved, cleanupCleared, highNameToken bool
	themeAddress, nameAddress, roomAddress, hallAddress    uintptr
	words                                                  [279]uint32
	roomName, hallName, exitName, linkData                 string
}

func mapgenThemeFixtureR310(name string, legacy bool) mapgenThemeResultR310 {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var flag C.int
	if legacy {
		flag = 1
	}
	out := C.nox_r310_read_theme(cname, flag)
	result := mapgenThemeResultR310{
		result: out.result != 0, guardsPreserved: out.guards_preserved != 0,
		cleanupCleared: out.cleanup_cleared != 0, highNameToken: out.high_name_token != 0,
		themeAddress: uintptr(out.theme_address), nameAddress: uintptr(out.name_address),
		roomAddress: uintptr(out.room_address), hallAddress: uintptr(out.hall_address),
		roomName: C.GoString(&out.room_name[0]), hallName: C.GoString(&out.hall_name[0]),
		exitName: C.GoString(&out.exit_name[0]), linkData: C.GoString(&out.link_data[0]),
	}
	for i := range result.words {
		result.words[i] = uint32(out.words[i])
	}
	return result
}
