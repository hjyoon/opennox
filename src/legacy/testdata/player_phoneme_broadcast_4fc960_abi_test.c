#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../player_phoneme_broadcast_4fc960.h"

typedef int32_t (*player_phoneme_broadcast_fn)(nox_object_t*, int8_t);
typedef int32_t (*spell_get_phoneme_fn)(uint32_t, int8_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int8_t) == 1, "phoneme must remain exact signed int8");
_Static_assert(sizeof(int32_t) == 4, "result must remain exact signed int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4FC960, player_phoneme_broadcast_fn: 1, default: 0),
	"004FC960 must preserve one native object pointer and signed byte");
_Static_assert(
	_Generic(&nox_xxx_spellGetPhoneme_4FE1C0,
		spell_get_phoneme_fn: 1, default: 0),
	"004FE1C0 must preserve exact dword code and signed byte");

struct nox_object_t {
	uint32_t marker;
};

static nox_object_t* observed_source;
static int8_t observed_broadcast_phoneme;
static uint32_t observed_net_code;
static int8_t observed_lookup_phoneme;

int32_t sub_4FC960(nox_object_t* source, int8_t phoneme) {
	observed_source = source;
	observed_broadcast_phoneme = phoneme;
	return INT32_MIN;
}

int32_t nox_xxx_spellGetPhoneme_4FE1C0(uint32_t net_code, int8_t phoneme) {
	observed_net_code = net_code;
	observed_lookup_phoneme = phoneme;
	return INT32_MAX;
}

int main(void) {
	nox_object_t source = {0};
	player_phoneme_broadcast_fn const broadcast = sub_4FC960;
	spell_get_phoneme_fn const lookup = nox_xxx_spellGetPhoneme_4FE1C0;

	assert(broadcast(&source, INT8_MIN) == INT32_MIN);
	assert(observed_source == &source);
	assert(observed_broadcast_phoneme == INT8_MIN);
	assert(lookup(UINT32_C(0xfedcba98), INT8_MIN) == INT32_MAX);
	assert(observed_net_code == UINT32_C(0xfedcba98));
	assert(observed_lookup_phoneme == INT8_MIN);
	return 0;
}
