#ifndef NOX_STATS_START_NATIVE_H
#define NOX_STATS_START_NATIVE_H

#include <stdint.h>

// Quest's packed header keeps eleven DWORD slots. Native pointers live in a
// separate context; the shared original column count remains a DWORD.
typedef struct nox_stats_quest_columns_native {
	char** names;
	uint32_t* ips;
	uint8_t* classes;
	uint32_t* scores[8];
} nox_stats_quest_columns_native;

void nox_stats_start_header_native(void);
void* nox_stats_quest_columns_clear_native(void);
int nox_stats_session_report_native(int mode);

#endif
