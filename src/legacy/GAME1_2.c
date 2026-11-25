// For network byte-order conversions.
#ifdef _WIN32
#include <winsock.h>
#else
#include <arpa/inet.h>
#endif

#include <math.h>

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "GAME5.h"
#include "GAME5_2.h"
#include "client__system__parsecmd.h"
#include "common__net_list.h"
#include "common__system__settings.h"
#include "common__system__team.h"
#include "common__crypt.h"

#include "client__drawable__drawable.h"
#include "client__gui__gamewin__gamewin.h"
#include "client__gui__guiggovr.h"
#include "client__gui__guiquit.h"
#include "client__gui__window.h"
#include "client__shell__noxworld.h"
#include "client__shell__selchar.h"
#include "client__system__ctrlevnt.h"
#include "client__video__draw_common.h"

#include "client__draw__partscrn.h"
#include "client__gui__guicon.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME5_2.h"
#include "client__draw__fx.h"
#include "client__gui__guiinv.h"
#include "client__gui__guimeter.h"
#include "client__gui__guishop.h"
#include "client__gui__guispell.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"
#include "client__io__win95__focus.h"
#include "client__shell__optsback.h"
#include "client__system__ctrlevnt.h"
#include "client__video__draw_common.h"
#include "common/fs/nox_fs.h"
#include "common__binfile.h"
#include "common__log.h"
#include "common__magic__speltree.h"
#include "defs.h"
#include "input.h"
#include "input_common.h"
#include "operators.h"
#include "server__script__builtin.h"
#include "server__script__script.h"

#include <time.h>

extern uint32_t nox_xxx_useAudio_587000_80772;
extern uint32_t dword_5d4594_811904;
extern uint32_t dword_5d4594_805820;
extern uint32_t dword_5d4594_741648;
extern uint32_t dword_5d4594_815052;
extern uint32_t dword_5d4594_741652;
extern uint32_t dword_5d4594_3807116;
extern uint32_t dword_5d4594_3807152;
extern uint32_t dword_5d4594_3807136;
extern uint32_t dword_5d4594_814548;
extern uint32_t dword_5d4594_3807140;
extern uint32_t nox_client_connError_814552;
extern uint32_t dword_5d4594_815056;
extern uint32_t dword_5d4594_741356;
extern uint32_t dword_5d4594_741332;
extern uint32_t dword_5d4594_741364;
extern nox_gui_animation* nox_wnd_xxx_815040;
extern uint32_t dword_587000_87408;
extern uint32_t dword_5d4594_528252;
extern void* nox_alloc_screenParticles_806044;
extern uint32_t dword_5d4594_815044;
extern uint32_t nox_wol_server_result_cnt_815088;
extern uint32_t dword_5d4594_528256;
extern nox_window* dword_5d4594_815032;
extern nox_window* dword_5d4594_815020;
extern nox_window* dword_5d4594_815024;
extern nox_window* dword_5d4594_815028;
extern nox_window* dword_5d4594_814984;
extern nox_window* dword_5d4594_815016;
extern uint32_t nox_game_createOrJoin_815048;
extern uint32_t dword_587000_87412;
extern nox_window* dword_5d4594_815000;
extern nox_window* nox_wol_wnd_gameList_815012;
extern nox_window* dword_5d4594_815004;
extern nox_window* nox_wol_wnd_world_814980;
extern nox_window* dword_5d4594_1307716;
extern uint32_t nox_color_white_2523948;
extern uintptr_t dword_8531A0_2576;

int nox_win_width = 0;
int nox_win_height = 0;


obj_5D4594_754088_t* ptr_5D4594_754088 = 0;
int ptr_5D4594_754088_cnt = 0;

obj_5D4594_754088_t* ptr_5D4594_754092 = 0;
int ptr_5D4594_754092_cnt = 0;


nox_screenParticle* nox_screenParticles_head = 0;
nox_screenParticle* dword_5d4594_806052 = 0;

nox_gui_server_ent_t* dword_5d4594_814624 = 0;

void* dword_5d4594_805984 = 0;

//----- (004281F0) --------------------------------------------------------
#if 0 // Restored by gui_geometry_native_entry.go.
int nox_xxx_pointInRect_4281F0(int2* a1, int4* a2) {
	int v2;      // ecx
	bool result; // al

	result = 0;
	if (a1->field_0 >= a2->field_0 && a1->field_0 <= a2->field_8) {
		v2 = a1->field_4;
		if (v2 >= a2->field_4 && v2 <= a2->field_C) {
			result = 1;
		}
	}
	return result;
}
#endif

//----- (00428220) --------------------------------------------------------
#if 0 // Restored by gui_geometry_native_entry.go.
int sub_428220(float2* a1, float4* a2) {
	return a1->field_0 >= (double)a2->field_0 && a1->field_0 <= (double)a2->field_8 &&
		   a1->field_4 >= (double)a2->field_4 && a1->field_4 <= (double)a2->field_C;
}
#endif

//----- (004282D0) --------------------------------------------------------
char* sub_4282D0(char* a1, int a2) {
	char* result; // eax

	result = strncpy((char*)getMemAt(0x5D4594, 741316), a1, 0x10u);
	*getMemU32Ptr(0x5D4594, 741304) = a2;
	return result;
}

//----- (004282F0) --------------------------------------------------------
void* sub_4282F0(int a1, int a2, size_t a3) {
	int v3;         // ebx
	size_t v4;      // edi
	unsigned int i; // esi
	size_t v6;      // ebp
	void* result;   // eax
	size_t v8;      // ebp
	int v9;         // edx
	int v10;        // esi
	int v11;        // eax
	int j;          // [esp+14h] [ebp+4h]

	v3 = a1;
	v4 = 0;
	if (*(uint32_t*)(a1 + 608)) {
		for (i = 0; i < *(int*)&dword_5d4594_741332; ++i) {
			free(*(void**)(*(uint32_t*)(a1 + 608) + 4 * i));
			*(uint32_t*)(*(uint32_t*)(a1 + 608) + 4 * i) = 0;
		}
		free(*(void**)(a1 + 608));
		*(uint32_t*)(a1 + 608) = 0;
	}
	if (*(uint32_t*)(a1 + 612)) {
		free(*(void**)(a1 + 612));
		*(uint32_t*)(a1 + 612) = 0;
	}
	if (*(uint32_t*)(a1 + 616)) {
		free(*(void**)(a1 + 616));
		*(uint32_t*)(a1 + 616) = 0;
	}
	if (*(uint32_t*)(a1 + 620)) {
		free(*(void**)(a1 + 620));
		*(uint32_t*)(a1 + 620) = 0;
	}
	if (*(uint32_t*)(a1 + 624)) {
		free(*(void**)(a1 + 624));
		*(uint32_t*)(a1 + 624) = 0;
	}
	if (*(uint32_t*)(a1 + 628)) {
		free(*(void**)(a1 + 628));
		*(uint32_t*)(a1 + 628) = 0;
	}
	if (*(uint32_t*)(a1 + 632)) {
		free(*(void**)(a1 + 632));
		*(uint32_t*)(a1 + 632) = 0;
	}
	v6 = 4 * a3;
	*(uint32_t*)(a1 + 608) = calloc(a3, 4);
	if (a3) {
		do {
			*(uint32_t*)(*(uint32_t*)(a1 + 608) + 4 * v4++) = calloc(1, 0xAu);
		} while (v4 < a3);
	}
	*(uint32_t*)(a1 + 612) = calloc(1, v6);
	*(uint32_t*)(a1 + 616) = calloc(1, v6);
	*(uint32_t*)(a1 + 620) = calloc(1, a3);
	*(uint32_t*)(a1 + 624) = calloc(1, a3);
	*(uint32_t*)(a1 + 628) = calloc(1, v6);
	result = calloc(1, a3);
	v8 = 0;
	*(uint32_t*)(a1 + 632) = result;
	if (a3) {
		v9 = a2 + 16;
		for (j = a2 + 16;; v9 = j) {
			strcpy(*(char**)(*(uint32_t*)(v3 + 608) + 4 * v8), (const char*)(v9 - 16));
			*(uint32_t*)(*(uint32_t*)(v3 + 612) + 4 * v8) = *(uint32_t*)(v9 - 4);
			*(uint32_t*)(*(uint32_t*)(v3 + 616) + 4 * v8) = *(uint32_t*)v9;
			*(uint8_t*)(*(uint32_t*)(v3 + 620) + v8) = *(uint8_t*)(v9 + 4);
			*(uint8_t*)(*(uint32_t*)(v3 + 624) + v8) = *(uint8_t*)(v9 + 5);
			*(uint8_t*)(*(uint32_t*)(v3 + 632) + v8) = *(uint8_t*)(v9 + 12);
			v10 = *(uint32_t*)(v9 + 8);
			v11 = time(0) - v10;
			++v8;
			*(uint32_t*)(j + 8) = v11;
			*(uint32_t*)(*(uint32_t*)(v3 + 628) + 4 * v8 - 4) = v11;
			result = (void*)a3;
			j += 32;
			if (v8 >= a3) {
				break;
			}
		}
		*(uint16_t*)(v3 + 6) = a3;
		dword_5d4594_741332 = a3;
	} else {
		*(uint16_t*)(a1 + 6) = 0;
		dword_5d4594_741332 = 0;
	}
	return result;
}

//----- (00428540) --------------------------------------------------------
unsigned int sub_428540(int a1, char* a2, int a3) {
	unsigned int result; // eax
	char* v4;            // ecx
	char v5;             // bl

	if (*(uint32_t*)(a1 + 636)) {
		free(*(void**)(a1 + 636));
		*(uint32_t*)(a1 + 636) = 0;
	}
	*(uint32_t*)(a1 + 636) = calloc(a3, 2);
	result = 0;
	if (2 * a3) {
		v4 = a2;
		do {
			v5 = *v4;
			v4 += 2;
			*(uint8_t*)(result + *(uint32_t*)(a1 + 636)) = v5;
			*(uint8_t*)(result + *(uint32_t*)(a1 + 636) + 1) = *(v4 - 1);
			result += 2;
		} while (result < 2 * a3);
		*getMemU32Ptr(0x5D4594, 741308) = a3;
	} else {
		*getMemU32Ptr(0x5D4594, 741308) = a3;
	}
	return result;
}

//----- (004285C0) --------------------------------------------------------
char* sub_4285C0(short* a1) {
	char* result;    // eax
	short* v2;       // ebx
	void* v3;        // eax
	unsigned int v4; // esi
	bool v5;         // zf
	void* v6;        // eax
	void* v7;        // eax
	void* v8;        // eax
	void* v9;        // eax
	void* v10;       // eax
	void* v11;       // eax
	char* v12;       // ebp
	int v13;         // eax
	size_t v14;      // [esp-24h] [ebp-30h]
	size_t v15;      // [esp-20h] [ebp-2Ch]
	size_t v16;      // [esp-18h] [ebp-24h]
	size_t v17;      // [esp-14h] [ebp-20h]
	size_t v18;      // [esp-Ch] [ebp-18h]
	size_t v19;      // [esp-8h] [ebp-14h]
	int v20;         // [esp+10h] [ebp+4h]

	result = (char*)nox_xxx_player_4E3CE0();
	v2 = a1;
	*a1 = (short)result;
	if ((short)result <= 0) {
		dword_5d4594_741332 = *a1;
	} else {
		v3 = calloc((short)result, 4);
		v4 = 0;
		v5 = *a1 == 0;
		*((uint32_t*)a1 + 134) = v3;
		if (!v5) {
			do {
				*(uint32_t*)(*((uint32_t*)a1 + 134) + 4 * v4++) = calloc(1, 0xAu);
			} while (v4 < *a1);
		}
		v6 = calloc(*a1, 4);
		v19 = *a1;
		*((uint32_t*)a1 + 135) = v6;
		v7 = calloc(1, v19);
		v18 = 4 * *a1;
		*((uint32_t*)a1 + 144) = v7;
		*((uint32_t*)a1 + 136) = calloc(1, v18);
		v8 = calloc(*a1, 4);
		v17 = 4 * *a1;
		*((uint32_t*)a1 + 137) = v8;
		v9 = calloc(1, v17);
		v16 = 4 * *a1;
		*((uint32_t*)a1 + 138) = v9;
		*((uint32_t*)a1 + 139) = calloc(1, v16);
		v10 = calloc(*a1, 4);
		v15 = 4 * *a1;
		*((uint32_t*)a1 + 140) = v10;
		v11 = calloc(1, v15);
		v14 = 4 * *a1;
		*((uint32_t*)a1 + 141) = v11;
		*((uint32_t*)a1 + 142) = calloc(1, v14);
		*((uint32_t*)a1 + 143) = calloc(*a1, 4);
		v20 = 0;
		result = nox_common_playerInfoGetFirst_416EA0();
		v12 = result;
		if (result) {
			do {
				if (*((uint32_t*)v12 + 1198)) {
					strcpy(*(char**)(*((uint32_t*)v2 + 134) + 4 * v20), v12 + 2096);
					v13 = nox_xxx_net_getIP_554200((unsigned char)v12[2064] + 1);
					*(uint32_t*)(*((uint32_t*)v2 + 135) + 4 * v20) = htonl(v13);
					*(uint8_t*)(v20 + *((uint32_t*)v2 + 144)) = v12[2251];
					*(uint32_t*)(*((uint32_t*)v2 + 136) + 4 * v20) = *((uint32_t*)v12 + 1172);
					*(uint32_t*)(*((uint32_t*)v2 + 137) + 4 * v20) = *((uint32_t*)v12 + 1172);
					*(uint32_t*)(*((uint32_t*)v2 + 138) + 4 * v20) = *((uint32_t*)v12 + 1166);
					*(uint32_t*)(*((uint32_t*)v2 + 139) + 4 * v20) = *((uint32_t*)v12 + 1165);
					*(uint32_t*)(*((uint32_t*)v2 + 140) + 4 * v20) = *((uint32_t*)v12 + 1167);
					*(uint32_t*)(*((uint32_t*)v2 + 141) + 4 * v20) = *((uint32_t*)v12 + 1168);
					*(uint32_t*)(*((uint32_t*)v2 + 142) + 4 * v20) = *((uint32_t*)v12 + 1171);
					*(uint32_t*)(*((uint32_t*)v2 + 143) + 4 * v20++) = sub_4D6540((unsigned char)v12[2064]);
				}
				v12 = nox_common_playerInfoGetNext_416EE0((int)v12);
			} while (v12);
			result = (char*)*v2;
			dword_5d4594_741332 = *v2;
		} else {
			dword_5d4594_741332 = *v2;
		}
	}
	return result;
}

//----- (00428810) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_428810
int sub_428810(int a1, int a2) {
	uint16_t* v2; // esi
	uint16_t* v3; // edi
	int v4;       // esi
	char v6[72];  // [esp+8h] [ebp-48h]

	v2 = sub_42ADA0(a1, a2, *getMemI16Ptr(0x5D4594, 741308), getMemUintPtr(0x5D4594, 741312));
	v3 = sub_42A8B0(v2, getMemIntPtr(0x5D4594, 741312));
	free(v2);
	v4 = 1;
	if (sub_420360(v6, &a2)) {
		abort();
	}
	free(v3);
	return v4;
}
#endif


//----- (00428890) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_428890
int sub_428890(short* a1) {
	uint16_t* v1; // esi
	uint16_t* v2; // edi
	int v3;       // esi
	char v5[72];  // [esp+8h] [ebp-48h]

	v1 = sub_42B810(a1, getMemUintPtr(0x5D4594, 741300));
	v2 = sub_42A8B0(v1, getMemIntPtr(0x5D4594, 741300));
	free(v1);
	v3 = 1;
	if (sub_420360(v5, &a1)) {
		abort();
	}
	free(v2);
	return v3;
}
#endif


//----- (004289D0) --------------------------------------------------------
void* sub_4289D0(void** a1) {
	unsigned int i; // edi
	void* result;   // eax

	if (a1[134]) {
		for (i = 0; i < *(int*)&dword_5d4594_741332; ++i) {
			free(*((void**)a1[134] + i));
			*((uint32_t*)a1[134] + i) = 0;
		}
		free(a1[134]);
		a1[134] = 0;
	}
	if (a1[135]) {
		free(a1[135]);
		a1[135] = 0;
	}
	if (a1[144]) {
		free(a1[144]);
		a1[144] = 0;
	}
	if (a1[136]) {
		free(a1[136]);
		a1[136] = 0;
	}
	if (a1[137]) {
		free(a1[137]);
		a1[137] = 0;
	}
	if (a1[138]) {
		free(a1[138]);
		a1[138] = 0;
	}
	if (a1[139]) {
		free(a1[139]);
		a1[139] = 0;
	}
	if (a1[140]) {
		free(a1[140]);
		a1[140] = 0;
	}
	if (a1[141]) {
		free(a1[141]);
		a1[141] = 0;
	}
	if (a1[142]) {
		free(a1[142]);
		a1[142] = 0;
	}
	result = a1[143];
	if (result) {
		free(a1[143]);
		a1[143] = 0;
	}
	return result;
}

// 428B85: variable 'v1' is possibly undefined
// 428BAB: variable 'v4' is possibly undefined
// 428B30: using guessed type char var_100[256];

//----- (0042A0F0) --------------------------------------------------------
// Ported to Go in wall.go so the foreach callback receives a native-width
// wall pointer instead of truncating it through the original PE32 int ABI.

//----- (0042A8B0) --------------------------------------------------------
uint16_t* sub_42A8B0(uint8_t* a1, int* a2) {
	// This packet has one CNTL field. Keep the original node lifecycle and
	// serializer here: their int-based entry points cannot carry LP64 pointers.
	typedef struct packet_field {
		char name[4];
		uint16_t type, length;
		void* data;
		struct packet_field* next;
	} packet_field;
	struct {
		uint16_t length, type;
		packet_field* first;
	} packet = {0};
	uint32_t* length = (uint32_t*)a2;
	uint32_t initial_length = *length;
	uint8_t* compressed = calloc(1, (uint32_t)(initial_length * 2u));
	sub_42A970(a1, compressed, a2);
	uint8_t* encoded = sub_42AC50(compressed, (size_t*)a2);
	if (compressed) {
		free(compressed);
	}
	if (!encoded) {
		return 0; // The original empty-list destructor has no service calls.
	}

	uint16_t field_length = (uint16_t)*length;
	packet_field* field = calloc(1, sizeof(*field));
	if (field) {
		field->data = 0;
		free(field->data);
		strcpy(field->name, (const char*)getMemAt(0x5D4594, 741688));
		field->type = field->length = 0;
		field->data = 0;
		field->next = 0;
		const char* name = (const char*)getMemAt(0x587000, 71480);
		// Keep all four wire bytes and strncpy's padding for a short tag.
		for (size_t i = 0; i < sizeof(field->name); ++i) {
			field->name[i] = *name;
			if (*name) ++name;
		}
		field->type = 20;
		field->length = field_length;
		field->data = calloc(1, field_length);
		memcpy(field->data, encoded, field->length);
		field->next = 0;
	}
	// Preserve the original unguarded link after a failed node allocation.
	field->next = packet.first;
	packet.first = field;
	free(encoded);

	*length = 4;
	for (field = packet.first; field; field = field->next) {
		*length += 8;
		*length = field->length + *length + ((0u - ((field->length + *length) & 3u)) & 3u);
	}
	uint16_t* result = calloc(1, *length);
	result[0] = htons((uint16_t)*length);
	result[1] = htons(packet.type);
	uint8_t* output = (uint8_t*)(result + 2);
	for (field = packet.first; field; field = field->next) {
		field->type = htons(field->type);
		field->length = htons(field->length);
		memcpy(output, field->name, 8);
		output += 8;
		memcpy(output, field->data, ntohs(field->length));
		output += ntohs(field->length) + ((0u - (ntohs(field->length) & 3u)) & 3u);
		field->type = ntohs(field->type);
		field->length = ntohs(field->length);
	}
	for (field = packet.first; field;) {
		packet_field* next = field->next;
		free(field->data);
		strcpy(field->name, (const char*)getMemAt(0x5D4594, 741688));
		field->type = field->length = 0;
		field->data = 0;
		field->next = 0;
		free(field);
		field = next;
	}
	return result;
}

//----- (0042A970) --------------------------------------------------------
int sub_42A970(uint8_t* a1, uint8_t* a2, int* a3) {
	int v3;            // eax
	char v4;           // bl
	int v5;            // edx
	int v6;            // ecx
	int v7;            // esi
	int v8;            // eax
	int* v9;           // edx
	uint8_t* v10;      // ebp
	int v11;           // edi
	char v12;          // cl
	int result;        // eax
	int v14;           // edx
	int v15;           // eax
	int v16;           // ebp
	char v17;          // [esp+13h] [ebp-405h]
	unsigned char v18; // [esp+14h] [ebp-404h]
	int v19[256];      // [esp+18h] [ebp-400h]

	v3 = 0;
	v4 = 1;
	memset(v19, 0, sizeof(v19));
	v18 = 1;
	v5 = *a3;
	if (*a3 > 0) {
		do {
			v6 = (unsigned char)a1[v3++];
			++v19[v6];
		} while (v3 < v5);
	}
	v7 = v19[0];
	v17 = 0;
	v8 = 0;
	v9 = v19;
	while (*v9) {
		if (*v9 < v7) {
			v7 = *v9;
			v17 = v8;
		}
		++v8;
		++v9;
		if (v8 >= 256) {
			goto LABEL_10;
		}
	}
	v17 = v8;
LABEL_10:
	v10 = a1;
	v11 = 1;
	*a2 = v17;
	v12 = *a1;
	result = 1;
	v14 = *a3;
	if ((int)*a3 < 1) {
		*a3 = 1;
	} else {
		do {
			if (v10[v11] != (uint8_t)v12 || v11 >= v14) {
				if ((unsigned char)v4 <= 3u) {
					if ((int)v18 > 0) {
						v16 = v18;
						do {
							if (v12 == v17) {
								a2[result++] = v12;
							}
							a2[result++] = v12;
							--v16;
						} while (v16);
					}
				} else {
					a2[result] = v17;
					v15 = result + 1;
					a2[v15++] = v4;
					a2[v15] = v12;
					result = v15 + 1;
				}
				v10 = a1;
				v4 = 1;
				v12 = a1[v11];
			} else {
				++v4;
			}
			++v11;
			v18 = v4;
			v14 = *a3;
		} while (v11 <= *a3);
		*a3 = result;
	}
	return result;
}
// 42A970: using guessed type int var_400[256];

//----- (0042AAA0) --------------------------------------------------------
// GAME.EXE uses precision53/toward-zero. Reproduce this multiply locally;
// no global FPU state or the independent 4096-word gameplay RNG is changed.
static double nox_stats_mul_chop53_42AAA0(double a, double b) {
 double product = a * b;
 double error = fma(a, b, -product);
 if ((product > 0.0 && error < 0.0) || (product < 0.0 && error > 0.0)) {
  return nextafter(product, 0.0);
 }
 return product;
}

double sub_42AAA0(int* a1) {
	int v1;            // eax
	int v2;            // esi
	int v3;            // edi
	int v4;            // ecx
	int v5;            // edi
	int v6;            // esi
	unsigned char* v7; // ecx
	int v8;            // eax
	int v9;            // edx
	int v10;           // ecx
	int v11;           // eax
	int v13;           // [esp+Ch] [ebp+4h]

	if ((int)*a1 < 0 || !*getMemU32Ptr(0x5D4594, 741656)) {
		*getMemU32Ptr(0x5D4594, 741656) = 1;
		v1 = *a1;
		if ((int)*a1 < 0) {
			v1 = (int32_t)(0u - (uint32_t)v1);
		}
		v2 = 1;
		v3 = 21;
		v4 = (int32_t)(UINT32_C(161803398) - (uint32_t)v1) % INT32_C(1000000000);
		*getMemU32Ptr(0x5D4594, 741604) = (int32_t)(UINT32_C(161803398) - (uint32_t)v1) % INT32_C(1000000000);
		do {
			*getMemU32Ptr(0x5D4594, 741384 + 4 * (v3 % 55)) = v2;
			v2 = v4 - v2 + (v4 - v2 < 0 ? 0x3B9ACA00 : 0);
			v4 = *getMemU32Ptr(0x5D4594, 741384 + 4 * (v3 % 55));
			v3 += 21;
		} while (v3 <= 1134);
		v5 = 4;
		do {
			v6 = 1;
			v7 = getMemAt(0x5D4594, 741388);
			do {
				v8 = *(uint32_t*)v7 - *getMemU32Ptr(0x5D4594, 741388 + 4 * ((v6 + 30) % 55));
				*(uint32_t*)v7 = v8;
				if (v8 < 0) {
					*(uint32_t*)v7 = v8 + 1000000000;
				}
				v7 += 4;
				++v6;
			} while (v6 <= 55);
			--v5;
		} while (v5);
		dword_5d4594_741648 = 0;
		dword_5d4594_741652 = 31;
		*a1 = 1;
	}
	v9 = ++dword_5d4594_741648;
	if (dword_5d4594_741648 == 56) {
		v9 = 1;
		dword_5d4594_741648 = 1;
	}
	v10 = ++dword_5d4594_741652;
	if (dword_5d4594_741652 == 56) {
		v10 = 1;
		dword_5d4594_741652 = 1;
	}
	v11 = *getMemU32Ptr(0x5D4594, 741384 + 4 * v9) - *getMemU32Ptr(0x5D4594, 741384 + 4 * v10);
	v13 = v11;
	if (v11 < 0) {
		v11 += 1000000000;
		v13 = v11;
	}
	*getMemU32Ptr(0x5D4594, 741384 + 4 * v9) = v11;
	return nox_stats_mul_chop53_42AAA0((double)v13, *getMemDoublePtr(0x581450, 8368));
}

//----- (0042ABF0) --------------------------------------------------------
void sub_42ABF0(int a1, int a2, int a3) {
	double v3; // st7
	int i;     // esi

	if (a3 > 0) {
		a3 = -a3;
	}
	v3 = sub_42AAA0(&a3);
	for (i = 0; i < a2; ++i) {
		int v5 = (int)(v3 * 255.0);
		if (v5 < 0) {
			v5 = -v5;
		}
		*(uint8_t*)(i + a1) = v5;
		v3 = sub_42AAA0(&a3);
	}
}

//----- (0042AC50) --------------------------------------------------------
uint8_t* sub_42AC50(uint8_t* a1, size_t* a2) {
	// The legacy signature carries a DWORD length, not a native size_t field.
	uint32_t* length = (uint32_t*)a2;
	uint8_t* mask = calloc(1, *length);
	uint8_t* result = calloc(1, (uint32_t)(*length + 5u));
	if ((int32_t)*length < 15) {
		*length = UINT32_C(0xfffffffe);
		free(mask);
		free(result);
		return 0;
	}

	int32_t maskSeed = (int32_t)(uint32_t)time(0);
	if (maskSeed > 0) {
		maskSeed = (int32_t)(0u - (uint32_t)maskSeed);
	}
	uint32_t key = htonl((uint32_t)maskSeed);
	uint32_t maskCount = *length;
	// Inline 0042ABF0's loop: its int buffer argument truncates native pointers.
	// Preserve the initial RNG draw and the draw after the final mask byte.
	double random = sub_42AAA0(&maskSeed);
	for (uint32_t i = 0; (int32_t)i < (int32_t)maskCount; ++i) {
		int32_t value = (int32_t)(uint32_t)(int64_t)(random * 255.0);
		if (value < 0) {
			value = (int32_t)(0u - (uint32_t)value);
		}
		mask[i] = (uint8_t)value;
		random = sub_42AAA0(&maskSeed);
	}

	int32_t selectorSeed = (int32_t)(uint32_t)time(0);
	if (selectorSeed > 0) {
		selectorSeed = (int32_t)(0u - (uint32_t)selectorSeed);
	}
	float factor = (int32_t)*length >= 241 ? 241.0f : (float)(int32_t)(*length - 14u);
	uint8_t selector = (uint8_t)((uint64_t)(int64_t)(sub_42AAA0(&selectorSeed) * factor) + 10u);
	uint32_t pos = 0;
	for (uint32_t i = 0; (int32_t)i < (int32_t)*length; ++i) {
		if (pos == 5) {
			result[pos++] = selector;
		}
		if (pos == selector) {
			memcpy(result + pos, &key, sizeof(key));
			pos += 4;
		}
		result[pos++] = a1[i] ^ mask[i];
	}
	*length += 5u;
	free(mask);
	return result;
}

//----- (0042ADA0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42ADA0
uint16_t* sub_42ADA0(int a1, int a2, short a3, unsigned int* a4) {
	int v4;           // esi
	int v5;           // edi
	void* v6;         // eax
	int v7;           // eax
	int v8;           // edi
	void* v9;         // eax
	int v10;          // eax
	int v11;          // edi
	void* v12;        // eax
	int v13;          // eax
	int v14;          // edi
	void* v15;        // eax
	int v16;          // eax
	void* v17;        // eax
	int v18;          // eax
	int v19;          // edi
	void* v20;        // eax
	int v21;          // eax
	int v22;          // edi
	void* v23;        // eax
	int v24;          // eax
	void* v25;        // eax
	int v26;          // eax
	int v27;          // edi
	void* v28;        // eax
	int v29;          // eax
	int v30;          // edi
	void* v31;        // eax
	int v32;          // eax
	int v33;          // edi
	void* v34;        // eax
	int v35;          // eax
	int v36;          // edi
	void* v37;        // eax
	int v38;          // eax
	void* v39;        // eax
	int v40;          // eax
	void* v41;        // eax
	int v42;          // eax
	void* v43;        // eax
	int v44;          // eax
	void* v45;        // eax
	int v46;          // eax
	void* v47;        // eax
	int v48;          // eax
	int v49;          // edi
	void* v50;        // eax
	int v51;          // eax
	int v52;          // edi
	void* v53;        // eax
	int v54;          // eax
	int v55;          // edi
	void* v56;        // eax
	int v57;          // eax
	int v58;          // edi
	void* v59;        // eax
	int v60;          // eax
	void* v61;        // eax
	int v62;          // eax
	void* v63;        // eax
	int v64;          // eax
	void* v65;        // eax
	int v66;          // eax
	int v67;          // edi
	void* v68;        // eax
	int v69;          // eax
	int v70;          // edi
	void* v71;        // eax
	int v72;          // eax
	void* v73;        // eax
	int v74;          // eax
	void* v75;        // eax
	int v76;          // eax
	void* v77;        // eax
	int v78;          // eax
	void* v79;        // eax
	int v80;          // eax
	void* v81;        // eax
	int v82;          // eax
	int v83;          // eax
	char v84;         // di
	void* v85;        // eax
	int v86;          // eax
	void* v87;        // eax
	int v88;          // eax
	short v89;        // di
	void* v90;        // eax
	int v91;          // eax
	int v92;          // eax
	const char* v93;  // edi
	void* v94;        // eax
	int v95;          // eax
	int v96;          // edi
	void* v97;        // eax
	int v98;          // eax
	int v99;          // edi
	void* v100;       // eax
	int v101;         // eax
	void* v102;       // eax
	int v103;         // eax
	void* v104;       // eax
	int v105;         // eax
	int v106;         // edi
	void* v107;       // eax
	int v108;         // eax
	void* v109;       // eax
	int v110;         // eax
	const void* v111; // esi
	void* v112;       // eax
	int v113;         // eax
	int v114;         // eax
	uint16_t* v115;   // esi
	char v117;        // [esp+Ch] [ebp-10h]
	char v118[12];    // [esp+10h] [ebp-Ch]
	char v119;        // [esp+20h] [ebp+4h]
	char v120;        // [esp+20h] [ebp+4h]
	char v121;        // [esp+20h] [ebp+4h]
	char v122;        // [esp+20h] [ebp+4h]
	char v123;        // [esp+20h] [ebp+4h]
	char v124;        // [esp+20h] [ebp+4h]
	char v125;        // [esp+20h] [ebp+4h]
	char v126;        // [esp+20h] [ebp+4h]
	char v127;        // [esp+20h] [ebp+4h]
	char v128;        // [esp+20h] [ebp+4h]
	char v129;        // [esp+24h] [ebp+8h]

	v4 = a1;
	v5 = *(uint32_t*)(a1 + 8);
	*(uint16_t*)v118 = 0;
	*(uint16_t*)&v118[2] = 0;
	*(uint32_t*)&v118[4] = 0;
	v6 = calloc(1, 0x10u);
	if (v6) {
		v7 = sub_42C8B0((int)v6, "MXPL", v5);
	} else {
		v7 = 0;
	}
	sub_42C360(v118, v7);
	v8 = *(uint32_t*)(a1 + 12);
	v9 = calloc(1, 0x10u);
	if (v9) {
		v10 = sub_42C8B0((int)v9, "IDNO", v8);
	} else {
		v10 = 0;
	}
	sub_42C360(v118, v10);
	v11 = *(uint32_t*)(a1 + 16);
	v12 = calloc(1, 0x10u);
	if (v12) {
		v13 = sub_42C8B0((int)v12, "GSKU", v11);
	} else {
		v13 = 0;
	}
	sub_42C360(v118, v13);
	v14 = *(uint32_t*)(a1 + 20);
	v15 = calloc(1, 0x10u);
	if (v15) {
		v16 = sub_42C8B0((int)v15, "GSTY", v14);
	} else {
		v16 = 0;
	}
	sub_42C360(v118, v16);
	v119 = *(uint8_t*)(a1 + 24);
	v17 = calloc(1, 0x10u);
	if (v17) {
		v18 = sub_42C7F0((int)v17, "CLGM", v119);
	} else {
		v18 = 0;
	}
	sub_42C360(v118, v18);
	v19 = *(uint32_t*)(v4 + 32);
	v20 = calloc(1, 0x10u);
	if (v20) {
		v21 = sub_42C8B0((int)v20, "LIMT", v19);
	} else {
		v21 = 0;
	}
	sub_42C360(v118, v21);
	v22 = *(uint32_t*)(v4 + 36);
	v23 = calloc(1, 0x10u);
	if (v23) {
		v24 = sub_42C8B0((int)v23, "TLMT", v22);
	} else {
		v24 = 0;
	}
	sub_42C360(v118, v24);
	v120 = *(uint8_t*)(v4 + 40);
	v25 = calloc(1, 0x10u);
	if (v25) {
		v26 = sub_42C7F0((int)v25, "RSTC", v120);
	} else {
		v26 = 0;
	}
	sub_42C360(v118, v26);
	v27 = *(uint32_t*)(v4 + 44);
	v28 = calloc(1, 0x10u);
	if (v28) {
		v29 = sub_42C8B0((int)v28, "MINE", v27);
	} else {
		v29 = 0;
	}
	sub_42C360(v118, v29);
	v30 = *(uint32_t*)(v4 + 48);
	v31 = calloc(1, 0x10u);
	if (v31) {
		v32 = sub_42C8B0((int)v31, "MAXE", v30);
	} else {
		v32 = 0;
	}
	sub_42C360(v118, v32);
	v33 = *(uint32_t*)(v4 + 52);
	v34 = calloc(1, 0x10u);
	if (v34) {
		v35 = sub_42C8B0((int)v34, "MINP", v33);
	} else {
		v35 = 0;
	}
	sub_42C360(v118, v35);
	v36 = *(uint32_t*)(v4 + 56);
	v37 = calloc(1, 0x10u);
	if (v37) {
		v38 = sub_42C8B0((int)v37, "MAXP", v36);
	} else {
		v38 = 0;
	}
	sub_42C360(v118, v38);
	v121 = *(uint8_t*)(v4 + 60);
	v39 = calloc(1, 0x10u);
	if (v39) {
		v40 = sub_42C7F0((int)v39, "VIDM", v121);
	} else {
		v40 = 0;
	}
	sub_42C360(v118, v40);
	v122 = *(uint8_t*)(v4 + 61);
	v41 = calloc(1, 0x10u);
	if (v41) {
		v42 = sub_42C7F0((int)v41, "SVRS", v122);
	} else {
		v42 = 0;
	}
	sub_42C360(v118, v42);
	v123 = *(uint8_t*)(v4 + 25);
	v43 = calloc(1, 0x10u);
	if (v43) {
		v44 = sub_42C7F0((int)v43, "NTMS", v123);
	} else {
		v44 = 0;
	}
	sub_42C360(v118, v44);
	v45 = calloc(1, 0x10u);
	if (v45) {
		v46 = sub_42C8E0((int)v45, "SCEN", (const char*)(v4 + 96));
	} else {
		v46 = 0;
	}
	sub_42C360(v118, v46);
	v47 = calloc(1, 0x10u);
	if (v47) {
		v48 = sub_42C8E0((int)v47, "GNAM", (const char*)(v4 + 352));
	} else {
		v48 = 0;
	}
	sub_42C360(v118, v48);
	v49 = *(uint32_t*)(v4 + 64);
	v50 = calloc(1, 0x10u);
	if (v50) {
		v51 = sub_42C8B0((int)v50, "SPL1", v49);
	} else {
		v51 = 0;
	}
	sub_42C360(v118, v51);
	v52 = *(uint32_t*)(v4 + 68);
	v53 = calloc(1, 0x10u);
	if (v53) {
		v54 = sub_42C8B0((int)v53, "SPL2", v52);
	} else {
		v54 = 0;
	}
	sub_42C360(v118, v54);
	v55 = *(uint32_t*)(v4 + 72);
	v56 = calloc(1, 0x10u);
	if (v56) {
		v57 = sub_42C8B0((int)v56, "SPL3", v55);
	} else {
		v57 = 0;
	}
	sub_42C360(v118, v57);
	v58 = *(uint32_t*)(v4 + 88);
	v59 = calloc(1, 0x10u);
	if (v59) {
		v60 = sub_42C8B0((int)v59, "ARMR", v58);
	} else {
		v60 = 0;
	}
	sub_42C360(v118, v60);
	v124 = *(uint8_t*)(v4 + 84);
	v61 = calloc(1, 0x10u);
	if (v61) {
		v62 = sub_42C7F0((int)v61, "WPN1", v124);
	} else {
		v62 = 0;
	}
	sub_42C360(v118, v62);
	v125 = *(uint8_t*)(v4 + 85);
	v63 = calloc(1, 0x10u);
	if (v63) {
		v64 = sub_42C7F0((int)v63, "WPN2", v125);
	} else {
		v64 = 0;
	}
	sub_42C360(v118, v64);
	v126 = *(uint8_t*)(v4 + 86);
	v65 = calloc(1, 0x10u);
	if (v65) {
		v66 = sub_42C7F0((int)v65, "WPN3", v126);
	} else {
		v66 = 0;
	}
	sub_42C360(v118, v66);
	v67 = *(uint32_t*)(v4 + 92);
	v68 = calloc(1, 0x10u);
	if (v68) {
		v69 = sub_42C8B0((int)v68, "STAF", v67);
	} else {
		v69 = 0;
	}
	sub_42C360(v118, v69);
	v70 = *(uint32_t*)(v4 + 28);
	v71 = calloc(1, 0x10u);
	if (v71) {
		v72 = sub_42C8B0((int)v71, "DURA", v70);
	} else {
		v72 = 0;
	}
	sub_42C360(v118, v72);
	v73 = calloc(1, 0x10u);
	if (v73) {
		v74 = sub_42C7F0((int)v73, "FINI", 1);
	} else {
		v74 = 0;
	}
	sub_42C360(v118, v74);
	v127 = *(uint8_t*)(v4 + 26);
	v75 = calloc(1, 0x10u);
	if (v75) {
		v76 = sub_42C7F0((int)v75, "TRNY", v127);
	} else {
		v76 = 0;
	}
	sub_42C360(v118, v76);
	switch (a2) {
	case 0:
		*getMemU32Ptr(0x5D4594, 741668) = 0;
		v77 = calloc(1, 0x10u);
		if (v77) {
			v78 = sub_42C8B0((int)v77, "SEQU", 0);
		} else {
			v78 = 0;
		}
		sub_42C360(v118, v78);
		v79 = calloc(1, 0x10u);
		if (v79) {
			v80 = sub_42C7F0((int)v79, "ENDF", 0);
		} else {
			v80 = 0;
		}
		sub_42C360(v118, v80);
		v81 = calloc(1, 0x10u);
		if (v81) {
			v82 = sub_42C820((int)v81, (char*)getMemAt(0x587000, 71872), -1);
			sub_42C360(v118, v82);
			break;
		}
		sub_42C360(v118, 0);
		break;
	case 2:
		v83 = *getMemU32Ptr(0x5D4594, 741668) + 1;
		*getMemU32Ptr(0x5D4594, 741668) = v83;
		v84 = v83;
		v85 = calloc(1, 0x10u);
		if (v85) {
			v86 = sub_42C8B0((int)v85, "SEQU", v84);
		} else {
			v86 = 0;
		}
		sub_42C360(v118, v86);
		v87 = calloc(1, 0x10u);
		if (v87) {
			v88 = sub_42C7F0((int)v87, "ENDF", 1);
		} else {
			v88 = 0;
		}
		sub_42C360(v118, v88);
		v89 = *(uint16_t*)(v4 + 6);
		v90 = calloc(1, 0x10u);
		if (v90) {
			v91 = sub_42C820((int)v90, (char*)getMemAt(0x587000, 71896), v89);
		} else {
			v91 = 0;
		}
		sub_42C360(v118, v91);
		v92 = 0;
		for (*getMemU32Ptr(0x5D4594, 741660) = 0; v92 < *(short*)(v4 + 6); *getMemU32Ptr(0x5D4594, 741660) = v92) {
			*getMemU8Ptr(0x587000, 71491) = v92 + 48;
			v93 = *(const char**)(*(uint32_t*)(v4 + 608) + 4 * v92);
			v94 = calloc(1, 0x10u);
			if (v94) {
				v95 = sub_42C8E0((int)v94, "LGL?", v93);
			} else {
				v95 = 0;
			}
			sub_42C360(v118, v95);
			*getMemU8Ptr(0x587000, 71499) = getMemByte(0x5D4594, 741660) + 48;
			v96 = *(uint32_t*)(*(uint32_t*)(v4 + 612) + 4 * *getMemU32Ptr(0x5D4594, 741660));
			v97 = calloc(1, 0x10u);
			if (v97) {
				v98 = sub_42C8B0((int)v97, "IPL?", v96);
			} else {
				v98 = 0;
			}
			sub_42C360(v118, v98);
			*getMemU8Ptr(0x587000, 71515) = getMemByte(0x5D4594, 741660) + 48;
			v99 = *(uint32_t*)(*(uint32_t*)(v4 + 616) + 4 * *getMemU32Ptr(0x5D4594, 741660));
			v100 = calloc(1, 0x10u);
			if (v100) {
				v101 = sub_42C8B0((int)v100, "CNL?", v99);
			} else {
				v101 = 0;
			}
			sub_42C360(v118, v101);
			*getMemU8Ptr(0x587000, 71507) = getMemByte(0x5D4594, 741660) + 48;
			v128 = *(uint8_t*)(*getMemU32Ptr(0x5D4594, 741660) + *(uint32_t*)(v4 + 620));
			v102 = calloc(1, 0x10u);
			if (v102) {
				v103 = sub_42C7F0((int)v102, "CLL?", v128);
			} else {
				v103 = 0;
			}
			sub_42C360(v118, v103);
			*getMemU8Ptr(0x587000, 71523) = getMemByte(0x5D4594, 741660) + 48;
			v129 = *(uint8_t*)(*(uint32_t*)(v4 + 624) + *getMemU32Ptr(0x5D4594, 741660));
			v104 = calloc(1, 0x10u);
			if (v104) {
				v105 = sub_42C7F0((int)v104, "CMP?", v129);
			} else {
				v105 = 0;
			}
			sub_42C360(v118, v105);
			*getMemU8Ptr(0x587000, 71531) = getMemByte(0x5D4594, 741660) + 48;
			v106 = *(uint32_t*)(*(uint32_t*)(v4 + 628) + 4 * *getMemU32Ptr(0x5D4594, 741660));
			v107 = calloc(1, 0x10u);
			if (v107) {
				v108 = sub_42C8B0((int)v107, "DUR?", v106);
			} else {
				v108 = 0;
			}
			sub_42C360(v118, v108);
			*getMemU8Ptr(0x587000, 71539) = getMemByte(0x5D4594, 741660) + 48;
			v117 = *(uint8_t*)(*(uint32_t*)(v4 + 632) + *getMemU32Ptr(0x5D4594, 741660));
			v109 = calloc(1, 0x10u);
			if (v109) {
				v110 = sub_42C7F0((int)v109, "PAR?", v117);
			} else {
				v110 = 0;
			}
			sub_42C360(v118, v110);
			v92 = *getMemU32Ptr(0x5D4594, 741660) + 1;
		}
		v111 = *(const void**)(v4 + 636);
		v112 = calloc(1, 0x10u);
		if (v112) {
			v113 = sub_42C910((int)v112, (char*)getMemAt(0x587000, 71904), v111, 2 * a3);
			sub_42C360(v118, v113);
			break;
		}
		sub_42C360(v118, 0);
		break;
	case 1:
		++*getMemU32Ptr(0x5D4594, 741668);
		sub_42BDC0(v118, "SEQU", getMemByte(0x5D4594, 741668));
		sub_42BCE0(v118, "ENDF", 0);
		sub_42BD50(v118, (char*)getMemAt(0x587000, 71928), *(uint16_t*)(v4 + 6));
		v114 = 0;
		for (*getMemU32Ptr(0x5D4594, 741660) = 0; v114 < *(short*)(v4 + 6); *getMemU32Ptr(0x5D4594, 741660) = v114) {
			*getMemU8Ptr(0x587000, 71491) = v114 + 48;
			sub_42BE30(v118, (char*)getMemAt(0x587000, 71488), *(const char**)(*(uint32_t*)(v4 + 608) + 4 * v114));
			*getMemU8Ptr(0x587000, 71499) = getMemByte(0x5D4594, 741660) + 48;
			sub_42BDC0(v118, "IPL?", *(uint32_t*)(*(uint32_t*)(v4 + 612) + 4 * *getMemU32Ptr(0x5D4594, 741660)));
			*getMemU8Ptr(0x587000, 71515) = getMemByte(0x5D4594, 741660) + 48;
			sub_42BDC0(v118, "CNL?", *(uint32_t*)(*(uint32_t*)(v4 + 616) + 4 * *getMemU32Ptr(0x5D4594, 741660)));
			*getMemU8Ptr(0x587000, 71507) = getMemByte(0x5D4594, 741660) + 48;
			sub_42BCE0(v118, "CLL?", *(uint8_t*)(*getMemU32Ptr(0x5D4594, 741660) + *(uint32_t*)(v4 + 620)));
			*getMemU8Ptr(0x587000, 71523) = getMemByte(0x5D4594, 741660) + 48;
			sub_42BCE0(v118, "CMP?", *(uint8_t*)(*(uint32_t*)(v4 + 624) + *getMemU32Ptr(0x5D4594, 741660)));
			*getMemU8Ptr(0x587000, 71531) = getMemByte(0x5D4594, 741660) + 48;
			sub_42BDC0(v118, "DUR?", *(uint32_t*)(*(uint32_t*)(v4 + 628) + 4 * *getMemU32Ptr(0x5D4594, 741660)));
			*getMemU8Ptr(0x587000, 71539) = getMemByte(0x5D4594, 741660) + 48;
			sub_42BCE0(v118, "PAR?", *(uint8_t*)(*(uint32_t*)(v4 + 632) + *getMemU32Ptr(0x5D4594, 741660)));
			v114 = *getMemU32Ptr(0x5D4594, 741660) + 1;
		}
		sub_42BEA0(v118, (char*)getMemAt(0x587000, 71936), *(const void**)(v4 + 636), 2 * a3);
		break;
	}
	v115 = sub_42C480(v118, a4);
	sub_42C330(v118);
	return v115;
}
#endif


//----- (0042B810) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42B810
uint16_t* sub_42B810(short* a1, unsigned int* a2) {
	short* v2;       // esi
	int v3;          // edi
	void* v4;        // eax
	int v5;          // eax
	int v6;          // edi
	void* v7;        // eax
	int v8;          // eax
	int v9;          // edi
	void* v10;       // eax
	int v11;         // eax
	void* v12;       // eax
	int v13;         // eax
	void* v14;       // eax
	int v15;         // eax
	int v16;         // edi
	void* v17;       // eax
	int v18;         // eax
	void* v19;       // eax
	int v20;         // eax
	short v21;       // di
	void* v22;       // eax
	int v23;         // eax
	char v24;        // di
	void* v25;       // eax
	int v26;         // eax
	int v27;         // eax
	const char* v28; // edi
	void* v29;       // eax
	int v30;         // eax
	int v31;         // edi
	void* v32;       // eax
	int v33;         // eax
	void* v34;       // eax
	int v35;         // eax
	int v36;         // edi
	void* v37;       // eax
	int v38;         // eax
	int v39;         // edi
	void* v40;       // eax
	int v41;         // eax
	int v42;         // edi
	void* v43;       // eax
	int v44;         // eax
	int v45;         // edi
	void* v46;       // eax
	int v47;         // eax
	int v48;         // edi
	void* v49;       // eax
	int v50;         // eax
	int v51;         // edi
	void* v52;       // eax
	int v53;         // eax
	int v54;         // edi
	void* v55;       // eax
	int v56;         // eax
	uint16_t* v57;   // esi
	char v59[12];    // [esp+8h] [ebp-Ch]
	char v60;        // [esp+18h] [ebp+4h]
	char v61;        // [esp+18h] [ebp+4h]

	v2 = a1;
	v3 = *((uint32_t*)a1 + 1);
	*(uint16_t*)v59 = 0;
	*(uint16_t*)&v59[2] = 0;
	*(uint32_t*)&v59[4] = 0;
	v4 = calloc(1, 0x10u);
	if (v4) {
		v5 = sub_42C8B0((int)v4, "IDNO", v3);
	} else {
		v5 = 0;
	}
	sub_42C360(v59, v5);
	v6 = *((uint32_t*)a1 + 2);
	v7 = calloc(1, 0x10u);
	if (v7) {
		v8 = sub_42C8B0((int)v7, "GSKU", v6);
	} else {
		v8 = 0;
	}
	sub_42C360(v59, v8);
	v9 = *((uint32_t*)a1 + 3);
	v10 = calloc(1, 0x10u);
	if (v10) {
		v11 = sub_42C8B0((int)v10, "GSTY", v9);
	} else {
		v11 = 0;
	}
	sub_42C360(v59, v11);
	v12 = calloc(1, 0x10u);
	if (v12) {
		v13 = sub_42C8E0((int)v12, "SCEN", (const char*)a1 + 24);
	} else {
		v13 = 0;
	}
	sub_42C360(v59, v13);
	v14 = calloc(1, 0x10u);
	if (v14) {
		v15 = sub_42C8E0((int)v14, "GNAM", (const char*)a1 + 280);
	} else {
		v15 = 0;
	}
	sub_42C360(v59, v15);
	v16 = *((uint32_t*)a1 + 5);
	v17 = calloc(1, 0x10u);
	if (v17) {
		v18 = sub_42C8B0((int)v17, "DURA", v16);
	} else {
		v18 = 0;
	}
	sub_42C360(v59, v18);
	v60 = *((uint8_t*)a1 + 16);
	v19 = calloc(1, 0x10u);
	if (v19) {
		v20 = sub_42C7F0((int)v19, "TRNY", v60);
	} else {
		v20 = 0;
	}
	sub_42C360(v59, v20);
	v21 = *v2;
	v22 = calloc(1, 0x10u);
	if (v22) {
		v23 = sub_42C820((int)v22, (char*)getMemAt(0x587000, 72000), v21);
	} else {
		v23 = 0;
	}
	sub_42C360(v59, v23);
	v24 = getMemByte(0x5D4594, 741672);
	v25 = calloc(1, 0x10u);
	if (v25) {
		v26 = sub_42C8B0((int)v25, "SEQU", v24);
	} else {
		v26 = 0;
	}
	sub_42C360(v59, v26);
	v27 = 0;
	for (*getMemU32Ptr(0x5D4594, 741664) = 0; v27 < *v2; *getMemU32Ptr(0x5D4594, 741664) = v27) {
		*getMemU8Ptr(0x587000, 71547) = v27 + 48;
		v28 = *(const char**)(*((uint32_t*)v2 + 134) + 4 * v27);
		v29 = calloc(1, 0x10u);
		if (v29) {
			v30 = sub_42C8E0((int)v29, "LGLS", v28);
		} else {
			v30 = 0;
		}
		sub_42C360(v59, v30);
		*getMemU8Ptr(0x587000, 71555) = getMemByte(0x5D4594, 741664) + 48;
		v31 = *(uint32_t*)(*((uint32_t*)v2 + 135) + 4 * *getMemU32Ptr(0x5D4594, 741664));
		v32 = calloc(1, 0x10u);
		if (v32) {
			v33 = sub_42C8B0((int)v32, "IPLS", v31);
		} else {
			v33 = 0;
		}
		sub_42C360(v59, v33);
		*getMemU8Ptr(0x587000, 71563) = getMemByte(0x5D4594, 741664) + 48;
		v61 = *(uint8_t*)(*getMemU32Ptr(0x5D4594, 741664) + *((uint32_t*)v2 + 144));
		v34 = calloc(1, 0x10u);
		if (v34) {
			v35 = sub_42C7F0((int)v34, "CLLS", v61);
		} else {
			v35 = 0;
		}
		sub_42C360(v59, v35);
		*getMemU8Ptr(0x587000, 71571) = getMemByte(0x5D4594, 741664) + 48;
		v36 = *(uint32_t*)(*((uint32_t*)v2 + 136) + 4 * *getMemU32Ptr(0x5D4594, 741664));
		v37 = calloc(1, 0x10u);
		if (v37) {
			v38 = sub_42C8B0((int)v37, "CSTS", v36);
		} else {
			v38 = 0;
		}
		sub_42C360(v59, v38);
		*getMemU8Ptr(0x587000, 71579) = getMemByte(0x5D4594, 741664) + 48;
		v39 = *(uint32_t*)(*((uint32_t*)v2 + 137) + 4 * *getMemU32Ptr(0x5D4594, 741664));
		v40 = calloc(1, 0x10u);
		if (v40) {
			v41 = sub_42C8B0((int)v40, "HSTS", v39);
		} else {
			v41 = 0;
		}
		sub_42C360(v59, v41);
		*getMemU8Ptr(0x587000, 71587) = getMemByte(0x5D4594, 741664) + 48;
		v42 = *(uint32_t*)(*((uint32_t*)v2 + 138) + 4 * *getMemU32Ptr(0x5D4594, 741664));
		v43 = calloc(1, 0x10u);
		if (v43) {
			v44 = sub_42C8B0((int)v43, "MKLS", v42);
		} else {
			v44 = 0;
		}
		sub_42C360(v59, v44);
		*getMemU8Ptr(0x587000, 71595) = getMemByte(0x5D4594, 741664) + 48;
		v45 = *(uint32_t*)(*((uint32_t*)v2 + 139) + 4 * *getMemU32Ptr(0x5D4594, 741664));
		v46 = calloc(1, 0x10u);
		if (v46) {
			v47 = sub_42C8B0((int)v46, "ANKS", v45);
		} else {
			v47 = 0;
		}
		sub_42C360(v59, v47);
		*getMemU8Ptr(0x587000, 71603) = getMemByte(0x5D4594, 741664) + 48;
		v48 = *(uint32_t*)(*((uint32_t*)v2 + 140) + 4 * *getMemU32Ptr(0x5D4594, 741664));
		v49 = calloc(1, 0x10u);
		if (v49) {
			v50 = sub_42C8B0((int)v49, "GNDS", v48);
		} else {
			v50 = 0;
		}
		sub_42C360(v59, v50);
		*getMemU8Ptr(0x587000, 71611) = getMemByte(0x5D4594, 741664) + 48;
		v51 = *(uint32_t*)(*((uint32_t*)v2 + 141) + 4 * *getMemU32Ptr(0x5D4594, 741664));
		v52 = calloc(1, 0x10u);
		if (v52) {
			v53 = sub_42C8B0((int)v52, "SECS", v51);
		} else {
			v53 = 0;
		}
		sub_42C360(v59, v53);
		*getMemU8Ptr(0x587000, 71619) = getMemByte(0x5D4594, 741664) + 48;
		v54 = *(uint32_t*)(*((uint32_t*)v2 + 142) + 4 * *getMemU32Ptr(0x5D4594, 741664));
		v55 = calloc(1, 0x10u);
		if (v55) {
			v56 = sub_42C8B0((int)v55, "BPTS", v54);
		} else {
			v56 = 0;
		}
		sub_42C360(v59, v56);
		*getMemU8Ptr(0x587000, 71627) = getMemByte(0x5D4594, 741664) + 48;
		sub_42BDC0(v59, "SCRS", *(uint32_t*)(*((uint32_t*)v2 + 143) + 4 * *getMemU32Ptr(0x5D4594, 741664)));
		v27 = *getMemU32Ptr(0x5D4594, 741664) + 1;
	}
	v57 = sub_42C480(v59, a2);
	++*getMemU32Ptr(0x5D4594, 741672);
	sub_42C330(v59);
	return v57;
}
#endif


//----- (0042BCE0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42BCE0
int sub_42BCE0(uint32_t* this, char* a2, char a3) {
	uint32_t* v3; // esi
	void* v4;     // eax
	int v5;       // eax

	v3 = this;
	v4 = calloc(1, 0x10u);
	if (v4) {
		v5 = sub_42C7F0((int)v4, a2, a3);
	} else {
		v5 = 0;
	}
	return sub_42C360(v3, v5);
}
#endif


//----- (0042BD50) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42BD50
int sub_42BD50(uint32_t* this, char* a2, char a3) {
	uint32_t* v3; // esi
	void* v4;     // eax
	int v5;       // eax

	v3 = this;
	v4 = calloc(1, 0x10u);
	if (v4) {
		v5 = sub_42C820((int)v4, a2, a3);
	} else {
		v5 = 0;
	}
	return sub_42C360(v3, v5);
}
#endif


//----- (0042BDC0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42BDC0
int sub_42BDC0(uint32_t* this, char* a2, char a3) {
	uint32_t* v3; // esi
	void* v4;     // eax
	int v5;       // eax

	v3 = this;
	v4 = calloc(1, 0x10u);
	if (v4) {
		v5 = sub_42C8B0((int)v4, a2, a3);
	} else {
		v5 = 0;
	}
	return sub_42C360(v3, v5);
}
#endif


//----- (0042BE30) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42BE30
int sub_42BE30(uint32_t* this, char* a2, const char* a3) {
	uint32_t* v3; // esi
	void* v4;     // eax
	int v5;       // eax

	v3 = this;
	v4 = calloc(1, 0x10u);
	if (v4) {
		v5 = sub_42C8E0((int)v4, a2, a3);
	} else {
		v5 = 0;
	}
	return sub_42C360(v3, v5);
}
#endif


//----- (0042BEA0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42BEA0
int sub_42BEA0(uint32_t* this, char* a2, const void* a3, unsigned short a4) {
	uint32_t* v4; // esi
	void* v5;     // eax
	int v6;       // eax

	v4 = this;
	v5 = calloc(1, 0x10u);
	if (v5) {
		v6 = sub_42C910((int)v5, a2, a3, a4);
	} else {
		v6 = 0;
	}
	return sub_42C360(v4, v6);
}
#endif


//----- (0042C330) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C330
void sub_42C330(uint32_t* this) {
	uint32_t* v1; // esi
	uint32_t* v2; // edi

	v1 = (uint32_t*)this[1];
	if (v1) {
		do {
			v2 = (uint32_t*)v1[3];
			if (v1) {
				sub_42CC50(v1);
				free(v1);
			}
			v1 = v2;
		} while (v2);
	}
}
#endif

// 42CC50: using guessed type int sub_42CC50(uint32_t);

//----- (0042C360) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C360
int sub_42C360(uint32_t* this, int a2) {
	int result; // eax

	result = a2;
	*(uint32_t*)(a2 + 12) = this[1];
	this[1] = a2;
	return result;
}
#endif


//----- (0042C480) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C480
uint16_t* sub_42C480(uint32_t* this, unsigned int* a2) {
	uint32_t* v2;    // ebx
	int i;           // eax
	unsigned int v4; // ecx
	uint16_t* v5;    // edi
	int v6;          // ebx
	uint32_t* v7;    // ebp
	char* v8;        // ebp
	int v9;          // ebp

	v2 = this;
	*a2 = 4;
	for (i = this[1]; i; i = *(uint32_t*)(i + 12)) {
		v4 = *a2 + 8;
		*a2 = v4;
		*a2 = *(unsigned short*)(i + 6) + v4 + (-(((unsigned char)*(uint16_t*)(i + 6) + (uint8_t)v4) & 3) & 3);
	}
	v5 = (uint16_t*)calloc(1, *a2);
	*v5 = htons(*(uint16_t*)a2);
	v5[1] = htons(*((uint16_t*)v2 + 1));
	v6 = v2[1];
	v7 = v5 + 2;
	if (!v6) {
		return v5;
	}
	do {
		sub_42CC70(v6);
		*v7 = *(uint32_t*)v6;
		v7[1] = *(uint32_t*)(v6 + 4);
		v8 = (char*)(v7 + 2);
		memcpy(v8, *(const void**)(v6 + 8), ntohs(*(uint16_t*)(v6 + 6)));
		v9 = (int)&v8[ntohs(*(uint16_t*)(v6 + 6))];
		v7 = (uint32_t*)((-(ntohs(*(uint16_t*)(v6 + 6)) & 3) & 3) + v9);
		sub_42CCE0((uint16_t*)v6);
		v6 = *(uint32_t*)(v6 + 12);
	} while (v6);
	return v5;
}
#endif


//----- (0042C770) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C770
int sub_42C770(void** this) {
	void** v1;  // ebx
	int result; // eax

	v1 = this;
	free(this[2]);
	result = 0;
	strcpy((char*)v1, (const char*)getMemAt(0x5D4594, 741688));
	*((uint16_t*)v1 + 2) = 0;
	*((uint16_t*)v1 + 3) = 0;
	v1[2] = 0;
	v1[3] = 0;
	return result;
}
#endif


//----- (0042C7F0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C7F0
int sub_42C7F0(int this, char* a2, char a3) {
	int v3; // esi

	v3 = this;
	*(uint32_t*)(this + 8) = 0;
	sub_42C9A0((void**)this, a2, a3);
	return v3;
}
#endif


//----- (0042C820) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C820
int sub_42C820(int this, char* a2, char a3) {
	int v3; // esi

	v3 = this;
	*(uint32_t*)(this + 8) = 0;
	sub_42CA00((void**)this, a2, a3);
	return v3;
}
#endif


//----- (0042C8B0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C8B0
int sub_42C8B0(int this, char* a2, char a3) {
	int v3; // esi

	v3 = this;
	*(uint32_t*)(this + 8) = 0;
	sub_42CB20((void**)this, a2, a3);
	return v3;
}
#endif


//----- (0042C8E0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C8E0
int sub_42C8E0(int this, char* a2, const char* a3) {
	int v3; // esi

	v3 = this;
	*(uint32_t*)(this + 8) = 0;
	sub_42CB80((void**)this, a2, a3);
	return v3;
}
#endif


//----- (0042C910) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C910
int sub_42C910(int this, char* a2, const void* a3, unsigned short a4) {
	int v4; // esi

	v4 = this;
	*(uint32_t*)(this + 8) = 0;
	sub_42CBF0((void**)this, a2, a3, a4);
	return v4;
}
#endif


//----- (0042C9A0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42C9A0
char* sub_42C9A0(void** this, char* a2, char a3) {
	void** v3;       // ebx
	char* result;    // eax
	unsigned int v5; // edx

	v3 = this;
	sub_42C770(this);
	strncpy((char*)v3, a2, 4u);
	*((uint16_t*)v3 + 2) = 2;
	*((uint16_t*)v3 + 3) = 1;
	result = (char*)calloc(1, 1u);
	v5 = *((unsigned short*)v3 + 3);
	v3[2] = result;
	memcpy(result, &a3, v5);
	v3[3] = 0;
	return result;
}
#endif


//----- (0042CA00) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42CA00
char* sub_42CA00(void** this, char* a2, char a3) {
	void** v3;       // ebx
	char* result;    // eax
	unsigned int v5; // edx

	v3 = this;
	sub_42C770(this);
	strncpy((char*)v3, a2, 4u);
	*((uint16_t*)v3 + 2) = 3;
	*((uint16_t*)v3 + 3) = 2;
	result = (char*)calloc(1, 2u);
	v5 = *((unsigned short*)v3 + 3);
	v3[2] = result;
	int a3i = a3;
	memcpy(result, &a3i, v5);
	v3[3] = 0;
	return result;
}
#endif


//----- (0042CB20) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42CB20
char* sub_42CB20(void** this, char* a2, char a3) {
	void** v3;       // ebx
	char* result;    // eax
	unsigned int v5; // edx

	v3 = this;
	sub_42C770(this);
	strncpy((char*)v3, a2, 4u);
	*((uint16_t*)v3 + 2) = 6;
	*((uint16_t*)v3 + 3) = 4;
	result = (char*)calloc(1, 4u);
	v5 = *((unsigned short*)v3 + 3);
	v3[2] = result;
	int a3i = a3;
	memcpy(result, &a3i, v5);
	v3[3] = 0;
	return result;
}
#endif


//----- (0042CB80) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42CB80
char* sub_42CB80(void** this, char* a2, const char* a3) {
	void** v3;         // ebx
	unsigned int v4;   // kr08_4
	char* result;      // eax
	unsigned short v6; // cx

	v3 = this;
	sub_42C770(this);
	strncpy((char*)v3, a2, 4u);
	*((uint16_t*)v3 + 2) = 7;
	v4 = strlen(a3) + 1;
	*((uint16_t*)v3 + 3) = v4;
	result = (char*)calloc(1, (unsigned short)v4);
	v6 = *((uint16_t*)v3 + 3);
	v3[2] = result;
	memcpy(result, a3, v6);
	v3[3] = 0;
	return result;
}
#endif


//----- (0042CBF0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42CBF0
void* sub_42CBF0(void** this, char* a2, const void* a3, unsigned short a4) {
	void** v4;       // ebx
	void* result;    // eax
	unsigned int v6; // edx

	v4 = this;
	sub_42C770(this);
	strncpy((char*)v4, a2, 4u);
	*((uint16_t*)v4 + 2) = 20;
	*((uint16_t*)v4 + 3) = a4;
	result = calloc(1, a4);
	v6 = *((unsigned short*)v4 + 3);
	v4[2] = result;
	memcpy(result, a3, v6);
	v4[3] = 0;
	return result;
}
#endif


//----- (0042CC70) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42CC70
uint16_t sub_42CC70(int this) {
	int v1;          // esi
	uint16_t v2;     // ax
	uint16_t v3;     // cx
	uint16_t result; // ax

	v1 = this;
	switch (*(unsigned short*)(this + 4)) {
	case 3u:
	case 4u:
		**(uint16_t**)(this + 8) = htons(**(uint16_t**)(this + 8));
		break;
	case 5u:
	case 6u:
		**(uint32_t**)(this + 8) = htonl(**(uint32_t**)(this + 8));
		break;
	default:
		break;
	}
	v2 = htons(*(uint16_t*)(v1 + 4));
	v3 = *(uint16_t*)(v1 + 6);
	*(uint16_t*)(v1 + 4) = v2;
	result = htons(v3);
	*(uint16_t*)(v1 + 6) = result;
	return result;
}
#endif


//----- (0042CCE0) --------------------------------------------------------
#if 0 // Closed legacy WOL component: see split-wol-unused-cleanup-20261115-v1 review; sub_42CCE0
short sub_42CCE0(uint16_t* this) {
	uint16_t* v1; // esi
	uint16_t v2;  // ax
	uint16_t v3;  // cx
	u_long v4;    // eax

	v1 = this;
	v2 = ntohs(this[2]);
	v3 = v1[3];
	v1[2] = v2;
	v1[3] = ntohs(v3);
	LOWORD(v4) = v1[2] - 3;
	switch (v1[2]) {
	case 3u:
	case 4u:
		LOWORD(v4) = ntohs(**((uint16_t**)v1 + 2));
		**((uint16_t**)v1 + 2) = v4;
		break;
	case 5u:
	case 6u:
		v4 = ntohl(**((uint32_t**)v1 + 2));
		**((uint32_t**)v1 + 2) = v4;
		break;
	default:
		return v4;
	}
	return v4;
}
#endif


//----- (00430AF0) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int nox_client_mousePriKey_430AF0() { return dword_5d4594_805820; }
#endif

//----- (00430B00) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int nox_xxx_cursor_430B00() { return nox_xxx_useAudio_587000_80772; }
#endif

//----- (00431270) --------------------------------------------------------
#if 0 // Native audio context owner: sub_431270
void sub_431270() {
	if (dword_5d4594_805984) {
		sub_487680(dword_5d4594_805984);
		dword_5d4594_805984 = 0;
	}
}
#endif


//----- (00431290) --------------------------------------------------------
#if 0 // Native audio context owner: sub_431290
void sub_431290() {
	if (dword_5d4594_805984) {
		sub_487970(dword_5d4594_805984, -1);
	}
}
#endif


// 4335F8: variable 'v12' is possibly undefined

//----- (004357A0) --------------------------------------------------------
#if 0 // C reference for the observer visibility fixture.
int nox_xxx_cliToggleObsWindow_4357A0() {
	int result; // eax
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;

	if (player && (player->field_3680 & 1)) {
		result = nox_xxx_showObserverWindow_48CA70(0);
	} else {
		result = nox_xxx_showObserverWindow_48CA70(1);
	}
	return result;
}

#endif
//----- (004372E0) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int nox_xxx_clientIsObserver_4372E0() {
	int result; // eax
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;

	if (player && player->active == 1) {
		result = (player->field_3680 & 3) != 0;
	} else {
		result = 0;
	}
	return result;
}
#endif

// 439385: variable 'v2' is possibly undefined

// 43A3CE: variable 'v22' is possibly undefined

//----- (0043AF30) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int sub_43AF30() { return dword_5d4594_815052; }
#endif

//----- (0043AF40) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int sub_43AF40() { return nox_game_createOrJoin_815048; }
#endif

//----- (0043AF80) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int sub_43AF80() { return dword_5d4594_814548; }
#endif
