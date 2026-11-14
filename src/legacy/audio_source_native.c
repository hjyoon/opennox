#include "audio_source_native.h"
#include "GAME1_3.h"

void *sub_487150_go(int index, void *format);
void *sub_4871C0_go(void *device, int index, void *format);
void *sub_487590_go(void *context, void *format);

void *sub_487150(int index, const void *format) {
    return sub_487150_go(index, (void *)format);
}
void *sub_4871C0(void *device, int index, const void *format) {
    return sub_4871C0_go(device, index, (void *)format);
}
void *sub_487590(void *context, const void *format) {
    return sub_487590_go(context, (void *)format);
}

void nox_audio_source_callbacks(void *source) {
    audio_native_voice *v = source;
    v->refill = sub_4BD8C0;
    v->restart = sub_4BD940;
    v->finish = sub_4BD9B0;
}
void *nox_audio_context_update_callback(void) { return sub_4873C0; }

static int audio_native_device_create(void *device) { return sub_43E940(device); }
static void audio_native_device_destroy(void *device) { (void)device; sub_43E9F0(); }
static void audio_native_context_destroy(void *context) { (void)context; sub_43EC10(); }
static void audio_native_voice_destroy(void *voice) { sub_43ECB0(voice); }
static void audio_native_voice_stop(void *voice) { sub_43EFD0(voice); }
static void audio_native_voice_refresh(void *voice) { sub_43F060(voice); }
static void audio_native_voice_restart(void *voice) { (void)voice; sub_43F0D0(); }
static int audio_native_voice_resume(void *voice) { (void)voice; return sub_43F030(0); }
static int audio_native_voice_status(void *voice) { (void)voice; return sub_43F050(); }

void *nox_audio_builtin_class(void) {
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
