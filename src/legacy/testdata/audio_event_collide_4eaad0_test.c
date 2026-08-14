// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert AudioEventCollide's native object, callback, and parser ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>
#include <string.h>

_Static_assert(sizeof(nox_audio_event_collide_data_t) == 4,
	"AudioEventCollide data size");
_Static_assert(offsetof(nox_audio_event_collide_data_t, sound) == 0,
	"AudioEventCollide sound offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140),
	"object collision timestamp offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&sub_4EAAD0),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"AudioEventCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&sub_536DA0),
		int (*)(char*, nox_audio_event_collide_data_t*)),
	"AudioEventCollide parser data width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void sub_4EAAD0(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int nox_xxx_utilFindSound_40AF50(char* name) {
	if (strcmp(name, "Bell") == 0) {
		return 417;
	}
	if (strcmp(name, "SignedBits") == 0) {
		return -7;
	}
	return 0;
}

static uint32_t audio_event_reference(
	nox_object_t* source,
	nox_object_t* target,
	uint32_t frame,
	int* stored_before_data) {
	if (target == NULL || (target->obj_class & UINT32_C(4)) == 0) {
		return 0;
	}
	if (frame <= source->field_34 + UINT32_C(30)) {
		return 0;
	}
	source->field_34 = frame;
	*stored_before_data = source->field_34 == frame;
	return ((nox_audio_event_collide_data_t*)source->collide_data)->sound;
}

int main(void) {
	nox_audio_event_collide_data_t data = {.sound = UINT32_C(0x11223344)};
	nox_object_t source = {.collide_data = &data};
	nox_object_t player = {.obj_class = UINT32_C(0x84)};
	nox_object_t non_player = {.obj_class = UINT32_C(0x80)};
	float collision[2] = {3.5f, -8.25f};
	int stored_before_data = 0;
	char bell[] = "Bell trailing";
	char missing[] = "Missing";
	char signed_bits[] = "SignedBits";

	sub_4EAAD0(&source, &player, collision);
	if (seen_source != &source || seen_target != &player || seen_collision != collision) {
		return 1;
	}
	if (audio_event_reference(&source, &non_player, 31, &stored_before_data) != 0 ||
		source.field_34 != 0 || stored_before_data != 0) {
		return 2;
	}
	if (audio_event_reference(&source, &player, 30, &stored_before_data) != 0 ||
		source.field_34 != 0) {
		return 3;
	}
	if (audio_event_reference(&source, &player, 31, &stored_before_data) != UINT32_C(0x11223344) ||
		source.field_34 != 31 || stored_before_data != 1) {
		return 4;
	}
	source.field_34 = UINT32_MAX;
	stored_before_data = 0;
	if (audio_event_reference(&source, &player, 29, &stored_before_data) != 0 ||
		source.field_34 != UINT32_MAX) {
		return 5;
	}
	if (audio_event_reference(&source, &player, 30, &stored_before_data) != UINT32_C(0x11223344) ||
		source.field_34 != 30 || stored_before_data != 1) {
		return 6;
	}
	if (sub_536DA0(bell, &data) != 1 || data.sound != 417) {
		return 7;
	}
	if (sub_536DA0(missing, &data) != 0 || data.sound != 0) {
		return 8;
	}
	if (sub_536DA0(signed_bits, &data) != 1 || data.sound != UINT32_MAX - 6) {
		return 9;
	}
	return 0;
}
