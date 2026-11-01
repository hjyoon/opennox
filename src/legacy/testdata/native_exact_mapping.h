#ifndef OPENNOX_NATIVE_TEST_EXACT_MAPPING_H
#define OPENNOX_NATIVE_TEST_EXACT_MAPPING_H

#ifdef __APPLE__
#include <stddef.h>
#include <stdint.h>
#include <mach/mach.h>
#include <mach/mach_vm.h>

/* Darwin mmap addresses are hints. Reserve the required native test address
 * with Mach instead. VM_FLAGS_FIXED refuses an occupied range; deliberately
 * omit VM_FLAGS_OVERWRITE, preserving other mappings and all exact-address
 * assertions in the original-derived fixtures. */
static void* opennox_native_test_exact_mapping(uintptr_t address, size_t size) {
	mach_vm_address_t result = (mach_vm_address_t)address;
	int flags = address ? VM_FLAGS_FIXED : VM_FLAGS_ANYWHERE;
	if (mach_vm_allocate(mach_task_self(), &result, (mach_vm_size_t)size, flags) != KERN_SUCCESS)
		return NULL;
	return (void*)(uintptr_t)result;
}

static void opennox_native_test_release_mapping(void* address, size_t size) {
	mach_vm_deallocate(mach_task_self(), (mach_vm_address_t)(uintptr_t)address, (mach_vm_size_t)size);
}
#endif

#endif
