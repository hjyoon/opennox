package legacy

/*
#include "GAME4_2.h"
#include "mapgen_decor_records.h"
#include "mapgen_legacy_ptr.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

const (
	mapGenRoomDecorationTokenOffset            = 372
	mapGenThemeDecorModifierSetsOffset         = 1100
	mapGenDecorFreePointScale          float32 = 0.9
)

type mapGenDecorEntry = C.nox_mapgen_decor_entry_legacy
type mapGenDecorSet = C.nox_mapgen_decor_set_legacy
type mapGenDecorDefinition = C.nox_mapgen_decor_definition_legacy

func mapGenPopulateDecor522340(definition *mapGenDecorDefinition, resolve func(uint32) *mapGenDecorSet,
	populate func(*mapGenDecorSet)) {
	if definition == nil {
		return
	}
	for token := definition.decor_sets_token; token != 0; {
		set := resolve(uint32(token))
		if set == nil {
			return
		}
		populate(set)
		token = set.next_token
	}
}

func mapGenPopulateDecorSet522370(set *mapGenDecorSet, roll func(int32, int32) int32,
	resolve func(uint32) *mapGenDecorEntry, populate func(*mapGenDecorEntry)) {
	if set.entries_token == 0 {
		return
	}
	count := roll(int32(set.min_count), int32(set.max_count))
	count = min(count, int32(set.entry_count))
	headToken := set.entries_token
	head := resolve(uint32(headToken))
	if head == nil {
		return
	}
	tailToken, tail := headToken, head
	head.shuffle_next_token, head.shuffle_prev_token = 0, 0
	token := head.next_token
	for i := int32(1); token != 0 && i < int32(set.entry_count); i++ {
		entry := resolve(uint32(token))
		if entry == nil {
			break
		}
		nearTail := roll(1, 100) >= 50
		secondDraw := roll(1, 100) >= 50
		var before, after C.uint32_t
		switch {
		case !nearTail && !secondDraw:
			after = headToken
		case !nearTail:
			before, after = headToken, head.shuffle_next_token
		case !secondDraw:
			before, after = tail.shuffle_prev_token, tailToken
		default:
			before = tailToken
		}
		entry.shuffle_next_token, entry.shuffle_prev_token = after, before
		if before == 0 {
			headToken, head = token, entry
		} else if previous := resolve(uint32(before)); previous != nil {
			previous.shuffle_next_token = token
		}
		if after == 0 {
			tailToken, tail = token, entry
		} else if next := resolve(uint32(after)); next != nil {
			next.shuffle_prev_token = token
		}
		token = entry.next_token
	}
	for i := int32(0); i < count && head != nil; i++ {
		populate(head)
		head = resolve(uint32(head.shuffle_next_token))
	}
}

type mapGenDecorModifierSet struct {
	choices uint32
	count   int32
}

type mapGenDecorPlacement struct {
	roll      func(int32, int32) int32
	spawn     func(*mapGenDecorEntry) *server.Object
	give      func(*server.Object, uint32)
	lookup    func(*mapGenDecorEntry) int32
	tryObject func(*mapGenDecorEntry, int32) bool
	clear     func()
	find      func(float32, *types.Pointf) bool
	enchanted func(*mapGenDecorEntry, mapGenDecorModifierSet) *server.Object
	book      func(*mapGenDecorEntry) *server.Object
	move      func(*server.Object, *types.Pointf)
}

func mapGenPopulateDecorEntry5224B0(entry *mapGenDecorEntry, roomSize *[2]C.int32_t,
	modifiers *[2]mapGenDecorModifierSet, place mapGenDecorPlacement) {
	low, high := int32(entry.min_count), int32(entry.max_count)
	var count int32
	if entry.density_enabled == 0 {
		count = place.roll(low, high)
	} else {
		area := int32(roomSize[0]) * int32(roomSize[1])
		count = x87TruncQwordLow(mulFloat64TowardZero(float64(area), float64(entry.density)))
		if count < low {
			count = low
		} else if count > high {
			count = high
		}
	}
	kind := int(entry._type)
	switch kind {
	case 0:
		for range count {
			if unit := place.spawn(entry); unit != nil {
				place.give(unit, uint32(entry.choices_token))
			}
		}
	case 1:
		index := place.lookup(entry)
		if index < 0 {
			return
		}
		for range count {
			for range 3 {
				if place.tryObject(entry, index) {
					break
				}
			}
		}
	case 2:
		place.clear()
	case 3, 4, 5:
		var point types.Pointf
		if !place.find(mapGenDecorFreePointScale, &point) {
			return
		}
		var unit *server.Object
		if kind == 5 {
			unit = place.book(entry)
		} else {
			unit = place.enchanted(entry, modifiers[kind-3])
		}
		if unit != nil {
			place.move(unit, &point)
		}
	}
}

//export nox_mapgenPopulateDecorNative_522340
func nox_mapgenPopulateDecorNative_522340(theme, room *C.uint8_t) {
	if theme == nil || room == nil {
		return
	}
	roomPointer := unsafe.Pointer(room)
	token := *(*C.uint32_t)(unsafe.Add(roomPointer, mapGenRoomDecorationTokenOffset))
	definition := (*mapGenDecorDefinition)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
	roomSize := &(*mapgenRoomGeometry)(roomPointer).cellDimensions
	modifiers := (*[2]mapGenDecorModifierSet)(unsafe.Add(unsafe.Pointer(theme), mapGenThemeDecorModifierSetsOffset))
	roll := func(low, high int32) int32 {
		return int32(nox_xxx_mapGenRandFunc_526AC0(C.int(low), C.int(high)))
	}
	resolveEntry := func(token uint32) *mapGenDecorEntry {
		return (*mapGenDecorEntry)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
	}
	place := mapGenDecorPlacement{
		roll: roll,
		spawn: func(entry *mapGenDecorEntry) *server.Object {
			return asObjectS(nox_xxx_mapGenMakeMonsterInRoom_522810(room, &entry.name[0]))
		},
		give: func(unit *server.Object, token uint32) {
			nox_xxx_mapgen_521FE0(theme, asObjectC(unit), C.uintptr_t(token))
		},
		lookup: func(entry *mapGenDecorEntry) int32 {
			return mapgenFindAreaRecord(GoString(&entry.name[0]))
		},
		tryObject: func(entry *mapGenDecorEntry, index int32) bool {
			return nox_mapgenPlaceDecorObjectNative_521CB0(theme, room, entry, C.int(index)) != 0
		},
		clear: func() { C.nox_mapgenClearTransientOccupiedRectsNative_521C10(room) },
		find: func(scale float32, point *types.Pointf) bool {
			return sub_5226D0(room, C.float(scale), (*C.float2)(unsafe.Pointer(point))) != 0
		},
		enchanted: func(entry *mapGenDecorEntry, modifiers mapGenDecorModifierSet) *server.Object {
			return asObjectS(nox_mapgenMakeEnchantedItemNative_5221A0(&entry.name[0], C.uintptr_t(modifiers.choices), C.int(modifiers.count)))
		},
		book: func(entry *mapGenDecorEntry) *server.Object {
			name := (*C.nox_mapgen_spell_name_t)(unsafe.Pointer(&entry.name[0]))
			return asObjectS(nox_xxx_mapGenMakeSpellbook_5220E0(theme, name))
		},
		move: func(unit *server.Object, point *types.Pointf) {
			nox_xxx_mapGenMoveObject_527A10(asObjectC(unit), (*C.float2)(unsafe.Pointer(point)))
		},
	}
	mapGenPopulateDecor522340(definition, func(token uint32) *mapGenDecorSet {
		return (*mapGenDecorSet)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
	}, func(set *mapGenDecorSet) {
		mapGenPopulateDecorSet522370(set, roll, resolveEntry, func(entry *mapGenDecorEntry) {
			mapGenPopulateDecorEntry5224B0(entry, roomSize, modifiers, place)
		})
	})
}
