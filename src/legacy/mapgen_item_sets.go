package legacy

/*
#include "GAME4_2.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const mapgenMaxItemModifiers = 256

type mapgenItemModifierName [60]byte

func mapgenResolveItemSetEntry(token C.uint32_t) *C.nox_mapgen_item_set_entry_legacy {
	return (*C.nox_mapgen_item_set_entry_legacy)(resolveMapgenPointer(uintptr(token)))
}

func mapgenFreeItemSetModifiers(entry *C.nox_mapgen_item_set_entry_legacy) {
	for i, token := range entry.modifier_tokens {
		if ptr := resolveMapgenPointer(uintptr(token)); ptr != nil {
			mapgenReleaseDecorRecord(ptr)
		}
		entry.modifier_tokens[i], entry.modifier_counts[i] = 0, 0
	}
}

func mapgenFreeItemSetEntry(entry *C.nox_mapgen_item_set_entry_legacy) {
	if entry == nil {
		return
	}
	mapgenFreeItemSetModifiers(entry)
	mapgenReleaseDecorRecord(unsafe.Pointer(entry))
}

//export nox_xxx_mapGenFreeStr_51F1F0
func nox_xxx_mapGenFreeStr_51F1F0(ref unsafe.Pointer) {
	ptr := resolveMapgenPointer(uintptr(ref))
	mapgenFreeItemSetEntry((*C.nox_mapgen_item_set_entry_legacy)(ptr))
}

func mapgenClearItemSetTemplate() {
	entry := mapgenResolveItemSetEntry(archiveAndMapgenStorage.item_set_template)
	archiveAndMapgenStorage.item_set_template = 0
	mapgenFreeItemSetEntry(entry)
}

func mapgenItemModifierSlot(word string) int {
	switch word {
	case "EFFECTIVENESS", "QUALITY":
		return 0
	case "MATERIAL":
		return 1
	case "PRIMARY_ENCHANTMENT":
		return 2
	case "SECONDARY_ENCHANTMENT":
		return 3
	default:
		return -1
	}
}

func mapgenMergeItemModifiers(base *C.nox_mapgen_item_set_entry_legacy, slot int, changes []mapgenItemModifierName) ([]mapgenItemModifierName, bool) {
	var merged []mapgenItemModifierName
	if base != nil {
		n := uint32(base.modifier_counts[slot])
		values := (*mapgenItemModifierName)(resolveMapgenPointer(uintptr(base.modifier_tokens[slot])))
		if n > mapgenMaxItemModifiers || n != 0 && values == nil {
			return nil, false
		}
		merged = append(merged, unsafe.Slice(values, int(n))...)
		for i := range merged {
			merged[i][len(merged[i])-1] = 0
		}
	}

	for _, change := range changes {
		remove := change[0] == '-'
		match := change[:]
		if remove {
			match = match[1:]
		}
		word := mapgenUppercaseToken(alloc.GoStringS(match))
		found := -1
		for i := range merged {
			if mapgenUppercaseToken(alloc.GoStringS(merged[i][:])) == word {
				found = i
				break
			}
		}
		if remove {
			if found >= 0 {
				copy(merged[found:], merged[found+1:])
				merged = merged[:len(merged)-1]
			}
		} else if found < 0 {
			if len(merged) >= mapgenMaxItemModifiers {
				return nil, false
			}
			merged = append(merged, change)
		}
	}
	return merged, true
}

func mapgenInstallItemModifiers(entry *C.nox_mapgen_item_set_entry_legacy, parsed *[4][]mapgenItemModifierName) bool {
	base := mapgenResolveItemSetEntry(archiveAndMapgenStorage.item_set_template)
	var replacement C.nox_mapgen_item_set_entry_legacy
	for i, changes := range parsed {
		names, ok := mapgenMergeItemModifiers(base, i, changes)
		if !ok {
			mapgenFreeItemSetModifiers(&replacement)
			return false
		}
		if len(names) == 0 {
			continue
		}
		size := C.size_t(len(names)) * C.size_t(unsafe.Sizeof(mapgenItemModifierName{}))
		ptr, token := mapgenAllocateDecorRecord(size)
		if ptr == nil {
			mapgenFreeItemSetModifiers(&replacement)
			return false
		}
		copy(unsafe.Slice((*mapgenItemModifierName)(ptr), len(names)), names)
		replacement.modifier_tokens[i] = token
		replacement.modifier_counts[i] = C.uint32_t(len(names))
	}
	mapgenFreeItemSetModifiers(entry)
	entry.modifier_tokens, entry.modifier_counts = replacement.modifier_tokens, replacement.modifier_counts
	return true
}

func mapgenReadItemModifiers(entry *C.nox_mapgen_item_set_entry_legacy, file *C.FILE) C.int {
	if entry == nil || file == nil {
		return 0
	}
	var parsed [4][]mapgenItemModifierName
	scratch := mapgenThemeTokenBuffer()
	text := (*C.char)(unsafe.Pointer(scratch))
	read := func() bool { return nox_xxx_mapGenReadLine_51E540(file, scratch) != 0 }
	for read() {
		section := mapgenUppercaseToken(GoString(text))
		if section == "END" {
			return C.int(bool2int(mapgenInstallItemModifiers(entry, &parsed)))
		}
		slot := mapgenItemModifierSlot(section)
		if slot < 0 {
			return 0
		}
		for {
			if !read() {
				return 0
			}
			word := GoString(text)
			if mapgenUppercaseToken(word) == "END" {
				break
			}
			var raw mapgenItemModifierName
			if len(parsed[slot]) >= mapgenMaxItemModifiers || !copyMapgenName(raw[:], word) {
				return 0
			}
			parsed[slot] = append(parsed[slot], raw)
		}
	}
	return 0
}

//export sub_51F230
func sub_51F230(raw *C.uint8_t, file *C.FILE) C.int {
	return mapgenReadItemModifiers((*C.nox_mapgen_item_set_entry_legacy)(unsafe.Pointer(raw)), file)
}

func mapgenReadItemSet(raw *C.uint8_t, file *C.FILE, section string, index int) C.int {
	if raw == nil || file == nil {
		return 0
	}
	theme := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(raw))
	mapgenClearItemSetTemplate()
	defer mapgenClearItemSetTemplate()
	scratch := mapgenThemeTokenBuffer()
	for nox_xxx_mapGenReadLine_51E540(file, scratch) != 0 {
		switch mapgenUppercaseToken(GoStringP(unsafe.Pointer(scratch))) {
		case "END":
			return 1
		case section:
		default:
			continue
		}
		record, token := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_item_set_entry_legacy)
		if record == nil {
			return 0
		}
		entry := (*C.nox_mapgen_item_set_entry_legacy)(record)
		if !mapgenReadDecorName(file, &entry.object_name) {
			mapgenFreeItemSetEntry(entry)
			return 0
		}
		isTemplate := mapgenUppercaseToken(GoString(&entry.object_name[0])) == "TEMPLATE"
		if (!isTemplate && !mapgenReadDecorName(file, &entry.lookup_name)) ||
			mapgenReadItemModifiers(entry, file) == 0 {
			mapgenFreeItemSetEntry(entry)
			return 0
		}

		if isTemplate {
			mapgenClearItemSetTemplate()
			archiveAndMapgenStorage.item_set_template = token
			continue
		}
		link := &theme[index]
		for *link != 0 {
			link = &mapgenResolveItemSetEntry(*link).next_token
		}
		*link = token
		theme[index+1]++
	}
	return 0
}

//export nox_xxx_genReadWeaponSet_51F030
func nox_xxx_genReadWeaponSet_51F030(raw *C.uint8_t, file *C.FILE) C.int {
	return mapgenReadItemSet(raw, file, "WEAPON", 275)
}

//export nox_xxx_genReadArmorSet_51F640
func nox_xxx_genReadArmorSet_51F640(raw *C.uint8_t, file *C.FILE) C.int {
	return mapgenReadItemSet(raw, file, "ARMOR", 277)
}
