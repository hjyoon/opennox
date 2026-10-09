#include <arpa/inet.h>
#include <assert.h>
#include <inttypes.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static uint32_t initial_length, controlled_length, choice, cases, allocation_calls;
static uint32_t rng_calls, clock_calls, compressed_length, encoded_length;
static int chain, nil_scramble, recording_fault;
static const char* failure;
static void* blocks[5];
static char events[16384];
static size_t event_length;
static unsigned char names[][5] = {"CNTL", {'Q', 0, 'x', 'y', 0}, {0xff, 0x80, 0x41, 0x7f, 0}};
static unsigned char empty_name[] = {0};

static void event(char value) {
	assert(event_length + 1 < sizeof(events));
	events[event_length++] = value;
	events[event_length] = 0;
	if (recording_fault) { fputc(value, stderr); fflush(stderr); }
}

static void* getMemAt(uint32_t base, uint32_t offset) {
	if (base == 0x587000 && offset == 71480) return names[choice % 3];
	assert(base == 0x5d4594 && offset == 741688);
	return empty_name;
}

static void* fixture_calloc(size_t count, size_t size) {
	assert(allocation_calls < 5);
	unsigned index = allocation_calls++;
	size_t bytes = count * size;
	if (index == 0) {
		assert(bytes == (uint32_t)(initial_length * 2u));
		event('A');
	} else if (index == 1) {
		assert(chain && count == 1 && bytes == compressed_length);
		event('M');
	} else if (index == 2) {
		assert(chain && count == 1 && bytes == compressed_length + 5u);
		event('E');
	} else assert(0);
	if (failure && !strcmp(failure, "compression_alloc") && index == 0) return NULL;
	void* ptr = calloc(1, bytes ? bytes : 1);
	assert(ptr && (sizeof(void*) == 4 || (uintptr_t)ptr > UINT32_MAX));
	blocks[index] = ptr;
	return ptr;
}

static void* envelope_calloc(size_t count, size_t size) {
	// Compression and scramble use fixture_calloc; the envelope's later
	// allocations have a native node object but unchanged WORD payload size.
	if (!allocation_calls) return fixture_calloc(count, size);
	assert(count == 1);
	char kind;
	unsigned index;
	if (!blocks[3]) { kind = 'N'; index = 3; assert(size == 8 + 2 * sizeof(void*)); }
	else if (!blocks[4]) { kind = 'D'; index = 4; assert(size == (uint16_t)encoded_length); }
	else { kind = 'P'; index = 5; assert(size == 12u + (uint16_t)encoded_length + ((0u - (uint16_t)encoded_length) & 3u)); }
	event(kind);
	if (failure && ((!strcmp(failure, "node_alloc") && kind == 'N') ||
		(!strcmp(failure, "data_alloc") && kind == 'D') ||
		(!strcmp(failure, "packet_alloc") && kind == 'P'))) return NULL;
	void* ptr = calloc(1, size ? size : 1);
	assert(ptr && (sizeof(void*) == 4 || (uintptr_t)ptr > UINT32_MAX));
	if (index < 5) blocks[index] = ptr;
	return ptr;
}

static void fixture_free(void* ptr) {
	event(ptr ? 'F' : 'Z');
	free(ptr);
}

static time_t fixture_time(time_t* destination) {
	assert(destination == NULL);
	event('T');
	return (time_t)(123456789 + clock_calls++);
}

static double sub_42AAA0(int* seed) {
	*seed = 1;
	event('R');
	return clock_calls == 1 ? (double)((rng_calls++ + choice) & 3u) / 4.0 : 0.5;
}

#define calloc fixture_calloc
#define free fixture_free
#define time fixture_time
#define sub_42A970 actual_compression
// PRODUCTION_COMPRESSION
#undef sub_42A970
#define sub_42AC50 actual_scramble
// PRODUCTION_SCRAMBLE
#undef sub_42AC50
#undef time
#undef free
#undef calloc

static uint8_t data_byte(uint32_t index) { return (uint8_t)(index * 37u + choice * 11u); }

static int sub_42A970(uint8_t* input, uint8_t* output, int* length) {
	event('C');
	assert((uint32_t)*length == initial_length);
	// The original envelope calls both stages without guarding nil buffers.
	volatile uint8_t byte = *input;
	output[0] = byte;
	if (chain) {
		int result = actual_compression(input, output, length);
		compressed_length = (uint32_t)*length;
		return result;
	}
	*length = (int32_t)controlled_length;
	return *length;
}

static uint8_t* sub_42AC50(uint8_t* input, size_t* length) {
	event('S');
	if (chain) {
		uint8_t* result = actual_scramble(input, length);
		encoded_length = *(uint32_t*)length;
		return result;
	}
	(void)input;
	*(uint32_t*)length = controlled_length;
	encoded_length = controlled_length;
	if (nil_scramble) return NULL;
	uint32_t bytes = (uint16_t)controlled_length;
	uint8_t* result = calloc(1, bytes + 8);
	assert(result && (sizeof(void*) == 4 || (uintptr_t)result > UINT32_MAX));
	for (uint32_t i = 0; i < bytes + 8; i++) result[i] = data_byte(i);
	return result;
}

// Old ABI32 dependencies must not receive any truncated native allocation.
int sub_42C910(int node, char* name, const void* data, unsigned short size) {
	(void)node; (void)name; (void)data; (void)size;
	fputs("node allocation crossed an ABI32 pointer boundary\n", stderr);
	exit(1);
}
int sub_42C360(void* packet, int node) { (void)packet; (void)node; abort(); }
void sub_42C330(void* packet) { (void)packet; }
uint16_t* sub_42C480(void* packet, unsigned int* size) { (void)packet; (void)size; abort(); }

#define calloc envelope_calloc
#define free fixture_free
// PRODUCTION_ENVELOPE
#undef free
#undef calloc

static uint8_t input_byte(uint32_t index, uint32_t pattern) {
	if (pattern == 0) return (uint8_t)(index * 37u + choice * 11u);
	if (pattern == 1) return 0;
	if (pattern == 2) return (uint8_t)(index / 7u);
	return (uint8_t)(index & 1u);
}

static void run_case(uint32_t initial, uint32_t length, uint32_t variant, int real_chain, uint32_t pattern, int nil, int json) {
	struct length_box { uint32_t length, next; uint64_t guard; };
	struct length_box* box = malloc(sizeof(*box));
	uint32_t input_bytes = real_chain ? initial : 32;
	uint8_t* input = malloc(input_bytes + 8);
	assert(box && input && (sizeof(void*) == 4 || ((uintptr_t)box > UINT32_MAX && (uintptr_t)input > UINT32_MAX)));
	box->length = initial; box->next = 0xa5c3e791; box->guard = UINT64_C(0x12345678fedcba98);
	initial_length = initial; controlled_length = length; choice = variant;
	chain = real_chain; nil_scramble = nil; compressed_length = encoded_length = 0;
	rng_calls = clock_calls = allocation_calls = 0;
	event_length = 0; events[0] = 0; memset(blocks, 0, sizeof(blocks));
	for (uint32_t i = 0; i < input_bytes + 8; i++) input[i] = input_byte(i, pattern);
	uint16_t* result = sub_42A8B0(input, (int*)box);
	assert(box->next == 0xa5c3e791 && box->guard == UINT64_C(0x12345678fedcba98));
	for (uint32_t i = 0; i < input_bytes + 8; i++) assert(input[i] == input_byte(i, pattern));
	if (nil || (chain && compressed_length < 15)) {
		assert(result == NULL && box->length == (chain ? UINT32_C(0xfffffffe) : length));
		assert(!blocks[3] && !blocks[4]);
	} else {
		uint32_t data_length = (uint16_t)encoded_length;
		uint32_t packet_length = 12u + data_length + ((0u - data_length) & 3u);
		assert(result && (sizeof(void*) == 4 || (uintptr_t)result > UINT32_MAX));
		assert(box->length == packet_length && ntohs(result[0]) == (uint16_t)packet_length && result[1] == 0);
		uint8_t name[4] = {0};
		strncpy((char*)name, (const char*)names[choice % 3], 4);
		assert(!memcmp((uint8_t*)result + 4, name, 4));
		assert(ntohs(result[4]) == 20 && ntohs(result[5]) == data_length);
		if (!chain) for (uint32_t i = 0; i < data_length; i++) assert(((uint8_t*)result)[12 + i] == data_byte(i));
		for (uint32_t i = 12 + data_length; i < packet_length; i++) assert(((uint8_t*)result)[i] == 0);
		assert(strstr(events, "NZDFP") && !strcmp(events + event_length - 2, "FF"));
	}
	if (json) {
		printf("{\"initial\":%u,\"length\":%u,\"variant\":%u,\"chain\":%d,\"pattern\":%u,\"nil\":%d,\"result_length\":%u,\"events\":\"%s\",\"output\":\"", initial, length, variant, chain, pattern, nil, box->length, events);
		if (result) for (uint32_t i = 0; i < box->length; i++) printf("%02x", ((uint8_t*)result)[i]);
		puts("\"}");
	}
	if (result) free(result);
	free(input); free(box); cases++;
}

int main(int argc, char** argv) {
	int json = argc == 2 && !strcmp(argv[1], "--json");
	if (argc == 2 && !strncmp(argv[1], "--fault=", 8)) {
		failure = argv[1] + 8; recording_fault = 1;
		initial_length = controlled_length = encoded_length = 16;
		uint8_t input[32] = {0}; int length = 16;
		sub_42A8B0(!strcmp(failure, "nil_input") ? NULL : input,
			!strcmp(failure, "nil_length") ? NULL : &length);
		return 3;
	}
	const uint32_t lengths[] = {0, 1, 2, 3, 4, 14, 15, 16, 240, 241, 242, 255, 256, 257,
		4096, 32768, 65520, 65521, 65522, 65523, 65524, 65525, 65535, 65536, 65537, 65538, 65539,
		65540, 90000, UINT32_C(0xfffffffd), UINT32_C(0xffffffff)};
	for (uint32_t v = 0; v < 3; v++) {
		for (size_t i = 0; i < sizeof(lengths) / sizeof(lengths[0]); i++) run_case(16, lengths[i], v, 0, 0, 0, json);
		for (uint32_t i = 0; i < 3; i++) run_case(UINT32_C(0x80000001) + i, 16 + i, v, 0, 0, 0, json);
		for (uint32_t i = 0; i < 3; i++) run_case(16, lengths[i] ^ UINT32_MAX, v, 0, 0, 1, json);
	}
	const uint32_t real_lengths[] = {8, 12, 16, 32, 240, 241, 242, 256, 257, 512, 1024};
	for (uint32_t p = 0; p < 4; p++) for (size_t i = 0; i < sizeof(real_lengths) / sizeof(real_lengths[0]); i++) {
		run_case(real_lengths[i], 0, p % 3, 1, p, 0, json);
	}
	if (!json) printf("%u native envelope/real-compression-chain cases passed\n", cases);
	return 0;
}
