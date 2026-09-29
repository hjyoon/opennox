package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4_2.h"
#include "common__binfile.h"
#include "mapgen_legacy_ptr.h"

extern uint32_t dword_5d4594_2487524;

typedef struct nox_test_mapgen_item_sets_51f030_result {
	int result;
	int count;
	int cleanup_cleared;
	int template_cleared;
	uintptr_t theme_address;
	uint32_t theme_token;
	uintptr_t entry_addresses[2];
	uint32_t entry_tokens[2];
	uintptr_t next_addresses[2];
	char lookup_names[2][60];
	char object_names[2][60];
	uintptr_t modifier_addresses[2][4];
	uint32_t modifier_tokens[2][4];
	uint32_t modifier_counts[2][4];
	char modifier_names[2][4][3][60];
} nox_test_mapgen_item_sets_51f030_result;

static nox_test_mapgen_item_sets_51f030_result nox_test_mapgen_item_sets_51f030(
	const char* path, int armor) {
	nox_test_mapgen_item_sets_51f030_result out = {0};
	uint8_t* theme = (uint8_t*)calloc(1, 0x45C);
	if (!theme) {
		return out;
	}
	out.theme_address = (uintptr_t)theme;
	out.theme_token = nox_mapgenLegacyPtrRegister(theme);
	if (!out.theme_token) {
		free(theme);
		return out;
	}

	FILE* file = nox_binfile_open_408CC0((char*)path, NOX_BINFILE_READ);
	if (file) {
		out.result = armor ? nox_xxx_genReadArmorSet_51F640(theme, file) :
							   nox_xxx_genReadWeaponSet_51F030(theme, file);
		nox_binfile_close_408D90(file);
	}

	size_t head_offset = armor ? 1108 : 1100;
	size_t count_offset = armor ? 1112 : 1104;
	out.count = *(uint32_t*)(theme + count_offset);
	uint32_t entry_token = *(uint32_t*)(theme + head_offset);
	for (int entry_index = 0; entry_index < 2 && entry_token; ++entry_index) {
		uint8_t* entry = (uint8_t*)nox_mapgenLegacyPtrResolve(entry_token);
		if (!entry) {
			break;
		}
		out.entry_tokens[entry_index] = entry_token;
		out.entry_addresses[entry_index] = (uintptr_t)entry;
		memcpy(out.lookup_names[entry_index], entry, 60);
		memcpy(out.object_names[entry_index], entry + 60, 60);
		for (int slot = 0; slot < 4; ++slot) {
			uint32_t modifier_token = *(uint32_t*)(entry + 120 + 4 * slot);
			uint32_t modifier_count = *(uint32_t*)(entry + 136 + 4 * slot);
			char* modifiers = (char*)nox_mapgenLegacyPtrResolve(modifier_token);
			out.modifier_tokens[entry_index][slot] = modifier_token;
			out.modifier_counts[entry_index][slot] = modifier_count;
			out.modifier_addresses[entry_index][slot] = (uintptr_t)modifiers;
			for (uint32_t i = 0; modifiers && i < modifier_count && i < 3; ++i) {
				memcpy(out.modifier_names[entry_index][slot][i], modifiers + 60 * i, 60);
			}
		}
		entry_token = *(uint32_t*)(entry + 152);
		out.next_addresses[entry_index] =
			(uintptr_t)nox_mapgenLegacyPtrResolve(entry_token);
	}
	out.template_cleared = dword_5d4594_2487524 == 0;

	sub_520D50((uint32_t*)theme);
	out.cleanup_cleared = *(uint32_t*)(theme + 1100) == 0 &&
		*(uint32_t*)(theme + 1104) == 0 && *(uint32_t*)(theme + 1108) == 0 &&
		*(uint32_t*)(theme + 1112) == 0;
	nox_mapgenLegacyPtrForget(theme);
	free(theme);
	return out;
}

static const char* nox_test_mapgen_item_sets_lookup_51f030(
	const nox_test_mapgen_item_sets_51f030_result* out, int entry) {
	return out->lookup_names[entry];
}

static const char* nox_test_mapgen_item_sets_object_51f030(
	const nox_test_mapgen_item_sets_51f030_result* out, int entry) {
	return out->object_names[entry];
}

static const char* nox_test_mapgen_item_sets_modifier_51f030(
	const nox_test_mapgen_item_sets_51f030_result* out, int entry, int slot, int modifier) {
	return out->modifier_names[entry][slot][modifier];
}
*/
import "C"

import "unsafe"

type mapgenItemSetsResult51F030 struct {
	result            bool
	count             int
	cleanupCleared    bool
	templateCleared   bool
	themeAddress      uintptr
	themeToken        uint32
	entryAddresses    [2]uintptr
	entryTokens       [2]uint32
	nextAddresses     [2]uintptr
	lookupNames       [2]string
	objectNames       [2]string
	modifierAddresses [2][4]uintptr
	modifierTokens    [2][4]uint32
	modifierCounts    [2][4]uint32
	modifierNames     [2][4][3]string
}

func mapgenItemSetsFixture51F030(path string, armor bool) mapgenItemSetsResult51F030 {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	var carmor C.int
	if armor {
		carmor = 1
	}
	value := C.nox_test_mapgen_item_sets_51f030(cpath, carmor)
	out := mapgenItemSetsResult51F030{
		result:          value.result != 0,
		count:           int(value.count),
		cleanupCleared:  value.cleanup_cleared != 0,
		templateCleared: value.template_cleared != 0,
		themeAddress:    uintptr(value.theme_address),
		themeToken:      uint32(value.theme_token),
	}
	for entry := range out.entryAddresses {
		out.entryAddresses[entry] = uintptr(value.entry_addresses[entry])
		out.entryTokens[entry] = uint32(value.entry_tokens[entry])
		out.nextAddresses[entry] = uintptr(value.next_addresses[entry])
		out.lookupNames[entry] = C.GoString(
			C.nox_test_mapgen_item_sets_lookup_51f030(&value, C.int(entry)))
		out.objectNames[entry] = C.GoString(
			C.nox_test_mapgen_item_sets_object_51f030(&value, C.int(entry)))
		for slot := range out.modifierAddresses[entry] {
			out.modifierAddresses[entry][slot] = uintptr(value.modifier_addresses[entry][slot])
			out.modifierTokens[entry][slot] = uint32(value.modifier_tokens[entry][slot])
			out.modifierCounts[entry][slot] = uint32(value.modifier_counts[entry][slot])
			for modifier := range out.modifierNames[entry][slot] {
				out.modifierNames[entry][slot][modifier] = C.GoString(
					C.nox_test_mapgen_item_sets_modifier_51f030(
						&value, C.int(entry), C.int(slot), C.int(modifier)))
			}
		}
	}
	return out
}
