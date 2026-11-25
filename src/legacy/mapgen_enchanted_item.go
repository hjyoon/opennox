package legacy

/*
#include "GAME4_2.h"
#include "noxstring.h"

typedef const nox_modifier_attrs_t object_field_const_attrs_t;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func mapgenSelectEnchantedItem(name *C.char, ref C.uintptr_t, count C.int) (*C.char, *C.nox_mapgen_item_set_entry_legacy) {
	entry := (*C.nox_mapgen_item_set_entry_legacy)(resolveMapgenPointer(uintptr(ref)))
	if GoString(name) == "*" {
		if entry == nil || count <= 0 {
			return nil, nil
		}
		index := nox_xxx_mapGenRandFunc_526AC0(0, count-1)
		for ; entry != nil && index > 0; index-- {
			entry = mapgenResolveItemSetEntry(entry.next_token)
		}
		if entry == nil {
			return nil, nil
		}
		return &entry.object_name[0], entry
	}
	for ; entry != nil; entry = mapgenResolveItemSetEntry(entry.next_token) {
		if C.nox_strcmpi(&entry.lookup_name[0], name) == 0 {
			return &entry.object_name[0], entry
		}
	}
	return name, nil
}

func mapgenEnchantedItemAttrs(entry *C.nox_mapgen_item_set_entry_legacy) server.ModifierInitData {
	var attrs server.ModifierInitData
	for index, count := range entry.modifier_counts {
		if count == 0 {
			continue
		}
		roll := nox_xxx_mapGenRandFunc_526AC0(1, 100)
		if uint32(roll) > memmap.Uint32(0x587000, uintptr(254688+4*index)) {
			continue
		}
		list := resolveMapgenPointer(uintptr(entry.modifier_tokens[index]))
		if list == nil {
			continue
		}
		selected := nox_xxx_mapGenRandFunc_526AC0(0, C.int(count-1))
		modifier := (*C.char)(unsafe.Add(list, int(selected)*len(mapgenItemModifierName{})))
		if C.nox_strcmpi(internCStr("none"), modifier) != 0 {
			id := nox_xxx_modifGetIdByName_413290(modifier)
			attrs.Modifiers[index] = (*server.ModifierEff)(nox_xxx_modifGetDescById_413330(id))
		}
	}
	if attrs.Modifiers[2] == nil {
		attrs.Modifiers[3] = nil
	}
	return attrs
}

//export nox_mapgenMakeEnchantedItemNative_5221A0
func nox_mapgenMakeEnchantedItemNative_5221A0(name *C.char, ref C.uintptr_t, count C.int) *C.nox_object_t {
	if name == nil {
		return nil
	}
	typeName, entry := mapgenSelectEnchantedItem(name, ref, count)
	if typeName == nil {
		return nil
	}
	object := nox_xxx_newObjectByTypeID_4E3810(typeName)
	if object != nil && entry != nil {
		attrs := mapgenEnchantedItemAttrs(entry)
		nox_xxx_modifSetItemAttrs_4E4990(object, (*C.object_field_const_attrs_t)(unsafe.Pointer(&attrs)))
	}
	return object
}

//export nox_xxx_mapGenMakeEnchantedItem_5221A0
func nox_xxx_mapGenMakeEnchantedItem_5221A0(name, list *C.char, count C.int) *C.uint32_t {
	item := nox_mapgenMakeEnchantedItemNative_5221A0(name, C.uintptr_t(uintptr(unsafe.Pointer(list))), count)
	return (*C.uint32_t)(unsafe.Pointer(item))
}
