package cfixture

/*
#cgo CPPFLAGS: -I${SRCDIR}/../.. -DNOX_CGO
#cgo CFLAGS: -fshort-wchar -fno-strict-aliasing -fno-strict-overflow
#cgo arm arm64 CFLAGS: -fsigned-char
#cgo CFLAGS: -Wno-pointer-to-int-cast -Wno-int-to-pointer-cast
#cgo CFLAGS: -Wno-incompatible-pointer-types -Wno-int-conversion
#include "memfile.h"
#include <stddef.h>
#include <stdint.h>
void* thing_fixture_save_411540(void);
void thing_fixture_restore_411540(void* p);
size_t thing_fixture_state_size_411540(void);
void thing_fixture_seed_411540(uint32_t floor, uint32_t edge);
void* thing_fixture_memfile_411540(const unsigned char* data,int n,int start);
void thing_fixture_memfile_free_411540(void* p);
intptr_t thing_fixture_offset_411540(void* p);
void* thing_fixture_input_411540(void* p);
uintptr_t thing_fixture_pointer_411540(void* p);
void thing_fixture_free_411540(void* p);
size_t thing_fixture_layout_411540(int i);
int thing_fixture_signed_char_411540(void);
uint32_t thing_fixture_rgb_411540(uint8_t r,uint8_t g,uint8_t b);
int thing_fixture_audio_real_415660(nox_memfile* f,void* p);
int thing_fixture_call_411540(uint32_t entry,int reference,nox_memfile* f,void* p);
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func ThingSave411540() (unsafe.Pointer, func()) {
	p := C.thing_fixture_save_411540()
	if p == nil {
		panic("thing fixture state allocation")
	}
	return p, func() { C.thing_fixture_restore_411540(p); C.thing_fixture_free_411540(p) }
}
func ThingState411540() []byte {
	p := C.thing_fixture_save_411540()
	if p == nil {
		panic("thing fixture snapshot allocation")
	}
	defer C.thing_fixture_free_411540(p)
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(C.thing_fixture_state_size_411540()))...)
}
func ThingSeed411540(floor, edge uint32) {
	C.thing_fixture_seed_411540(C.uint32_t(floor), C.uint32_t(edge))
}
func ThingMemfile411540(input []byte, start int) (unsafe.Pointer, func()) {
	var data *C.uchar
	if len(input) != 0 {
		data = (*C.uchar)(unsafe.Pointer(&input[0]))
	}
	p := C.thing_fixture_memfile_411540(data, C.int(len(input)), C.int(start))
	runtime.KeepAlive(input)
	if p == nil {
		panic("thing fixture input allocation")
	}
	return p, func() { C.thing_fixture_memfile_free_411540(p) }
}
func ThingInput411540(f unsafe.Pointer, n int) []byte {
	p := C.thing_fixture_input_411540(unsafe.Pointer(f))
	return append([]byte(nil), unsafe.Slice((*byte)(p), n+32)...)
}
func ThingOffset411540(f unsafe.Pointer) int {
	return int(C.thing_fixture_offset_411540(unsafe.Pointer(f)))
}
func ThingCall411540(entry uint32, reference bool, f unsafe.Pointer, p unsafe.Pointer) int32 {
	r := C.int(0)
	if reference {
		r = 1
	}
	v := int32(C.thing_fixture_call_411540(C.uint32_t(entry), r, (*C.nox_memfile)(f), p))
	runtime.KeepAlive(f)
	runtime.KeepAlive(p)
	return v
}
func ThingPointer411540(p unsafe.Pointer) uintptr {
	return uintptr(C.thing_fixture_pointer_411540(p))
}
func ThingLayout411540() []uintptr {
	r := make([]uintptr, 11)
	for i := range r {
		r[i] = uintptr(C.thing_fixture_layout_411540(C.int(i)))
	}
	return r
}
func ThingSignedChar411540() bool { return C.thing_fixture_signed_char_411540() != 0 }
func ThingRGB411540(r, g, b byte) uint32 {
	return uint32(C.thing_fixture_rgb_411540(C.uint8_t(r), C.uint8_t(g), C.uint8_t(b)))
}
func ThingAudioReal415660(f unsafe.Pointer, p unsafe.Pointer) int32 {
	return int32(C.thing_fixture_audio_real_415660((*C.nox_memfile)(f), p))
}
func ThingGoMemfile411540(b []byte) unsafe.Pointer {
	f := new(C.nox_memfile)
	f.data = (*C.char)(unsafe.Pointer(&b[0]))
	f.cur = f.data
	f.size = C.int(len(b))
	f.end = (*C.char)(unsafe.Add(unsafe.Pointer(f.data), len(b)))
	return unsafe.Pointer(f)
}
