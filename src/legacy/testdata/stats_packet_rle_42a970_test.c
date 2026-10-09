#include <assert.h>
#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// PRODUCTION_COMPRESSION

static uint32_t cases;

static uint8_t input_byte(uint32_t i, uint32_t kind, uint32_t value) {
	if (!kind) return (uint8_t)value;
	if (kind == 1) return (uint8_t)i;
	if (kind == 2) return (uint8_t)(i / 7u);
	if (kind == 3) return (uint8_t)(255u - i);
	return i % 16u < 8u ? 0 : (uint8_t)(i % 16u);
}

static uint32_t reference(const uint8_t* input, uint32_t count, uint8_t* output) {
	uint32_t frequency[256] = {0};
	for (uint32_t i = 0; i < count; i++) frequency[input[i]]++;
	uint8_t escape = 0;
	for (uint32_t i = 0; i < 256; i++) {
		if (!frequency[i]) { escape = (uint8_t)i; break; }
		if (frequency[i] < frequency[escape]) escape = (uint8_t)i;
	}
	output[0] = escape;
	uint32_t pos = 1;
	for (uint32_t first = 0; first < count;) {
		uint32_t end = first + 1;
		while (end < count && input[end] == input[first]) end++;
		// The original BYTE count wraps rather than splitting a long run.
		uint8_t run = (uint8_t)(end - first);
		if (run > 3) {
			output[pos++] = escape;
			output[pos++] = run;
			output[pos++] = input[first];
		} else for (uint32_t i = 0; i < run; i++) {
			if (input[first] == escape) output[pos++] = escape;
			output[pos++] = input[first];
		}
		first = end;
	}
	return pos;
}

static void run_case(uint32_t length, uint32_t kind, uint32_t value, int json) {
	struct length_box { uint32_t length, next; uint64_t guard; };
	uint32_t count = (int32_t)length > 0 ? length : 0;
	uint32_t capacity = count * 2u + 16u;
	struct length_box* box = malloc(sizeof(*box));
	uint8_t* input = malloc(count + 8u);
	uint8_t* output = malloc(capacity);
	uint8_t* expected = malloc(capacity);
	assert(box && input && output && expected);
	assert(sizeof(void*) == 4 || ((uintptr_t)box > UINT32_MAX && (uintptr_t)input > UINT32_MAX && (uintptr_t)output > UINT32_MAX));
	box->length = length; box->next = 0xa5c3e791; box->guard = UINT64_C(0x12345678fedcba98);
	for (uint32_t i = 0; i < count + 8u; i++) input[i] = input_byte(i, kind, value);
	memset(output, 0xcc, capacity); memset(expected, 0xcc, capacity);
	uint32_t size = reference(input, count, expected);
	int result = sub_42A970(input, output, (int*)box);
	if (result != (int)size || box->length != size || memcmp(output, expected, capacity)) {
		fprintf(stderr, "case %u: length=%u kind=%u value=%u output=%d want=%u\n", cases, length, kind, value, result, size);
		exit(1);
	}
	assert(box->next == 0xa5c3e791 && box->guard == UINT64_C(0x12345678fedcba98));
	for (uint32_t i = 0; i < count + 8u; i++) assert(input[i] == input_byte(i, kind, value));
	if (json) {
		printf("{\"length\":%u,\"kind\":%u,\"value\":%u,\"result_length\":%u,\"result\":%d,\"output\":\"", length, kind, value, box->length, result);
		for (uint32_t i = 0; i < size; i++) printf("%02x", output[i]);
		puts("\"}");
	}
	free(expected); free(output); free(input); free(box); cases++;
}

int main(int argc, char** argv) {
	int json = argc == 2 && !strcmp(argv[1], "--json");
	const uint32_t runs[] = {0, 1, 2, 3, 4, 7, 255, 256, 257, 511, 512};
	for (uint32_t value = 0; value < 256; value++) for (size_t i = 0; i < sizeof(runs) / sizeof(runs[0]); i++) run_case(runs[i], 0, value, json);
	const uint32_t lengths[] = {0, 1, 2, 3, 4, 8, 12, 16, 240, 241, 242, 256, 257, 512, 1024};
	for (uint32_t kind = 1; kind < 5; kind++) for (size_t i = 0; i < sizeof(lengths) / sizeof(lengths[0]); i++) run_case(lengths[i], kind, 0, json);
	const uint32_t negative[] = {UINT32_C(0x80000000), UINT32_C(0x80000001), UINT32_C(0xffffffff)};
	for (size_t i = 0; i < sizeof(negative) / sizeof(negative[0]); i++) run_case(negative[i], 0, 0x80, json);
	if (!json) printf("%u native BYTE/run-wrap/length-guard compression cases passed\n", cases);
	return 0;
}
