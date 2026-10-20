#include <assert.h>
#include <limits.h>
#include <stdio.h>
#include "../defs.h"
#include "../GAME1.h"
#include "../GAME1_2.h"
#include "../GAME1_3.h"
#include "../common__strman.h"

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wsign-compare"
#ifdef __clang__
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types-discards-qualifiers"
#else
#pragma GCC diagnostic ignored "-Wdiscarded-qualifiers"
#endif
#include "../noxstring.c"
#pragma GCC diagnostic pop

char* nox_itoa(int value, char* dest, int radix) {
	if (radix == 10) sprintf(dest, "%d", value);
	else if (radix == 8) sprintf(dest, "%o", (unsigned int)value);
	else sprintf(dest, "%x", (unsigned int)value);
	return dest;
}

uintptr_t dword_8531A0_2576;
uint32_t nox_color_white_2523948 = 0x12345678;
static nox_drawable* local_player;
static nox_playerInfo* profile;
static nox_draw_viewport_t viewport;
static struct { uint32_t before; wchar2_t text[128]; uint32_t after; } buffer;
static const uint32_t pe32_classes[] = {0x58E904, 0x58E90C, 0x58E914, 0xABCDEF12};
static char* native_classes[3];
static const char* class_names[] = {"Warrior", "Wizard", "Conjurer"};
static wchar2_t* class_labels[] = {(wchar2_t*)L"전사", (wchar2_t*)L"마법사", (wchar2_t*)L"소환술사"};
static wchar2_t* expected_info[] = {
	(wchar2_t*)L"레벨:17 직업:전사", (wchar2_t*)L"레벨:17 직업:마법사", (wchar2_t*)L"레벨:17 직업:소환술사"
};
static unsigned selected_class, draws, class_reads, locale_reads;
static int has_local, has_profile;
static const int draw_result = INT_MIN + 71;

void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base == 0x5D4594 && off == 811120) return buffer.text;
	if (base == 0x587000 && off >= 29456 && off <= 29464) {
		// A raw 64-bit read at Wizard's slot reproduces 0x58e9140058e90c.
		return (unsigned char*)pe32_classes + off - 29456;
	}
	abort();
}
void* mem_getPtrValue(uintptr_t base, uintptr_t off) {
	if (base == 0x852978 && off == 8) return has_local ? local_player : NULL;
	if (base == 0x587000 && off == 29456 + 4 * selected_class) {
		++class_reads;
		return native_classes[selected_class];
	}
	abort();
}
int16_t* mem_getI16Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base, off); }
int nox_xxx_guiFontHeightMB_43F320(void* font) { assert(!font); return 13; }
nox_draw_viewport_t* nox_draw_getViewport_437250(void) { return &viewport; }
void nox_xxx_drawSetTextColor_434390(int color) { assert((uint32_t)color == nox_color_white_2523948); }
char* nox_server_currentMapGetFilename_409B30(void) { return "War02b"; }
wchar2_t* nox_strman_loadString_40F1D0(char* key, char** out, char* file, int line) {
	assert(!out && !strcmp(file, "C:\\NoxPost\\src\\client\\System\\client.c"));
	assert(has_local && has_profile);
	if (line == 1357) {
		// Compare identity before dereferencing the reported noncanonical pointer.
		if (key != native_classes[selected_class]) {
			fprintf(stderr, "class key pointer=0x%llx, expected native pointer=0x%llx\n",
				(unsigned long long)(uintptr_t)key,
				(unsigned long long)(uintptr_t)native_classes[selected_class]);
			exit(42);
		}
		assert(!strcmp(key, class_names[selected_class]));
		++locale_reads;
		return class_labels[selected_class];
	}
	assert(line == 1355 && !strcmp(key, "PlayerInfo") && locale_reads == 1);
	++locale_reads;
	return (wchar2_t*)L"레벨:%d 직업:%s";
}
int nox_xxx_drawString_43F6E0(void* font, wchar2_t* text, int x, int y) {
	assert(!font && text == buffer.text && x == 3 && y == 107 + 13 * (int)draws);
	if (draws == 0) assert(!nox_wcscmp(text, (wchar2_t*)L"War02b"));
	else if (draws == 1) assert(!nox_wcscmp(text, (wchar2_t*)L"X:-123\tY:456"));
	else {
		assert(draws == 2 && locale_reads == 2);
		assert(!nox_wcscmp(text, expected_info[selected_class]));
	}
	++draws;
	return draw_result;
}

#pragma GCC diagnostic push
// The retained drawer passes its original signed UTF-16 memory view to the
// unsigned UTF-16 drawing API. This is independent of the pointer-table fault.
#pragma GCC diagnostic ignored "-Wpointer-sign"
// COMPLETE_PLAYER_INFO_DRAWER
#pragma GCC diagnostic pop

int main(int argc, char** argv) {
	assert(argc == 2);
	selected_class = !strcmp(argv[1], "Conjurer") ? 2 : !strcmp(argv[1], "Warrior") ? 0 : 1;
	has_local = strcmp(argv[1], "NoLocalPlayer") != 0;
	has_profile = strcmp(argv[1], "NoPlayerInfo") != 0;
	local_player = calloc(1, sizeof(*local_player));
	profile = calloc(1, sizeof(*profile));
	assert(local_player && profile);
	for (unsigned i = 0; i < 3; ++i) {
		native_classes[i] = malloc(16);
		assert(native_classes[i]);
		strcpy(native_classes[i], class_names[i]);
		if (sizeof(void*) > 4) assert((uintptr_t)native_classes[i] > UINT32_MAX);
	}
	dword_8531A0_2576 = has_profile ? (uintptr_t)profile : 0;
	profile->info.playerClass = selected_class;
	profile->field_3684 = 17;
	local_player->pos.x = -123;
	local_player->pos.y = 456;
	viewport.x1 = -7; viewport.y1 = 17;
	buffer.before = 0x13579BDF; buffer.after = 0x2468ACE0;
	int result = sub_436F50();
	assert(result == (has_local && has_profile ? draw_result : has_local));
	assert(draws == (has_local && has_profile ? 3u : 1u));
	assert(class_reads == (has_local && has_profile ? 1u : 0u));
	assert(locale_reads == (has_local && has_profile ? 2u : 0u));
	assert(buffer.before == 0x13579BDF && buffer.after == 0x2468ACE0);
	assert(pe32_classes[0] == 0x58E904 && pe32_classes[1] == 0x58E90C && pe32_classes[2] == 0x58E914);
	assert(profile->field_3684 == 17 && profile->info.playerClass == selected_class);
	printf("player info %s: PASS (native pointer bytes=%zu, draws=%u)\n", argv[1], sizeof(void*), draws);
	for (unsigned i = 0; i < 3; ++i) free(native_classes[i]);
	free(local_player); free(profile);
	return 0;
}
