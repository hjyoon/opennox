#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "../quest_map_buffer_4e8e50.c"

static char quest_map_storage[256];
static void* mapped_pointer;
static uintptr_t mapped_base;
static uintptr_t mapped_offset;
static unsigned int mapping_calls;

void* mem_getPtr(uintptr_t base, uintptr_t offset) {
	++mapping_calls;
	mapped_base = base;
	mapped_offset = offset;
	return mapped_pointer;
}

static char* (*const quest_map_buffer_signature_4e8e50)(void) = sub_4E8E50;

int main(void) {
	mapped_pointer = &quest_map_storage[37];
	assert(quest_map_buffer_signature_4e8e50() == &quest_map_storage[37]);
	assert(mapping_calls == 1);
	assert(mapped_base == UINT32_C(0x5D4594));
	assert(mapped_offset == UINT32_C(1567844));
	assert(mapped_base + mapped_offset == UINT32_C(0x7531F8));

	mapped_pointer = &quest_map_storage[191];
	assert(quest_map_buffer_signature_4e8e50() == &quest_map_storage[191]);
	assert(mapping_calls == 2);
	return 0;
}
