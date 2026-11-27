package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME3_2.h"
#include "native_callback_calls.h"
static uintptr_t modifier_predicate_transport_seen_4e03d0[6];
static int32_t modifier_predicate_transport_stub_4e03d0(void* effect,nox_object_t* item,uintptr_t context,nox_object_t* target,uintptr_t other,int32_t* out){
 uintptr_t values[6]={(uintptr_t)effect,(uintptr_t)item,context,(uintptr_t)target,other,(uintptr_t)out};
 for(int i=0;i<6;i++)modifier_predicate_transport_seen_4e03d0[i]=values[i];
 *out=INT32_C(0x13579bdf);return INT32_C(0x2468ace0);
}
static void* modifier_predicate_transport_pointer_4e03d0(void){return (void*)modifier_predicate_transport_stub_4e03d0;}
static uintptr_t modifier_predicate_transport_get_4e03d0(int index){return modifier_predicate_transport_seen_4e03d0[index];}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// These field accesses and callback transports are frozen before any native
// predicate/traversal engine. The observer is absent in ordinary production.
type modifierPredicateValue4E03D0 struct {
	Word    uint32
	Pointer unsafe.Pointer
}

var modifierPredicateObserve4E03D0 func(write, complete bool, base unsafe.Pointer, kind, index int, value modifierPredicateValue4E03D0)

func modifierPredicateRead4E03D0(base unsafe.Pointer, kind, index int) (value modifierPredicateValue4E03D0) {
	if modifierPredicateObserve4E03D0 != nil {
		modifierPredicateObserve4E03D0(false, false, base, kind, index, value)
	}
	switch kind {
	case 0:
		value.Word = uint32((*server.ModifierEff)(base).DefendCollide88.Val)
	case 1:
		value.Word = uint32((*server.Object)(base).ObjClass)
	case 2:
		value.Pointer = (*server.Object)(base).InitData
	case 3:
		value.Pointer = unsafe.Pointer((*server.ModifierInitData)(base).Modifiers[index])
	case 4:
		value.Pointer = (*server.ModifierEff)(base).DefendCollide88.Fnc
	case 5:
		value.Word = *(*uint32)(base)
	default:
		panic("undeclared modifier predicate field")
	}
	if modifierPredicateObserve4E03D0 != nil {
		modifierPredicateObserve4E03D0(false, true, base, kind, index, value)
	}
	return
}

func modifierPredicateStore4E03D0(output *int32, word uint32) {
	value := modifierPredicateValue4E03D0{Word: word}
	if modifierPredicateObserve4E03D0 != nil {
		modifierPredicateObserve4E03D0(true, false, unsafe.Pointer(output), 5, 0, value)
	}
	*output = int32(word)
	if modifierPredicateObserve4E03D0 != nil {
		modifierPredicateObserve4E03D0(true, true, unsafe.Pointer(output), 5, 0, value)
	}
}

type modifierPredicateTrace4E03D0 struct {
	Function              uintptr
	Args                  [6]uintptr
	Phase                 int
	Before, After, Result int32
}

var modifierPredicateTraceState4E03D0 struct {
	enabled bool
	count   int
	records [32]modifierPredicateTrace4E03D0
}

func modifierPredicateTracing4E03D0(enabled bool) {
	state := &modifierPredicateTraceState4E03D0
	if enabled {
		state.count = 0
	}
	state.enabled = enabled
}

func modifierPredicateTraceRecords4E03D0() []modifierPredicateTrace4E03D0 {
	state := &modifierPredicateTraceState4E03D0
	if state.count > len(state.records) {
		panic("modifier predicate trace overflow")
	}
	return append([]modifierPredicateTrace4E03D0(nil), state.records[:state.count]...)
}

func recordModifierPredicateCall4E03D0(fn, first, output unsafe.Pointer, opaque [4]uintptr, phase int, before, after, result int32) {
	state := &modifierPredicateTraceState4E03D0
	if !state.enabled {
		return
	}
	i := state.count
	state.count++
	if i < len(state.records) {
		state.records[i] = modifierPredicateTrace4E03D0{
			Function: uintptr(fn),
			Args:     [6]uintptr{uintptr(first), opaque[0], opaque[1], opaque[2], opaque[3], uintptr(output)},
			Phase:    phase, Before: before, After: after, Result: result,
		}
	}
}

func modifierPredicateRootPointer4E03D0(which int) unsafe.Pointer {
	switch int32(which) {
	case 0, 1:
		return modifierPredicatesCallback4E03D0(which)
	default:
		return C.nox_xxx_unusedCheckGripEffect_4E03F0
	}
}

func modifierPredicatesTracedPublic4E03D0(which int, argument unsafe.Pointer, opaque [4]uint32, output unsafe.Pointer) int32 {
	var args [4]uintptr
	for i, value := range opaque {
		args[i] = uintptr(value)
	}
	fn := modifierPredicateRootPointer4E03D0(which)
	recordModifierPredicateCall4E03D0(fn, argument, output, args, 0, 0, 0, 0)
	result := modifierPredicatesPublic4E03D0(which, argument, opaque, output)
	recordModifierPredicateCall4E03D0(fn, argument, output, args, 1, 0, 0, result)
	return result
}

func modifierPredicateInvoke4E03D0(fn, effect unsafe.Pointer, opaque [4]uintptr, output *int32) int32 {
	var before int32
	if modifierPredicateTraceState4E03D0.enabled {
		before = *output
		recordModifierPredicateCall4E03D0(fn, effect, unsafe.Pointer(output), opaque, 2, before, 0, 0)
	}
	result := int32(C.nox_legacy_call_modifier_predicate(fn, effect,
		C.uintptr_t(opaque[0]), C.uintptr_t(opaque[1]), C.uintptr_t(opaque[2]), C.uintptr_t(opaque[3]),
		(*C.int32_t)(unsafe.Pointer(output))))
	if modifierPredicateTraceState4E03D0.enabled {
		recordModifierPredicateCall4E03D0(fn, effect, unsafe.Pointer(output), opaque, 3, before, *output, result)
	}
	return result
}

func modifierPredicateTransportControl4E03D0(effect unsafe.Pointer, opaque [4]uintptr, output *int32) (result int32, actual [6]uintptr) {
	result = modifierPredicateInvoke4E03D0(C.modifier_predicate_transport_pointer_4e03d0(), effect, opaque, output)
	for i := range actual {
		actual[i] = uintptr(C.modifier_predicate_transport_get_4e03d0(C.int(i)))
	}
	return
}
