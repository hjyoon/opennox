package legacy

/*
#include <stddef.h>
#include <stdint.h>

typedef struct nox_respawn_record_t nox_respawn_record_t;

extern void* nox_alloc_respawn_1568020;
extern nox_respawn_record_t* dword_5d4594_1568024;
extern uint32_t nox_xxx_respawnAllow_587000_205200;

static uint32_t nox_respawn_add_load_allow_4EC5E0(void) {
	return nox_xxx_respawnAllow_587000_205200;
}

static void nox_respawn_add_store_allow_4EC5E0(uint32_t value) {
	nox_xxx_respawnAllow_587000_205200 = value;
}

static void* nox_respawn_add_load_allocator_4EC5E0(void) {
	return nox_alloc_respawn_1568020;
}

static void nox_respawn_add_store_allocator_4EC5E0(void* value) {
	nox_alloc_respawn_1568020 = value;
}

static void* nox_respawn_add_load_head_4EC5E0(void) {
	return dword_5d4594_1568024;
}

static void nox_respawn_add_store_head_4EC5E0(void* value) {
	dword_5d4594_1568024 = (nox_respawn_record_t*)value;
}
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func respawnAddLoadAllow4EC5E0() uint32 {
	return uint32(C.nox_respawn_add_load_allow_4EC5E0())
}

func respawnAddStoreAllow4EC5E0(value uint32) {
	C.nox_respawn_add_store_allow_4EC5E0(C.uint32_t(value))
}

func respawnAddLoadAllocator4EC5E0() unsafe.Pointer {
	return C.nox_respawn_add_load_allocator_4EC5E0()
}

func respawnAddStoreAllocator4EC5E0(value unsafe.Pointer) {
	C.nox_respawn_add_store_allocator_4EC5E0(value)
}

func respawnAddLoadHead4EC5E0() *respawnRecord4EC5E0 {
	return (*respawnRecord4EC5E0)(C.nox_respawn_add_load_head_4EC5E0())
}

func respawnAddStoreHead4EC5E0(value *respawnRecord4EC5E0) {
	C.nox_respawn_add_store_head_4EC5E0(unsafe.Pointer(value))
}

// Nox_xxx_respawnAdd_4EC5E0 restores GAME.EXE 004EC5E0 while retaining the
// native list representation consumed by the adjacent legacy routines.
func Nox_xxx_respawnAdd_4EC5E0(obj *server.Object) {
	server.RespawnAdd4EC5E0(obj, server.RespawnAddHooks4EC5E0[
		*server.Object, *respawnRecord4EC5E0, unsafe.Pointer, unsafe.Pointer,
	]{
		LoadAllow: respawnAddLoadAllow4EC5E0,
		LoadAllocator: func() unsafe.Pointer {
			return respawnAddLoadAllocator4EC5E0()
		},
		AllocZero: func(allocator unsafe.Pointer) *respawnRecord4EC5E0 {
			return (*respawnRecord4EC5E0)(alloc.AsClass(allocator).NewObject())
		},
		LoadTypeInd: func(obj *server.Object) uint16 {
			return obj.TypeInd
		},
		StoreObject: func(rec *respawnRecord4EC5E0, obj *server.Object) {
			rec.Object = obj
		},
		StoreTypeInd: func(rec *respawnRecord4EC5E0, value uint32) {
			rec.TypeInd = value
		},
		LoadPositionXBits: func(obj *server.Object) uint32 {
			return math.Float32bits(obj.PosVec.X)
		},
		StorePositionXBits: func(rec *respawnRecord4EC5E0, value uint32) {
			rec.X = math.Float32frombits(value)
		},
		LoadPositionYBits: func(obj *server.Object) uint32 {
			return math.Float32bits(obj.PosVec.Y)
		},
		StorePositionYBits: func(rec *respawnRecord4EC5E0, value uint32) {
			rec.Y = math.Float32frombits(value)
		},
		LoadDirection: func(obj *server.Object) uint16 {
			return uint16(obj.Direction1)
		},
		StoreDirection: func(rec *respawnRecord4EC5E0, value uint16) {
			rec.Direction = value
		},
		LoadClass: func(obj *server.Object) uint32 {
			return uint32(obj.ObjClass)
		},
		CopyModifierAttrs: func(rec *respawnRecord4EC5E0, obj *server.Object) {
			src := (*server.ModifierInitData)(obj.InitData)
			for i := range src.Modifiers {
				rec.Attrs.Modifiers[i] = src.Modifiers[i]
			}
			rec.Attrs.Field16 = src.Field16
		},
		WeaponEquipFlags: objectNPCWeaponEquipFlags,
		LoadUseData: func(obj *server.Object) unsafe.Pointer {
			return obj.UseData.Ptr
		},
		LoadUseByte: func(data unsafe.Pointer, index uint32) uint8 {
			return *(*uint8)(unsafe.Add(data, uintptr(index)))
		},
		StoreCharge1: func(rec *respawnRecord4EC5E0, value uint8) {
			rec.Charge1 = value
		},
		StoreCharge0: func(rec *respawnRecord4EC5E0, value uint8) {
			rec.Charge0 = value
		},
		LoadHead: respawnAddLoadHead4EC5E0,
		StoreNext: func(rec *respawnRecord4EC5E0, next *respawnRecord4EC5E0) {
			rec.Next = next
		},
		StorePrev: func(rec *respawnRecord4EC5E0, prev *respawnRecord4EC5E0) {
			rec.Prev = prev
		},
		StoreHead: respawnAddStoreHead4EC5E0,
	})
}
