#include <arpa/inet.h>
#include <assert.h>
#include <inttypes.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static uint32_t input_length, variant, cases, rng_calls, clock_calls;
static uint32_t allocation_calls, free_calls;
static int64_t epoch;
static void* blocks[2];
static char events[8192];
static size_t event_length;

static void event(char value) {
	assert(event_length + 1 < sizeof(events));
	events[event_length++] = value;
	events[event_length] = 0;
}

static void* fixture_calloc(size_t count, size_t size) {
	assert(allocation_calls < 2 && count == 1);
	size_t expected = input_length + (allocation_calls == 1 ? 5 : 0);
	if (size != expected) {
		fprintf(stderr, "case %u: DWORD length was widened: allocation=%zu want=%zu\n", cases, size, expected);
		exit(1);
	}
	void* ptr = calloc(1, size);
	assert(ptr && (sizeof(void*) == 4 || (uintptr_t)ptr > UINT32_MAX));
	blocks[allocation_calls++] = ptr;
	event('A');
	return ptr;
}

static void fixture_free(void* ptr) {
	assert(free_calls < 2 && ptr == blocks[free_calls]);
	free_calls++;
	event('F');
	free(ptr);
}

static int32_t seed_for(int64_t value) {
	int32_t word = (int32_t)(uint32_t)value;
	return word > 0 ? (int32_t)(0u - (uint32_t)word) : word;
}

static time_t fixture_time(time_t* destination) {
	assert(destination == NULL && clock_calls < 2);
	event('T');
	return (time_t)(epoch + clock_calls++);
}

static double mask_value(uint32_t index) {
	// Exact binary fractions exercise the original absolute/truncate path.
	static const double values[] = {0.0, 0.25, -0.5, 0.75};
	return values[(index + variant) & 3];
}

static double selector_value(void) {
	static const double values[] = {0.0, 0.25, 0.5, 0.75, 255.0 / 256.0};
	return values[variant];
}

static double sub_42AAA0(int* seed) {
	assert(sizeof(*seed) == 4);
	assert(clock_calls == 1 || clock_calls == 2);
	if (clock_calls == 1) {
		assert(*seed == (rng_calls ? 1 : seed_for(epoch)));
	} else {
		assert(rng_calls == input_length + 1 && *seed == seed_for(epoch + 1));
	}
	*seed = 1;
	event('R');
	double result = clock_calls == 1 ? mask_value(rng_calls) : selector_value();
	rng_calls++;
	return result;
}

// The old body sends a truncated native allocation to this PE32 dependency.
// A successful native body must instead generate its mask at full width.
void sub_42ABF0(int destination, int count, int seed) {
	(void)destination; (void)count; (void)seed;
	fprintf(stderr, "case %u: mask buffer crossed an ABI32 address boundary\n", cases);
	exit(1);
}

#define calloc fixture_calloc
#define free fixture_free
#define time fixture_time

// PRODUCTION_SCRAMBLE

#undef calloc
#undef free
#undef time

static int run_case(uint32_t length, uint32_t choice, int64_t clock, int json) {
	struct length_box { uint32_t length; uint32_t next; uint64_t guard; };
	struct length_box* box = malloc(sizeof(*box));
	uint8_t* input = malloc(length + 8);
	uint8_t* expected = calloc(1, length + 5);
	assert(box && input && expected);
	assert(sizeof(void*) == 4 || ((uintptr_t)box > UINT32_MAX && (uintptr_t)input > UINT32_MAX));
	box->length = length;
	box->next = 0xa5c3e791;
	box->guard = UINT64_C(0x12345678fedcba98);
	for (uint32_t i = 0; i < length + 8; i++) input[i] = (uint8_t)(i * 37u + choice * 11u);
	input_length = length; variant = choice; epoch = clock;
	rng_calls = clock_calls = allocation_calls = free_calls = 0;
	event_length = 0; events[0] = 0;
	uint8_t* output = sub_42AC50(input, (size_t*)box);
	assert(box->next == 0xa5c3e791 && box->guard == UINT64_C(0x12345678fedcba98));
	assert(allocation_calls == 2);
	for (uint32_t i = 0; i < length + 8; i++) assert(input[i] == (uint8_t)(i * 37u + choice * 11u));
	if (length < 15) {
		assert(output == NULL && box->length == UINT32_C(0xfffffffe));
		assert(clock_calls == 0 && rng_calls == 0 && free_calls == 2 && !strcmp(events, "AAFF"));
	} else {
		uint32_t factor = length >= 241 ? 241 : length - 14;
		uint32_t selector = 10 + (choice == 4 ? factor * 255 / 256 : factor * choice / 4);
		uint32_t key = htonl((uint32_t)seed_for(clock));
		uint32_t pos = 0;
		for (uint32_t i = 0; i < length; i++) {
			if (pos == 5) expected[pos++] = (uint8_t)selector;
			if (pos == selector) { memcpy(expected + pos, &key, 4); pos += 4; }
			int32_t mask = (int32_t)(mask_value(i) * 255.0);
			if (mask < 0) mask = -mask;
			expected[pos++] = input[i] ^ (uint8_t)mask;
		}
		// For 241/242-byte input, selector 250 is never reached. The original
		// still reports length+5; our explicitly zeroed allocator leaves its
		// untouched tail zero. Do not "repair" that packet selection behavior.
		assert((pos == length + 5 || pos == length + 1) && output == blocks[1] && box->length == length + 5);
		assert(!memcmp(output, expected, length + 5));
		assert(clock_calls == 2 && rng_calls == length + 2 && free_calls == 1);
		assert(!strncmp(events, "AAT", 3));
		for (uint32_t i = 0; i < length + 1; i++) assert(events[i + 3] == 'R');
		assert(!strcmp(events + length + 4, "TRF"));
	}
	if (json) {
		printf("{\"length\":%u,\"variant\":%u,\"epoch\":%" PRId64 ",\"result_length\":%u,\"events\":\"%s\",\"output\":\"", length, choice, clock, box->length, events);
		if (output) for (uint32_t i = 0; i < length + 5; i++) printf("%02x", output[i]);
		puts("\"}");
	}
	if (output) free(output);
	free(expected); free(input); free(box);
	cases++;
	return 0;
}

int main(int argc, char** argv) {
	int json = argc == 2 && !strcmp(argv[1], "--json");
	const int64_t clocks[] = {0, 123456789, -123456789, INT32_MAX, INT32_MIN, INT64_C(0x10000002a)};
	const uint32_t long_lengths[] = {15, 16, 20, 240, 241, 242, 256, 257, 400, 1024};
	for (size_t c = 0; c < sizeof(clocks) / sizeof(clocks[0]); c++) {
		for (uint32_t v = 0; v < 5; v++) {
			for (uint32_t n = 0; n < 15; n++) run_case(n, v, clocks[c], json);
			for (size_t n = 0; n < sizeof(long_lengths) / sizeof(long_lengths[0]); n++) run_case(long_lengths[n], v, clocks[c], json);
		}
	}
	if (!json) printf("%u native DWORD/high-address scrambler cases passed\n", cases);
	return 0;
}
