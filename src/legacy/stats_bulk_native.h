#ifndef NOX_STATS_BULK_NATIVE_H
#define NOX_STATS_BULK_NATIVE_H

#include <stdint.h>

// Keep the report's scalar prefix and eight ABI32 memmap slots unchanged.
// Address-bearing columns live separately, at the host's pointer width.
typedef struct nox_stats_columns_native {
	char** names;
	uint32_t* ips;
	uint32_t* teams;
	uint8_t* classes;
	uint8_t* active;
	uint32_t* durations;
	uint8_t* participants;
	uint8_t* pairs;
} nox_stats_columns_native;

char* nox_stats_bulk_flush_native(void);

#endif
