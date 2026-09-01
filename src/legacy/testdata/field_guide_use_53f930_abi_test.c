#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../GAME4_3.h"
#include "../server__magic__plyrgide.h"

typedef int32_t (*beast_guide_award_fn)(nox_object_t*, int32_t, int32_t);
typedef int32_t (*field_guide_use_fn)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "guide scalars must remain exact signed int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide,
		beast_guide_award_fn: 1, default: 0),
	"beast guide award must use a native object pointer and signed int32 scalars");
_Static_assert(
	_Generic(&sub_53F930, field_guide_use_fn: 1, default: 0),
	"FieldGuideUse must preserve both native object pointers");

static nox_object_t* observed_owner;
static nox_object_t* observed_item;
static int32_t observed_guide;
static int32_t observed_notify;
static unsigned int observed_award_calls;
static unsigned int observed_use_calls;

int32_t nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide(
		nox_object_t* unit, int32_t guide, int32_t notify) {
	observed_owner = unit;
	observed_guide = guide;
	observed_notify = notify;
	++observed_award_calls;
	return notify;
}

int32_t sub_53F930(nox_object_t* owner, nox_object_t* item) {
	observed_owner = owner;
	observed_item = item;
	++observed_use_calls;
	return INT32_MIN;
}

int main(void) {
	nox_object_t owner = {0};
	nox_object_t item = {0};
	beast_guide_award_fn const award = nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide;
	field_guide_use_fn const use = sub_53F930;

	assert(award(&owner, INT32_MAX, INT32_MIN) == INT32_MIN);
	assert(observed_owner == &owner);
	assert(observed_guide == INT32_MAX);
	assert(observed_notify == INT32_MIN);
	assert(award(NULL, INT32_MIN, INT32_MAX) == INT32_MAX);
	assert(observed_owner == NULL);
	assert(observed_award_calls == 2);

	assert(use(&owner, &item) == INT32_MIN);
	assert(observed_owner == &owner);
	assert(observed_item == &item);
	assert(use(NULL, NULL) == INT32_MIN);
	assert(observed_owner == NULL);
	assert(observed_item == NULL);
	assert(observed_use_calls == 2);
	return 0;
}
