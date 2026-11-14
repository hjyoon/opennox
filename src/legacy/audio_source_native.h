#ifndef OPENNOX_AUDIO_SOURCE_NATIVE_H
#define OPENNOX_AUDIO_SOURCE_NATIVE_H
#include <stdint.h>
#include <stddef.h>
#include "game2_audio_native.h"
#include "client/audio/ail/compat_mss.h"

/* Private native owners, never overlays of the old 312-byte voice or packed
 * method tables. Callback slots and all links contain complete C pointers. */
#if defined(_WIN32) && UINTPTR_MAX == UINT32_MAX
#pragma pack(push, 4)
#endif
typedef struct audio_native_link {
    struct audio_native_link *next, *prev, *head;
} audio_native_link;
typedef struct audio_native_chunk {
    audio_native_link link;
    unsigned char *data;
    uint32_t size;
    void *buffer;
} audio_native_chunk;
typedef struct audio_native_buffer {
    unsigned char *data;
    uint32_t size;
    audio_native_link chunks;
    uint32_t *format;
    uint32_t field24;
} audio_native_buffer;
typedef int (*audio_native_status_callback)(void *);
typedef void (*audio_native_notification_callback)(void *);
typedef struct audio_native_methods {
    void *reserved;
    audio_native_status_callback create;
    audio_native_notification_callback destroy;
    audio_native_status_callback play;
    audio_native_notification_callback stop;
    audio_native_status_callback pause, resume, status;
    audio_native_notification_callback refresh, restart;
} audio_native_methods;
typedef struct audio_native_class {
    void *name;
    uint32_t field4, flags, state;
    void *reserved;
    audio_native_status_callback create;
    audio_native_notification_callback destroy;
    audio_native_status_callback context_create;
    audio_native_notification_callback context_destroy;
    audio_native_methods *methods;
} audio_native_class;
typedef struct audio_native_device {
    audio_native_link link;
    audio_native_class *class_;
    uint32_t count, limit;
    void *contexts[16];
} audio_native_device;
typedef struct audio_native_context {
    audio_native_link link;
    int32_t kind;
    uint32_t users;
    audio_native_device *device;
    int32_t index;
    uint32_t field28;
    uint32_t format32[7], format60[7];
    game2_audio_group group;
    game2_audio_group *master;
    int32_t priority, count, limit;
    audio_native_link voices;
    int32_t busy;
    audio_native_status_callback update;
    uint32_t field220;
    uint64_t interval, delta, peak, last;
    audio_native_methods *methods;
    uint32_t field260;
} audio_native_context;
typedef struct audio_native_voice {
    audio_native_link link;
    int32_t kind;
    game2_audio_group group;
    game2_audio_group *owner_group, *master;
    int32_t priority;
    uint32_t flags;
    int32_t repeat;
    audio_native_context *context;
    audio_native_status_callback on_refill;
    audio_native_notification_callback on_restart;
    audio_native_status_callback on_end, on_stop;
    void *user_data;
    uint32_t reserved[4];
    audio_native_methods *methods;
    game2_audio_group mixed;
    void *sample;
    audio_native_status_callback refill, restart, finish;
    audio_native_buffer *buffer;
    audio_native_chunk *chunk;
    unsigned char *cursor;
    uint32_t available, remaining, field308;
} audio_native_voice;
typedef struct audio_native_global {
    audio_native_link devices, contexts;
    int32_t busy;
    game2_audio_group group;
} audio_native_global;
typedef struct audio_native_sample {
    uintptr_t driver;
    audio_native_voice *source;
    uintptr_t handle;
    uint32_t eof;
    void *format;
    unsigned char *scratch0, *scratch1;
    uint32_t finished;
} audio_native_sample;
typedef struct audio_native_cache {
    void *bank, *chunks, *samples;
    audio_native_link lru;
    int32_t chunk_size;
} audio_native_cache;
typedef struct audio_native_cached_sample {
    audio_native_link link;
    uint32_t references;
    int32_t id;
    uint32_t loaded;
    audio_native_buffer buffer;
    audio_native_cache *cache;
    uint32_t format[7];
} audio_native_cached_sample;
#if defined(_WIN32) && UINTPTR_MAX == UINT32_MAX
#pragma pack(pop)
#endif

void *nox_audio_builtin_class(void);
void nox_audio_source_callbacks(void *voice);
void *nox_audio_context_update_callback(void);

void sub_43EDB0(HSAMPLE sample);
int sub_43EE00(void *sample);
void *sub_4BD280(int count, int size);
void sub_4BD2D0(void *pool);
void *sub_4BD2E0(void *pool);
void *sub_4BD300(void *pool, void *item);
void *sub_4BD340(void *bank, int bytes, int count, int size);
void sub_4BD3C0(void *cache);
void *sub_4BD420(void *cache, int id);
void *sub_4BD470(void *cache, int id);
int sub_4BD600(void *cache);
void *sub_4BD650(void *sample);
int sub_4BD660(void *sample);
int sub_4BD680(void *sample);
void *sub_4BD690(void *sample);
void *sub_4BD710(void *sample);
void *sub_4BD720(void *context);
void sub_4BD7A0(void *voice);
void *sub_4BD7C0(void *voice);
void sub_4BD840(void *voice);
int sub_4BD8C0(void *voice);
int sub_4BD940(void *voice);
int sub_4BD9B0(void *voice);
void sub_4BDA60(void *voice);
int sub_4BDA80(void *voice);
void *sub_4BDB20(void *voice);
void *sub_4BDB30(void *voice);
int sub_4BDB40(void *voice);
void sub_4BDB90(void *voice, void *buffer);
void *sub_4BDC00(void *priority);
void *sub_4866D0(void *bank, int id);
unsigned int sub_486A10(void *bank, void *name);
int sub_486AA0(void *bank, int id, uint32_t *format);
int sub_486B60(void *bank, int id);
int sub_486DB0(void *bank, char *data, int size);
void *sub_486E00(void *bank);
int sub_486E30(void *context, void *voice);
int sub_486E90(void *voice);
void *sub_486FA0(void *class_);
void *sub_486FE0(void *class_);
void sub_487030(void *device);
void sub_487050(void *device);
void sub_487070(void *device);
void sub_487090(void *device);
void sub_4870A0(void);
void *sub_4870E0(void *out);
void *sub_487100(void *out);
void *sub_487150(int index, const void *format);
void *sub_4871C0(void *device, int index, const void *format);
void sub_4872C0(void *context);
int sub_487310(void *context);
void *sub_487360(int index, void *device, int *slot);
int sub_4873C0(void *context);
void *sub_487590(void *context, const void *format);
void *sub_4875B0(void *out);
void *sub_4875D0(void *out);
int sub_4875F0(void);
void sub_487680(void *context);
void *sub_4876A0(void *context);
void *sub_487750(void *context);
int sub_487790(void *context, int count);
void *sub_4877D0(void *context, void *out);
void *sub_4877F0(void *out);
void *sub_487810(void *context, int kind);
int sub_487910(void *context, int kind);
void *sub_487970(void *context, int kind);
void sub_487C30(void *buffer);
int sub_487C50(void *buffer, void *chunk);
void *sub_487C80(void *buffer);
int sub_487D00(uint32_t *format);
void *sub_487D30(void *chunk, void *data, int size);
void *sub_487D60(void *chunk);
#endif
