#include <assert.h>
#include <stdint.h>
#include <string.h>

#include "../quest_journal_500540.h"

typedef double (*quest_journal_get_float_fn)(char*);

_Static_assert(sizeof(double) == 8,
	"00500770 result must remain an eight-byte C double");
_Static_assert(
	_Generic(&sub_500770, quest_journal_get_float_fn: 1, default: 0),
	"00500770 must preserve its mutable C-string, double-result ABI");

static uint32_t value_bits;

static float float_from_bits(uint32_t bits) {
	float value;
	memcpy(&value, &bits, sizeof(value));
	return value;
}

static uint32_t float_bits(float value) {
	uint32_t bits;
	memcpy(&bits, &value, sizeof(bits));
	return bits;
}

double sub_500770(char* name) {
	if (name[0] == '\0') {
		return 0.0;
	}
	return (double)float_from_bits(value_bits);
}

int main(void) {
	char missing[] = "";
	char name[] = "War01a:Value";
	quest_journal_get_float_fn const get_float = sub_500770;

	assert(float_bits((float)get_float(missing)) == UINT32_C(0x00000000));
	value_bits = UINT32_C(0x80000000);
	assert(float_bits((float)get_float(name)) == UINT32_C(0x80000000));
	value_bits = UINT32_C(0x00000001);
	assert(float_bits((float)get_float(name)) == UINT32_C(0x00000001));
	value_bits = UINT32_C(0x7f7fffff);
	assert(float_bits((float)get_float(name)) == UINT32_C(0x7f7fffff));
	value_bits = UINT32_C(0x7fc12345);
	assert(float_bits((float)get_float(name)) == UINT32_C(0x7fc12345));
	return 0;
}
