#ifndef OPENNOX_DRAW_ALLOCATION_H
#define OPENNOX_DRAW_ALLOCATION_H
#include "defs.h"
void* nox_draw_memory_calloc(size_t count, size_t size);
void nox_draw_memory_free(void* ptr);
int nox_draw_parse_with_policy(void* fn, nox_thing* thing, nox_memfile* file, char* scratch, int high);
#endif
