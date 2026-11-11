#ifndef OPENNOX_NATIVE_FIXTURE_ALLOC_H
#define OPENNOX_NATIVE_FIXTURE_ALLOC_H

#include <stdint.h>
#include <stdlib.h>
#if defined(_WIN64)
#include <windows.h>
#endif

/* Explicit fixture policy: exercise native bits above DWORD on Windows/amd64,
 * whose race executable can have a low CRT heap. Product calloc/free and
 * fixture scalar widths are unchanged. Every caller must use the matched
 * release below; this memory must never be handed to an owning CRT free. */
static inline void* nox_native_fixture_calloc_high(size_t count, size_t size) {
#if defined(_WIN64)
	if (!count || !size || count > SIZE_MAX / size) return NULL;
	void* ptr = VirtualAlloc(NULL, count * size,
	                         MEM_RESERVE | MEM_COMMIT | MEM_TOP_DOWN, PAGE_READWRITE);
	if (ptr && (uintptr_t)ptr <= UINT32_MAX) {
		VirtualFree(ptr, 0, MEM_RELEASE);
		return NULL;
	}
	return ptr;
#else
	return calloc(count, size);
#endif
}

static inline void nox_native_fixture_free_high(void* ptr) {
#if defined(_WIN64)
	if (ptr && !VirtualFree(ptr, 0, MEM_RELEASE)) abort();
#else
	free(ptr);
#endif
}

#endif
