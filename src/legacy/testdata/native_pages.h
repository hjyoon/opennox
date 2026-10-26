#ifndef OPENNOX_NATIVE_TEST_PAGES_H
#define OPENNOX_NATIVE_TEST_PAGES_H

#include <stddef.h>
#ifdef _WIN32
#ifndef WIN32_LEAN_AND_MEAN
#define WIN32_LEAN_AND_MEAN
#endif
#include <windows.h>
static inline size_t native_fixture_page_size(void) {
	SYSTEM_INFO info;
	GetSystemInfo(&info);
	return (size_t)info.dwPageSize;
}
static inline void* native_fixture_reserve(size_t size) {
	return VirtualAlloc(NULL,size,MEM_RESERVE,PAGE_NOACCESS);
}
static inline int native_fixture_commit(void* address, size_t size) {
	return VirtualAlloc(address,size,MEM_COMMIT,PAGE_READWRITE)==address ? 0 : -1;
}
static inline int native_fixture_release(void* address, size_t size) {
	(void)size;
	return VirtualFree(address,0,MEM_RELEASE) ? 0 : -1;
}
#else
#include <sys/mman.h>
#include <unistd.h>
static inline size_t native_fixture_page_size(void) {
	return (size_t)sysconf(_SC_PAGESIZE);
}
static inline void* native_fixture_reserve(size_t size) {
	void* address=mmap(NULL,size,PROT_NONE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
	return address==MAP_FAILED ? NULL : address;
}
static inline int native_fixture_commit(void* address, size_t size) {
	return mprotect(address,size,PROT_READ|PROT_WRITE);
}
static inline int native_fixture_release(void* address, size_t size) {
	return munmap(address,size);
}
#endif

#endif
