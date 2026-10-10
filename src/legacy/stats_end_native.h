#ifndef NOX_STATS_END_NATIVE_H
#define NOX_STATS_END_NATIVE_H

#include "stats_start_native.h"

nox_stats_quest_columns_native* nox_stats_quest_columns_native_get(void);
void* nox_stats_field_native_prepend(void* head, const char* name, uint16_t type,
									uint16_t length, const void* data);
uint16_t* nox_stats_packet_native_finish(void* head, uint32_t* length, uint32_t* sequence);
int nox_stats_quest_end_native(void);

#endif
