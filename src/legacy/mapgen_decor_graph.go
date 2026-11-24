package legacy

/*
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
*/
import "C"

import "unsafe"

func mapgenResolveDecorChoice(ref C.uintptr_t) *mapGenChoiceRecord {
	return (*mapGenChoiceRecord)(C.nox_mapgenLegacyPtrResolve(ref))
}

//export nox_mapgenChoiceNextNative_520380
func nox_mapgenChoiceNextNative_520380(ref C.uintptr_t) *C.uint32_t {
	choice := mapgenResolveDecorChoice(ref)
	if choice == nil {
		return nil
	}
	return (*C.uint32_t)(unsafe.Pointer(mapgenResolveDecorChoice(C.uintptr_t(choice.next))))
}

func mapgenFreeChoiceList(ref C.uint32_t) {
	for node := mapgenResolveDecorChoice(C.uintptr_t(ref)); node != nil; {
		next := mapgenResolveDecorChoice(C.uintptr_t(node.next))
		mapgenReleaseDecorRecord(unsafe.Pointer(node))
		node = next
	}
}

func mapgenFreeForeachList(ref C.uint32_t) {
	for node := mapGenForeachLookup(uintptr(ref)); node != nil; {
		next := mapGenForeachLookup(uintptr(node.next))
		mapgenFreeChoiceList(C.uint32_t(node.choices))
		mapgenReleaseDecorRecord(unsafe.Pointer(node))
		node = next
	}
}

func mapgenFreeDecorFillList(ref C.uint32_t) {
	for node := mapgenResolveDecorFill(ref); node != nil; {
		next := mapgenResolveDecorFill(node.next_token)
		mapgenReleaseDecorRecord(unsafe.Pointer(node))
		node = next
	}
}

func mapgenResolveWallFloor(token C.uint32_t) *C.nox_mapgen_wall_floor_legacy {
	return (*C.nox_mapgen_wall_floor_legacy)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
}

func mapgenFreeWallFloorList(ref C.uint32_t) {
	for node := mapgenResolveWallFloor(ref); node != nil; {
		next := mapgenResolveWallFloor(node.next_token)
		mapgenFreeDecorFillList(node.fills_token)
		mapgenReleaseDecorRecord(unsafe.Pointer(node))
		node = next
	}
}

func mapgenResolveDecorEntry(token C.uint32_t) *C.nox_mapgen_decor_entry_legacy {
	return (*C.nox_mapgen_decor_entry_legacy)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
}

func mapgenFreeDecorEntryList(ref C.uint32_t) {
	for node := mapgenResolveDecorEntry(ref); node != nil; {
		next := mapgenResolveDecorEntry(node.next_token)
		mapgenFreeChoiceList(node.choices_token)
		mapgenFreeForeachList(node.foreach_token)
		mapgenReleaseDecorRecord(unsafe.Pointer(node))
		node = next
	}
}

func mapgenResolveDecorSet(token C.uint32_t) *C.nox_mapgen_decor_set_legacy {
	return (*C.nox_mapgen_decor_set_legacy)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
}

func mapgenFreeDecorSetList(ref C.uint32_t) {
	for node := mapgenResolveDecorSet(ref); node != nil; {
		next := mapgenResolveDecorSet(node.next_token)
		if node.shares_entries == 0 {
			mapgenFreeDecorEntryList(node.entries_token)
		}
		mapgenReleaseDecorRecord(unsafe.Pointer(node))
		node = next
	}
}

func mapgenFreeDecorDefinitionList(ref C.uint32_t) {
	for node := mapgenResolveDecorDefinition(ref); node != nil; {
		next := mapgenResolveDecorDefinition(node.next_token)
		mapgenFreeWallFloorList(node.wall_floors_token)
		mapgenFreeDecorSetList(node.decor_sets_token)
		mapgenReleaseDecorRecord(unsafe.Pointer(node))
		node = next
	}
}

func mapgenResolveItemSetEntry(token C.uint32_t) *C.nox_mapgen_item_set_entry_legacy {
	return (*C.nox_mapgen_item_set_entry_legacy)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
}

//export sub_520D50
func sub_520D50(raw *C.uint32_t) *C.uint32_t {
	if raw == nil {
		return nil
	}
	theme := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(raw))
	// COPY shares entries with earlier definitions and with templates.
	for _, index := range [...]int{22, 30, 46, 38} {
		mapgenFreeDecorDefinitionList(theme[index])
		theme[index], theme[index+1] = 0, 0
	}
	for _, index := range [...]int{275, 277} {
		for item := mapgenResolveItemSetEntry(theme[index]); item != nil; {
			next := mapgenResolveItemSetEntry(item.next_token)
			C.nox_xxx_mapGenFreeStr_51F1F0(unsafe.Pointer(item))
			item = next
		}
		theme[index], theme[index+1] = 0, 0
	}
	resolvePrefab := func(token C.uint32_t) *[40]C.uint32_t {
		return (*[40]C.uint32_t)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
	}
	for prefab := resolvePrefab(theme[20]); prefab != nil; {
		next := resolvePrefab(prefab[39])
		mapgenFreeForeachList(prefab[38])
		mapgenReleaseDecorRecord(unsafe.Pointer(prefab))
		prefab = next
	}
	theme[20], theme[21] = 0, 0
	return nil
}
