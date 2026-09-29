package legacy

/*
#include <stdint.h>
#include <stdlib.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
#include "memmap.h"

extern uint32_t dword_5d4594_2487656;

typedef struct nox_test_mapgen_initial_prefabs_525d20_result {
	uintptr_t theme_address;
	uintptr_t prefab_addresses[8];
	int result;
	int calls;
	int order[8];
	uint32_t placed_mask;
} nox_test_mapgen_initial_prefabs_525d20_result;

static uint8_t* nox_test_mapgen_prefabs_525d20[8];
static int nox_test_mapgen_calls_525d20;
static int nox_test_mapgen_order_525d20[8];
static int nox_test_mapgen_fail_at_525d20;

static int nox_test_mapgen_place_prefab_525d20(uint8_t* theme, uint8_t* prefab) {
	(void)theme;
	int call = ++nox_test_mapgen_calls_525d20;
	for (int i = 0; i < 8; ++i) {
		if (nox_test_mapgen_prefabs_525d20[i] == prefab) {
			nox_test_mapgen_order_525d20[call - 1] = i;
			break;
		}
	}
	return call != nox_test_mapgen_fail_at_525d20;
}

static nox_test_mapgen_initial_prefabs_525d20_result nox_test_mapgen_initial_prefabs_525d20(
	int count, uint32_t required_mask, int fail_at
) {
	nox_test_mapgen_initial_prefabs_525d20_result out = {0};
	if (count < 0 || count > 8) {
		return out;
	}
	for (int i = 0; i < 8; ++i) {
		nox_test_mapgen_prefabs_525d20[i] = NULL;
	}
	uint8_t* theme = (uint8_t*)calloc(1, 0x45C);
	if (!theme || !nox_mapgenLegacyPtrRegister(theme)) {
		free(theme);
		return out;
	}
	*(float*)(theme + 64) = 128.0f;
	*(uint32_t*)(theme + 84) = (uint32_t)count;

	for (int i = 0; i < count; ++i) {
		nox_test_mapgen_prefabs_525d20[i] = (uint8_t*)calloc(1, 160);
		if (!nox_test_mapgen_prefabs_525d20[i]) {
			count = i;
			break;
		}
		uint32_t token = nox_mapgenLegacyPtrRegister(nox_test_mapgen_prefabs_525d20[i]);
		if (!token) {
			free(nox_test_mapgen_prefabs_525d20[i]);
			nox_test_mapgen_prefabs_525d20[i] = NULL;
			count = i;
			break;
		}
		*(uint32_t*)(nox_test_mapgen_prefabs_525d20[i] + 72) = (required_mask >> i) & 1u;
		if (i > 0) {
			*(uint32_t*)(nox_test_mapgen_prefabs_525d20[i - 1] + 156) = token;
		} else {
			*(uint32_t*)(theme + 80) = token;
		}
	}

	nox_test_mapgen_calls_525d20 = 0;
	nox_test_mapgen_fail_at_525d20 = fail_at;
	for (int i = 0; i < 8; ++i) {
		nox_test_mapgen_order_525d20[i] = -1;
	}
	uint32_t old_north = dword_5d4594_2487656;
	uint32_t old_south = *getMemU32Ptr(0x5D4594, 2487660);
	uint32_t old_east = *getMemU32Ptr(0x5D4594, 2487664);
	uint32_t old_west = *getMemU32Ptr(0x5D4594, 2487668);
	dword_5d4594_2487656 = 1;
	*getMemU32Ptr(0x5D4594, 2487660) = 2;
	*getMemU32Ptr(0x5D4594, 2487664) = 3;
	*getMemU32Ptr(0x5D4594, 2487668) = 4;
	out.result = nox_mapgenInitialPrefabsWithCallback_525D20(
		theme, nox_test_mapgen_place_prefab_525d20);
	dword_5d4594_2487656 = old_north;
	*getMemU32Ptr(0x5D4594, 2487660) = old_south;
	*getMemU32Ptr(0x5D4594, 2487664) = old_east;
	*getMemU32Ptr(0x5D4594, 2487668) = old_west;
	out.theme_address = (uintptr_t)theme;
	out.calls = nox_test_mapgen_calls_525d20;
	for (int i = 0; i < 8; ++i) {
		out.order[i] = nox_test_mapgen_order_525d20[i];
		out.prefab_addresses[i] = (uintptr_t)nox_test_mapgen_prefabs_525d20[i];
		if (nox_test_mapgen_prefabs_525d20[i] &&
			*(uint32_t*)(nox_test_mapgen_prefabs_525d20[i] + 76)) {
			out.placed_mask |= 1u << i;
		}
	}

	for (int i = 0; i < count; ++i) {
		nox_mapgenLegacyPtrForget(nox_test_mapgen_prefabs_525d20[i]);
		free(nox_test_mapgen_prefabs_525d20[i]);
		nox_test_mapgen_prefabs_525d20[i] = NULL;
	}
	nox_mapgenLegacyPtrForget(theme);
	free(theme);
	return out;
}
*/
import "C"

type mapgenInitialPrefabsResult525D20 struct {
	themeAddress    uintptr
	prefabAddresses [8]uintptr
	result          bool
	calls           int
	order           [8]int
	placedMask      uint32
}

func mapgenInitialPrefabsFixture525D20(count int, requiredMask uint32, failAt int) mapgenInitialPrefabsResult525D20 {
	got := C.nox_test_mapgen_initial_prefabs_525d20(C.int(count), C.uint32_t(requiredMask), C.int(failAt))
	out := mapgenInitialPrefabsResult525D20{
		themeAddress: uintptr(got.theme_address),
		result:       got.result != 0,
		calls:        int(got.calls),
		placedMask:   uint32(got.placed_mask),
	}
	for i := range out.order {
		out.order[i] = int(got.order[i])
		out.prefabAddresses[i] = uintptr(got.prefab_addresses[i])
	}
	return out
}
