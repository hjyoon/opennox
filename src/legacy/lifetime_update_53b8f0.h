#ifndef NOX_LIFETIME_UPDATE_53B8F0_H
#define NOX_LIFETIME_UPDATE_53B8F0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_lifetime_update_data_t {
	uint32_t duration;
} nox_lifetime_update_data_t;

_Static_assert(offsetof(nox_lifetime_update_data_t, duration) == 0,
	"LifetimeUpdate duration moved");
_Static_assert(sizeof(nox_lifetime_update_data_t) == 4,
	"LifetimeUpdate data must remain four bytes");

void nox_xxx_updateLifetime_53B8F0(nox_object_t* source);

#endif // NOX_LIFETIME_UPDATE_53B8F0_H
