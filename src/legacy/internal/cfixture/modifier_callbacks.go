package cfixture

/*
#include "GAME3_2.h"
#include "common__object__modifier.h"

typedef float* (*modifier_defend_fn)(void*, nox_object_t*, uintptr_t, nox_object_t*, uintptr_t, float*);
typedef float* (*modifier_multiply_fn)(void*, nox_object_t*, nox_object_t*, nox_object_t*, float*);
typedef nox_object_t* (*modifier_projectile_fn)(void*, nox_object_t*, nox_object_t*, nox_object_t*, nox_object_t*);
_Static_assert(_Generic(&sub_4E0370, modifier_defend_fn: 1, default: 0), "armor multiplier ABI");
_Static_assert(_Generic(&sub_4E0380, modifier_defend_fn: 1, default: 0), "durability multiplier ABI");
_Static_assert(_Generic(&nox_xxx_effectDamageMultiplier_4E04C0, modifier_multiply_fn: 1, default: 0), "damage multiplier ABI");
_Static_assert(_Generic(&nox_xxx_effectProjectileSpeed_4E09B0, modifier_projectile_fn: 1, default: 0), "projectile speed ABI");

static int modifier_noop_callbacks_distinct(void) {
	typedef void (*callback_fn)(void);
	callback_fn volatile callbacks[] = {
		nullsub_22, nullsub_36, nullsub_38, nullsub_39, nullsub_40,
		nullsub_41, nullsub_42, nullsub_43, nullsub_44,
	};
	for (size_t i = 0; i < sizeof(callbacks) / sizeof(callbacks[0]); ++i) {
		callback_fn callback = callbacks[i];
		for (size_t j = 0; j < i; ++j) {
			if (callback == callbacks[j]) {
				return 0;
			}
		}
		callback();
	}
	return 1;
}
*/
import "C"

func ModifierReservedCallbacksDistinct() bool {
	return C.modifier_noop_callbacks_distinct() != 0
}
