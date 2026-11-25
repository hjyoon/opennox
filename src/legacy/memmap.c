#define NOX_IN_MEMMAP
#include <inttypes.h>
#include <stdio.h>
#include <stdlib.h>

// Width used by the generic legacy accessor.
#define MEMLOG_UNK_SIZE sizeof(int)

#ifndef NOX_CGO_MEMMAP
extern void* nox_legacy_memory_pointer(uintptr_t base, uintptr_t off, uintptr_t size);

void* mem_getPtrSize(uintptr_t base, uintptr_t off, uintptr_t size) {
	void* ptr = nox_legacy_memory_pointer(base, off, size);
	if (ptr) {
		return ptr;
	}
	fprintf(stderr, "Invalid memory access! Requested = %" PRIxPTR "+%" PRIuPTR "[%" PRIuPTR "]\n",
	        base, off, size);
	abort();
}
#else
void* mem_getPtrSize(uintptr_t base, uintptr_t off, uintptr_t size);
#endif

// defined in the header, emits the signature; we need an implementation now
#undef MEM_FUNC_PTR

#ifndef NOX_CGO_MEMMAP
void* mem_getPtr(uintptr_t base, uintptr_t off) { return mem_getPtrSize(base, off, MEMLOG_UNK_SIZE); }
#endif // NOX_CGO_MEMMAP
#define MEM_FUNC_PTR(T, NAME, SIZE)                                                                                    \
	T* NAME(uintptr_t base, uintptr_t off) { return (T*)mem_getPtrSize(base, off, SIZE); }

MEM_FUNC_PTR(uint8_t, mem_getU8Ptr, 1)
MEM_FUNC_PTR(int8_t, mem_getI8Ptr, 1)
MEM_FUNC_PTR(uint16_t, mem_getU16Ptr, 2)
MEM_FUNC_PTR(int16_t, mem_getI16Ptr, 2)
MEM_FUNC_PTR(uint32_t, mem_getU32Ptr, 4)
MEM_FUNC_PTR(int32_t, mem_getI32Ptr, 4)
MEM_FUNC_PTR(uint64_t, mem_getU64Ptr, 8)
MEM_FUNC_PTR(int64_t, mem_getI64Ptr, 8)
MEM_FUNC_PTR(float, mem_getFloatPtr, 4)
MEM_FUNC_PTR(double, mem_getDoublePtr, 8)
