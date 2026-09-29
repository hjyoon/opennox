package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4_2.h"
#include "common__binfile.h"
#include "mapgen_legacy_ptr.h"

typedef struct nox_test_mapgen_decor_51f9f0_result {
	int template_parsed;
	int backdrop_parsed;
	int settings_valid;
	int selected;
	int cleanup_cleared;
	uintptr_t theme_address;
	uintptr_t template_address;
	uintptr_t backdrop_address;
	uintptr_t template_wall_address;
	uintptr_t backdrop_wall_address;
	uintptr_t template_set_address;
	uintptr_t backdrop_set_address;
	uintptr_t selected_address;
	uint32_t theme_token;
	uint32_t template_token;
	uint32_t backdrop_token;
	uint32_t template_wall_token;
	uint32_t backdrop_wall_token;
	uint32_t template_set_token;
	uint32_t backdrop_set_token;
	uint32_t selected_token;
	uint32_t template_wall_count;
	uint32_t backdrop_wall_count;
	uint32_t template_set_count;
	uint32_t backdrop_set_count;
	uint32_t template_set_shares_entries;
	uint32_t backdrop_set_shares_entries;
	uint32_t frequency;
	int32_t min_room_size;
	int32_t max_room_size;
	char template_name[60];
	char backdrop_name[60];
	char wall_name[60];
	char floor_name[60];
	char door_name[60];
	char double_door_name[60];
} nox_test_mapgen_decor_51f9f0_result;

static int nox_test_mapgen_read_decor_51f9f0(uint8_t* theme, const char* path) {
	FILE* file = nox_binfile_open_408CC0((char*)path, NOX_BINFILE_READ);
	if (!file) {
		return 0;
	}
	int result = nox_xxx_genReadDecor_51F9F0((uint32_t*)theme, file);
	nox_binfile_close_408D90(file);
	return result;
}

static nox_test_mapgen_decor_51f9f0_result nox_test_mapgen_decor_51f9f0(
	const char* template_path, const char* backdrop_path) {
	nox_test_mapgen_decor_51f9f0_result out = {0};
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

	out.template_parsed = nox_test_mapgen_read_decor_51f9f0(theme, template_path);
	out.backdrop_parsed = nox_test_mapgen_read_decor_51f9f0(theme, backdrop_path);
	out.template_token = *(uint32_t*)(theme + 152);
	out.backdrop_token = *(uint32_t*)(theme + 184);
	uint8_t* template_definition =
		(uint8_t*)nox_mapgenLegacyPtrResolve(out.template_token);
	uint8_t* backdrop_definition =
		(uint8_t*)nox_mapgenLegacyPtrResolve(out.backdrop_token);
	out.template_address = (uintptr_t)template_definition;
	out.backdrop_address = (uintptr_t)backdrop_definition;

	if (template_definition) {
		memcpy(out.template_name, template_definition, sizeof(out.template_name));
		out.template_wall_token = *(uint32_t*)(template_definition + 84);
		out.template_wall_count = *(uint32_t*)(template_definition + 88);
		out.template_set_token = *(uint32_t*)(template_definition + 92);
		out.template_set_count = *(uint32_t*)(template_definition + 96);
	}
	if (backdrop_definition) {
		memcpy(out.backdrop_name, backdrop_definition, sizeof(out.backdrop_name));
		out.frequency = *(uint32_t*)(backdrop_definition + 72);
		out.min_room_size = *(int32_t*)(backdrop_definition + 76);
		out.max_room_size = *(int32_t*)(backdrop_definition + 80);
		out.backdrop_wall_token = *(uint32_t*)(backdrop_definition + 84);
		out.backdrop_wall_count = *(uint32_t*)(backdrop_definition + 88);
		out.backdrop_set_token = *(uint32_t*)(backdrop_definition + 92);
		out.backdrop_set_count = *(uint32_t*)(backdrop_definition + 96);
		memcpy(out.door_name, backdrop_definition + 100, sizeof(out.door_name));
		memcpy(out.double_door_name, backdrop_definition + 160,
			sizeof(out.double_door_name));
	}

	uint8_t* template_wall =
		(uint8_t*)nox_mapgenLegacyPtrResolve(out.template_wall_token);
	uint8_t* backdrop_wall =
		(uint8_t*)nox_mapgenLegacyPtrResolve(out.backdrop_wall_token);
	uint8_t* template_set =
		(uint8_t*)nox_mapgenLegacyPtrResolve(out.template_set_token);
	uint8_t* backdrop_set =
		(uint8_t*)nox_mapgenLegacyPtrResolve(out.backdrop_set_token);
	out.template_wall_address = (uintptr_t)template_wall;
	out.backdrop_wall_address = (uintptr_t)backdrop_wall;
	out.template_set_address = (uintptr_t)template_set;
	out.backdrop_set_address = (uintptr_t)backdrop_set;
	if (backdrop_wall) {
		memcpy(out.wall_name, backdrop_wall, sizeof(out.wall_name));
		memcpy(out.floor_name, backdrop_wall + 60, sizeof(out.floor_name));
	}
	if (template_set) {
		out.template_set_shares_entries = *(uint32_t*)(template_set + 16);
	}
	if (backdrop_set) {
		out.backdrop_set_shares_entries = *(uint32_t*)(backdrop_set + 16);
	}

	out.settings_valid = nox_xxx_mapgenCheckSettings_520AD0((int*)(theme + 184));
	uint8_t* room = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(4, 4);
	if (room) {
		*(uint32_t*)(room + 364) = 1;
		out.selected = nox_mapgenSelectBackdropNative_524070(theme, room);
		out.selected_token = *(uint32_t*)(room + 372);
		out.selected_address =
			(uintptr_t)nox_mapgenLegacyPtrResolve(out.selected_token);
		sub_521A10(room);
	}

	sub_520D50((uint32_t*)theme);
	out.cleanup_cleared = *(uint32_t*)(theme + 152) == 0 &&
		*(uint32_t*)(theme + 156) == 0 && *(uint32_t*)(theme + 184) == 0 &&
		*(uint32_t*)(theme + 188) == 0;
	nox_mapgenLegacyPtrForget(theme);
	free(theme);
	return out;
}

static const char* nox_test_mapgen_decor_name_51f9f0(
	const nox_test_mapgen_decor_51f9f0_result* out, int index) {
	switch (index) {
	case 0:
		return out->template_name;
	case 1:
		return out->backdrop_name;
	case 2:
		return out->wall_name;
	case 3:
		return out->floor_name;
	case 4:
		return out->door_name;
	default:
		return out->double_door_name;
	}
}
*/
import "C"

import "unsafe"

type mapgenDecorResult51F9F0 struct {
	templateParsed           bool
	backdropParsed           bool
	settingsValid            bool
	selected                 bool
	cleanupCleared           bool
	themeAddress             uintptr
	templateAddress          uintptr
	backdropAddress          uintptr
	templateWallAddress      uintptr
	backdropWallAddress      uintptr
	templateSetAddress       uintptr
	backdropSetAddress       uintptr
	selectedAddress          uintptr
	themeToken               uint32
	templateToken            uint32
	backdropToken            uint32
	templateWallToken        uint32
	backdropWallToken        uint32
	templateSetToken         uint32
	backdropSetToken         uint32
	selectedToken            uint32
	templateWallCount        uint32
	backdropWallCount        uint32
	templateSetCount         uint32
	backdropSetCount         uint32
	templateSetSharesEntries bool
	backdropSetSharesEntries bool
	frequency                uint32
	minRoomSize              int32
	maxRoomSize              int32
	templateName             string
	backdropName             string
	wallName                 string
	floorName                string
	doorName                 string
	doubleDoorName           string
}

func mapgenDecorFixture51F9F0(templatePath, backdropPath string) mapgenDecorResult51F9F0 {
	ctemplatePath := C.CString(templatePath)
	defer C.free(unsafe.Pointer(ctemplatePath))
	cbackdropPath := C.CString(backdropPath)
	defer C.free(unsafe.Pointer(cbackdropPath))
	value := C.nox_test_mapgen_decor_51f9f0(ctemplatePath, cbackdropPath)
	names := [6]string{}
	for i := range names {
		names[i] = C.GoString(C.nox_test_mapgen_decor_name_51f9f0(&value, C.int(i)))
	}
	return mapgenDecorResult51F9F0{
		templateParsed:           value.template_parsed != 0,
		backdropParsed:           value.backdrop_parsed != 0,
		settingsValid:            value.settings_valid != 0,
		selected:                 value.selected != 0,
		cleanupCleared:           value.cleanup_cleared != 0,
		themeAddress:             uintptr(value.theme_address),
		templateAddress:          uintptr(value.template_address),
		backdropAddress:          uintptr(value.backdrop_address),
		templateWallAddress:      uintptr(value.template_wall_address),
		backdropWallAddress:      uintptr(value.backdrop_wall_address),
		templateSetAddress:       uintptr(value.template_set_address),
		backdropSetAddress:       uintptr(value.backdrop_set_address),
		selectedAddress:          uintptr(value.selected_address),
		themeToken:               uint32(value.theme_token),
		templateToken:            uint32(value.template_token),
		backdropToken:            uint32(value.backdrop_token),
		templateWallToken:        uint32(value.template_wall_token),
		backdropWallToken:        uint32(value.backdrop_wall_token),
		templateSetToken:         uint32(value.template_set_token),
		backdropSetToken:         uint32(value.backdrop_set_token),
		selectedToken:            uint32(value.selected_token),
		templateWallCount:        uint32(value.template_wall_count),
		backdropWallCount:        uint32(value.backdrop_wall_count),
		templateSetCount:         uint32(value.template_set_count),
		backdropSetCount:         uint32(value.backdrop_set_count),
		templateSetSharesEntries: value.template_set_shares_entries != 0,
		backdropSetSharesEntries: value.backdrop_set_shares_entries != 0,
		frequency:                uint32(value.frequency),
		minRoomSize:              int32(value.min_room_size),
		maxRoomSize:              int32(value.max_room_size),
		templateName:             names[0],
		backdropName:             names[1],
		wallName:                 names[2],
		floorName:                names[3],
		doorName:                 names[4],
		doubleDoorName:           names[5],
	}
}
