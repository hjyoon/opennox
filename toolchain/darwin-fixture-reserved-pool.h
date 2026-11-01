#ifndef OPENNOX_NATIVE_TEST_EXACT_MAPPING_H
#define OPENNOX_NATIVE_TEST_EXACT_MAPPING_H
#ifdef __APPLE__
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <mach/mach.h>
#include <mach/mach_vm.h>

static struct {
	uintptr_t address;
	void* memory;
	int used;
	kern_return_t error;
} opennox_native_test_reservations[] = {
	{0x37ffc0000ull, NULL, 0, 0}, {0x3fffc0000ull, NULL, 0, 0}, {0x440000000ull, NULL, 0, 0},
	{0x47ffc0000ull, NULL, 0, 0}, {0x4fffc0000ull, NULL, 0, 0}, {0x540000000ull, NULL, 0, 0},
};

/* Keep every original fixture address reserved throughout the process.
 * Returning a slot never releases its address to an unrelated allocator.
 * VM_FLAGS_FIXED deliberately omits VM_FLAGS_OVERWRITE. */
__attribute__((constructor)) static void opennox_native_test_reserve_addresses(void) {
	for (size_t i=0;i<sizeof(opennox_native_test_reservations)/sizeof(opennox_native_test_reservations[0]);i++) {
		mach_vm_address_t address=opennox_native_test_reservations[i].address;
		kern_return_t error=mach_vm_allocate(mach_task_self(),&address,0x80000,VM_FLAGS_FIXED);
		opennox_native_test_reservations[i].error=error;
		if (error==KERN_SUCCESS) opennox_native_test_reservations[i].memory=(void*)(uintptr_t)address;
	}
}

static void* opennox_native_test_exact_mapping(uintptr_t address,size_t size) {
	for (size_t i=0;i<sizeof(opennox_native_test_reservations)/sizeof(opennox_native_test_reservations[0]);i++) {
		if (address!=opennox_native_test_reservations[i].address) continue;
		if (size>0x80000||opennox_native_test_reservations[i].used) return NULL;
		void* memory=opennox_native_test_reservations[i].memory;
		if (!memory) {
			mach_vm_address_t region=address;mach_vm_size_t region_size=0;
			vm_region_extended_info_data_t info;mach_msg_type_number_t count=VM_REGION_EXTENDED_INFO_COUNT;
			mach_port_t object=MACH_PORT_NULL;
			kern_return_t error=mach_vm_region(mach_task_self(),&region,&region_size,VM_REGION_EXTENDED_INFO,(vm_region_info_t)&info,&count,&object);
			fprintf(stderr,"fixture reservation unavailable: address=%p allocate=%d region-error=%d region=%p size=%llu tag=%u\n",(void*)address,opennox_native_test_reservations[i].error,error,(void*)(uintptr_t)region,(unsigned long long)region_size,error==KERN_SUCCESS?info.user_tag:0);
			if(object!=MACH_PORT_NULL)mach_port_deallocate(mach_task_self(),object);
			return NULL;
		}
		opennox_native_test_reservations[i].used=1;
		memset(memory,0,size);
		return memory;
	}
	mach_vm_address_t result=address;
	if (mach_vm_allocate(mach_task_self(),&result,(mach_vm_size_t)size,address?VM_FLAGS_FIXED:VM_FLAGS_ANYWHERE)!=KERN_SUCCESS)return NULL;
	return (void*)(uintptr_t)result;
}

static void opennox_native_test_release_mapping(void* address,size_t size) {
	for (size_t i=0;i<sizeof(opennox_native_test_reservations)/sizeof(opennox_native_test_reservations[0]);i++) {
		if (address==opennox_native_test_reservations[i].memory&&address!=NULL) {
			opennox_native_test_reservations[i].used=0;
			return;
		}
	}
	mach_vm_deallocate(mach_task_self(),(mach_vm_address_t)(uintptr_t)address,(mach_vm_size_t)size);
}
#endif
#endif
