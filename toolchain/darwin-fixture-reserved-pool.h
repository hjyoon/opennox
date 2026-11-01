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
static uintptr_t opennox_native_test_address_delta;
#ifndef OPENNOX_NATIVE_TEST_MIN_LANE
#define OPENNOX_NATIVE_TEST_MIN_LANE 0
#endif

#define OPENNOX_NATIVE_TEST_ADDRESS(address) ((uintptr_t)(address)+opennox_native_test_address_delta)

/* Preserve each low DWORD, relative placement and 4-GiB crossing. Newer
 * Darwin loaders can already own the old synthetic native addresses.
 * Select one common whole-4-GiB translation, then retain every reservation.
 * VM_FLAGS_FIXED deliberately omits VM_FLAGS_OVERWRITE. */
__attribute__((constructor)) static void opennox_native_test_reserve_addresses(void) {
	for (unsigned lane=OPENNOX_NATIVE_TEST_MIN_LANE;lane<256;lane++) {
		uintptr_t delta=(uintptr_t)lane<<32;
		int complete=1;
		for (size_t i=0;i<sizeof(opennox_native_test_reservations)/sizeof(opennox_native_test_reservations[0]);i++) {
			mach_vm_address_t address=opennox_native_test_reservations[i].address+delta;
			kern_return_t error=mach_vm_allocate(mach_task_self(),&address,0x80000,VM_FLAGS_FIXED);
			opennox_native_test_reservations[i].error=error;
			if(error!=KERN_SUCCESS){complete=0;break;}
			opennox_native_test_reservations[i].memory=(void*)(uintptr_t)address;
		}
		if(complete){
			opennox_native_test_address_delta=delta;
			for(size_t i=0;i<sizeof(opennox_native_test_reservations)/sizeof(opennox_native_test_reservations[0]);i++)opennox_native_test_reservations[i].address+=delta;
			return;
		}
		/* Release only reservations made by this attempt, never another map. */
		for(size_t i=0;i<sizeof(opennox_native_test_reservations)/sizeof(opennox_native_test_reservations[0]);i++){
			if(opennox_native_test_reservations[i].memory)mach_vm_deallocate(mach_task_self(),(mach_vm_address_t)(uintptr_t)opennox_native_test_reservations[i].memory,0x80000);
			opennox_native_test_reservations[i].memory=NULL;
		}
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
#else
#define OPENNOX_NATIVE_TEST_ADDRESS(address) ((uintptr_t)(address))
#endif
#endif
