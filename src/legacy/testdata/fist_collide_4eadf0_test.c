// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert only the records and callback used by GAME.EXE 004EADF0.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <math.h>
#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_fist_update_data_t) == 4, "Fist update-data size");
_Static_assert(offsetof(nox_fist_update_data_t, damage) == 0, "Fist damage offset");
_Static_assert(offsetof(nox_object_t, z) == (sizeof(void*) == 4 ? 104 : 108), "object Z offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object Damage offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(__builtin_types_compatible_p(__typeof__(&nox_xxx_collideFist_4EADF0),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"FistCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideFist_4EADF0(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

static nox_object_t* parent_result;
static nox_object_t* parent_source;
static nox_object_t* mutation_target;
static nox_fist_update_data_t* mutation_data;
static nox_fist_update_data_t replacement_data;
static int event_index;
static int parent_event;
static int damage_event;

static nox_object_t* damage_target;
static nox_object_t* damage_parent;
static nox_object_t* damage_source;
static int32_t damage_value;
static int32_t damage_type;

static int fist_damage_new(
	nox_object_t* target,
	nox_object_t* parent,
	nox_object_t* source,
	int32_t damage,
	int32_t type) {
	damage_event = ++event_index;
	damage_target = target;
	damage_parent = parent;
	damage_source = source;
	damage_value = damage;
	damage_type = type;
	return -INT32_C(0x1234567);
}

static int fist_damage_old(
	nox_object_t* target,
	nox_object_t* parent,
	nox_object_t* source,
	int32_t damage,
	int32_t type) {
	(void)target;
	(void)parent;
	(void)source;
	(void)damage;
	(void)type;
	return INT32_C(0x7654321);
}

static nox_object_t* fist_parent_reference(nox_object_t* source) {
	parent_event = ++event_index;
	parent_source = source;
	mutation_data->damage = -9;
	source->data_update = &replacement_data;
	mutation_target->func_damage = fist_damage_new;
	return parent_result;
}

static void fist_reference(nox_object_t* source, nox_object_t* target) {
	if (!(source->z > 0.0f)) {
		nox_fist_update_data_t* data = source->data_update;
		if (target != NULL) {
			int32_t damage = data->damage;
			nox_object_t* parent = fist_parent_reference(source);
			int (*damage_fn)(nox_object_t*, nox_object_t*, nox_object_t*, int32_t, int32_t) =
				target->func_damage;
			(void)damage_fn(target, parent, source, damage, 2);
		}
	}
}

int main(void) {
	nox_fist_update_data_t data = {.damage = INT32_C(0x71234567)};
	nox_object_t source = {.z = NAN, .data_update = &data};
	nox_object_t target = {.func_damage = fist_damage_old};
	nox_object_t parent = {0};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideFist_4EADF0(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}

	parent_result = &parent;
	mutation_target = &target;
	mutation_data = &data;
	replacement_data.damage = -10;
	fist_reference(&source, &target);
	if (parent_event != 1 || damage_event != 2 || parent_source != &source ||
		damage_target != &target || damage_parent != &parent || damage_source != &source ||
		damage_value != INT32_C(0x71234567) || damage_type != 2 ||
		target.func_damage != fist_damage_new || source.data_update != &replacement_data) {
		return 2;
	}

	source.z = 1.0f;
	source.data_update = NULL;
	event_index = 0;
	fist_reference(&source, &target);
	if (event_index != 0) {
		return 3;
	}

	source.z = -0.0f;
	fist_reference(&source, NULL);
	if (event_index != 0 || collision[0] != 3.5f || collision[1] != -8.25f) {
		return 4;
	}
	return 0;
}
