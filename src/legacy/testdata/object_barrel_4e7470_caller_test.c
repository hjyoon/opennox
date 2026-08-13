#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5.h"

#include <assert.h>
#include <stdint.h>

enum {
	EVENT_NAME = 1,
	EVENT_NEW,
	EVENT_CREATE,
	EVENT_AUDIO,
	EVENT_SPAWN,
	EVENT_DELETE,
};

static uint32_t cached_breaking_type;
static int events[8];
static int event_count;
static nox_object_t* next_object;
static nox_object_t* expected_source;
static nox_object_t* expected_effect;

static void event(int value) { events[event_count++] = value; }

uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) {
	assert(base == UINT32_C(0x5D4594));
	assert(off == 2491696);
	return &cached_breaking_type;
}

int nox_xxx_getNameId_4E3AA0(char* name) {
	assert(strcmp(name, "BarrelBreaking") == 0);
	event(EVENT_NAME);
	return 77;
}

nox_object_t* nox_xxx_newObjectWithTypeInd_4E3450(int ind) {
	assert(ind == 77);
	event(EVENT_NEW);
	return next_object;
}

void nox_xxx_createAt_4DAA50(nox_object_t* obj, nox_object_t* owner, float x, float y) {
	assert(obj == expected_effect);
	assert(owner == NULL);
	assert(x == expected_source->x);
	assert(y == expected_source->y);
	event(EVENT_CREATE);
}

void nox_xxx_aud_501960(int sound, nox_object_t* obj, int a3, int a4) {
	assert(sound == 286);
	assert(obj == expected_source);
	assert(a3 == 0);
	assert(a4 == 0);
	event(EVENT_AUDIO);
}

void nox_xxx_spawnSomeBarrel_4E7470(nox_object_t* source, float2* pos) {
	assert(source == expected_source);
	assert((void*)pos == (void*)&expected_source->x);
	event(EVENT_SPAWN);
}

void nox_xxx_delayedDeleteObject_4E5CC0(nox_object_t* obj) {
	assert(obj == expected_source);
	event(EVENT_DELETE);
}

static void reset_events(void) {
	event_count = 0;
	memset(events, 0, sizeof(events));
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t effect = {0};
	source.x = 12.5f;
	source.y = -34.25f;
	expected_source = &source;
	expected_effect = &effect;

	cached_breaking_type = 0;
	next_object = &effect;
	reset_events();
	nox_xxx_dieBarrel_54DFA0(&source);
	const int first[] = {EVENT_NAME, EVENT_NEW, EVENT_CREATE, EVENT_AUDIO, EVENT_SPAWN, EVENT_DELETE};
	assert(event_count == (int)(sizeof(first) / sizeof(first[0])));
	assert(memcmp(events, first, sizeof(first)) == 0);
	assert(cached_breaking_type == 77);

	next_object = NULL;
	reset_events();
	nox_xxx_dieBarrel_54DFA0(&source);
	const int second[] = {EVENT_NEW, EVENT_AUDIO, EVENT_SPAWN, EVENT_DELETE};
	assert(event_count == (int)(sizeof(second) / sizeof(second[0])));
	assert(memcmp(events, second, sizeof(second)) == 0);
	return 0;
}
