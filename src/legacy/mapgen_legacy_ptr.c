#include "mapgen_legacy_ptr.h"

#include <stdlib.h>

typedef struct nox_mapgen_legacy_ptr_entry {
	uint32_t token;
	void* ptr;
	struct nox_mapgen_legacy_ptr_entry* next;
} nox_mapgen_legacy_ptr_entry;

static nox_mapgen_legacy_ptr_entry* nox_mapgen_legacy_ptrs;
static uint32_t nox_mapgen_legacy_next_token = UINT32_C(0xF0000000);

static nox_mapgen_legacy_ptr_entry* nox_mapgenLegacyPtrFindToken(uint32_t token) {
	for (nox_mapgen_legacy_ptr_entry* it = nox_mapgen_legacy_ptrs; it; it = it->next) {
		if (it->token == token) {
			return it;
		}
	}
	return NULL;
}

uint32_t nox_mapgenLegacyPtrRegister(void* ptr) {
	if (!ptr) {
		return 0;
	}
	for (nox_mapgen_legacy_ptr_entry* it = nox_mapgen_legacy_ptrs; it; it = it->next) {
		if (it->ptr == ptr) {
			return it->token;
		}
	}

	uint32_t token = (uint32_t)(uintptr_t)ptr;
	if (!token || nox_mapgenLegacyPtrFindToken(token)) {
		do {
			token = nox_mapgen_legacy_next_token++;
		} while (!token || nox_mapgenLegacyPtrFindToken(token));
	}

	nox_mapgen_legacy_ptr_entry* entry = (nox_mapgen_legacy_ptr_entry*)malloc(sizeof(*entry));
	if (!entry) {
		return 0;
	}
	entry->token = token;
	entry->ptr = ptr;
	entry->next = nox_mapgen_legacy_ptrs;
	nox_mapgen_legacy_ptrs = entry;
	return token;
}

void* nox_mapgenLegacyPtrResolve(uintptr_t value) {
	if (!value) {
		return NULL;
	}
#if UINTPTR_MAX > UINT32_MAX
	if (value > UINT32_MAX) {
		return (void*)value;
	}
#endif
	nox_mapgen_legacy_ptr_entry* entry = nox_mapgenLegacyPtrFindToken((uint32_t)value);
	return entry ? entry->ptr : (void*)value;
}

void nox_mapgenLegacyPtrForget(void* ptr) {
	if (!ptr) {
		return;
	}
	nox_mapgen_legacy_ptr_entry** link = &nox_mapgen_legacy_ptrs;
	while (*link) {
		if ((*link)->ptr == ptr) {
			nox_mapgen_legacy_ptr_entry* entry = *link;
			*link = entry->next;
			free(entry);
			return;
		}
		link = &(*link)->next;
	}
}
