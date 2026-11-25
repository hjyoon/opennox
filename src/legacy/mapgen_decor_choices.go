package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

func mapgenReadChoiceItems(file *C.FILE, choice *mapGenChoiceRecord) string {
	scratch := mapgenThemeTokenBuffer()
	for nox_xxx_mapGenReadLine_51E540(file, scratch) != 0 {
		word := mapgenUppercaseToken(GoStringP(unsafe.Pointer(scratch)))
		if word == "END" || word == "OR" {
			return word
		}
		kind, ok := mapgenDecorKeywordIndex(mapgenDecorEntryNames, word)
		if !ok {
			return ""
		}
		entry := &choice.items[choice.count]
		entry.kind = uint32(kind)
		if mapgenReadFixedName(file, (*[60]C.char)(unsafe.Pointer(&entry.name))) == 0 {
			return ""
		}
		// The last slot remains scratch after 31 stored choices.
		if choice.count < int32(len(choice.items)-1) {
			choice.count++
		}
	}
	return ""
}

func mapgenReadChoices(file *C.FILE) (result C.uint32_t) {
	if file == nil {
		return 0
	}
	var head C.uint32_t
	var tail *mapGenChoiceRecord
	defer func() {
		if result == 0 {
			mapgenFreeChoiceList(head)
		}
	}()
	scratch := mapgenThemeTokenBuffer()
	text := (*C.char)(unsafe.Pointer(scratch))
	remaining, wildcards := int32(100), int32(0)
	for {
		record, token := mapgenAllocateDecorRecord(C.size_t(unsafe.Sizeof(mapGenChoiceRecord{})))
		if record == nil {
			return 0
		}
		choice := (*mapGenChoiceRecord)(record)
		if tail == nil {
			head = token
		} else {
			tail.next = uint32(token)
		}
		tail = choice
		if nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 {
			return 0
		}
		if GoString(text) == "*" {
			choice.weight = -1
			wildcards++
		} else {
			choice.weight = int32(parseLegacyDecimal32(text))
			remaining -= choice.weight
		}
		switch mapgenReadChoiceItems(file, choice) {
		case "OR":
			continue
		case "END":
			if wildcards != 0 {
				for cur := mapgenResolveDecorChoice(C.uintptr_t(head)); cur != nil; cur = mapgenResolveDecorChoice(C.uintptr_t(cur.next)) {
					if cur.weight < 0 {
						cur.weight = remaining/wildcards + 1
					}
				}
			}
			return head
		default:
			return 0
		}
	}
}

func mapgenReadForeach(file *C.FILE) (*mapGenForeachRecord, C.uint32_t) {
	scratch := mapgenThemeTokenBuffer()
	if file == nil || nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 {
		return nil, 0
	}
	record, token := mapgenAllocateDecorRecord(C.size_t(unsafe.Sizeof(mapGenForeachRecord{})))
	if record == nil {
		return nil, 0
	}
	entry := (*mapGenForeachRecord)(record)
	text := (*C.char)(unsafe.Pointer(scratch))
	entry.typeID = uint32(nox_xxx_getNameId_4E3AA0(text))
	if nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 || mapgenUppercaseToken(GoString(text)) != "CONTAINS" {
		mapgenReleaseDecorRecord(record)
		return nil, 0
	}
	entry.choices = uint32(mapgenReadChoices(file))
	if entry.choices == 0 {
		mapgenReleaseDecorRecord(record)
		return nil, 0
	}
	return entry, token
}

//export nox_xxx_gen_520380
func nox_xxx_gen_520380(file *C.FILE) *C.uint32_t {
	head := mapgenReadChoices(file)
	return (*C.uint32_t)(unsafe.Pointer(mapgenResolveDecorChoice(C.uintptr_t(head))))
}

//export nox_xxx_gen_5205B0
func nox_xxx_gen_5205B0(file *C.FILE) *C.uint32_t {
	entry, _ := mapgenReadForeach(file)
	return (*C.uint32_t)(unsafe.Pointer(entry))
}
