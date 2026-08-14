#ifndef NOX_CHAKRAM_UPDATE_53DCC0_H
#define NOX_CHAKRAM_UPDATE_53DCC0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_chakram_update_data_t {
	uint32_t field_0;
	uint8_t reflections;
	uint8_t padding_5[3];
	nox_object_t* return_target;
	nox_object_t* last_hit;
	float owner_x;
	float owner_y;
	uint8_t return_state;
} nox_chakram_update_data_t;

_Static_assert(offsetof(nox_chakram_update_data_t, field_0) == 0,
	"wrong offset of Chakram field 0!");
_Static_assert(offsetof(nox_chakram_update_data_t, reflections) == 4,
	"wrong offset of Chakram reflections!");
_Static_assert(offsetof(nox_chakram_update_data_t, return_target) == 8,
	"wrong offset of Chakram return target!");
_Static_assert(offsetof(nox_chakram_update_data_t, last_hit) == 8 + sizeof(void*),
	"wrong offset of Chakram last hit!");
_Static_assert(offsetof(nox_chakram_update_data_t, owner_x) == 8 + 2 * sizeof(void*),
	"wrong offset of Chakram owner X!");
_Static_assert(offsetof(nox_chakram_update_data_t, owner_y) == 12 + 2 * sizeof(void*),
	"wrong offset of Chakram owner Y!");
_Static_assert(offsetof(nox_chakram_update_data_t, return_state) == 16 + 2 * sizeof(void*),
	"wrong offset of Chakram return state!");

#if UINTPTR_MAX == UINT32_MAX
_Static_assert(sizeof(nox_chakram_update_data_t) == 28,
	"wrong 32-bit size of Chakram update data!");
#elif UINTPTR_MAX == UINT64_MAX
_Static_assert(sizeof(nox_chakram_update_data_t) == 40,
	"wrong 64-bit size of Chakram update data!");
#else
#error unsupported pointer width for Chakram update data
#endif

void nox_xxx_updateChakramInMotion_53DCC0(nox_object_t* source);

#endif // NOX_CHAKRAM_UPDATE_53DCC0_H
