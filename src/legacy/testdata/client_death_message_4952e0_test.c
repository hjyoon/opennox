#include <stdarg.h>
#include <stdio.h>
#include "../defs.h"
#include "../client__gui__guicon.h"
#include "../common__strman.h"

// Keep the real UTF-16 formatting and concatenation. Interpose only the
// one-name varargs entry to reject a bad pointer before a diagnostic crash.
#define nox_swprintf fixture_production_swprintf
#pragma GCC diagnostic push
#ifdef __clang__
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types-discards-qualifiers"
#else
#pragma GCC diagnostic ignored "-Wdiscarded-qualifiers"
#endif
#include "../noxstring.c"
#pragma GCC diagnostic pop
#undef nox_swprintf

char* nox_itoa(int value, char* dst, int radix) {
	if (radix == 10) sprintf(dst, "%d", value);
	else if (radix == 8) sprintf(dst, "%o", (unsigned int)value);
	else sprintf(dst, "%x", (unsigned int)value);
	return dst;
}

static nox_playerInfo* players[3];
static uint16_t codes[3];
static wchar2_t attacker_format[] = {'<','%','s','>',0};
static wchar2_t victim_format[] = {'[','%','s',']',0};
static wchar2_t nasty[] = {0xb098,0xc05c,0xac83,0};
static wchar2_t final_format[] = {'%','s',' ','%','s',0};
static uint32_t ring[600], guards[2] = {0x12345678, 0x89abcdef};
static uint32_t remap[4];
static uint32_t dword_5d4594_1203836, dword_5d4594_1203840;
static int trace[16], trace_count, bad_pointer, print_count, print_result;
static int mutate_packet, mutate_name, cases;
static uint16_t* live_words;
static wchar2_t printed[512];
static unsigned char printed_color;

#define CHECK(value) do { if (!(value)) { \
	fprintf(stderr, "case %d line %d: %s bad-pointer=%d\n", cases, __LINE__, #value, bad_pointer); return 1; \
} } while (0)

char* nox_common_playerInfoGetByID_417040(int code) {
	trace[trace_count++] = code;
	for (int i = 0; i < 3; i++)
		if (code == codes[i] && players[i]->active) return (char*)players[i];
	return NULL;
}

wchar2_t* nox_strman_loadString_40F1D0(char* name, char** extra, char* file, int line) {
	(void)extra;
	(void)file;
	trace[trace_count++] = -line;
	if (mutate_packet && line == 48) live_words[2] = codes[1];
	if (mutate_name && line == 72) players[2]->name_final[0] = 0xd53c;
	if (line == 65 && strcmp(name, "die.c:AttackerNasty") == 0) return nasty;
	if (line == 72 && strcmp(name, "die.c:LocalizeVictim") == 0) return victim_format;
	if ((line == 48 || line == 64) && strcmp(name, "die.c:LocalizeAttacker") == 0) return attacker_format;
	abort();
}

int nox_swprintf(wchar2_t* dst, const wchar2_t* format, ...) {
	va_list ap, check;
	va_start(ap, format);
	va_copy(check, ap);
	const wchar2_t* name = va_arg(check, const wchar2_t*);
	va_end(check);
	int valid = name == nasty;
	for (int i = 0; i < 3; i++) valid |= name == players[i]->name_final;
	if (!valid) {
		bad_pointer++;
		dst[0] = 0;
		va_end(ap);
		return 0;
	}
	int result = nox_vswprintf(dst, format, ap);
	va_end(ap);
	return result;
}

int nox_gui_console_Print_450B90(unsigned char color, wchar2_t* message) {
	print_count++;
	printed_color = color;
	nox_wcscpy(printed, message);
	trace[trace_count++] = -450;
	return print_result;
}

void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base == 0x587000 && off == 161668) return final_format;
	abort();
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) {
	if (base != 0x5d4594) abort();
	if (off >= 1201428 && off < 1201428 + sizeof(ring) && off % 4 == 0)
		return &ring[(off - 1201428) / 4];
	if (off == 1203844) return &remap[0];
	if (off == 1203848) return &remap[1];
	abort();
}
uint16_t* mem_getU16Ptr(uintptr_t base, uintptr_t off) {
	if (base != 0x5d4594) abort();
	if (off == 1203856) return (uint16_t*)&remap[2];
	if (off == 1203852) return (uint16_t*)&remap[3];
	abort();
}
uint32_t gameFrame(void) { return UINT32_C(0xf1234567); }

// PRODUCTION_DEATH_MESSAGE
#pragma GCC diagnostic push
#ifdef __clang__
#pragma GCC diagnostic ignored "-Wc23-extensions"
#endif
// PRODUCTION_DEATH_PACKET
#pragma GCC diagnostic pop

static void reset(void) {
	trace_count = bad_pointer = print_count = mutate_packet = mutate_name = 0;
	memset(printed, 0, sizeof(printed));
	for (int i = 0; i < 3; i++) {
		memset(players[i], 0x5a, sizeof(*players[i]));
		players[i]->active = 1;
		// Native name and obsolete PE32 name bytes deliberately disagree.
		if (offsetof(nox_playerInfo, name_final) != 4704)
			memset((uint8_t*)players[i] + 4704, 0, 56);
		players[i]->name_final[0] = (wchar2_t)(0xac00 + i);
		players[i]->name_final[1] = 0xd83d;
		players[i]->name_final[2] = 0xde80;
		players[i]->name_final[3] = 0;
	}
}

static int run_message(int attacker, int assist, int mutation, int result) {
	cases++;
	reset();
	uint16_t words[6] = {0xa90e, codes[0], codes[1], codes[2], 0x5678, 0x42};
	if (attacker == 0) words[1] = 0;
	if (attacker == 1) players[0]->active = 0;
	if (assist == 0 || mutation) words[2] = 0;
	if (assist == 1) players[1]->active = 0;
	mutate_packet = mutate_name = mutation;
	live_words = words;
	print_result = result;
	nox_playerInfo before[3];
	for (int i = 0; i < 3; i++) memcpy(&before[i], players[i], sizeof(before[i]));
	uint16_t words_before[6];
	memcpy(words_before, words, sizeof(words));
	wchar2_t expected[512], attacker_text[64], assistant_text[32], victim_text[32];
	nox_swprintf(attacker_text, attacker_format, attacker == 2 ? players[0]->name_final : nasty);
	if (attacker == 2 && (assist == 2 || mutation)) {
		nox_swprintf(assistant_text, L" + %s", players[1]->name_final);
		nox_wcscat(attacker_text, assistant_text);
	}
	wchar2_t victim_name[28];
	memcpy(victim_name, players[2]->name_final, sizeof(victim_name));
	if (mutation) victim_name[0] = 0xd53c;
	fixture_production_swprintf(victim_text, victim_format, victim_name);
	fixture_production_swprintf(expected, final_format, victim_text, attacker_text);
	CHECK(sub_4952E0(words) == result);
	CHECK(bad_pointer == 0 && print_count == 1 && printed_color == NOX_CONSOLE_WHITE);
	CHECK(nox_wcscmp(printed, expected) == 0);
	int expected_trace[16], count = 0;
	if (attacker != 0) expected_trace[count++] = codes[0];
	if (attacker == 2) {
		expected_trace[count++] = -48;
		if (assist != 0 || mutation) expected_trace[count++] = codes[1];
	} else { expected_trace[count++] = -65; expected_trace[count++] = -64; }
	expected_trace[count++] = codes[2];
	expected_trace[count++] = -72;
	expected_trace[count++] = -450;
	CHECK(trace_count == count && memcmp(trace, expected_trace, count * sizeof(int)) == 0);
	if (mutation) { words_before[2] = codes[1]; before[2].name_final[0] = 0xd53c; }
	CHECK(memcmp(words, words_before, sizeof(words)) == 0);
	for (int i = 0; i < 3; i++) CHECK(memcmp(players[i], &before[i], sizeof(before[i])) == 0);
	return 0;
}

static int run_packet(int index, int oldest, int rewrite) {
	cases++;
	reset();
	memset(ring, 0xa5, sizeof(ring));
	uint32_t before[600];
	memcpy(before, ring, sizeof(ring));
	dword_5d4594_1203836 = index;
	dword_5d4594_1203840 = oldest;
	remap[0] = 0x2468; remap[1] = 0x1357; remap[2] = 0xabcd; remap[3] = 0xfedc;
	uint8_t bytes[15] = {0x91,0x82,0xa9,0x0e};
	uint8_t* packet = bytes + 2; // genuine unaligned little-endian packet
	packet[2] = codes[0]; packet[3] = codes[0] >> 8;
	packet[4] = codes[1]; packet[5] = codes[1] >> 8;
	packet[6] = codes[2]; packet[7] = codes[2] >> 8;
	uint16_t weapon = rewrite == 2 ? 0x1357 : 0x2468;
	packet[8] = weapon; packet[9] = weapon >> 8;
	packet[10] = rewrite ? 1 : 2;
	bytes[13] = 0x73; bytes[14] = 0x64;
	uint8_t expected[15]; memcpy(expected, bytes, sizeof(bytes));
	if (rewrite) { weapon = rewrite == 2 ? 0xfedc : 0xabcd; expected[10] = weapon; expected[11] = weapon >> 8; }
	print_result = -91;
	CHECK(sub_495210(packet) == -91);
	CHECK(bad_pointer == 0 && print_count == 1 && printed_color == NOX_CONSOLE_WHITE);
	uint32_t row[] = {codes[0],codes[1],codes[2],weapon,packet[10],UINT32_C(0xf1234567)};
	memcpy(before + 6 * index, row, sizeof(row));
	CHECK(memcmp(ring, before, sizeof(ring)) == 0 && memcmp(bytes, expected, sizeof(bytes)) == 0);
	CHECK(dword_5d4594_1203836 == (uint32_t)((index + 1) % 100));
	CHECK(dword_5d4594_1203840 == (uint32_t)(oldest == (index + 1) % 100 ? (oldest + 1) % 100 : oldest));
	CHECK(guards[0] == 0x12345678 && guards[1] == 0x89abcdef);
	return 0;
}

int main(void) {
	for (int i = 0; i < 3; i++) {
		players[i] = malloc(sizeof(*players[i]));
		CHECK(players[i] != NULL);
		if (sizeof(void*) == 8) CHECK((uintptr_t)players[i] > UINT32_MAX && offsetof(nox_playerInfo, name_final) != 4704);
	}
	const uint16_t code_sets[][3] = {{1,2,3},{0x7fff,0x8000,0xffff}};
	const int results[] = {0,1,-91};
	for (int c = 0; c < 2; c++) {
		memcpy(codes, code_sets[c], sizeof(codes));
		for (int a = 0; a < 3; a++) for (int b = 0; b < 3; b++) for (int r = 0; r < 3; r++)
			CHECK(run_message(a, b, 0, results[r]) == 0);
		CHECK(run_message(2, 0, 1, 7) == 0);
	}
	for (int rewrite = 0; rewrite < 3; rewrite++) {
		CHECK(run_packet(0, 57, rewrite) == 0);
		CHECK(run_packet(99, 0, rewrite) == 0);
	}
	printf("Death message: %d native cases; UTF-16/surrogate names; original lookup/localization order, live callback reads, console result and unaligned packet ring preserved; Player=%p name-offset=%zu\n",
		cases, (void*)players[0], offsetof(nox_playerInfo, name_final));
	for (int i = 0; i < 3; i++) free(players[i]);
	return 0;
}
