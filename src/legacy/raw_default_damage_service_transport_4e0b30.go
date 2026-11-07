//go:build darwin && arm64

package legacy

/*
#include <stdint.h>
#include <string.h>
#include "defs.h"
typedef struct rdd_service_capture_4e0b30 {
 uintptr_t words[5]; uint32_t response; uint32_t mutation;
 unsigned char packet[10]; unsigned int packet_length;
} rdd_service_capture_4e0b30;
static _Thread_local rdd_service_capture_4e0b30* rdd_service_active_4e0b30;
// Exact five outgoing operands, including the real int32_t* in fourth slot.
// This is a transport control, never the registered PreDamage/Defend helper.
typedef uint32_t (*rdd_service5_4e0b30)(nox_object_t*,nox_object_t*,nox_object_t*,int32_t*,uint32_t);
static uint32_t rdd_service_capture5_4e0b30(nox_object_t*t,nox_object_t*s,nox_object_t*w,int32_t*d,uint32_t k) {
 rdd_service_capture_4e0b30*c=rdd_service_active_4e0b30;
 c->words[0]=(uintptr_t)t;c->words[1]=(uintptr_t)s;c->words[2]=(uintptr_t)w;c->words[3]=(uintptr_t)d;c->words[4]=k;
 memcpy(d,&c->mutation,sizeof(c->mutation));return c->response;
}
static uint32_t rdd_service_transport5_4e0b30(void*t,void*s,void*w,int32_t*d,uint32_t k,uint32_t response,uint32_t mutation,uintptr_t*out) {
 rdd_service_capture_4e0b30 frame={0};rdd_service_capture_4e0b30*parent=rdd_service_active_4e0b30;
 frame.response=response;frame.mutation=mutation;rdd_service_active_4e0b30=&frame;
 rdd_service5_4e0b30 fn=&rdd_service_capture5_4e0b30;uint32_t result=fn(t,s,w,d,k);
 memcpy(out,frame.words,sizeof(frame.words));rdd_service_active_4e0b30=parent;return result;
}
// Raw x87 service input bytes are transported, never evaluated as host float.
static void rdd_packet_transport_4e0b30(const unsigned char*input,unsigned int length,unsigned char*output) {
 rdd_service_capture_4e0b30 frame={0};frame.packet_length=length;
 if(length<=sizeof(frame.packet)){memcpy(frame.packet,input,length);memcpy(output,frame.packet,length);}
}
static void* rdd_pointer_return_4e0b30(void*pointer) {return pointer;}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

func rawDefaultDamageCServiceTransport4E0B30(target, source, weapon unsafe.Pointer, damage *int32, kind, response, mutation uint32) ([5]uintptr, uint32) {
	var out [5]uintptr
	got := C.rdd_service_transport5_4e0b30(target, source, weapon, (*C.int32_t)(unsafe.Pointer(damage)), C.uint32_t(kind), C.uint32_t(response), C.uint32_t(mutation), (*C.uintptr_t)(unsafe.Pointer(&out[0])))
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
	runtime.KeepAlive(damage)
	return out, uint32(got)
}
func rawDefaultDamageCPacketTransport4E0B30(format string, input []byte) ([]byte, error) {
	want, ok := map[string]int{"f32": 4, "f64": 8, "ext80": 10}[format]
	if !ok || len(input) != want {
		return nil, fmt.Errorf("exact raw x87 packet domain")
	}
	output := make([]byte, len(input))
	C.rdd_packet_transport_4e0b30((*C.uchar)(unsafe.Pointer(&input[0])), C.uint(len(input)), (*C.uchar)(unsafe.Pointer(&output[0])))
	runtime.KeepAlive(input)
	return output, nil
}
func rawDefaultDamageCPointerTransport4E0B30(pointer unsafe.Pointer) unsafe.Pointer {
	result := C.rdd_pointer_return_4e0b30(pointer)
	runtime.KeepAlive(pointer)
	return result
}
