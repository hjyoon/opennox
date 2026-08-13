#include <assert.h>
#include <stddef.h>
#include <stdint.h>

static uint32_t gate_words[3];
static unsigned int memory_calls;

uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t offset) {
	assert(base == UINT32_C(0x5D4594));
	assert(offset == UINT32_C(1567712));
	++memory_calls;
	return &gate_words[1];
}

#include "../unit_freeze_gate_4e79b0.c"

static uint32_t (*const setter_signature)(uint32_t) = sub_4E79B0;

int main(void) {
	static const uint32_t values[] = {
		UINT32_C(0),
		UINT32_C(1),
		UINT32_C(0x7fffffff),
		UINT32_C(0x80000000),
		UINT32_C(0xffffffff),
	};

	for (size_t i = 0; i < sizeof(values) / sizeof(values[0]); ++i) {
		gate_words[0] = UINT32_C(0x11223344);
		gate_words[1] = ~values[i];
		gate_words[2] = UINT32_C(0xaabbccdd);
		memory_calls = 0;

		assert(setter_signature(values[i]) == values[i]);
		assert(memory_calls == 1);
		assert(gate_words[0] == UINT32_C(0x11223344));
		assert(gate_words[1] == values[i]);
		assert(gate_words[2] == UINT32_C(0xaabbccdd));
	}
	return 0;
}
