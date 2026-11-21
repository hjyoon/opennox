//go:build safe
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

void* mem_getPtr_go(unsigned int base, unsigned int off, bool* ok);
void* mem_getPtr(unsigned int base, unsigned int off) {
	bool ok = true;
	void* p = mem_getPtr_go(base, off, &ok);
	if (!ok) {
		abort();
	}
	return p;
}

void* mem_getPtrSize_go(unsigned int base, unsigned int off, unsigned int size, bool* ok);
void* mem_getPtrSize(unsigned int base, unsigned int off, unsigned int size) {
	bool ok = true;
	void* p = mem_getPtrSize_go(base, off, size, &ok);
	if (!ok) {
		abort();
	}
	return p;
}

// Keep the destination pointer on the C side of the callback.
void nox_memset_go(void* dst, int val, size_t size);
void* nox_memset(void* dst, int val, size_t size) {
	nox_memset_go(dst, val, size);
	return dst;
}

void nox_memcpy_go(void* dst, void* src, size_t size);
void* nox_memcpy(void* dst, const void* src, size_t size) {
	nox_memcpy_go(dst, (void*)src, size);
	return dst;
}

int nox_memcmp_go(void* ptr1, void* ptr2, size_t size);
int nox_memcmp(const void* ptr1, const void* ptr2, size_t size) { return nox_memcmp_go((void*)ptr1, (void*)ptr2, size); }

size_t nox_strlen_go(char* ptr);
size_t nox_strlen(const char* ptr) { return nox_strlen_go((char*)ptr); }

void nox_strcpy_go(char* dst, char* src);
char* nox_strcpy(char* dst, const char* src) {
	nox_strcpy_go(dst, (char*)src);
	return dst;
}

void nox_strcat_go(char* dst, char* src);
char* nox_strcat(char* dst, const char* src) {
	nox_strcat_go(dst, (char*)src);
	return dst;
}

int nox_strcmp_go(char* str1, char* str2);
int nox_strcmp(const char* str1, const char* str2) { return nox_strcmp_go((char*)str1, (char*)str2); }
