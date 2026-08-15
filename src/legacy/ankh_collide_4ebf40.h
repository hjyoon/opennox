#ifndef NOX_ANKH_COLLIDE_4EBF40_H
#define NOX_ANKH_COLLIDE_4EBF40_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_ankh_history_record_t {
	uint16_t name[25];
	uint8_t player_class;
	uint8_t serial[25];
	uint32_t frame;
} nox_ankh_history_record_t;

typedef struct nox_ankh_init_data_t {
	nox_ankh_history_record_t records[64];
	uint8_t next;
	uint8_t reserved[3];
} nox_ankh_init_data_t;

// Native-width view of the Player tail beginning at original offset 4792.
// The five object references retain their original Win32 offsets and widen
// naturally on 64-bit targets.
typedef struct nox_ankh_player_tail_t {
	uint32_t quest_state;
	nox_object_t* quest_ankhs[5];
	uint32_t reserved[3];
} nox_ankh_player_tail_t;

// PlayerUpdateData through ExtraLives. The shared prefix ends immediately
// before original offset 320; its native 64-bit form ends at offset 400.
typedef struct nox_ankh_player_update_prefix_t {
	uint8_t prefix[sizeof(void*) == 4 ? 320 : 400];
	uint32_t extra_lives;
} nox_ankh_player_update_prefix_t;

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

void nox_xxx_collideAnkhQuest_4EBF40(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_ANKH_COLLIDE_4EBF40_H
