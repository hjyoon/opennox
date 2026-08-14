#ifndef NOX_AUDIO_EVENT_COLLIDE_4EAAD0_H
#define NOX_AUDIO_EVENT_COLLIDE_4EAAD0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_audio_event_collide_data_t {
	uint32_t sound;
} nox_audio_event_collide_data_t;
_Static_assert(sizeof(nox_audio_event_collide_data_t) == 4,
	"wrong size of AudioEventCollide data structure!");
_Static_assert(offsetof(nox_audio_event_collide_data_t, sound) == 0,
	"wrong offset of AudioEventCollide sound!");

void sub_4EAAD0(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);
int sub_536DA0(char* args, nox_audio_event_collide_data_t* data);

#endif // NOX_AUDIO_EVENT_COLLIDE_4EAAD0_H
