// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert only TeleportWake's native object, data and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_teleport_wake_collide_data_t) == 8,
	"TeleportWake collide-data size");
_Static_assert(offsetof(nox_teleport_wake_collide_data_t, destination_x) == 0,
	"TeleportWake X offset");
_Static_assert(offsetof(nox_teleport_wake_collide_data_t, destination_y) == 4,
	"TeleportWake Y offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, buffs) == (sizeof(void*) == 4 ? 340 : 344),
	"object buffs offset");
_Static_assert(offsetof(nox_object_t, owner) == (sizeof(void*) == 4 ? 508 : 552),
	"object owner offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideTeleportWake_4EAE30),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"TeleportWake callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideTeleportWake_4EAE30(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

enum teleport_wake_event {
	EVENT_DATA = 1,
	EVENT_ANCHORED,
	EVENT_QUEST,
	EVENT_OWNER,
	EVENT_OWNER_CLASS,
	EVENT_TARGET_CLASS,
	EVENT_INVISIBLE_PRE,
	EVENT_POSITION_PRE,
	EVENT_FX_PRE,
	EVENT_AUDIO_PRE,
	EVENT_TELEPORT,
	EVENT_INVISIBLE_POST,
	EVENT_POSITION_POST,
	EVENT_FX_POST,
	EVENT_AUDIO_POST,
};

static int events[32];
static size_t event_count;
static int reference_error;
static int invisible_calls;
static int audio_calls;
static nox_object_t* reference_source;
static nox_object_t* reference_target;
static nox_object_t player_owner;
static nox_teleport_wake_collide_data_t* cached_data;
static nox_teleport_wake_collide_data_t replacement_data;

static void record_event(int event) {
	events[event_count++] = event;
}

static int reference_has_enchant(nox_object_t* obj, uint32_t enchant) {
	if (enchant == UINT32_C(14)) {
		record_event(EVENT_ANCHORED);
		return (obj->buffs & (UINT32_C(1) << 14)) != 0;
	}
	if (enchant == 0) {
		invisible_calls++;
		record_event(invisible_calls == 1 ? EVENT_INVISIBLE_PRE : EVENT_INVISIBLE_POST);
		return (obj->buffs & UINT32_C(1)) != 0;
	}
	reference_error = 1;
	return 0;
}

static int reference_quest_mode(void) {
	record_event(EVENT_QUEST);
	reference_source->owner = &player_owner;
	reference_target->obj_class = UINT32_C(0x03001016);
	return 1;
}

static void reference_point_fx(uint32_t id, nox_pointf* pos) {
	if (pos != (nox_pointf*)&reference_target->x) {
		reference_error = 2;
	}
	if (id == UINT32_C(138)) {
		record_event(EVENT_FX_PRE);
		reference_source->collide_data = &replacement_data;
		reference_target->obj_class = 0;
		reference_target->buffs |= UINT32_C(1);
	} else if (id == UINT32_C(137)) {
		record_event(EVENT_FX_POST);
		if (pos->x != -123.5f || pos->y != 456.25f) {
			reference_error = 3;
		}
	} else {
		reference_error = 4;
	}
}

static void reference_audio(uint32_t id, nox_object_t* obj) {
	audio_calls++;
	record_event(audio_calls == 1 ? EVENT_AUDIO_PRE : EVENT_AUDIO_POST);
	if (id != UINT32_C(147) || obj != reference_target) {
		reference_error = 5;
	}
	if (audio_calls == 1) {
		cached_data->destination_x = -123.5f;
		cached_data->destination_y = 456.25f;
	}
}

static void reference_teleport(nox_object_t* obj, nox_teleport_wake_collide_data_t* destination) {
	record_event(EVENT_TELEPORT);
	if (obj != reference_target || destination != cached_data ||
		destination->destination_x != -123.5f || destination->destination_y != 456.25f) {
		reference_error = 6;
	}
	obj->x = destination->destination_x;
	obj->y = destination->destination_y;
	obj->buffs &= ~UINT32_C(1);
}

static void teleport_wake_reference(nox_object_t* source, nox_object_t* target) {
	nox_teleport_wake_collide_data_t* destination;
	nox_object_t* owner;

	record_event(EVENT_DATA);
	destination = source->collide_data;
	if (target == NULL) {
		return;
	}
	if (reference_has_enchant(target, UINT32_C(14))) {
		return;
	}
	if (reference_quest_mode()) {
		record_event(EVENT_OWNER);
		owner = source->owner;
		if (owner != NULL) {
			record_event(EVENT_OWNER_CLASS);
			if (((uint8_t)owner->obj_class & UINT8_C(4)) == 0) {
				return;
			}
		}
	}
	record_event(EVENT_TARGET_CLASS);
	if ((target->obj_class & UINT32_C(0x03001016)) == 0) {
		return;
	}
	if (!reference_has_enchant(target, 0)) {
		record_event(EVENT_POSITION_PRE);
		reference_point_fx(UINT32_C(138), (nox_pointf*)&target->x);
	}
	reference_audio(UINT32_C(147), target);
	reference_teleport(target, destination);
	if (!reference_has_enchant(target, 0)) {
		record_event(EVENT_POSITION_POST);
		reference_point_fx(UINT32_C(137), (nox_pointf*)&target->x);
	}
	reference_audio(UINT32_C(147), target);
}

int main(void) {
	static const int expected[] = {
		EVENT_DATA, EVENT_ANCHORED, EVENT_QUEST, EVENT_OWNER, EVENT_OWNER_CLASS,
		EVENT_TARGET_CLASS, EVENT_INVISIBLE_PRE, EVENT_POSITION_PRE, EVENT_FX_PRE,
		EVENT_AUDIO_PRE, EVENT_TELEPORT, EVENT_INVISIBLE_POST, EVENT_POSITION_POST,
		EVENT_FX_POST, EVENT_AUDIO_POST,
	};
	nox_teleport_wake_collide_data_t data = {
		.destination_x = 10.0f,
		.destination_y = -20.0f,
	};
	nox_object_t old_owner = {.obj_class = UINT32_C(0x8)};
	nox_object_t source = {.owner = &old_owner, .collide_data = &data};
	nox_object_t target = {.x = 1.5f, .y = -2.5f};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideTeleportWake_4EAE30(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}

	player_owner.obj_class = UINT32_C(0x91020304);
	reference_source = &source;
	reference_target = &target;
	cached_data = &data;
	replacement_data.destination_x = 30.0f;
	replacement_data.destination_y = -40.0f;
	teleport_wake_reference(&source, &target);
	if (reference_error != 0 || event_count != sizeof(expected) / sizeof(expected[0])) {
		return 2;
	}
	for (size_t i = 0; i < event_count; i++) {
		if (events[i] != expected[i]) {
			return 3;
		}
	}
	if (source.collide_data != &replacement_data || target.obj_class != 0 ||
		target.x != -123.5f || target.y != 456.25f ||
		collision[0] != 3.5f || collision[1] != -8.25f) {
		return 4;
	}

	event_count = 0;
	source.collide_data = NULL;
	teleport_wake_reference(&source, NULL);
	if (event_count != 1 || events[0] != EVENT_DATA) {
		return 5;
	}
	return 0;
}
