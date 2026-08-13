#include "../defs.h"

#ifdef NOX_REPLAY_CHECKSUM_4E76C0_NATIVE_LAYOUT
// Native 64-bit probes suppress unrelated Win32-only assertions while the
// header is parsed, then re-enable only the fields consumed by 004E7700.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8), "object type index offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(offsetof(nox_object_t, field_5) == (sizeof(void*) == 4 ? 20 : 24), "object field 5 offset");
_Static_assert(offsetof(nox_object_t, net_code) == (sizeof(void*) == 4 ? 36 : 40), "object net code offset");
_Static_assert(offsetof(nox_object_t, extent) == (sizeof(void*) == 4 ? 40 : 44), "object extent offset");
_Static_assert(offsetof(nox_object_t, script_id) == (sizeof(void*) == 4 ? 44 : 48), "object script id offset");
_Static_assert(offsetof(nox_object_t, field_13) == (sizeof(void*) == 4 ? 52 : 56), "object team id byte offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60), "object position x offset");
_Static_assert(offsetof(nox_object_t, y) == (sizeof(void*) == 4 ? 60 : 64), "object position y offset");
_Static_assert(offsetof(nox_object_t, new_x) == (sizeof(void*) == 4 ? 64 : 68), "object new position x offset");
_Static_assert(offsetof(nox_object_t, new_y) == (sizeof(void*) == 4 ? 68 : 72), "object new position y offset");
_Static_assert(offsetof(nox_object_t, prev_x) == (sizeof(void*) == 4 ? 72 : 76), "object previous position x offset");
_Static_assert(offsetof(nox_object_t, prev_y) == (sizeof(void*) == 4 ? 76 : 80), "object previous position y offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84), "object velocity x offset");
_Static_assert(offsetof(nox_object_t, vel_y) == (sizeof(void*) == 4 ? 84 : 88), "object velocity y offset");
_Static_assert(offsetof(nox_object_t, force_x) == (sizeof(void*) == 4 ? 88 : 92), "object force x offset");
_Static_assert(offsetof(nox_object_t, force_y) == (sizeof(void*) == 4 ? 92 : 96), "object force y offset");
_Static_assert(offsetof(nox_object_t, float_24) == (sizeof(void*) == 4 ? 96 : 100), "object position 24 x offset");
_Static_assert(offsetof(nox_object_t, float_25) == (sizeof(void*) == 4 ? 100 : 104), "object position 24 y offset");
_Static_assert(offsetof(nox_object_t, z) == (sizeof(void*) == 4 ? 104 : 108), "object z offset");
_Static_assert(offsetof(nox_object_t, field_27) == (sizeof(void*) == 4 ? 108 : 112), "object field 27 offset");
_Static_assert(offsetof(nox_object_t, mass) == (sizeof(void*) == 4 ? 120 : 124), "object mass offset");
_Static_assert(offsetof(nox_object_t, direction1) == (sizeof(void*) == 4 ? 124 : 128), "object direction 1 offset");
_Static_assert(offsetof(nox_object_t, direction2) == (sizeof(void*) == 4 ? 126 : 130), "object direction 2 offset");
_Static_assert(offsetof(nox_object_t, field_32) == (sizeof(void*) == 4 ? 128 : 132), "object field 32 offset");
_Static_assert(offsetof(nox_object_t, field_33) == (sizeof(void*) == 4 ? 132 : 136), "object field 33 offset");
_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140), "object field 34 offset");
_Static_assert(offsetof(nox_object_t, field_37) == (sizeof(void*) == 4 ? 148 : 152), "object field 37 offset");
_Static_assert(offsetof(nox_object_t, field_38) == (sizeof(void*) == 4 ? 152 : 156), "object field 38 offset");
_Static_assert(offsetof(nox_object_t, field_62) == (sizeof(void*) == 4 ? 248 : 252), "object field 62 offset");
_Static_assert(offsetof(nox_object_t, buffs) == (sizeof(void*) == 4 ? 340 : 344), "object buffs offset");
_Static_assert(offsetof(nox_object_t, object_next) == (sizeof(void*) == 4 ? 444 : 448), "object next offset");
_Static_assert(offsetof(nox_object_t, health_data) == (sizeof(void*) == 4 ? 556 : 616), "object health pointer offset");

typedef struct nox_replay_health_prefix_4e7700 {
	uint16_t cur;
	uint16_t field_2;
	uint16_t max;
	uint16_t field_6;
	uint32_t field_8;
	uint32_t field_12;
	uint32_t field_16;
} nox_replay_health_prefix_4e7700;
_Static_assert(sizeof(nox_replay_health_prefix_4e7700) == 20, "health data size");
_Static_assert(offsetof(nox_replay_health_prefix_4e7700, cur) == 0, "health current offset");
_Static_assert(offsetof(nox_replay_health_prefix_4e7700, field_2) == 2, "health field 2 offset");
_Static_assert(offsetof(nox_replay_health_prefix_4e7700, max) == 4, "health maximum offset");
#endif
