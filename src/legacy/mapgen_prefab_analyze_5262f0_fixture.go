package legacy

/*
#include <stdint.h>

#include "GAME4_2.h"

typedef struct nox_test_mapgen_prefab_entry_5262f0_result {
	uintptr_t theme_address;
	uintptr_t prefab_address;
	int result;
	int null_theme_result;
	int null_prefab_result;
} nox_test_mapgen_prefab_entry_5262f0_result;

static nox_test_mapgen_prefab_entry_5262f0_result nox_test_mapgen_prefab_entry_5262f0(void) {
	uint8_t theme[20] = {0};
	uint8_t prefab[160] = {0};
	return (nox_test_mapgen_prefab_entry_5262f0_result){
		.theme_address = (uintptr_t)&theme[0],
		.prefab_address = (uintptr_t)&prefab[0],
		.result = sub_5262F0(&theme[0], &prefab[0]),
		.null_theme_result = sub_5262F0(NULL, &prefab[0]),
		.null_prefab_result = sub_5262F0(&theme[0], NULL),
	};
}
*/
import "C"

type mapgenPrefabEntryResult5262F0 struct {
	themeAddress     uintptr
	prefabAddress    uintptr
	result           bool
	nullThemeResult  bool
	nullPrefabResult bool
}

func mapgenPrefabEntryFixture5262F0() mapgenPrefabEntryResult5262F0 {
	result := C.nox_test_mapgen_prefab_entry_5262f0()
	return mapgenPrefabEntryResult5262F0{
		themeAddress:     uintptr(result.theme_address),
		prefabAddress:    uintptr(result.prefab_address),
		result:           result.result != 0,
		nullThemeResult:  result.null_theme_result != 0,
		nullPrefabResult: result.null_prefab_result != 0,
	}
}
