#include <stdio.h>
#define _Static_assert(...)
#include "../defs.h"
#undef _Static_assert

static uint32_t pe_record[64];
static uint32_t dword_5d4594_1047520, dword_5d4594_1046652, dword_5d4594_1046648;
static uint32_t dword_5d4594_1046636, dword_5d4594_1046640;
static uint32_t dword_5d4594_1047524, dword_5d4594_1046852;
static float2 obj_5d4594_1046620;
static void* nox_xxx_aClosewoodengat_587000_133480;
static nox_window* window;
static nox_video_bag_image_t* icon;
static int nox_win_width = 1024;
static unsigned int clock_tick, spell_flags, expected_random, expected_particles, initial_burst;
static unsigned int random_calls, particle_calls, insert_calls, flag_calls, finish_calls;
static unsigned int image_calls, position_calls, normalize_calls, sound_calls, case_count;
static int failure;

#define CHECK(test) do { if (!(test)) { \
	fprintf(stderr, "line %d: %s (case=%u mode=%u trail=%u flags=%#x insert=%u flag_calls=%u)\n", \
		__LINE__, #test, case_count, pe_record[12], pe_record[13], spell_flags, insert_calls, flag_calls); return 1; \
} } while (0)

void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base != 0x5D4594 || off < 1046628 || off >= 1046628 + sizeof(pe_record)) abort();
	return (unsigned char*)pe_record + off - 1046628;
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base, off); }
int32_t* mem_getI32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base, off); }
float* mem_getFloatPtr(uintptr_t base, uintptr_t off) { return mem_getPtr(base, off); }

// PRODUCTION_SELECTED_ROW

static void nox_xxx_spellKeyPackSetSpell_45DC40(void* base, int spell, int slot) {
	if (base != nox_xxx_aClosewoodengat_587000_133480 || spell != (int)dword_5d4594_1047524 ||
		slot != (int)dword_5d4594_1046852 || particle_calls != expected_particles ||
		random_calls != expected_random || image_calls != 1 || position_calls || finish_calls) failure = 1;
	insert_calls++;
	unsigned char* row = nox_quickbar_selected_row(base);
	if (row && slot >= 0 && slot < 5) memcpy(row + 8 * slot, &spell, 4);
}
static bool nox_xxx_spellHasFlags_424A50(int spell, int flags) {
	if (insert_calls != 1 || finish_calls || spell != (int)dword_5d4594_1047524 || flags != 0x600) failure = 1;
	flag_calls++;
	return (spell_flags & (unsigned int)flags) != 0;
}
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wint-to-pointer-cast"
// PRODUCTION_BOOK_INSERT
#pragma GCC diagnostic pop

static unsigned int nox_xxx_bookGet_430B40_get_mouse_prev_seq(void) { return clock_tick; }
uint32_t gameFPS(void) { return 30; }
static int nox_float2int(float value) { return (int)value; }
int nox_client_wndGetPosition_46AA60(nox_window* win, unsigned int* x, unsigned int* y) {
	if (win != window || particle_calls || image_calls) failure = 1;
	*x = 100;
	*y = 300;
	return 1;
}
static int nox_common_randomIntMinMax_415FF0(int min, int max, const char* file, int line) {
	const int burst_lines[] = {1287, 1286, 1284, 1283, 1282, 1281};
	const int burst_min[] = {3, 2, -10, -10, 0, 0};
	const int burst_max[] = {6, 5, -1, 10, 30, 30};
	const int trail_lines[] = {1331, 1330, 1326, 1325};
	const int trail_min[] = {2, 1, 0, 0};
	const int trail_max[] = {4, 2, 30, 30};
	if (strcmp(file, "C:\\NoxPost\\src\\Client\\Gui\\guibook.c") || insert_calls || finish_calls) failure = 1;
	if (initial_burst && random_calls < 300) {
		unsigned int i = random_calls % 6;
		if (line != burst_lines[i] || min != burst_min[i] || max != burst_max[i]) failure = 1;
	} else {
		unsigned int i = (random_calls - 300 * initial_burst) % 4;
		if (line != trail_lines[i] || min != trail_min[i] || max != trail_max[i]) failure = 1;
	}
	random_calls++;
	return min;
}
static void nox_client_newScreenParticle_431540(int kind, int x, int y, int dx, int dy,
	int a6, int a7, int a8, int a9, int a10) {
	if (kind != (dword_5d4594_1046652 ? 3 : 0) || x != 100 || y != 300 ||
		a10 != 1 || insert_calls || finish_calls) failure = 1;
	if (initial_burst && particle_calls < 50) {
		if (random_calls != 6 * (particle_calls + 1) || dx != -10 || dy != -10 || a6 != 1 ||
			a7 != 2 || a8 != 3 || a9 != 2) failure = 1;
	} else if (random_calls != 300 * initial_burst + 4 * (particle_calls - 50 * initial_burst + 1) ||
		dx || dy || a6 || a7 != 1 || a8 != 2 || a9 != 1) failure = 1;
	particle_calls++;
}
static void nox_xxx_clientPlaySoundSpecial_452D80(int sound, int volume) {
	if (sound != 795 || volume != 100 || particle_calls != 50 || random_calls != 300) failure = 1;
	sound_calls++;
}
static nox_video_bag_image_t* nox_xxx_spellGetAbilityIcon_425310(int spell, int unused) {
	if (spell != (int)dword_5d4594_1047524 || unused || dword_5d4594_1046652 != 1) failure = 1;
	return icon;
}
static nox_video_bag_image_t* nox_xxx_spellIcon_424A90(int spell) {
	if (spell != (int)dword_5d4594_1047524 || dword_5d4594_1046652 == 1) failure = 1;
	return icon;
}
static void nox_client_drawImageAt_47D2C0(nox_video_bag_image_t* image, int x, int y) {
	if (image != icon || x != 100 || y != 300 || particle_calls != expected_particles ||
		random_calls != expected_random || insert_calls || finish_calls) failure = 1;
	image_calls++;
}
static void nox_xxx_utilNormalizeVector_509F20(float2* vector) {
	if (vector != &obj_5d4594_1046620 || vector->field_0 != 200 || vector->field_4 != 200) failure = 1;
	vector->field_0 = vector->field_4 = 0.70710677f;
	normalize_calls++;
}
static void sub_45D810(void) {
	if (insert_calls != 1 || image_calls != 1 || position_calls ||
		particle_calls != expected_particles || random_calls != expected_random) failure = 1;
	finish_calls++;
	dword_5d4594_1047520 = 0;
}
int nox_window_setPos_46A9B0(nox_window* win, int x, int y) {
	if (win != window || x != 110 || y != 314 || image_calls != 1) failure = 1;
	position_calls++;
	return 1;
}

// Preserve the existing uint32_t* window signature; native pointers are not
// narrowed by it. The old implicit window-type conversion is unrelated here.
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types"
#pragma GCC diagnostic ignored "-Wpointer-sign"
// PRODUCTION_DRAW_BODY
#pragma GCC diagnostic pop

// branch 0: arrive at target; 1: exhaust trail; 2: continue trail;
// branch 3: wait for the initial burst clock; 4: inactive animation.
static int check_draw(unsigned int mode, unsigned int trail, unsigned int flags,
	unsigned int row, int slot, unsigned int spell, unsigned int kind, unsigned int burst, unsigned int branch) {
	case_count++;
	unsigned char* bar = nox_xxx_aClosewoodengat_587000_133480;
	memset(bar, 0xA5, 256);
	bar[200] = (unsigned char)row;
	memset(pe_record, 0x5A, sizeof(pe_record));
	*mem_getU32Ptr(0x5D4594, 1046628) = 0;
	*mem_getU32Ptr(0x5D4594, 1046668) = branch == 0 ? 110 : 1000;
	*mem_getU32Ptr(0x5D4594, 1046672) = branch == 0 ? 314 : 1000;
	*mem_getU32Ptr(0x5D4594, 1046676) = mode;
	*mem_getU32Ptr(0x5D4594, 1046680) = trail;
	*mem_getFloatPtr(0x5D4594, 1046692) = 100;
	*mem_getFloatPtr(0x5D4594, 1046696) = 300;
	*mem_getFloatPtr(0x5D4594, 1046700) = 300;
	*mem_getFloatPtr(0x5D4594, 1046704) = 500;
	const float initial_x = 100, initial_y = 300;
	memcpy(&dword_5d4594_1046636, &initial_x, 4);
	memcpy(&dword_5d4594_1046640, &initial_y, 4);
	obj_5d4594_1046620 = (float2){10, 14};
	dword_5d4594_1047524 = spell;
	dword_5d4594_1046852 = (uint32_t)slot;
	dword_5d4594_1047520 = branch == 4 ? 0 : 1;
	dword_5d4594_1046652 = kind;
	dword_5d4594_1046648 = burst || branch == 3 ? 1 : 0;
	clock_tick = branch == 3 ? 30 : 400;
	spell_flags = flags;
	initial_burst = burst;
	expected_particles = 50 * burst + 2;
	expected_random = 300 * burst + 8;
	random_calls = particle_calls = insert_calls = flag_calls = finish_calls = 0;
	image_calls = position_calls = normalize_calls = sound_calls = 0;
	failure = 0;
	unsigned char expected[256];
	memcpy(expected, bar, sizeof(expected));
	unsigned int done = branch < 2;
	unsigned int valid = row < 5 && slot >= 0 && slot < 5;
	if (done && valid) {
		memcpy(expected + 40 * row + 8 * slot, &spell, 4);
		if (mode == 2) expected[40 * row + 8 * slot + 4] = spell && (flags & 0x600) ? 1 : 0;
	}
	uint32_t wanted_record[64];
	memcpy(wanted_record, pe_record, sizeof(wanted_record));
	if (branch == 1 || branch == 2) wanted_record[0] = 1;
	CHECK(nox_xxx_bookDrawFn_45C7D0((uint32_t*)window) == 1);
	CHECK(!failure);
	CHECK(insert_calls == done && finish_calls == done);
	CHECK(flag_calls == (done && valid && mode == 2 && spell ? 1U : 0U));
	CHECK(!memcmp(bar, expected, sizeof(expected)));
	CHECK(!memcmp(pe_record, wanted_record, sizeof(wanted_record)));
	CHECK(normalize_calls == (branch == 2 ? 1U : 0U));
	CHECK(image_calls == (branch < 3 ? 1U : 0U) && position_calls == image_calls);
	CHECK(particle_calls == (branch < 3 ? expected_particles : 0U));
	CHECK(random_calls == (branch < 3 ? expected_random : 0U));
	CHECK(sound_calls == (branch < 3 ? burst : 0U));
	CHECK(dword_5d4594_1047520 == (done || branch == 4 ? 0U : 1U));
	return 0;
}
int main(void) {
	nox_xxx_aClosewoodengat_587000_133480 = calloc(1, 256);
	window = calloc(1, sizeof(*window));
	icon = malloc(1);
	if (!nox_xxx_aClosewoodengat_587000_133480 || !window || !icon) return 2;
	if (sizeof(void*) == 8) CHECK((uintptr_t)nox_xxx_aClosewoodengat_587000_133480 > UINT32_MAX && (uintptr_t)window > UINT32_MAX);
	CHECK(check_draw(2, 20, 0x200, 0, 0, 1, 0, 0, 0) == 0);
	const unsigned int modes[] = {2, 3, 4, 0, 0xFFFFFFFF};
	const unsigned int trails[] = {0, 1, 20, 0xFFFFFFFF};
	const unsigned int flags[] = {0, 0x200, 0x400, 0x600, 0x1000, 0x2000, 0x200400, 0x800000};
	for (unsigned int m = 0; m < sizeof(modes) / sizeof(modes[0]); m++) {
		for (unsigned int n = 0; n < sizeof(trails) / sizeof(trails[0]); n++) {
			for (unsigned int f = 0; f < sizeof(flags) / sizeof(flags[0]); f++) {
				for (unsigned int row = 0; row < 5; row++) {
					for (int slot = 0; slot < 5; slot++) {
						for (unsigned int kind = 0; kind < 3; kind++) {
							for (unsigned int burst = 0; burst < 2; burst++) {
								CHECK(check_draw(modes[m], trails[n], flags[f], row, slot, 136, kind, burst, 0) == 0);
								CHECK(check_draw(modes[m], 1, flags[f], row, slot, 136, kind, burst, 1) == 0);
							}
						}
					}
				}
			}
		}
	}
	for (unsigned int branch = 2; branch < 5; branch++) {
		CHECK(check_draw(2, 20, 0x600, 4, 4, 12, 0, 0, branch) == 0);
	}
	for (unsigned int spell = 0; spell < 137; spell++) CHECK(check_draw(2, 20, 0x600, 4, 4, spell, 0, 0, 0) == 0);
	const int slots[] = {-1, 0, 4, 5, 255};
	const unsigned int rows[] = {0, 4, 5, 255};
	for (unsigned int r = 0; r < sizeof(rows) / sizeof(rows[0]); r++) {
		for (unsigned int s = 0; s < sizeof(slots) / sizeof(slots[0]); s++) {
			CHECK(check_draw(2, 20, 0x600, rows[r], slots[s], 12, 1, 1, 0) == 0);
		}
	}
	free(icon);
	free(window);
	free(nox_xxx_aClosewoodengat_587000_133480);
	printf("book reward draw: %u arrival/trail/default-self/slot cases passed\n", case_count);
	return 0;
}
