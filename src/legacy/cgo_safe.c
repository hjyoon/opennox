//go:build safe
#include <stdlib.h>
#include <string.h>

#include "_cgo_export.h"

// Keep the destination pointer on the C side of the callback.
void* nox_memset(void* dst, int val, size_t size) {
	nox_memset_go(dst, val, size);
	return dst;
}

void* nox_memcpy(void* dst, const void* src, size_t size) {
	nox_memcpy_go(dst, (void*)src, size);
	return dst;
}

char* nox_strcpy(char* dst, const char* src) {
	nox_strcpy_go(dst, (char*)src);
	return dst;
}

char* nox_strcat(char* dst, const char* src) {
	nox_strcat_go(dst, (char*)src);
	return dst;
}
