package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

func mapgenDecorSetTailLink(def *C.nox_mapgen_decor_definition_legacy) *C.uint32_t {
	link := &def.decor_sets_token
	for *link != 0 {
		link = &mapgenResolveDecorSet(*link).next_token
	}
	return link
}

func mapgenReadDecorEntry(entry *C.nox_mapgen_decor_entry_legacy, file *C.FILE) bool {
	switch entry._type {
	case 0, 1, 3, 4, 5:
		if !mapgenReadDecorName(file, &entry.name) {
			return false
		}
	}
	if entry._type > 1 {
		return true
	}
	scratch := mapgenThemeTokenBuffer()
	text := (*C.char)(unsafe.Pointer(scratch))
	read := func() bool { return nox_xxx_mapGenReadLine_51E540(file, scratch) != 0 }
	if !read() {
		return false
	}
	if mapgenUppercaseToken(GoString(text)) != "DENSITY" {
		entry.min_count = C.int32_t(C.atoi(text))
		if !read() {
			return false
		}
		entry.max_count = C.int32_t(C.atoi(text))
		return true
	}

	entry.density_enabled = 1
	if !read() {
		return false
	}
	entry.density = C.float(C.atof(text))
	if !read() {
		return false
	}
	if GoString(text) != "*" {
		entry.min_count = C.int32_t(C.atoi(text))
	}
	if !read() {
		return false
	}
	if GoString(text) == "*" {
		entry.max_count = 999999
	} else {
		entry.max_count = C.int32_t(C.atoi(text))
	}
	return true
}

func mapgenReadDecorSet(definition *C.nox_mapgen_decor_definition_legacy, file *C.FILE) C.int {
	if definition == nil || file == nil {
		return 0
	}
	record, token := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_decor_set_legacy)
	if record == nil {
		return 0
	}
	set := (*C.nox_mapgen_decor_set_legacy)(record)
	published := false
	defer func() {
		if !published {
			mapgenFreeDecorSetList(token)
		}
	}()
	scratch := mapgenThemeTokenBuffer()
	text := (*C.char)(unsafe.Pointer(scratch))
	read := func() bool { return nox_xxx_mapGenReadLine_51E540(file, scratch) != 0 }
	if !read() {
		return 0
	}
	set.min_count = C.int32_t(C.atoi(text))
	if !read() {
		return 0
	}
	set.max_count = C.int32_t(C.atoi(text))

	var current *C.nox_mapgen_decor_entry_legacy
	for read() {
		word := mapgenUppercaseToken(GoString(text))
		switch word {
		case "END":
			*mapgenDecorSetTailLink(definition) = token
			definition.decor_set_count++
			published = true
			return 1
		case "FOREACH":
			if current == nil || current._type != 1 {
				return 0
			}
			foreach, foreachToken := mapgenReadForeach(file)
			if foreach == nil {
				return 0
			}
			foreach.next = uint32(current.foreach_token)
			current.foreach_token = foreachToken
		case "CONTAINS":
			if current == nil || current._type != 0 {
				return 0
			}
			choices := mapgenReadChoices(file)
			if choices == 0 {
				return 0
			}
			mapgenFreeChoiceList(current.choices_token)
			current.choices_token = choices
		default:
			kind, ok := mapgenDecorKeywordIndex(mapgenDecorEntryNames, word)
			if !ok {
				return 0
			}
			record, entryToken := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_decor_entry_legacy)
			if record == nil {
				return 0
			}
			entry := (*C.nox_mapgen_decor_entry_legacy)(record)
			entry._type = C.uint32_t(kind)
			entry.next_token = set.entries_token
			set.entries_token = entryToken
			if !mapgenReadDecorEntry(entry, file) {
				return 0
			}
			set.entry_count++
			current = entry
		}
	}
	return 0
}

//export nox_xxx_genDecorReadDecorSet_51FFA0
func nox_xxx_genDecorReadDecorSet_51FFA0(token C.int, file *C.FILE) C.int {
	return mapgenReadDecorSet(mapgenResolveDecorDefinition(C.uint32_t(token)), file)
}
