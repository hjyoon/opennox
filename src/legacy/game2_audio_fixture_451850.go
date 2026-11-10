package legacy

/*
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
#include "game2_audio_native.h"
#include "game2_audio_reference_452890.h"
extern void* dword_587000_127004;
void *sub_4864A0(void*);
void sub_486570(void*,void*);
void sub_486620(void*);
// Typed transport of the exact 4BD840 service-call order. Input buffers have
// only timer scalars; no PE32 pointer-bearing voice/context is executed.
static void ga_fixture_mix(game2_audio_group*out,game2_audio_group*voice,game2_audio_group*owner,game2_audio_group*fx,game2_audio_group*base,game2_audio_group*adaptive){sub_4864A0(out);sub_486570(out,voice);sub_486620(voice);if(owner){sub_486570(out,owner);sub_486620(owner);}if(fx)sub_486570(out,fx);sub_486570(out,base);if(adaptive)sub_486570(out,adaptive);}
static void ga_fixture_mixed_parameters(game2_audio_group*in,uint32_t rate,int32_t*out){out[0]=(uint32_t)(127*(in->timers[0].current>>16))>>14;out[1]=(uint32_t)(127*(in->timers[2].current>>16))>>14;int32_t product=(uint32_t)rate*(in->timers[1].current>>16);out[2]=product/100;}
static void*ga_fixture_fx_owner(void*p,int set){void*old=dword_587000_127004;if(set)dword_587000_127004=p;return old;}
static size_t ga_fixture_layout(int i){switch(i){
case 0:return sizeof(game2_audio_event);case 1:return sizeof(game2_audio_voice);case 2:return sizeof(game2_audio_handle);
case 3:return offsetof(game2_audio_event,definition);case 4:return offsetof(game2_audio_event,sample_count);
case 5:return offsetof(game2_audio_event,voice);case 6:return offsetof(game2_audio_event,group);
case 7:return offsetof(game2_audio_event,audio_sample);case 8:return offsetof(game2_audio_event,serial);
case 9:return offsetof(game2_audio_event,due);case 10:return offsetof(game2_audio_event,buffer);
case 11:return offsetof(game2_audio_event,order);case 12:return offsetof(game2_audio_event,remaining);
case 13:return offsetof(game2_audio_event,shuffle_remaining);case 14:return sizeof(void*);
case 15:return offsetof(game2_audio_handle,definition);case 16:return sizeof(game2_audio_context);case 17:return offsetof(game2_audio_context,ail_driver);case 18:return offsetof(game2_audio_context,event_master);case 19:return offsetof(game2_audio_context,base_group);case 20:return offsetof(game2_audio_voice,mixed);default:return 0;}}
static uintptr_t ga_fixture_pointer(void *p){return (uintptr_t)p;}
static void ga_fixture_reference_reset(const char *name){const int fields[]={0,1,2,14,15,19,20,48,18,17,25,26};memset(ga_ref_words,0xa7,sizeof(ga_ref_words));for(size_t i=0;i<sizeof(fields)/sizeof(fields[0]);i++)ga_ref_words[fields[i]]=0;strncpy(ga_ref_name,name,255);ga_ref_name[255]=0;ga_ref_words[12]=1;ga_ref_words[16]=600;sub_4862E0(ga_ref_words+4,0x4000);}
static void* ga_fixture_reference_words(void){return ga_ref_words;}
static int ga_fixture_parse(int kind,int reference,nox_memfile *f,void *scratch){if(reference){switch(kind){case 0:return ga_ref_avnt(f,scratch);case 1:return ga_ref_aud(f,scratch);case 2:return ga_ref_skip(f);default:return ga_ref_skip_inner(f);}}switch(kind){case 0:return nox_thing_read_AVNT_452890(f,scratch);case 1:return sub_452BD0(f,scratch);case 2:return nox_thing_skip_AVNT_452B00(f);default:return nox_thing_skip_AVNT_inner_452B30(f);}}
static int ga_fixture_pan(int reference,int v){return reference?ga_ref_pan(v):sub_452FA0(v);}
static void *ga_fixture_allocate(void *p){return nox_xxx_draw_452300(p);}
static void *ga_fixture_set_handle(void*h,void*e){return sub_452E90(h,e);}
static void *ga_fixture_get_handle(void*h){return sub_452EB0(h);}
static int64_t ga_fixture_delay(void*e,int64_t delta,int next){return sub_452690(e,delta,next);}
static int ga_fixture_init(void *context,void*bank){return sub_451850(context,bank);}
static void ga_fixture_shutdown(void){sub_451970();}
static int ga_fixture_stop(void*e){return sub_4523D0(e);}
static void *ga_fixture_release(void*e){return sub_451FE0(e);}
static int ga_fixture_enabled(int set){if(set<0)return sub_453070();if(set)nox_xxx____setargv_9_453060();else sub_453050();return sub_453070();}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/timer"
	"runtime"
	"unsafe"
)

func game2AudioFixtureLayout() []uintptr {
	v := make([]uintptr, 21)
	for i := range v {
		v[i] = uintptr(C.ga_fixture_layout(C.int(i)))
	}
	return v
}
func game2AudioFixtureMix(out, voice, owner, fx, base, adaptive *timer.TimerGroup) {
	C.ga_fixture_mix((*C.game2_audio_group)(unsafe.Pointer(out)), (*C.game2_audio_group)(unsafe.Pointer(voice)), (*C.game2_audio_group)(unsafe.Pointer(owner)), (*C.game2_audio_group)(unsafe.Pointer(fx)), (*C.game2_audio_group)(unsafe.Pointer(base)), (*C.game2_audio_group)(unsafe.Pointer(adaptive)))
}
func game2AudioFixtureMixedParameters(group *timer.TimerGroup, rate uint32) [3]int32 {
	var out [3]int32
	C.ga_fixture_mixed_parameters((*C.game2_audio_group)(unsafe.Pointer(group)), C.uint32_t(rate), (*C.int32_t)(unsafe.Pointer(&out[0])))
	return out
}
func game2AudioFixtureFXOwner(p unsafe.Pointer, set bool) unsafe.Pointer {
	s := C.int(0)
	if set {
		s = 1
	}
	return C.ga_fixture_fx_owner(p, s)
}
func game2AudioFixturePointer(p unsafe.Pointer) uintptr { return uintptr(C.ga_fixture_pointer(p)) }
func game2AudioFixtureReferenceReset(name string)       { C.ga_fixture_reference_reset(internCStr(name)) }
func game2AudioFixtureReferenceWords() []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(C.ga_fixture_reference_words()), 200)...)
}
func game2AudioFixtureParse(kind int, reference bool, f *binfile.MemFile, scratch unsafe.Pointer) int32 {
	r := C.int(0)
	if reference {
		r = 1
	}
	ret := int32(C.ga_fixture_parse(C.int(kind), r, (*C.nox_memfile)(f.C()), scratch))
	runtime.KeepAlive(f)
	return ret
}
func game2AudioFixturePan(reference bool, v int32) int32 {
	r := C.int(0)
	if reference {
		r = 1
	}
	return int32(C.ga_fixture_pan(r, C.int(v)))
}
func game2AudioFixtureAllocate(p unsafe.Pointer) *game2AudioEvent {
	return (*game2AudioEvent)(C.ga_fixture_allocate(p))
}
func game2AudioFixtureSetHandle(h *game2AudioHandle, e *game2AudioEvent) unsafe.Pointer {
	return C.ga_fixture_set_handle(unsafe.Pointer(h), unsafe.Pointer(e))
}
func game2AudioFixtureGetHandle(h *game2AudioHandle) *game2AudioEvent {
	return (*game2AudioEvent)(C.ga_fixture_get_handle(unsafe.Pointer(h)))
}
func game2AudioFixtureDelay(e *game2AudioEvent, delta int64, next int32) int64 {
	return int64(C.ga_fixture_delay(unsafe.Pointer(e), C.int64_t(delta), C.int(next)))
}
func game2AudioFixtureInit(context, bank unsafe.Pointer) int32 {
	return int32(C.ga_fixture_init(context, bank))
}
func game2AudioFixtureShutdown() { C.ga_fixture_shutdown() }
func game2AudioFixtureStop(e *game2AudioEvent) int32 {
	return int32(C.ga_fixture_stop(unsafe.Pointer(e)))
}
func game2AudioFixtureRelease(e *game2AudioEvent) unsafe.Pointer {
	return C.ga_fixture_release(unsafe.Pointer(e))
}
func game2AudioFixtureEnabled(v int32) int32 { return int32(C.ga_fixture_enabled(C.int(v))) }
