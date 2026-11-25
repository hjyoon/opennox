#ifndef NOX_SECRET_WALL_H
#define NOX_SECRET_WALL_H

#include "defs.h"

// GAME.EXE uses a 32-byte record with pointers at offsets 0 and 12. Keep the
// scalar fields fixed and let only those pointers grow on native 64-bit builds.
typedef struct nox_secret_wall_t {
	struct nox_secret_wall_t* next;
	int32_t x;
	int32_t y;
	void* wall;
	uint32_t open_wait;
	uint8_t flags;
	uint8_t state;
	uint8_t open_delay;
	uint8_t reserved;
	uint32_t last_open;
	uint32_t player_bits;
} nox_secret_wall_t;

_Static_assert(offsetof(nox_secret_wall_t, next) == 0, "wrong native offset of secret-wall next");
_Static_assert(offsetof(nox_secret_wall_t, x) == (sizeof(void*) == 4 ? 4 : 8),
	"wrong native offset of secret-wall X");
_Static_assert(offsetof(nox_secret_wall_t, y) == (sizeof(void*) == 4 ? 8 : 12),
	"wrong native offset of secret-wall Y");
_Static_assert(offsetof(nox_secret_wall_t, wall) == (sizeof(void*) == 4 ? 12 : 16),
	"wrong native offset of secret-wall wall pointer");
_Static_assert(offsetof(nox_secret_wall_t, open_wait) == (sizeof(void*) == 4 ? 16 : 24),
	"wrong native offset of secret-wall open wait");
_Static_assert(offsetof(nox_secret_wall_t, flags) == (sizeof(void*) == 4 ? 20 : 28),
	"wrong native offset of secret-wall flags");
_Static_assert(offsetof(nox_secret_wall_t, state) == (sizeof(void*) == 4 ? 21 : 29),
	"wrong native offset of secret-wall state");
_Static_assert(offsetof(nox_secret_wall_t, open_delay) == (sizeof(void*) == 4 ? 22 : 30),
	"wrong native offset of secret-wall open delay");
_Static_assert(offsetof(nox_secret_wall_t, last_open) == (sizeof(void*) == 4 ? 24 : 32),
	"wrong native offset of secret-wall last-open frame");
_Static_assert(offsetof(nox_secret_wall_t, player_bits) == (sizeof(void*) == 4 ? 28 : 36),
	"wrong native offset of secret-wall player bits");
_Static_assert(sizeof(nox_secret_wall_t) == (sizeof(void*) == 4 ? 32 : 40),
	"wrong native size of secret-wall record");

#endif // NOX_SECRET_WALL_H
