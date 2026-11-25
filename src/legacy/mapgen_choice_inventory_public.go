package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// The choice and foreach lists are stored in the original fixed-width format.
// The list links are registry tokens, and object links use native pointers.
type mapGenChoiceEntry struct {
	kind uint32
	name [60]byte
}

type mapGenChoiceRecord struct {
	weight int32
	items  [32]mapGenChoiceEntry
	count  int32
	next   uint32
}

type mapGenForeachRecord struct {
	typeID  uint32
	choices uint32
	next    uint32
}

var (
	_ [2060]byte = [unsafe.Sizeof(mapGenChoiceRecord{})]byte{}
	_ [12]byte   = [unsafe.Sizeof(mapGenForeachRecord{})]byte{}
)

type mapGenChoiceSetup struct {
	weaponRef   uint32
	weaponCount int32
	armorRef    uint32
	armorCount  int32
}

type mapGenChoiceServices struct {
	roll      func(int32, int32) int32
	next      func(uint32) *mapGenChoiceRecord
	newObject func(*byte) *server.Object
	enchant   func(*byte, uint32, int32) *server.Object
	spellbook func(*byte) *server.Object
	attach    func(*server.Object, *server.Object)
}

func mapGenAttachInventory522300(holder, item *server.Object) *server.Object {
	item.Field125 = nil
	item.InvNextItem = holder.InvFirstItem
	if head := holder.InvFirstItem; head != nil {
		head.Field125 = item
	}
	holder.InvFirstItem = item
	item.InvHolder = holder
	return item
}

func mapGenApplyChoice521FE0(setup *mapGenChoiceSetup, holder *server.Object, choice *mapGenChoiceRecord, services mapGenChoiceServices) {
	if setup == nil || holder == nil || choice == nil {
		return
	}
	roll := services.roll(1, 100)
	for choice != nil {
		roll -= choice.weight
		if roll <= 0 {
			break
		}
		choice = services.next(choice.next)
	}
	if choice == nil {
		return
	}
	for i := int32(0); i < choice.count && i < int32(len(choice.items)); i++ {
		entry := &choice.items[i]
		name := &entry.name[0]
		var item *server.Object
		// Room placement directives have no inventory item.
		switch entry.kind {
		case 0:
			item = services.newObject(name)
		case 3:
			item = services.enchant(name, setup.weaponRef, setup.weaponCount)
		case 4:
			item = services.enchant(name, setup.armorRef, setup.armorCount)
		case 5:
			item = services.spellbook(name)
		}
		if item != nil {
			services.attach(holder, item)
		}
	}
}

func mapGenFindForeach521C60(ref uintptr, typeID uint32, lookup func(uintptr) *mapGenForeachRecord) *mapGenForeachRecord {
	for entry := lookup(ref); entry != nil; entry = lookup(uintptr(entry.next)) {
		if entry.typeID == typeID {
			return entry
		}
	}
	return nil
}

func mapGenApplyForeach521C60(
	first func() *server.Object,
	next func(*server.Object) *server.Object,
	lookup func(uintptr) *mapGenForeachRecord,
	apply func(*server.Object, uint32),
	foreachRef uintptr,
) {
	for object := first(); object != nil; object = next(object) {
		if entry := mapGenFindForeach521C60(foreachRef, uint32(object.TypeInd), lookup); entry != nil {
			apply(object, entry.choices)
		}
	}
}

//export sub_522300
func sub_522300(holder, item *C.nox_object_t) *C.nox_object_t {
	return asObjectC(mapGenAttachInventory522300(asObjectS(holder), asObjectS(item)))
}

//export nox_xxx_mapgen_521FE0
func nox_xxx_mapgen_521FE0(theme *C.uint8_t, holder *C.nox_object_t, choices C.uintptr_t) {
	var setup *mapGenChoiceSetup
	if theme != nil {
		setup = (*mapGenChoiceSetup)(unsafe.Add(unsafe.Pointer(theme), 1100))
	}
	choice := mapgenResolveDecorChoice(choices)
	mapGenApplyChoice521FE0(setup, asObjectS(holder), choice, mapGenChoiceServices{
		roll: func(low, high int32) int32 {
			return int32(nox_xxx_mapGenRandFunc_526AC0(C.int(low), C.int(high)))
		},
		next: func(next uint32) *mapGenChoiceRecord {
			return mapgenResolveDecorChoice(C.uintptr_t(next))
		},
		newObject: func(name *byte) *server.Object {
			return asObjectS(nox_xxx_newObjectByTypeID_4E3810((*C.char)(unsafe.Pointer(name))))
		},
		enchant: func(name *byte, ref uint32, count int32) *server.Object {
			return asObjectS(nox_mapgenMakeEnchantedItemNative_5221A0((*C.char)(unsafe.Pointer(name)), C.uintptr_t(ref), C.int(count)))
		},
		spellbook: func(name *byte) *server.Object {
			return asObjectS(nox_xxx_mapGenMakeSpellbook_5220E0(theme, (*C.nox_mapgen_spell_name_t)(unsafe.Pointer(name))))
		},
		attach: func(holder, item *server.Object) {
			mapGenAttachInventory522300(holder, item)
		},
	})
}

func mapGenForeachLookup(ref uintptr) *mapGenForeachRecord {
	return (*mapGenForeachRecord)(resolveMapgenPointer(uintptr(ref)))
}

//export nox_mapgenFindForeachChoicesNative_521C60
func nox_mapgenFindForeachChoicesNative_521C60(foreachRef C.uintptr_t, typeID C.uint16_t) *C.uint32_t {
	entry := mapGenFindForeach521C60(uintptr(foreachRef), uint32(typeID), mapGenForeachLookup)
	if entry == nil {
		return nil
	}
	return (*C.uint32_t)(resolveMapgenPointer(uintptr(entry.choices)))
}

//export nox_mapgenApplyForeachNative_521C60
func nox_mapgenApplyForeachNative_521C60(theme *C.uint8_t, foreachRef C.uintptr_t) {
	nox_xxx_mapgen_521C60(theme, foreachRef)
}

//export nox_xxx_mapgen_521C60
func nox_xxx_mapgen_521C60(theme *C.uint8_t, foreachRef C.uintptr_t) C.int {
	mapGenApplyForeach521C60(
		func() *server.Object { return asObjectS(sub_504980()) },
		func(object *server.Object) *server.Object { return asObjectS(sub_5049C0(asObjectC(object))) },
		mapGenForeachLookup,
		func(object *server.Object, choice uint32) {
			nox_xxx_mapgen_521FE0(theme, asObjectC(object), C.uintptr_t(choice))
		},
		uintptr(foreachRef),
	)
	return 0
}
