package legacy

/*
#include <stdint.h>
#include "readiness_raw_4e0960.h"
typedef int (*readiness_raw_trace_old_4e0960)(int);
typedef int32_t (*readiness_raw_trace_native_4e0960)(nox_object_t*);
typedef struct {uintptr_t function,item;int32_t phase,result;} readiness_raw_trace_record_4e0960;
static int readiness_raw_trace_enabled_4e0960,readiness_raw_trace_count_4e0960;
static readiness_raw_trace_record_4e0960 readiness_raw_trace_data_4e0960[2];
static void readiness_raw_trace_begin_4e0960(void){readiness_raw_trace_count_4e0960=0;readiness_raw_trace_enabled_4e0960=1;}
static void readiness_raw_trace_end_4e0960(void){
 readiness_raw_trace_enabled_4e0960=0;readiness_raw_trace_count_4e0960=0;
 for(int i=0;i<2;i++)readiness_raw_trace_data_4e0960[i]=(readiness_raw_trace_record_4e0960){0};
}
static void readiness_raw_trace_append_4e0960(uintptr_t function,nox_object_t* item,int phase,int32_t result){
 if(!readiness_raw_trace_enabled_4e0960)return;
 int i=readiness_raw_trace_count_4e0960++;
 if(i>=2)return;
 readiness_raw_trace_data_4e0960[i]=(readiness_raw_trace_record_4e0960){function,(uintptr_t)item,phase,result};
}
static int32_t readiness_raw_trace_call_old_4e0960(readiness_raw_trace_old_4e0960 fn,nox_object_t* item){return fn((int)(uintptr_t)item);}
static int32_t readiness_raw_trace_call_native_4e0960(readiness_raw_trace_native_4e0960 fn,nox_object_t* item){return fn(item);}
static int32_t readiness_raw_trace_public_4e0960(nox_object_t* item){
 uintptr_t function=(uintptr_t)&nox_xxx_itemCheckReadinessEffect_4E0960;
 readiness_raw_trace_append_4e0960(function,item,0,0);
 int32_t result=_Generic(&nox_xxx_itemCheckReadinessEffect_4E0960,
  readiness_raw_trace_old_4e0960:readiness_raw_trace_call_old_4e0960,
  readiness_raw_trace_native_4e0960:readiness_raw_trace_call_native_4e0960)
 (&nox_xxx_itemCheckReadinessEffect_4E0960,item);
 readiness_raw_trace_append_4e0960(function,item,1,result);return result;
}
static int32_t readiness_raw_transport_stub_4e0960(nox_object_t* item){(void)item;return -INT32_C(2023406815);}
static int32_t readiness_raw_transport_control_4e0960(nox_object_t* item){
 readiness_raw_trace_native_4e0960 fn=readiness_raw_transport_stub_4e0960;
 readiness_raw_trace_append_4e0960((uintptr_t)fn,item,0,0);
 int32_t result=fn(item);
 readiness_raw_trace_append_4e0960((uintptr_t)fn,item,1,result);return result;
}
static int readiness_raw_trace_size_4e0960(void){return readiness_raw_trace_count_4e0960;}
static readiness_raw_trace_record_4e0960 readiness_raw_trace_get_4e0960(int i){return readiness_raw_trace_data_4e0960[i];}
static uintptr_t readiness_raw_root_pointer_4e0960(void){return (uintptr_t)&nox_xxx_itemCheckReadinessEffect_4E0960;}
static int32_t readiness_raw_return_transport_4e0960(int32_t value){return value;}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type readinessRawValue4E0960 struct {
	Pointer unsafe.Pointer
	Word    uint32
}

// Pre-engine typed transports only; no admission, traversal or predicate root
// is implemented here. The observer is nil in ordinary production execution.
var readinessRawObserve4E0960 func(complete bool, base unsafe.Pointer, kind, index int, value readinessRawValue4E0960)

func readinessRawRead4E0960(base unsafe.Pointer, kind, index int) (value readinessRawValue4E0960) {
	if readinessRawObserve4E0960 != nil {
		readinessRawObserve4E0960(false, base, kind, index, value)
	}
	switch kind {
	case 0:
		value.Pointer = (*server.Object)(base).InitData
	case 1:
		value.Word = uint32((*server.Object)(base).ObjClass)
	case 2:
		value.Pointer = unsafe.Pointer((*server.ModifierInitData)(base).Modifiers[index])
	case 3:
		value.Pointer = (*server.ModifierEff)(base).Attack40.Fnc
	case 4:
		value.Word = uint32((*server.ModifierEff)(base).Attack40.Val)
	default:
		panic("undeclared raw Readiness field")
	}
	if readinessRawObserve4E0960 != nil {
		readinessRawObserve4E0960(true, base, kind, index, value)
	}
	return
}

type readinessRawTrace4E0960 struct {
	Function, Item uintptr
	Phase          int
	Result         int32
}

func readinessRawTracing4E0960(enabled bool) {
	if enabled {
		C.readiness_raw_trace_begin_4e0960()
	} else {
		C.readiness_raw_trace_end_4e0960()
	}
}

func readinessRawTraces4E0960() (out []readinessRawTrace4E0960) {
	n := int(C.readiness_raw_trace_size_4e0960())
	if n < 0 || n > 2 {
		panic("Readiness C trace overflow")
	}
	for i := 0; i < n; i++ {
		r := C.readiness_raw_trace_get_4e0960(C.int(i))
		out = append(out, readinessRawTrace4E0960{uintptr(r.function), uintptr(r.item), int(r.phase), int32(r.result)})
	}
	return
}

func readinessRawRootPointer4E0960() uintptr { return uintptr(C.readiness_raw_root_pointer_4e0960()) }

func readinessRawTransport4E0960(item unsafe.Pointer) int32 {
	return int32(C.readiness_raw_transport_control_4e0960((*C.nox_object_t)(item)))
}

func readinessRawReturnTransport4E0960(value int32) int32 {
	return int32(C.readiness_raw_return_transport_4e0960(C.int32_t(value)))
}

// This hook remains nil before the separately frozen native implementation.
// Its later engine must be the same function used by the public Go export.
var readinessRawEngine4E0960 func(item unsafe.Pointer) int32
var readinessRawObserveRoot4E0960 func(complete bool, item unsafe.Pointer, result int32)

func readinessRawInvoke4E0960(public bool, item unsafe.Pointer) (result int32) {
	if readinessRawObserveRoot4E0960 != nil {
		readinessRawObserveRoot4E0960(false, item, 0)
	}
	if public {
		result = int32(C.readiness_raw_trace_public_4e0960((*C.nox_object_t)(item)))
	} else {
		result = readinessRawEngine4E0960(item)
	}
	if readinessRawObserveRoot4E0960 != nil {
		readinessRawObserveRoot4E0960(true, item, result)
	}
	return
}
