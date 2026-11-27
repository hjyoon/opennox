#ifndef NOX_STATS_START_NATIVE_H
#define NOX_STATS_START_NATIVE_H

#include <stddef.h>
#include <stdint.h>

// Quest's packed header keeps eleven DWORD slots. Native pointers live in a
// separate context; the shared original column count remains a DWORD.
typedef struct nox_stats_quest_columns_native {
	char** names;
	uint32_t* ips;
	uint8_t* classes;
	uint32_t* scores[8];
} nox_stats_quest_columns_native;
_Static_assert(offsetof(nox_stats_quest_columns_native, scores) == 3 * sizeof(void*), "quest stats score columns");
_Static_assert(sizeof(nox_stats_quest_columns_native) == 11 * sizeof(void*), "quest stats columns size");

void* nox_stats_quest_columns_clear_native(void);
int nox_stats_session_report_native(int mode);

#endif
