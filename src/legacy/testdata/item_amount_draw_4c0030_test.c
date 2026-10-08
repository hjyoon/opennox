#include <stdio.h>

// Keep the production native layouts; unrelated PE32 aggregate assertions are
// not part of this fixture. Reassert the viewport ABI used by the real drawer.
#define _Static_assert(...)
#include "../defs.h"
#undef _Static_assert
_Static_assert(sizeof(nox_draw_viewport_t) == 13 * sizeof(void*), "native viewport size");
_Static_assert(offsetof(nox_draw_viewport_t, width) == 8 * sizeof(void*), "native viewport width");
_Static_assert(offsetof(nox_draw_viewport_t, height) == 9 * sizeof(void*), "native viewport height");

// PRODUCTION_IMAGE_ENUM

static uint32_t pe_record[40];
static nox_window* window;
static nox_window* children[4];
static nox_window* nox_gui_itemAmount_dialog_1319228;
static nox_drawable* nox_gui_itemAmount_item_1319256;
static nox_video_bag_image_t* nox_gui_itemAmount_images_1319196[8];
static uint32_t dword_5d4594_1319264;
static uint32_t dword_587000_183456;
static uint32_t dword_587000_183460;
static int nox_win_width = 1024;
static int nox_win_height = 768;
static unsigned int window_x;
static unsigned int window_y;
static unsigned int events[20];
static unsigned int event_count;
static unsigned int draw_count;
static unsigned int case_count;
static int mutate_viewport;
static int failure;

#define CHECK(test) do { if (!(test)) { \
	fprintf(stderr, "line %d: %s (case=%u tag=%u events=%u)\n", \
		__LINE__, #test, case_count, dword_5d4594_1319264, event_count); return 1; \
} } while (0)

void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base != 0x5D4594 || off < 1319108 || off >= 1319108 + sizeof(pe_record)) abort();
	return (unsigned char*)pe_record + off - 1319108;
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base, off); }
int32_t* mem_getI32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base, off); }

static void event(unsigned int id) {
	if (event_count >= sizeof(events) / sizeof(events[0])) abort();
	events[event_count++] = id;
}
static void nox_client_drawRectFilledAlpha_49CF10(int x, int y, int width, int height) {
	if (x || y || width != nox_win_width || height != nox_win_height) failure = 1;
	event(1);
}
int nox_gui_getWindowOffs_46AA20(nox_window* win, unsigned int* x, unsigned int* y) {
	if (win != window) failure = 1;
	*x = window_x;
	*y = window_y;
	event(2);
	return 1;
}
static void nox_client_drawImageAt_47D2C0(nox_video_bag_image_t* image, int x, int y) {
	if (x != (int)window_x || y != (int)window_y) failure = 1;
	for (unsigned int i = 0; i < 8; i++) {
		if (image == nox_gui_itemAmount_images_1319196[i]) {
			event(10 + i);
			return;
		}
	}
	failure = 1;
}
nox_window* nox_xxx_wndGetChildByID_46B0C0(nox_window* root, int id) {
	if (root != nox_gui_itemAmount_dialog_1319228) failure = 1;
	const int ids[] = {3603, 3602, 3604, 3605};
	for (unsigned int i = 0; i < 4; i++) {
		if (ids[i] == id) {
			event(30 + i);
			return children[i];
		}
	}
	abort();
}
static int draw_icon(uint32_t* raw, nox_drawable* item) {
	nox_draw_viewport_t* vp = (nox_draw_viewport_t*)raw;
	const int64_t actual[] = {vp->x1, vp->y1, vp->x2, vp->y2, vp->field_4, vp->field_5,
		vp->field_6, vp->field_7, vp->width, vp->height, (int64_t)vp->field_10,
		(int64_t)vp->field_11, vp->field_12};
	for (unsigned int i = 0; i < 13; i++) {
		int64_t expected = (i == 10 || i == 11) ? (int64_t)pe_record[i] : (int64_t)(int32_t)pe_record[i];
		if (actual[i] != expected) {
			fprintf(stderr, "native viewport field %u: %lld, want %lld\n", i,
				(long long)actual[i], (long long)expected);
			failure = 1;
		}
	}
	if (item != nox_gui_itemAmount_item_1319256 ||
		item->pos.x != (int)window_x + (int32_t)dword_587000_183456 ||
		item->pos.y != (int)window_y + (int32_t)dword_587000_183460) failure = 1;
	event(20);
	draw_count++;
	if (mutate_viewport) {
		vp->x1 = -99;
		vp->field_11 = UINTPTR_MAX;
	}
	return -17;
}

// PRODUCTION_DRAW_BODY

static int check_draw(const uint32_t words[13], unsigned int tagged, unsigned int pressed,
	unsigned int x, unsigned int y, int mutate) {
	case_count++;
	memset(pe_record, 0xA5, sizeof(pe_record));
	memcpy(pe_record, words, 13 * sizeof(uint32_t));
	dword_5d4594_1319264 = tagged;
	dword_587000_183456 = (uint32_t)-14;
	dword_587000_183460 = 17;
	window_x = x;
	window_y = y;
	mutate_viewport = mutate;
	failure = 0;
	event_count = 0;
	draw_count = 0;
	for (unsigned int i = 0; i < 4; i++) children[i]->draw_data.field_0 = (pressed & (1U << i)) ? 0xA4 : 0xA0;
	uint32_t old_record[40];
	memcpy(old_record, pe_record, sizeof(old_record));
	nox_drawable wanted_item = *nox_gui_itemAmount_item_1319256;
	wanted_item.pos.x = (int)x - 14;
	wanted_item.pos.y = (int)y + 17;
	nox_window old_window = *window;
	nox_window old_children[4];
	for (unsigned int i = 0; i < 4; i++) old_children[i] = *children[i];
	CHECK(sub_4C0030(window, (void*)(uintptr_t)0xDEADBEEF) == 1);
	CHECK(!failure && draw_count == 1);
	CHECK(!memcmp(pe_record, old_record, sizeof(old_record)));
	CHECK(!memcmp(nox_gui_itemAmount_item_1319256, &wanted_item, sizeof(wanted_item)));
	CHECK(!memcmp(window, &old_window, sizeof(old_window)));
	for (unsigned int i = 0; i < 4; i++) CHECK(!memcmp(children[i], &old_children[i], sizeof(old_children[i])));
	unsigned int expected[20] = {1, 2, 10 + (tagged ? NOX_ITEM_AMOUNT_IMAGE_BASE : NOX_ITEM_AMOUNT_IMAGE_BASE_NO_TAG), 20};
	unsigned int n = 4;
	const unsigned int images[] = {NOX_ITEM_AMOUNT_IMAGE_DOWN_LIT, NOX_ITEM_AMOUNT_IMAGE_UP_LIT,
		tagged ? NOX_ITEM_AMOUNT_IMAGE_YES_PRESSED : NOX_ITEM_AMOUNT_IMAGE_YES_PRESSED_NO_TAG,
		tagged ? NOX_ITEM_AMOUNT_IMAGE_NO_PRESSED : NOX_ITEM_AMOUNT_IMAGE_NO_PRESSED_NO_TAG};
	for (unsigned int i = 0; i < 4; i++) {
		expected[n++] = 30 + i;
		if (pressed & (1U << i)) expected[n++] = 10 + images[i];
	}
	CHECK(event_count == n && !memcmp(events, expected, n * sizeof(events[0])));
	return 0;
}
int main(void) {
	window = calloc(1, sizeof(*window));
	nox_gui_itemAmount_dialog_1319228 = window;
	nox_gui_itemAmount_item_1319256 = calloc(1, sizeof(*nox_gui_itemAmount_item_1319256));
	if (!window || !nox_gui_itemAmount_item_1319256) return 2;
	nox_gui_itemAmount_item_1319256->draw_func = draw_icon;
	for (unsigned int i = 0; i < 4; i++) {
		children[i] = calloc(1, sizeof(*children[i]));
		if (!children[i]) return 2;
	}
	for (unsigned int i = 0; i < 8; i++) {
		nox_gui_itemAmount_images_1319196[i] = malloc(1);
		if (!nox_gui_itemAmount_images_1319196[i]) return 2;
	}
	if (sizeof(void*) == 8) {
		CHECK((uintptr_t)window > UINT32_MAX && (uintptr_t)nox_gui_itemAmount_item_1319256 > UINT32_MAX);
		for (unsigned int i = 0; i < 4; i++) CHECK((uintptr_t)children[i] > UINT32_MAX);
		for (unsigned int i = 0; i < 8; i++) CHECK((uintptr_t)nox_gui_itemAmount_images_1319196[i] > UINT32_MAX);
	}
	const uint32_t viewports[][13] = {
		{0, 0, 1024, 768, 0, 0, 0, 0, 1024, 768, 0, 0, 0},
		{0, 0, 640, 480, 0, 0, 0, 0, 640, 480, 0, 0, 0},
		{10, 20, 300, 400, 50, 60, 700, 800, 290, 380, 0xFEDCBA98, 0x89ABCDEF, 12},
		{0x80000000, 0xFFFFFFFF, 0x7FFFFFFF, 0xFFFFFFFE, 0xFFFFFFF0, 0x80000001, 0x80000002,
			0x80000003, 0xFFFFFFE0, 0xFFFFFFD0, 0xFFFFFFFF, 0x80000000, 0xFFFFFFF4},
		{0, 0, 1280, 720, 0, 0, 0, 0, 1280, 720, 0, 0, 0},
	};
	const unsigned int positions[][2] = {{100, 200}, {810, 604}, {0, 0}};
	for (unsigned int v = 0; v < sizeof(viewports) / sizeof(viewports[0]); v++) {
		for (unsigned int tag = 0; tag < 2; tag++) {
			for (unsigned int press = 0; press < 16; press++) {
				for (unsigned int p = 0; p < sizeof(positions) / sizeof(positions[0]); p++) {
					for (int mutate = 0; mutate < 2; mutate++) {
						CHECK(check_draw(viewports[v], tag, press, positions[p][0], positions[p][1], mutate) == 0);
					}
				}
			}
		}
	}
	for (unsigned int i = 0; i < 8; i++) free(nox_gui_itemAmount_images_1319196[i]);
	for (unsigned int i = 0; i < 4; i++) free(children[i]);
	free(nox_gui_itemAmount_item_1319256);
	free(window);
	printf("item-amount drawer: %u native viewport/icon/overlay cases passed\n", case_count);
	return 0;
}
