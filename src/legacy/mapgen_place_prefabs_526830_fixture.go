package legacy

/*
#include <stdint.h>
#include <stdlib.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"

typedef struct nox_test_mapgen_place_prefabs_526830_result {
	uintptr_t theme_address;
	uintptr_t prefab_addresses[3];
	uintptr_t foreach_addresses[2];
	uintptr_t room_addresses[3];
	uintptr_t choice_addresses[2];
	uintptr_t callback_theme_addresses[2];
	uintptr_t callback_prefab_addresses[2];
	uintptr_t callback_foreach_addresses[2];
	uintptr_t callback_room_addresses[2];
	uintptr_t found_choice_address;
	uintptr_t next_choice_address;
	int result;
	int calls;
} nox_test_mapgen_place_prefabs_526830_result;

typedef struct nox_test_mapgen_place_prefabs_526830_buffers {
	uint8_t* theme;
	uint8_t* prefabs[3];
	uint8_t* rooms[3];
	uint32_t* foreach_entries[2];
	uint32_t* choices[2];
} nox_test_mapgen_place_prefabs_526830_buffers;

static nox_test_mapgen_place_prefabs_526830_result* nox_test_mapgen_place_prefabs_out_526830;

static void nox_test_mapgen_finalize_prefab_526830(
	uint8_t* theme, uint8_t* prefab, uint8_t* foreach_entries, uint8_t* room) {
	nox_test_mapgen_place_prefabs_526830_result* out = nox_test_mapgen_place_prefabs_out_526830;
	int call = out->calls++;
	if (call >= 2) {
		return;
	}
	out->callback_theme_addresses[call] = (uintptr_t)theme;
	out->callback_prefab_addresses[call] = (uintptr_t)prefab;
	out->callback_foreach_addresses[call] = (uintptr_t)foreach_entries;
	out->callback_room_addresses[call] = (uintptr_t)room;
}

static nox_test_mapgen_place_prefabs_526830_result nox_test_mapgen_place_prefabs_526830(
	const nox_test_mapgen_place_prefabs_526830_buffers* buffers) {
	nox_test_mapgen_place_prefabs_526830_result out = {0};
	uint8_t* theme = buffers->theme;
	uint8_t* const* prefabs = buffers->prefabs;
	uint8_t* const* rooms = buffers->rooms;
	uint32_t* const* foreach_entries = buffers->foreach_entries;
	uint32_t* const* choices = buffers->choices;
	if (!nox_mapgenLegacyPtrRegister(theme)) {
		return out;
	}
	out.theme_address = (uintptr_t)theme;

	for (int i = 0; i < 3; ++i) {
		if (!nox_mapgenLegacyPtrRegister(prefabs[i]) || !nox_mapgenLegacyPtrRegister(rooms[i])) {
			goto cleanup;
		}
		out.prefab_addresses[i] = (uintptr_t)prefabs[i];
		out.room_addresses[i] = (uintptr_t)rooms[i];
		*(uint32_t*)(prefabs[i] + 148) = nox_mapgenLegacyPtrRegister(rooms[i]);
		if (i == 0) {
			*(uint32_t*)(theme + 80) = nox_mapgenLegacyPtrRegister(prefabs[i]);
		} else {
			*(uint32_t*)(prefabs[i - 1] + 156) = nox_mapgenLegacyPtrRegister(prefabs[i]);
		}
	}
	*(uint32_t*)(prefabs[0] + 76) = 1;
	*(uint32_t*)(prefabs[1] + 76) = 0;
	*(uint32_t*)(prefabs[2] + 76) = 1;

	for (int i = 0; i < 2; ++i) {
		if (!nox_mapgenLegacyPtrRegister(foreach_entries[i]) || !nox_mapgenLegacyPtrRegister(choices[i])) {
			goto cleanup;
		}
		out.foreach_addresses[i] = (uintptr_t)foreach_entries[i];
		out.choice_addresses[i] = (uintptr_t)choices[i];
		foreach_entries[i][0] = (uint32_t)(0x120 + i);
		foreach_entries[i][1] = nox_mapgenLegacyPtrRegister(choices[i]);
	}
	foreach_entries[0][2] = nox_mapgenLegacyPtrRegister(foreach_entries[1]);
	choices[0][514] = nox_mapgenLegacyPtrRegister(choices[1]);
	*(uint32_t*)(prefabs[0] + 152) = nox_mapgenLegacyPtrRegister(foreach_entries[0]);
	*(uint32_t*)(prefabs[2] + 152) = nox_mapgenLegacyPtrRegister(foreach_entries[1]);

	out.found_choice_address = (uintptr_t)nox_mapgenFindForeachChoicesNative_521C60(
		nox_mapgenLegacyPtrRegister(foreach_entries[0]), 0x121);
	out.next_choice_address = (uintptr_t)nox_mapgenChoiceNextNative_520380(
		(uintptr_t)choices[0]);
	nox_test_mapgen_place_prefabs_out_526830 = &out;
	out.result = nox_mapgenPlacePrefabsWithCallback_526830(
		theme, nox_test_mapgen_finalize_prefab_526830);
	nox_test_mapgen_place_prefabs_out_526830 = NULL;

cleanup:
	for (int i = 0; i < 2; ++i) {
		nox_mapgenLegacyPtrForget(choices[i]);
		nox_mapgenLegacyPtrForget(foreach_entries[i]);
	}
	for (int i = 0; i < 3; ++i) {
		nox_mapgenLegacyPtrForget(rooms[i]);
		nox_mapgenLegacyPtrForget(prefabs[i]);
	}
	nox_mapgenLegacyPtrForget(theme);
	return out;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type mapgenPlacePrefabsResult526830 struct {
	themeAddress             uintptr
	prefabAddresses          [3]uintptr
	foreachAddresses         [2]uintptr
	roomAddresses            [3]uintptr
	choiceAddresses          [2]uintptr
	callbackThemeAddresses   [2]uintptr
	callbackPrefabAddresses  [2]uintptr
	callbackForeachAddresses [2]uintptr
	callbackRoomAddresses    [2]uintptr
	foundChoiceAddress       uintptr
	nextChoiceAddress        uintptr
	result                   bool
	calls                    int
}

func mapgenPlacePrefabsFixture526830() mapgenPlacePrefabsResult526830 {
	var releases []func()
	defer func() {
		for _, release := range releases {
			release()
		}
	}()
	allocate := func(size uintptr) unsafe.Pointer {
		pointer, release := alloc.MallocHigh(size)
		releases = append(releases, release)
		return pointer
	}
	var buffers C.nox_test_mapgen_place_prefabs_526830_buffers
	buffers.theme = (*C.uint8_t)(allocate(0x45c))
	for i := range buffers.prefabs {
		buffers.prefabs[i] = (*C.uint8_t)(allocate(160))
		buffers.rooms[i] = (*C.uint8_t)(allocate(64))
	}
	for i := range buffers.foreach_entries {
		buffers.foreach_entries[i] = (*C.uint32_t)(allocate(12))
		buffers.choices[i] = (*C.uint32_t)(allocate(0x80c))
	}
	got := C.nox_test_mapgen_place_prefabs_526830(&buffers)
	out := mapgenPlacePrefabsResult526830{
		themeAddress:       uintptr(got.theme_address),
		foundChoiceAddress: uintptr(got.found_choice_address),
		nextChoiceAddress:  uintptr(got.next_choice_address),
		result:             got.result != 0,
		calls:              int(got.calls),
	}
	for i := range out.prefabAddresses {
		out.prefabAddresses[i] = uintptr(got.prefab_addresses[i])
		out.roomAddresses[i] = uintptr(got.room_addresses[i])
	}
	for i := range out.foreachAddresses {
		out.foreachAddresses[i] = uintptr(got.foreach_addresses[i])
		out.choiceAddresses[i] = uintptr(got.choice_addresses[i])
		out.callbackThemeAddresses[i] = uintptr(got.callback_theme_addresses[i])
		out.callbackPrefabAddresses[i] = uintptr(got.callback_prefab_addresses[i])
		out.callbackForeachAddresses[i] = uintptr(got.callback_foreach_addresses[i])
		out.callbackRoomAddresses[i] = uintptr(got.callback_room_addresses[i])
	}
	return out
}
