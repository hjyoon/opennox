#include <stdio.h>
#include "../defs.h"

static uintptr_t dword_8531A0_2576;
static nox_window* dword_5d4594_1193712;
static unsigned int calls, cases;
static int last_hidden;
static nox_window* last_window;

#define CHECK(value) do { if (!(value)) { \
	fprintf(stderr, "case %u line %d: %s\n", cases, __LINE__, #value); return 1; \
} } while (0)

// Visibility is the only external GUI service in these selected bodies.
// The complete native Player is never written by this service.
int nox_window_set_hidden(nox_window* win, int hidden) {
	calls++;
	last_hidden = hidden;
	last_window = win;
	if (!win) return -2;
	if (hidden) win->flags |= NOX_WIN_HIDDEN;
	else win->flags &= ~NOX_WIN_HIDDEN;
	return 0;
}

int nox_xxx_showObserverWindow_48CA70(int a1);
// PRODUCTION_OBSERVER_TOGGLE
// PRODUCTION_OBSERVER_VISIBILITY

static int run_case(nox_playerInfo* player, nox_window* win, uint32_t flags,
					uint8_t active, uint8_t decoy, int initially_hidden) {
	cases++;
	nox_playerInfo before_player;
	if (player) {
		memset(player, 0x5a, sizeof(*player));
		player->field_3680 = flags;
		player->active = active;
		// Opposing PE32-byte values must not determine LP64 visibility.
		if (offsetof(nox_playerInfo, field_3680) != 3680)
			((uint8_t*)player)[3680] = decoy;
		memcpy(&before_player, player, sizeof(*player));
	}
	nox_window before_window;
	if (win) {
		memset(win, 0x3c, sizeof(*win));
		win->flags = (nox_window_flags)(0x88 |
			(initially_hidden ? NOX_WIN_HIDDEN : 0));
		memcpy(&before_window, win, sizeof(*win));
	}
	dword_8531A0_2576 = (uintptr_t)player;
	dword_5d4594_1193712 = win;
	calls = 0;
	last_hidden = -77;
	last_window = NULL;
	int hidden = !player || !(flags & 1);
	CHECK(nox_xxx_cliToggleObsWindow_4357A0() == (win ? 0 : -2));
	CHECK(calls == 1 && last_hidden == hidden && last_window == win);
	CHECK(dword_8531A0_2576 == (uintptr_t)player && dword_5d4594_1193712 == win);
	if (player) CHECK(memcmp(player, &before_player, sizeof(*player)) == 0);
	if (win) {
		if (hidden) before_window.flags |= NOX_WIN_HIDDEN;
		else before_window.flags &= ~NOX_WIN_HIDDEN;
		CHECK(memcmp(win, &before_window, sizeof(*win)) == 0);
	}
	return 0;
}

int main(void) {
	nox_playerInfo* player = malloc(sizeof(*player));
	nox_window* win = malloc(sizeof(*win));
	if (!player || !win) return 2;
	if (sizeof(void*) == 8) {
		CHECK((uintptr_t)player > UINT32_MAX && (uintptr_t)win > UINT32_MAX);
		CHECK(offsetof(nox_playerInfo, field_3680) != 3680);
	}
	const uint32_t flags[] = {0, 1, 2, 3, 0x10, 0x11, 0x400, 0x401, UINT32_MAX};
	const uint8_t active[] = {0, 1, 2, UINT8_MAX};
	const uint8_t decoy[] = {0, 1, UINT8_MAX};
	for (unsigned int f = 0; f < sizeof(flags)/sizeof(flags[0]); f++)
		for (unsigned int a = 0; a < sizeof(active)/sizeof(active[0]); a++)
			for (unsigned int d = 0; d < sizeof(decoy)/sizeof(decoy[0]); d++)
				for (int hidden = 0; hidden < 2; hidden++)
					for (int present = 0; present < 2; present++)
						CHECK(run_case(player, present ? win : NULL, flags[f], active[a], decoy[d], hidden) == 0);
	for (int hidden = 0; hidden < 2; hidden++)
		for (int present = 0; present < 2; present++)
			CHECK(run_case(NULL, present ? win : NULL, 0, 0, 0, hidden) == 0);
	// The ordinary ready -> observer -> ready round trip must read live status.
	CHECK(run_case(player, win, 0x10, 1, 1, 0) == 0);
	CHECK(run_case(player, win, 0x11, 1, 0, 1) == 0);
	CHECK(run_case(player, win, 0x10, 1, 1, 0) == 0);
	printf("Observer window: %u native cases; Player=%p Window=%p status-offset=%zu; original low-bit gate, no active gate or Player mutation\n",
		cases, (void*)player, (void*)win, offsetof(nox_playerInfo, field_3680));
	free(win);
	free(player);
	return 0;
}
