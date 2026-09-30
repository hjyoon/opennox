#ifndef NOX_PORT_QUEST_WIN_SCREEN_450770_H
#define NOX_PORT_QUEST_WIN_SCREEN_450770_H

#include <stddef.h>
#include <stdint.h>

// Native counterpart of the six PE32 records at 005D4594+832364. Wire
// fields retain their widths; only the player identity occupies native space.
typedef struct nox_quest_stats_row_450770 {
	struct nox_playerInfo* player;
	uint16_t kills;
	uint16_t generators;
	uint16_t secrets;
	uint16_t coop_secrets;
	uint32_t score;
} nox_quest_stats_row_450770;

_Static_assert(offsetof(nox_quest_stats_row_450770, kills) == sizeof(void*), "Quest kills offset");
_Static_assert(offsetof(nox_quest_stats_row_450770, generators) == sizeof(void*) + 2, "Quest generators offset");
_Static_assert(offsetof(nox_quest_stats_row_450770, secrets) == sizeof(void*) + 4, "Quest secrets offset");
_Static_assert(offsetof(nox_quest_stats_row_450770, coop_secrets) == sizeof(void*) + 6, "Quest cooperative secrets offset");
_Static_assert(offsetof(nox_quest_stats_row_450770, score) == sizeof(void*) + 8, "Quest score offset");
_Static_assert(sizeof(nox_quest_stats_row_450770) == (sizeof(void*) == 4 ? 16 : 24), "Quest native row size");

extern nox_quest_stats_row_450770 nox_quest_stats_450770[6];

int nox_xxx_clientQuestWinScreen_450770(const unsigned char* packet);

#endif
