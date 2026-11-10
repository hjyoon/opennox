#ifndef OPENNOX_GAME2_AUDIO_NATIVE_H
#define OPENNOX_GAME2_AUDIO_NATIVE_H
#include <stdint.h>
#include <stddef.h>
#include "memfile.h"

/* Private native layouts. The original sound-definition wire/scalar record
 * remains exactly 200 bytes. Its old pointer/list slots are NOT native slots.
 * Names and list membership live in the Go-owned sidecar, pointing only to
 * C-owned nodes and interned C strings. No Go pointer is kept by C storage. */
typedef struct { uint32_t flags,current,target,delta; uint64_t max_delta,last; } game2_audio_timer;
typedef struct { game2_audio_timer timers[3]; } game2_audio_group;
typedef struct game2_audio_event game2_audio_event;
typedef struct game2_audio_voice game2_audio_voice;
struct game2_audio_event {
    game2_audio_event *next,*prev,*rank_next,*rank_prev;
    void *definition;
    uint32_t flags,state,saved_state;
    void *samples[32];
    uint32_t sample_count;
    int32_t cursor;
    game2_audio_voice *voice;
    game2_audio_group group;
    void *audio_sample;
    uint32_t serial,delay;
    uint64_t due;
    void *buffer;
    int32_t category;
    uint32_t order[32],remaining,loops,shuffle[32],shuffle_remaining;
};
struct game2_audio_voice {
    game2_audio_event *owner;
    game2_audio_group group;
    game2_audio_group mixed;
    game2_audio_group *master;
    int32_t priority,repeat;
    uint32_t flags;
    void *buffer;
};
typedef struct { game2_audio_event *event; uint32_t serial; void *definition; } game2_audio_handle;
typedef struct { game2_audio_group base_group; game2_audio_group *event_master; uintptr_t ail_driver; } game2_audio_context;

/* Root must replace the corresponding GAME2.h signatures, not add conflicting
 * redeclarations. Pointer results below are real pointers, never raw EAX ints. */
int sub_451850(void *native_context, void *bank);
int sub_451920(void *definition);
void sub_451970(void); void sub_4519C0(void);
int sub_451BE0(void*); void *sub_451CA0(void*);
int sub_451F30(void*,int); int sub_451F90(void*); void *sub_451FE0(void*);
int sub_452010(void); void sub_452050(void*); int sub_452120(void*);
void sub_452190(void*); void *sub_4521A0(int);
uintptr_t sub_4521F0(void); void *sub_452230(void);
void *nox_xxx_draw_452270(int); int sub_4522A0(void*);
void *nox_xxx_draw_452300(void*); int sub_4523D0(void*);
void *sub_452410(void*); int sub_452490(void*); void sub_452510(void*);
int64_t sub_452690(void*,int64_t,int); int sub_4526D0(void*); int sub_4526F0(void*);
void *sub_452810(int,char);
int nox_thing_read_AVNT_452890(nox_memfile*,void*);
int nox_thing_skip_AVNT_452B00(nox_memfile*);
int nox_thing_skip_AVNT_inner_452B30(nox_memfile*);
int sub_452BD0(nox_memfile*,char*);
void sub_452DC0(int,int,int); void sub_452E10(int,int,int);
void *sub_452E90(void*,void*); void *sub_452EB0(void*);
int sub_452EE0(void*,int); unsigned int sub_452F10(void*,int);
int sub_452F50(void*,int); void *sub_452F80(void*,int);
int sub_452FA0(int); int sub_452FE0(void*,int);
void sub_453050(void); void nox_xxx____setargv_9_453060(void);
int sub_453070(void); int sub_453080(char);

/* Necessary packed-node dependencies formerly in client__audio__audevent.c.
 * Root guards those five original bodies and updates that private header. */
void *sub_451CF0(void*); int sub_451DC0(void*); int sub_451E80(void*);
int sub_452580(void*); int sub_452770(void*);
#endif
