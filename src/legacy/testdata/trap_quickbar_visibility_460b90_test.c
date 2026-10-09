#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../client__gui__window.h"

static unsigned char *main_data, *trap_data;
static void* nox_xxx_aClosewoodengat_587000_133480;
static nox_window *main_root, *trap_root, *buttons[2][5], *nuggets[2][5];
static nox_window *dword_5d4594_1049500, *dword_5d4594_1049504, *dword_5d4594_1049520;
static nox_window *dword_5d4594_1049508, *dword_5d4594_1049512;
static int dword_5d4594_1049484;
static uint32_t expanded;
static unsigned int trap_writes, toggle_calls, sounds, cases;

#define CHECK(value) do { if (!(value)) { \
	fprintf(stderr, "case %u line %d: %s (trap=%d expanded=%u writes=%u)\n", \
		cases, __LINE__, #value, dword_5d4594_1049484, expanded, trap_writes); return 1; \
} } while (0)

static void* getMemAt(uintptr_t base, uintptr_t offset) {
	if (base != 0x5D4594 || offset != 1047940) abort();
	return trap_data;
}
static uint32_t* getMemU32Ptr(uintptr_t base, uintptr_t offset) {
	if (base != 0x5D4594 || offset != 1049476) abort();
	return &expanded;
}
static nox_window* nox_quickbar_root(void* data) {
	if (data == main_data) return main_root;
	if (data == trap_data) return trap_root;
	abort();
}
static nox_window* nox_quickbar_button(void* data, int slot) {
	if (slot < 0 || slot >= 5) abort();
	if (data == main_data) return buttons[0][slot];
	if (data == trap_data) return buttons[1][slot];
	abort();
}
static nox_window* nox_quickbar_nugget(void* data, int slot) {
	if (slot < 0 || slot >= 5) abort();
	if (data == main_data) return nuggets[0][slot];
	if (data == trap_data) return nuggets[1][slot];
	abort();
}
int nox_window_set_hidden(nox_window* win, int hidden) {
	if (!win) return -2;
	if (win->parent == trap_root) trap_writes++;
	if (hidden) win->flags |= NOX_WIN_HIDDEN;
	else win->flags &= ~NOX_WIN_HIDDEN;
	return 0;
}
static void sub_460920(void) { toggle_calls++; expanded = 0; }
static void nox_xxx_quickBarClose_4606B0(void) { toggle_calls++; expanded = 0; }
static int sub_46AE10(nox_window* win, int active) {
	if (win != dword_5d4594_1049500) abort();
	win->draw_data.field_0 = active != 0;
	return 0;
}
static void nox_xxx_clientPlaySoundSpecial_452D80(int id, int volume) {
	if ((id != 796 && id != 797) || volume != 100) abort();
	sounds++;
}

// PRODUCTION_VISIBILITY
// PRODUCTION_CLOSE
// PRODUCTION_OPEN

static nox_window* new_window(nox_window* parent) {
	nox_window* win = calloc(1, sizeof(*win));
	if (!win) abort();
	win->parent = parent;
	win->flags = (nox_window_flags)1032;
	if (sizeof(void*) == 8 && (uintptr_t)win <= UINT32_MAX) abort();
	return win;
}
static int hidden(nox_window* win) { return (win->flags & NOX_WIN_HIDDEN) != 0; }
static int visible(nox_window* win) {
	for (; win; win = win->parent) if (hidden(win)) return 0;
	return 1;
}
static int configuration(int hud, int trap, unsigned int expansion, unsigned int mask) {
	cases++;
	dword_5d4594_1049484 = trap;
	expanded = expansion;
	trap_root->flags = (nox_window_flags)1048;
	for (int i = 0; i < 3; i++) {
		buttons[1][i]->flags = (nox_window_flags)(1032 | ((mask >> (2 * i)) & 1) * NOX_WIN_HIDDEN);
		nuggets[1][i]->flags = (nox_window_flags)(1160 | ((mask >> (2 * i + 1)) & 1) * NOX_WIN_HIDDEN);
	}
	unsigned char saved[256];
	memcpy(saved, trap_data, sizeof(saved));
	trap_writes = toggle_calls = sounds = 0;
	CHECK(sub_460B90(hud) == 1);
	CHECK(trap_writes == (hud && trap ? 6U : 0U));
	CHECK(toggle_calls == (!hud && expansion == 1 ? 1U : 0U));
	CHECK(sounds == 0 && dword_5d4594_1049484 == trap);
	CHECK(hidden(trap_root) == (!hud || !trap));
	for (int i = 0; i < 3; i++) {
		CHECK(buttons[1][i]->flags == (nox_window_flags)(1032 | (hud && trap ? 0 : ((mask >> (2 * i)) & 1) * NOX_WIN_HIDDEN)));
		CHECK(nuggets[1][i]->flags == (nox_window_flags)(1160 | (hud && trap ? 0 : ((mask >> (2 * i + 1)) & 1) * NOX_WIN_HIDDEN)));
	}
	for (int i = 0; i < 5; i++) {
		CHECK(hidden(buttons[0][i]) == !hud && hidden(nuggets[0][i]) == !hud);
	}
	CHECK(hidden(main_root) == !hud);
	CHECK(hidden(dword_5d4594_1049512) == (!hud || trap != 0));
	CHECK(!memcmp(saved, trap_data, sizeof(saved)));
	return 0;
}
static int lifecycle(void) {
	cases++;
	dword_5d4594_1049484 = 0;
	expanded = 0;
	for (int i = 0; i < 3; i++) buttons[1][i]->flags = (nox_window_flags)1032;
	trap_root->flags = (nox_window_flags)1048;
	CHECK(sub_460B90(0) == 1 && sub_460B90(1) == 1);
	sub_461060();
	for (int i = 0; i < 3; i++) CHECK(visible(buttons[1][i]));
	CHECK(dword_5d4594_1049484 == 1);
	CHECK(sub_460B90(0) == 1);
	for (int i = 0; i < 3; i++) CHECK(!visible(buttons[1][i]));
	CHECK(sub_460B90(1) == 1);
	for (int i = 0; i < 3; i++) CHECK(visible(buttons[1][i]));
	sub_461060();
	for (int i = 0; i < 3; i++) CHECK(!visible(buttons[1][i]) && !hidden(buttons[1][i]));
	sub_461060();
	for (int i = 0; i < 3; i++) CHECK(visible(buttons[1][i]));
	expanded = 1;
	sub_461060();
	sub_461060();
	CHECK(expanded == 0);
	for (int i = 0; i < 3; i++) CHECK(visible(buttons[1][i]));
	return 0;
}
int main(void) {
	main_data = calloc(1, 256);
	trap_data = malloc(256);
	if (!main_data || !trap_data) return 2;
	if (sizeof(void*) == 8 && ((uintptr_t)main_data <= UINT32_MAX || (uintptr_t)trap_data <= UINT32_MAX)) return 2;
	memset(trap_data, 0xA5, 256);
	nox_xxx_aClosewoodengat_587000_133480 = main_data;
	main_root = new_window(NULL);
	trap_root = new_window(NULL);
	dword_5d4594_1049500 = new_window(NULL);
	dword_5d4594_1049504 = new_window(NULL);
	dword_5d4594_1049520 = new_window(NULL);
	dword_5d4594_1049508 = new_window(NULL);
	dword_5d4594_1049512 = new_window(NULL);
	for (int i = 0; i < 5; i++) {
		buttons[0][i] = new_window(main_root);
		nuggets[0][i] = new_window(main_root);
		if (i < 3) {
			buttons[1][i] = new_window(trap_root);
			nuggets[1][i] = new_window(trap_root);
		}
	}
	// First reproduce the user-visible close/HUD/open sequence.
	CHECK(lifecycle() == 0);
	const int values[] = {0, 1, 2, -1};
	for (unsigned int h = 0; h < 4; h++)
		for (unsigned int t = 0; t < 4; t++)
			for (unsigned int e = 0; e < 2; e++)
				for (unsigned int m = 0; m < 64; m++)
					CHECK(configuration(values[h], values[t], e, m) == 0);
	printf("Trap Set: %u native visibility/lifecycle cases, 3 slots, numeric entries unchanged\n", cases);
	return 0;
}
