// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert the fixed record and pointer-native callback/parser boundaries.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_mana_drain_collide_data_t) == 8,
	"ManaDrainCollide data size");
_Static_assert(offsetof(nox_mana_drain_collide_data_t, amount) == 0,
	"ManaDrainCollide amount offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, field_542) == (sizeof(void*) == 4 ? 542 : 602),
	"object shared timer offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(offsetof(nox_player_update_data_t, mana_cur) == 4,
	"player-update current mana offset");
_Static_assert(offsetof(nox_player_update_data_t, mana_prev) == 6,
	"player-update previous mana offset");
_Static_assert(offsetof(nox_player_update_data_t, player) == (sizeof(void*) == 4 ? 276 : 336),
	"player-update player pointer offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideManadrain_4E9490),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"ManaDrainCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideManaDrainLoad_536E50),
		int (*)(char*, nox_mana_drain_collide_data_t*)),
	"ManaDrainCollide parser pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideManadrain_4E9490(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};
	nox_mana_drain_collide_data_t data = {
		.amount = UINT8_C(9),
		.reserved = {UINT8_C(1), UINT8_C(2), UINT8_C(3), UINT8_C(4),
			UINT8_C(5), UINT8_C(6), UINT8_C(7)},
	};
	char amount[] = "511 ignored";

	nox_xxx_collideManadrain_4E9490(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	if (!nox_xxx_collideManaDrainLoad_536E50(amount, &data)) {
		return 2;
	}
	if (data.amount != UINT8_C(255) || data.reserved[0] != UINT8_C(1) ||
		data.reserved[1] != UINT8_C(2) || data.reserved[2] != UINT8_C(3) ||
		data.reserved[3] != UINT8_C(4) || data.reserved[4] != UINT8_C(5) ||
		data.reserved[5] != UINT8_C(6) || data.reserved[6] != UINT8_C(7)) {
		return 3;
	}
	return 0;
}
