// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert only AnkhCollide's native callback and record boundaries.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, x) ==
	(sizeof(void*) == 4 ? 56 : 60), "object position offset");
_Static_assert(offsetof(nox_object_t, field_34) ==
	(sizeof(void*) == 4 ? 136 : 140), "object frame offset");
_Static_assert(offsetof(nox_object_t, init_data) ==
	(sizeof(void*) == 4 ? 692 : 760), "object InitData offset");
_Static_assert(offsetof(nox_object_t, func_pickup) ==
	(sizeof(void*) == 4 ? 708 : 792), "object Pickup offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object UpdateData offset");
_Static_assert(sizeof(nox_ankh_history_record_t) == 80,
	"Ankh history record size");
_Static_assert(offsetof(nox_ankh_history_record_t, player_class) == 50,
	"Ankh history class offset");
_Static_assert(offsetof(nox_ankh_history_record_t, serial) == 51,
	"Ankh history serial offset");
_Static_assert(offsetof(nox_ankh_history_record_t, frame) == 76,
	"Ankh history frame offset");
_Static_assert(sizeof(nox_ankh_init_data_t) == 5124,
	"Ankh InitData size");
_Static_assert(offsetof(nox_ankh_init_data_t, next) == 5120,
	"Ankh InitData next-index offset");
_Static_assert(offsetof(nox_ankh_player_tail_t, quest_ankhs) ==
	(sizeof(void*) == 4 ? 4 : 8), "Ankh Player slot offset");
_Static_assert(sizeof(nox_ankh_player_tail_t) ==
	(sizeof(void*) == 4 ? 36 : 64), "Ankh Player tail size");
_Static_assert(offsetof(nox_ankh_player_update_prefix_t, extra_lives) ==
	(sizeof(void*) == 4 ? 320 : 400), "Ankh extra-lives offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideAnkhQuest_4EBF40),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"AnkhCollide callback three-pointer ABI");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideAnkhQuest_4EBF40(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_ankh_init_data_t data = {
		.records[63] = {
			.player_class = UINT8_C(3),
			.serial = {UINT8_C(0xa5)},
			.frame = UINT32_C(0xfedcba98),
		},
		.next = UINT8_C(63),
	};
	nox_ankh_player_update_prefix_t update = {
		.extra_lives = UINT32_C(0x89abcdef),
	};
	nox_ankh_player_tail_t player_tail = {
		.quest_state = UINT32_C(0x11223344),
	};
	nox_object_t source = {
		.x = 12.5f,
		.y = -7.25f,
		.field_34 = UINT32_C(0x55667788),
		.init_data = &data,
	};
	nox_object_t target = {
		.obj_class = UINT32_C(0x40000004),
		.data_update = &update,
	};
	float collision[2] = {3.5f, -8.25f};

	player_tail.quest_ankhs[4] = &source;
	nox_xxx_collideAnkhQuest_4EBF40(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || source.init_data != &data ||
		target.data_update != &update || player_tail.quest_ankhs[4] != &source ||
		data.next != UINT8_C(63) || data.records[63].player_class != UINT8_C(3) ||
		data.records[63].serial[0] != UINT8_C(0xa5) ||
		data.records[63].frame != UINT32_C(0xfedcba98) ||
		update.extra_lives != UINT32_C(0x89abcdef) ||
		collision[0] != 3.5f || collision[1] != -8.25f) {
		return 1;
	}
	return 0;
}
