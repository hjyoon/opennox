#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../defs.h"
#include "../client__gui__window.h"

// Declared service boundaries match the bounded unchanged PE32 corpus. The
// handler, packed records, native-width capture pointers and rectangle are real.
static unsigned char *bars[2];
static nox_window *buttons[2][5], *capture, *dword_5d4594_1049532;
static void *dword_8531A0_2576, *dword_5d4594_1049692, *dword_5d4594_1049696;
static uint32_t targeting, dword_5d4594_1047932;
static int last_target;
static int4 rectangle = {352, 677, 672, 768};
static int bar, slot, target, x, y;
static uint32_t dropped;
static char trace[512];

#define REQUIRE(expr) do { if (!(expr)) { \
	fprintf(stderr, "line %d: %s\n", __LINE__, #expr); abort(); \
} } while (0)
static void record(const char* call) {
	REQUIRE(strlen(trace) + strlen(call) + 2 < sizeof(trace));
	if (*trace) strcat(trace, ",");
	strcat(trace, call);
}
static unsigned char* nox_quickbar_selected_row(void* data) {
	REQUIRE(data == bars[0] || data == bars[1]);
	unsigned char* bytes = data;
	REQUIRE(bytes[200] < 5);
	return bytes + 40 * bytes[200];
}
static void* nox_quickbar_data_for_window(nox_window* win) {
	for (int b = 0; b < 2; b++)
		for (int s = 0; s < 5; s++) if (buttons[b][s] == win) return bars[b];
	return NULL;
}
static nox_window* nox_quickbar_button(void* data, int s) {
	REQUIRE(s >= 0 && s < 5 && (data == bars[0] || data == bars[1]));
	return buttons[data == bars[1]][s];
}
void* getMemAt(uintptr_t base, uintptr_t offset) {
	if (base == 0x5D4594 && offset == 1047940) return bars[1];
	REQUIRE(base == 0x587000 && offset == 133656);
	return &rectangle;
}
uint32_t* getMemU32Ptr(uintptr_t base, uintptr_t offset) {
	REQUIRE(base == 0x5D4594 && offset == 1047928);
	return &targeting;
}
int* getMemIntPtr(uintptr_t base, uintptr_t offset) {
	REQUIRE(base == 0x5D4594 && offset == 1047556);
	return &last_target;
}
static int nox_client_localPlayerHasFlags30(int flags) { REQUIRE(flags == 2); return 0; }
static int nox_xxx_get_57AF20(void) { return 0; }
static int nox_xxx_spellIsEnabled_424B70(uint32_t spell) { REQUIRE(spell != 0); return 1; }
int nox_xxx_wndSetCaptureMain_46ADC0(nox_window* win) {
	REQUIRE(win == buttons[bar][slot]); capture = win; record("capture"); return 0;
}
nox_window* nox_xxx_wndGetCaptureMain_46AE00(void) { return capture; }
int nox_xxx_wndClearCaptureMain_46ADE0(nox_window* win) {
	REQUIRE(win == buttons[bar][slot]); capture = NULL; record("release"); return 0;
}
static void nox_xxx_bookSaveSpellForDragDrop_477640(uint32_t spell, int kind) {
	REQUIRE(spell == *(uint32_t*)(nox_quickbar_selected_row(bars[bar]) + 8 * slot) && kind == 2);
	record("drag");
}
static void nox_xxx_clientPlaySoundSpecial_452D80(int id, int volume) {
	REQUIRE(volume == 100 && (id == 793 || id == 794)); record(id == 793 ? "pickup" : "swap-sound");
}
static int nox_xxx_spellBoxPointToWnd_45DE60(void* data, int px, int py) {
	REQUIRE(data == bars[bar] && px == x && py == y); record("hit"); return target;
}
static unsigned int nox_xxx_bookSpellDrop_45DCA0(uint32_t spell, unsigned char flag, int px, int py, void* data) {
	unsigned char* row = nox_quickbar_selected_row(data);
	REQUIRE(data == bars[bar] && px == x && py == y);
	REQUIRE(spell == *(uint32_t*)(row + 8 * slot) && flag == row[8 * slot + 4]);
	record("drop"); return dropped;
}
static void nox_xxx_clientSwapQuickbarKeys_45F300(void* data, int first, int second) {
	REQUIRE(data == bars[bar] && first == slot && second == target);
	unsigned char* row = nox_quickbar_selected_row(data);
	unsigned char saved[8];
	memcpy(saved, row + 8 * first, 8);
	memcpy(row + 8 * first, row + 8 * second, 8);
	memcpy(row + 8 * second, saved, 8); record("swap");
}
static void nox_xxx_updateSpellIcons_45DDF0(void* data) { REQUIRE(data == bars[bar]); record("icons"); }
static void nox_xxx_guiSpellSetCursor_45DF60(uint32_t spell, unsigned char flag) {
	unsigned char* row = nox_quickbar_selected_row(bars[bar]);
	REQUIRE(spell == *(uint32_t*)(row + 8 * slot) && flag == row[8 * slot + 4]); record("cursor");
}
static void sub_45DFC0(uint32_t spell) {
	REQUIRE(spell == *(uint32_t*)(nox_quickbar_selected_row(bars[bar]) + 8 * slot)); record("ability");
}
static void nox_xxx_clientStoreLastButton_45DAD0(int s) { REQUIRE(s == slot); record("store"); }
static void sub_4611B0(void) { abort(); }
static void nox_xxx_guiSpellTargetClickSet_45D9D0(int value) { (void)value; abort(); }
static void nox_xxx_bookSpellDnDclear_477660(void) { record("clear-dnd"); }

// PRODUCTION_RECTANGLE
// PRODUCTION_EVENT

int main(int argc, char** argv) {
	REQUIRE(argc == 12);
	bar = atoi(argv[1]); int selected = atoi(argv[2]); slot = atoi(argv[3]); int flag = atoi(argv[4]);
	int event = atoi(argv[5]); x = atoi(argv[6]); y = atoi(argv[7]); target = atoi(argv[8]);
	dropped = (uint32_t)strtoul(argv[9], NULL, 10); int mode = atoi(argv[10]), player_class = atoi(argv[11]);
	REQUIRE(bar >= 0 && bar < 2 && selected >= 0 && selected < 5 && slot >= 0 && slot < 5);
	for (int b = 0; b < 2; b++) {
		bars[b] = malloc(256); REQUIRE(bars[b]); memset(bars[b], 0xA5, 256);
		if (sizeof(void*) == 8) REQUIRE((uintptr_t)bars[b] > UINT32_MAX);
		for (int r = 0; r < 5; r++) for (int s = 0; s < 5; s++) {
			*(uint32_t*)(bars[b] + r * 40 + s * 8) = 27 + r * 5 + s;
			bars[b][r * 40 + s * 8 + 4] = (r + s) & 1;
		}
		bars[b][200] = b == bar ? selected : 0;
		for (int s = 0; s < 5; s++) {
			buttons[b][s] = calloc(1, sizeof(nox_window)); REQUIRE(buttons[b][s]);
			if (sizeof(void*) == 8) REQUIRE((uintptr_t)buttons[b][s] > UINT32_MAX);
		}
	}
	bars[bar][selected * 40 + slot * 8 + 4] = flag;
	nox_playerInfo* player = calloc(1, sizeof(*player)); REQUIRE(player);
	if (sizeof(void*) == 8) REQUIRE((uintptr_t)player > UINT32_MAX);
	player->info.playerClass = player_class; dword_8531A0_2576 = player;
	int press = mode == 3 ? -1 : nox_xxx_quickBarWnd_45EF50(buttons[bar][slot], 5, (743U << 16) | 438U);
	if (mode != 3) REQUIRE(dword_5d4594_1049692 == bars[bar] && dword_5d4594_1049696 == bars[bar] + 40 * selected);
	if (mode == 1) dword_5d4594_1049696 = NULL;
	if (mode == 2) bars[bar][200] = (selected + 1) % 5;
	int release = nox_xxx_quickBarWnd_45EF50(buttons[bar][slot], event, (unsigned)x | ((unsigned)y << 16));
	for (int b = 0; b < 2; b++) {
		for (int i = 0; i < 201; i++) printf("%02x", bars[b][i]);
		putchar('\n');
		for (int i = 201; i < 256; i++) REQUIRE(bars[b][i] == 0xA5);
		for (int s = 0; s < 5; s++) {
			const unsigned char* bytes = (const unsigned char*)buttons[b][s];
			for (size_t i = 0; i < sizeof(nox_window); i++) REQUIRE(bytes[i] == 0);
		}
	}
	puts(trace);
	printf("%d %d %d %d\n", press, release, capture != NULL, dword_5d4594_1049532 == NULL);
	for (int b = 0; b < 2; b++) { for (int s = 0; s < 5; s++) free(buttons[b][s]); free(bars[b]); }
	free(player);
	return 0;
}
