//go:build darwin && arm64

package legacy

/*
#include <stdint.h>
#include <string.h>
#include "defs.h"
typedef int32_t (*rdd_typed5_4e0b30)(nox_object_t*,nox_object_t*,nox_object_t*,int32_t,int32_t);
typedef struct rdd_context_4e0b30 {
 struct rdd_context_4e0b30* parent;
 uintptr_t words[5],nested_words[5];
 int32_t response,nested_response,nested_return;
 int nested,restored;
} rdd_context_4e0b30;
static _Thread_local rdd_context_4e0b30* rdd_active_4e0b30;
static int32_t rdd_capture5_4e0b30(nox_object_t*t,nox_object_t*s,nox_object_t*w,int32_t d,int32_t k) {
 rdd_context_4e0b30* c=rdd_active_4e0b30;
 c->words[0]=(uintptr_t)t;c->words[1]=(uintptr_t)s;c->words[2]=(uintptr_t)w;
 c->words[3]=(uint32_t)d;c->words[4]=(uint32_t)k;
 if(c->nested) {
  rdd_context_4e0b30 inner={0};inner.parent=c;inner.response=c->nested_response;
  rdd_active_4e0b30=&inner;
  c->nested_return=rdd_capture5_4e0b30(s,w,t,(int32_t)((uint32_t)d^0x80000000u),(int32_t)((uint32_t)k^0xffffffffu));
  memcpy(c->nested_words,inner.words,sizeof(inner.words));
  rdd_active_4e0b30=inner.parent;c->restored=rdd_active_4e0b30==c;
 }
 return c->response;
}
static void rdd_transport5_4e0b30(void*t,void*s,void*w,int32_t d,int32_t k,int32_t response,int nested,int32_t nested_response,uintptr_t*out) {
 rdd_context_4e0b30 frame={0};frame.parent=rdd_active_4e0b30;
 frame.response=response;frame.nested=nested;frame.nested_response=nested_response;rdd_active_4e0b30=&frame;
 rdd_typed5_4e0b30 fn=&rdd_capture5_4e0b30;
 int32_t result=fn(t,s,w,d,k);
 memcpy(out,frame.words,sizeof(frame.words));memcpy(out+5,frame.nested_words,sizeof(frame.nested_words));
 out[10]=(uint32_t)result;out[11]=(uint32_t)frame.nested_return;
 rdd_active_4e0b30=frame.parent;out[12]=(rdd_active_4e0b30==frame.parent)&&(!nested||frame.restored);
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Legacy transport and independent foreign-input controls. The native raw
// engine registration lives with the public export on every architecture.
type rawDefaultDamageAccess4E0B30 struct {
	Site        uint32
	Address     unsafe.Pointer
	NativeWidth int
	Value       uint64
	Write       bool
}
type rawDefaultDamageCall4E0B30 struct {
	Site     uint32
	Pointers [3]unsafe.Pointer
	Words    [5]uint32
	Damage   *int32
	Callback unsafe.Pointer
}
type rawDefaultDamageResponse4E0B30 struct {
	Word         uint32
	Pointer      unsafe.Pointer
	FloatFormat  string
	FloatPayload [10]byte
}
type rawDefaultDamageContext4E0B30 struct {
	AfterRead  func(rawDefaultDamageAccess4E0B30)
	AfterWrite func(rawDefaultDamageAccess4E0B30)
	Call       func(rawDefaultDamageCall4E0B30) rawDefaultDamageResponse4E0B30
	Frame      func(site uint32, width int) uint32
}

// Serial test-controlled context stack only; no t.Parallel or concurrent/raw
// production dispatch is qualified here. C TLS transport is a separate scope.
var rawDefaultDamageCurrentContext4E0B30 *rawDefaultDamageContext4E0B30

type rawDefaultDamageContextFrame4E0B30 struct {
	Context *rawDefaultDamageContext4E0B30
	Parent  *rawDefaultDamageContextFrame4E0B30
}

var rawDefaultDamageContextStack4E0B30 *rawDefaultDamageContextFrame4E0B30

func rawDefaultDamagePushContext4E0B30(next *rawDefaultDamageContext4E0B30) func() {
	previous := rawDefaultDamageCurrentContext4E0B30
	previousFrame := rawDefaultDamageContextStack4E0B30
	frame := &rawDefaultDamageContextFrame4E0B30{Context: next, Parent: previousFrame}
	rawDefaultDamageContextStack4E0B30 = frame
	rawDefaultDamageCurrentContext4E0B30 = next
	return func() {
		if rawDefaultDamageContextStack4E0B30 != frame || rawDefaultDamageCurrentContext4E0B30 != next {
			panic("raw damage context restoration order")
		}
		rawDefaultDamageContextStack4E0B30 = previousFrame
		rawDefaultDamageCurrentContext4E0B30 = previous
	}
}

// Foreign C-owned initial/reference graph reader. It never converts a Goheap
// pointer through uintptr. The load completes before an after-read mutation.
// An actual protected-field panic occurs before the observer is invoked.
func rawDefaultDamageForeignRead4E0B30(base unsafe.Pointer, offset uintptr, width int, site uint32) uint64 {
	loadWidth := width
	if width == 8 {
		loadWidth = 0
	}
	value := uint64(rawParentInitialGoBits4E0B30(base, offset, loadWidth))
	if c := rawDefaultDamageCurrentContext4E0B30; c != nil && c.AfterRead != nil {
		c.AfterRead(rawDefaultDamageAccess4E0B30{Site: site, Address: unsafe.Add(base, offset), NativeWidth: width, Value: value})
	}
	runtime.KeepAlive(base)
	return value
}

// Explicit input-effect transport only; no root branch/return is computed.
// Native WORD stores retain platform behavior: no fragment synthesis.
func rawDefaultDamageForeignStore4E0B30(base unsafe.Pointer, offset uintptr, width int, value uint64, site uint32) {
	if base == nil {
		_ = *(*byte)(base)
	}
	address := unsafe.Add(base, offset)
	switch width {
	case 1:
		value = uint64(uint8(value))
		*(*uint8)(address) = uint8(value)
	case 2:
		value = uint64(uint16(value))
		*(*uint16)(address) = uint16(value)
	case 4:
		value = uint64(uint32(value))
		*(*uint32)(address) = uint32(value)
	default:
		panic("non-scalar input effect width")
	}
	if c := rawDefaultDamageCurrentContext4E0B30; c != nil && c.AfterWrite != nil {
		c.AfterWrite(rawDefaultDamageAccess4E0B30{Site: site, Address: address, NativeWidth: width, Value: value, Write: true})
	}
	runtime.KeepAlive(base)
}

type rawDefaultDamageTransport4E0B30 struct {
	Outer, Inner        [5]uintptr
	Return, InnerReturn int32
	Restored            bool
}

func rawDefaultDamageCTransport4E0B30(target, source, weapon unsafe.Pointer, damage, kind, response int32, nested bool, nestedResponse int32) rawDefaultDamageTransport4E0B30 {
	var out [13]uintptr
	flag := 0
	if nested {
		flag = 1
	}
	C.rdd_transport5_4e0b30(target, source, weapon, C.int32_t(damage), C.int32_t(kind), C.int32_t(response), C.int(flag), C.int32_t(nestedResponse), (*C.uintptr_t)(unsafe.Pointer(&out[0])))
	var result rawDefaultDamageTransport4E0B30
	copy(result.Outer[:], out[:5])
	copy(result.Inner[:], out[5:10])
	result.Return = int32(uint32(out[10]))
	result.InnerReturn = int32(uint32(out[11]))
	result.Restored = out[12] == 1
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
	return result
}
