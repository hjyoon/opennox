//go:build nox_audio_methods_fixture

package legacy

/*
#include "audio_source_native.h"
#include "GAME1_3.h"
static int audio_native_device_create(void *device) { return sub_43E940(device); }
static void audio_native_device_destroy(void *device) { (void)device; sub_43E9F0(); }
static void audio_native_context_destroy(void *context) { (void)context; sub_43EC10(); }
static void audio_native_voice_destroy(void *voice) { sub_43ECB0(voice); }
static void audio_native_voice_stop(void *voice) { sub_43EFD0(voice); }
static void audio_native_voice_refresh(void *voice) { sub_43F060(voice); }
static void audio_native_voice_restart(void *voice) { (void)voice; sub_43F0D0(); }
static int audio_native_voice_resume(void *voice) { (void)voice; return sub_43F030(0); }
static int audio_native_voice_status(void *voice) { (void)voice; return sub_43F050(); }

static audio_native_class *am_class(void) {
    static audio_native_methods methods = {
        .create = sub_43EC30, .destroy = audio_native_voice_destroy, .play = sub_43ED00,
        .stop = audio_native_voice_stop, .pause = sub_43F010, .resume = audio_native_voice_resume,
        .status = audio_native_voice_status, .refresh = audio_native_voice_refresh, .restart = audio_native_voice_restart,
    };
    static audio_native_class class_ = {
        .name = "Audio", .flags = 2,
        .create = audio_native_device_create, .destroy = audio_native_device_destroy,
        .context_create = sub_43EA20, .context_destroy = audio_native_context_destroy,
        .methods = &methods,
    };
    return &class_;
}

static int am_context_create(void *p) { return am_class()->context_create(p); }
static void am_context_destroy(void *p) { am_class()->context_destroy(p); }
static void *am_create(void *p) { return sub_4BD720(p); }
static void am_init_buffer(void *p) { sub_487C30(p); }
static void am_select(void *p,void *b) { sub_4BDB90(p,b); }
static int am_play(void *p) { return sub_4BDB40(p); }
static void am_refresh(void *p) { ((audio_native_voice *)p)->methods->refresh(p); }
static int am_pause(void *p) { return ((audio_native_voice *)p)->methods->pause(p); }
static int am_stop(void *p) { return sub_4BDA80(p); }
static void am_destroy(void *p) { sub_4BD7A0(p); }
static uintptr_t am_driver(void) { return sub_43F130(); }
*/
import "C"

import "unsafe"

func AudioMethodsFixtureClass() *AudioClassNative {
	return (*AudioClassNative)(unsafe.Pointer(C.am_class()))
}
func AudioMethodsFixtureContextCreate(p *AudioContextNative) int32 {
	return int32(C.am_context_create(unsafe.Pointer(p)))
}
func AudioMethodsFixtureContextDestroy(p *AudioContextNative) {
	C.am_context_destroy(unsafe.Pointer(p))
}
func AudioMethodsFixtureCreate(p *AudioContextNative) *AudioVoiceNative {
	return (*AudioVoiceNative)(C.am_create(unsafe.Pointer(p)))
}
func AudioMethodsFixtureInitBuffer(p *AudioBufferNative) { C.am_init_buffer(unsafe.Pointer(p)) }
func AudioMethodsFixtureSelect(p *AudioVoiceNative, b *AudioBufferNative) {
	C.am_select(unsafe.Pointer(p), unsafe.Pointer(b))
}
func AudioMethodsFixturePlay(p *AudioVoiceNative) int32  { return int32(C.am_play(unsafe.Pointer(p))) }
func AudioMethodsFixtureRefresh(p *AudioVoiceNative)     { C.am_refresh(unsafe.Pointer(p)) }
func AudioMethodsFixturePause(p *AudioVoiceNative) int32 { return int32(C.am_pause(unsafe.Pointer(p))) }
func AudioMethodsFixtureStop(p *AudioVoiceNative) int32  { return int32(C.am_stop(unsafe.Pointer(p))) }
func AudioMethodsFixtureDestroy(p *AudioVoiceNative)     { C.am_destroy(unsafe.Pointer(p)) }
func AudioMethodsFixtureDriver() uintptr                 { return uintptr(C.am_driver()) }
