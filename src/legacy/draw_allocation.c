#include "draw_allocation.h"
#include <stdint.h>
#include <stdlib.h>

extern void* nox_draw_memory_calloc_high_go(size_t count, size_t size);
extern void nox_draw_memory_release_go(void* ptr);

// The policy is scoped to one synchronous parser call on its actual C thread.
// The normal parser still uses CRT calloc. Destruction follows actual ownership.
static _Thread_local int nox_draw_high_policy;

void* nox_draw_memory_calloc(size_t count, size_t size) {
	if (!nox_draw_high_policy || !count || !size || count > SIZE_MAX / size) {
		return calloc(count, size);
	}
	return nox_draw_memory_calloc_high_go(count, size);
}

void nox_draw_memory_free(void* ptr) {
	nox_draw_memory_release_go(ptr);
}

int nox_draw_parse_with_policy(void* fn, nox_thing* thing, nox_memfile* file, char* scratch, int high) {
	int previous = nox_draw_high_policy;
	nox_draw_high_policy = high;
	int result = ((bool (*)(nox_thing*, nox_memfile*, char*))fn)(thing, file, scratch);
	nox_draw_high_policy = previous;
	return result;
}
