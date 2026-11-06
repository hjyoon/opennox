package legacy

/*
#include <stdint.h>
#include "regeneration_raw_4e01d0.h"
typedef void (*regeneration_raw_transport_fn_4e01d0)(void*,nox_object_t*);
static uintptr_t regeneration_raw_transport_seen_4e01d0[2];
static void regeneration_raw_transport_stub_4e01d0(void* effect,nox_object_t* item) {
 regeneration_raw_transport_seen_4e01d0[0]=(uintptr_t)effect;
 regeneration_raw_transport_seen_4e01d0[1]=(uintptr_t)item;
}
static void regeneration_raw_transport_control_4e01d0(void* effect,nox_object_t* item) {
 regeneration_raw_transport_fn_4e01d0 fn=regeneration_raw_transport_stub_4e01d0;
 fn(effect,item);
}
static uintptr_t regeneration_raw_transport_get_4e01d0(int index) {
 return regeneration_raw_transport_seen_4e01d0[index];
}
static uintptr_t regeneration_raw_root_pointer_4e01d0(void) {
 return (uintptr_t)&nox_xxx_effectRegeneration_4E01D0;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type regenerationRawValue4E01D0 struct {
	Pointer unsafe.Pointer
	Word    uint32
}

// Fields 0..7 match the independently frozen initial-memory correspondence.
// Fields 8/9 are the actual frame/FPS service reads. Observers are nil normally.
var regenerationRawObserveRead4E01D0 func(complete bool, base unsafe.Pointer, kind int, value regenerationRawValue4E01D0)

var regenerationRawFrame4E01D0 = func() uint32 { return GetServer().S().Frame() }
var regenerationRawFPS4E01D0 = func() uint32 { return GetServer().S().TickRate() }

func regenerationRawRead4E01D0(base unsafe.Pointer, kind int) (value regenerationRawValue4E01D0) {
	if regenerationRawObserveRead4E01D0 != nil {
		regenerationRawObserveRead4E01D0(false, base, kind, value)
	}
	switch kind {
	case 0:
		value.Pointer = unsafe.Pointer((*server.Object)(base).InvHolder)
	case 1:
		value.Word = uint32((*server.Object)(base).ObjClass)
	case 2:
		value.Pointer = unsafe.Pointer((*server.Object)(base).HealthData)
	case 3:
		value.Word = (*server.Object)(base).Frame134
	case 4:
		value.Word = uint32((*server.Object)(base).ObjFlags)
	case 5:
		value.Word = uint32((*server.ModifierEff)(base).Update100.Val)
	case 6:
		value.Word = uint32((*server.HealthData)(base).Cur)
	case 7:
		value.Word = uint32((*server.HealthData)(base).Max)
	case 8:
		value.Word = regenerationRawFrame4E01D0()
	case 9:
		value.Word = regenerationRawFPS4E01D0()
	default:
		panic("undeclared raw Regeneration read")
	}
	if regenerationRawObserveRead4E01D0 != nil {
		regenerationRawObserveRead4E01D0(true, base, kind, value)
	}
	return
}

type regenerationRawCall4E01D0 struct {
	Kind     string
	Function uintptr
	Args     [2]uintptr
	Complete bool
	Word     uint32
}

var regenerationRawObserveCall4E01D0 func(regenerationRawCall4E01D0)

func regenerationRawCallBegin4E01D0(kind string, first, second uintptr) regenerationRawCall4E01D0 {
	call := regenerationRawCall4E01D0{Kind: kind, Args: [2]uintptr{first, second}}
	if kind == "regeneration" {
		call.Function = uintptr(C.regeneration_raw_root_pointer_4e01d0())
	}
	if regenerationRawObserveCall4E01D0 != nil {
		regenerationRawObserveCall4E01D0(call)
	}
	return call
}

func regenerationRawCallEnd4E01D0(call regenerationRawCall4E01D0, word uint32) {
	if regenerationRawObserveCall4E01D0 != nil {
		call.Complete, call.Word = true, word
		regenerationRawObserveCall4E01D0(call)
	}
}

func regenerationRawMaximum4E01D0(owner *server.Object) uint16 {
	call := regenerationRawCallBegin4E01D0("maximum-hp", uintptr(unsafe.Pointer(owner)), 0)
	value := server.UnitGetMaxHP4EE7A0(owner)
	regenerationRawCallEnd4E01D0(call, uint32(value))
	return value
}

func regenerationRawCurrent4E01D0(owner *server.Object) uint16 {
	call := regenerationRawCallBegin4E01D0("current-hp", uintptr(unsafe.Pointer(owner)), 0)
	value := server.UnitGetHP4EE780(owner)
	regenerationRawCallEnd4E01D0(call, uint32(value))
	return value
}

var regenerationRawArmorService4E01D0 = func(item *server.Object) uint32 {
	return GetServer().S().Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70(item)
}
var regenerationRawAdjustService4E01D0 = unitAdjustHPCall4EE460

func regenerationRawArmor4E01D0(item *server.Object) uint32 {
	call := regenerationRawCallBegin4E01D0("armor-flags", uintptr(unsafe.Pointer(item)), 0)
	value := regenerationRawArmorService4E01D0(item)
	regenerationRawCallEnd4E01D0(call, value)
	return value
}

func regenerationRawAdjust4E01D0(owner *server.Object, delta int32) {
	call := regenerationRawCallBegin4E01D0("adjust-hp", uintptr(unsafe.Pointer(owner)), uintptr(uint32(delta)))
	regenerationRawAdjustService4E01D0(owner, delta)
	regenerationRawCallEnd4E01D0(call, 0) // The native service and root are void.
}

// Stage 0 is unsigned interval division (PE4E0279); stage 1 is remainder
// (PE4E0287). Real arithmetic faults remain real Go runtime errors.
var regenerationRawObserveArithmetic4E01D0 func(complete bool, stage int, dividend, divisor, result uint32)

func regenerationRawDivide4E01D0(stage int, dividend, divisor uint32) (result uint32) {
	if regenerationRawObserveArithmetic4E01D0 != nil {
		regenerationRawObserveArithmetic4E01D0(false, stage, dividend, divisor, 0)
	}
	switch stage {
	case 0:
		result = dividend / divisor
	case 1:
		result = dividend % divisor
	default:
		panic("undeclared raw Regeneration arithmetic stage")
	}
	if regenerationRawObserveArithmetic4E01D0 != nil {
		regenerationRawObserveArithmetic4E01D0(true, stage, dividend, divisor, result)
	}
	return
}

// The implementation is deliberately absent during consumer preparation.
// The later raw engine registers the same function used by its public export.
var regenerationRawEngine4E01D0 func(effect unsafe.Pointer, item *server.Object)

func regenerationRawInvoke4E01D0(public bool, effect, item unsafe.Pointer) {
	call := regenerationRawCallBegin4E01D0("regeneration", uintptr(effect), uintptr(item))
	if public {
		regenerationRawPublic4E01D0(effect, item)
	} else {
		regenerationRawEngine4E01D0(effect, (*server.Object)(item))
	}
	regenerationRawCallEnd4E01D0(call, 0)
}

func regenerationRawTransportControl4E01D0(effect, item unsafe.Pointer) (actual [2]uintptr) {
	C.regeneration_raw_transport_control_4e01d0(effect, (*C.nox_object_t)(item))
	for i := range actual {
		actual[i] = uintptr(C.regeneration_raw_transport_get_4e01d0(C.int(i)))
	}
	return
}
