#ifndef OPENNOX_MAPGEN_ALLOCATION_H
#define OPENNOX_MAPGEN_ALLOCATION_H

#include <stddef.h>

/* Transferred native graph records must use this matched allocation owner.
 * The default service is the ordinary CRT; high memory is fixture scoped. */
void* nox_mapgen_alloc_memory(size_t count, size_t size);
void nox_mapgen_release_memory(void* ptr);

#endif
