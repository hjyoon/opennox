#include <math.h>
#include <stdlib.h>
#include <string.h>

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
#include "GAME5_2.h"
#include "client__draw__animdraw.h"
#include "client__drawable__drawable.h"
#include "client__gui__guisumn.h"
#include "client__gui__servopts__advserv.h"
#include "common__net_list.h"
#include "common__system__team.h"

#include "client__gui__guiinput.h"
#include "client__gui__servopts__objlst.h"
#include "client__gui__servopts__spelllst.h"
#include "client__gui__tooltip.h"
#include "client__gui__window.h"
#include "client__shell__inputcfg__inputcfg.h"
#include "client__shell__mainmenu.h"
#include "client__shell__selcolor.h"

#include "client__draw__plasma.h"
#include "client__drawable__update__charmup.h"
#include "client__drawable__update__fireball.h"
#include "client__system__ctrlevnt.h"

#include "client__shell__optsback.h"
#include "client__video__draw_common.h"
#include "common/fs/nox_fs.h"
#include "common__strman.h"
#include "input_common.h"
#include "operators.h"

extern uint32_t dword_8531A0_2572;
extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_587000_180476;
extern uint32_t dword_5d4594_1319260;
extern uint32_t dword_5d4594_1321196;
extern uint32_t dword_5d4594_1313880;
extern uint32_t dword_5d4594_1320944;
extern uint32_t dword_5d4594_1320948;
extern uint32_t dword_587000_183460;
extern uint32_t dword_5d4594_1316412;
extern uint32_t dword_5d4594_1522968;
extern uint32_t dword_5d4594_1321520;
extern uint32_t dword_5d4594_1319268;
extern nox_video_bag_image_t* dword_5d4594_1321024;
extern uint32_t dword_587000_183456;
extern uint32_t dword_5d4594_1320988;
extern uint32_t dword_5d4594_1319248;
extern uint32_t dword_587000_180480;
extern nox_window* dword_5d4594_1319236;
extern uint32_t dword_5d4594_1316704;
extern uint32_t dword_5d4594_1321208;
extern uint32_t dword_5d4594_1321800;
extern nox_window* dword_5d4594_1321224;
extern uint32_t dword_5d4594_1319056;
extern nox_window* dword_5d4594_1321032;
extern uint32_t dword_5d4594_1319264;
extern nox_window* dword_5d4594_1321044;
extern nox_window* dword_5d4594_1319232;
extern nox_window* dword_5d4594_1321216;
extern nox_window* dword_5d4594_1321236;
extern nox_window* dword_5d4594_1321240;
extern nox_window* dword_5d4594_1321248;
extern nox_window* dword_5d4594_1321244;
nox_window* dword_5d4594_1522616 = 0;
nox_window* dword_5d4594_1522620 = 0;
nox_window* dword_5d4594_1522624 = 0;
nox_window* dword_5d4594_1522628 = 0;
extern uint32_t dword_5d4594_1320992;
extern uint32_t dword_5d4594_1316408;
extern uint64_t qword_581450_9512;
extern uint64_t qword_581450_9544;
extern nox_window* dword_5d4594_1319060;
extern nox_window* dword_5d4594_1522632;
extern nox_window* dword_5d4594_1321252;
extern nox_window* dword_5d4594_1522612;
extern nox_window* dword_5d4594_1522604;
extern nox_window* dword_5d4594_1321232;
extern nox_window* dword_5d4594_1316712;
extern uint32_t dword_5d4594_1320964;
extern nox_window* dword_5d4594_1316708;
extern nox_window* dword_5d4594_1321228;
extern nox_window* dword_5d4594_1321040;
extern nox_window* dword_5d4594_1316972;
extern nox_window* dword_5d4594_1320940;
extern uint32_t nox_player_netCode_85319C;
extern int nox_win_width;
extern int nox_win_height;

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_violet_2598268;
extern uint32_t nox_color_black_2650656;

nox_gui_animation* nox_wnd_xxx_1522608 = 0;

nox_drawable* nox_gui_itemAmount_item_1319256 = NULL;
nox_window* nox_gui_itemAmount_dialog_1319228 = NULL;
nox_video_bag_image_t* nox_gui_itemAmount_images_1319196[8] = {0};
nox_gui_item_amount_callback_t nox_gui_itemAmount_accept_1319160 = NULL;
nox_gui_item_amount_callback_t nox_gui_itemAmount_cancel_1319100 = NULL;

//----- (004BD280) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD280
uint32_t* sub_4BD280(int a1, int a2) {
	int v2;           // esi
	uint32_t* result; // eax
	uint32_t* v4;     // ecx
	int v5;           // edi

	v2 = a2 + 4;
	result = calloc(1, a1 * (a2 + 4) + 4);
	if (result) {
		v4 = result + 1;
		*result = result + 1;
		if (a1 != 1) {
			v5 = a1 - 1;
			do {
				--v5;
				*v4 = (char*)v4 + v2;
				v4 = (uint32_t*)((char*)v4 + v2);
			} while (v5);
		}
		*v4 = 0;
	}
	return result;
}
#endif


//----- (004BD2D0) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD2D0
void sub_4BD2D0(void* lpMem) { free(lpMem); }
#endif


//----- (004BD2E0) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD2E0
uint32_t* sub_4BD2E0(uint32_t** a1) {
	uint32_t* result; // eax
	uint32_t* v2;     // edx

	result = *a1;
	if (*a1) {
		v2 = (uint32_t*)*result;
		++result;
		*a1 = v2;
	}
	return result;
}
#endif


//----- (004BD300) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD300
int sub_4BD300(uint32_t* a1, int a2) {
	int result; // eax

	result = a2 - 4;
	*(uint32_t*)(a2 - 4) = *a1;
	*a1 = a2 - 4;
	return result;
}
#endif


//----- (004BD340) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD340
uint32_t* sub_4BD340(int a1, int a2, int a3, int a4) {
	uint32_t* v4; // esi

	v4 = calloc(1, 0x1Cu);
	memset(v4, 0, 0x1Cu);
	*v4 = a1;
	v4[6] = a4;
	v4[1] = sub_4BD280(a2 / (a4 + 24), a4 + 24);
	v4[2] = sub_4BD280(a3, 84);
	nox_common_list_clear_425760(v4 + 3);
	if (v4[1] && v4[2]) {
		return v4;
	}
	sub_4BD3C0(v4);
	return 0;
}
#endif


//----- (004BD3C0) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD3C0
void sub_4BD3C0(void* lpMem) {
	int i; // eax

	for (i = nox_common_list_getNext_425940((int*)lpMem + 3); i; i = nox_common_list_getNext_425940((int*)lpMem + 3)) {
		sub_4BD690(i);
	}
	if (*((uint32_t*)lpMem + 1)) {
		sub_4BD2D0(*((void**)lpMem + 1));
	}
	if (*((uint32_t*)lpMem + 2)) {
		sub_4BD2D0(*((void**)lpMem + 2));
	}
	free(lpMem);
}
#endif


//----- (004BD420) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD420
uint32_t* sub_4BD420(int a1, int a2) {
	uint32_t* result; // eax

	result = *(uint32_t**)(a1 + 12);
	if (result == (uint32_t*)(a1 + 12)) {
		return 0;
	}
	while (result[4] != a2 || !result[5]) {
		result = (uint32_t*)*result;
		if (result == (uint32_t*)(a1 + 12)) {
			return 0;
		}
	}
	return result;
}
#endif


//----- (004BD470) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD470
uint32_t* sub_4BD470(uint32_t** a1, int a2) {
	uint32_t* v2;  // eax
	uint32_t* v3;  // edi
	uint32_t* v5;  // ebx
	uint32_t* v6;  // eax
	uint32_t* v7;  // ebp
	char* v8;      // edi
	signed int v9; // eax
	uint32_t* v10; // [esp+18h] [ebp+8h]

	v2 = sub_4BD420((int)a1, a2);
	v3 = v2;
	if (v2) {
		nox_common_list_remove_425920((uint32_t**)v2);
		sub_425900(a1 + 3, v3);
		return v3;
	}
	if (!sub_486B60((int)*a1, a2)) {
		return 0;
	}
	v5 = sub_4BD2E0((uint32_t**)a1[2]);
	if (!v5) {
		sub_4BD600((int)a1);
		v5 = sub_4BD2E0((uint32_t**)a1[2]);
		if (!v5) {
			sub_486E00((int)*a1);
			return 0;
		}
	}
	v5[4] = a2;
	v5[13] = a1;
	sub_425770(v5);
	v5[3] = 0;
	sub_487C30(v5 + 6);
	nullsub_10(v5 + 14);
	v5[11] = v5 + 14;
	v6 = (uint32_t*)(*a1)[71];
	v10 = (uint32_t*)(*a1)[71];
	if (!v6) {
		sub_486AA0(*a1, v5[4], v5 + 14);
		sub_425900(a1 + 3, v5);
		v5[5] = 1;
		sub_486E00((int)*a1);
		return v5;
	}
	while (1) {
		v7 = a1[6];
		if ((int)v7 > (int)v6) {
			v7 = v6;
		}
		v8 = (char*)sub_4BD2E0((uint32_t**)a1[1]);
		if (!v8) {
			while (sub_4BD600((int)a1)) {
				v8 = (char*)sub_4BD2E0((uint32_t**)a1[1]);
				if (v8) {
					goto LABEL_17;
				}
			}
			sub_4BD690((int)v5);
			return 0;
		}
	LABEL_17:
		sub_487D30(v8, (int)(v8 + 24), (int)v7);
		sub_487C50((int)(v5 + 6), v8);
		v9 = sub_486DB0((int)*a1, v8 + 24, (signed int)v7);
		if ((uint32_t*)v9 != v7) {
			sub_4BD690((int)v5);
			return 0;
		}
		v10 = (uint32_t*)((char*)v10 - v9);
		if (!v10) {
			sub_486AA0(*a1, v5[4], v5 + 14);
			sub_425900(a1 + 3, v5);
			v5[5] = 1;
			sub_486E00((int)*a1);
			return v5;
		}
		v6 = v10;
	}
}
#endif

// 487CF0: using guessed type void  nullsub_10(uint32_t);

//----- (004BD600) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD600
int sub_4BD600(int a1) {
	int v1; // esi

	v1 = sub_425960(a1 + 12);
	if (!v1) {
		return 0;
	}
	while (sub_4BD680(v1)) {
		v1 = sub_425960(v1);
		if (!v1) {
			return 0;
		}
	}
	sub_4BD690(v1);
	return 1;
}
#endif


//----- (004BD650) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD650
int sub_4BD650(int a1) {
	int result; // eax

	result = a1;
	++*(uint32_t*)(a1 + 12);
	return result;
}
#endif


//----- (004BD660) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD660
int sub_4BD660(int a1) {
	int result; // eax

	result = *(uint32_t*)(a1 + 12) - 1;
	*(uint32_t*)(a1 + 12) = result;
	if (result < 0) {
		*(uint32_t*)(a1 + 12) = 0;
	}
	return result;
}
#endif


//----- (004BD680) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD680
int sub_4BD680(int a1) { return *(uint32_t*)(a1 + 12); }
#endif


//----- (004BD690) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD690
int sub_4BD690(int a1) {
	uint32_t** i; // esi

	if (*(uint32_t*)(a1 + 4) != a1) {
		nox_common_list_remove_425920((uint32_t**)a1);
	}
	for (i = (uint32_t**)nox_common_list_getNext_425940((int*)(a1 + 32)); i;
		 i = (uint32_t**)nox_common_list_getNext_425940((int*)(a1 + 32))) {
		nox_common_list_remove_425920(i);
		sub_487D60((int)i);
		sub_4BD300(*(uint32_t**)(*(uint32_t*)(a1 + 52) + 4), (int)i);
	}
	nullsub_9(a1 + 24);
	return sub_4BD300(*(uint32_t**)(*(uint32_t*)(a1 + 52) + 8), a1);
}
#endif

// 487CA0: using guessed type void  nullsub_9(uint32_t);

//----- (004BD710) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_cache_native.go:sub_4BD710
int sub_4BD710(int a1) { return a1 + 24; }
#endif


//----- (004BD720) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BD720
uint32_t* sub_4BD720(int a1) {
	uint32_t* v1; // esi

	v1 = calloc(1, 0x138u);
	memset(v1, 0, 0x138u);
	sub_425770(v1);
	sub_4BDC00((int)(v1 + 30));
	sub_4864A0(v1 + 44);
	sub_4BD7C0(v1);
	v1[33] = a1;
	v1[43] = *(uint32_t*)(a1 + 256);
	if (!(*(int (**)(uint32_t*))(*(uint32_t*)(a1 + 256) + 4))(v1)) {
		return v1;
	}
	if (v1) {
		sub_4BD7A0(v1);
	}
	return 0;
}
#endif


//----- (004BD7A0) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BD7A0
void sub_4BD7A0(void* lpMem) {
	(*(void (**)(void*))(*((uint32_t*)lpMem + 43) + 8))(lpMem);
	free(lpMem);
}
#endif


//----- (004BD7C0) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BD7C0
uint32_t* sub_4BD7C0(uint32_t* a1) {
	uint32_t* result; // eax

	a1[69] = sub_4BD8C0;
	a1[70] = sub_4BD940;
	a1[71] = sub_4BD9B0;
	a1[34] = 0;
	a1[35] = 0;
	a1[36] = 0;
	a1[38] = 0;
	a1[3] = 1;
	sub_4BDC00((int)(a1 + 30));
	a1[30] = 0;
	a1[29] = *getMemU32Ptr(0x5D4594, 1193340);
	a1[28] = 0;
	result = sub_4864A0(a1 + 4);
	a1[72] = 0;
	return result;
}
#endif


//----- (004BD840) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BD840
void sub_4BD840(int a3) {
	int v1;           // ebp
	unsigned int* v2; // edi
	uint32_t* v3;     // esi
	int result;       // eax

	v1 = *(uint32_t*)(a3 + 132);
	v2 = (unsigned int*)(a3 + 176);
	sub_4864A0((uint32_t*)(a3 + 176));
	sub_486570((unsigned int*)(a3 + 176), (uint32_t*)(a3 + 16));
	sub_486620((uint32_t*)(a3 + 16));
	if (*(uint32_t*)(a3 + 112)) {
		sub_486570(v2, *(uint32_t**)(a3 + 112));
		sub_486620(*(uint32_t**)(a3 + 112));
	}
	v3 = *(uint32_t**)(a3 + 116);
	if (v3) {
		sub_486570(v2, v3);
	}
	sub_486570(v2, (uint32_t*)(v1 + 88));
	result = *(uint32_t*)(v1 + 184);
	if (result) {
		 sub_486570(v2, *(uint32_t**)(v1 + 184));
	}
}
#endif


//----- (004BD8C0) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BD8C0
int sub_4BD8C0(int a1) {
	int (*v1)(int); // eax
	int result;     // eax
	int v3;         // eax
	int v4;         // eax

	v1 = *(int (**)(int))(a1 + 136);
	if (v1) {
		result = v1(a1);
		if (result) {
			*(uint32_t*)(a1 + 300) = 0;
			*(uint32_t*)(a1 + 304) = 0;
			*(uint32_t*)(a1 + 296) = 0;
			return result;
		}
	} else {
		if (*(uint32_t*)(a1 + 292)) {
			v3 = nox_common_list_getNext_425940(*(int**)(a1 + 292));
			*(uint32_t*)(a1 + 292) = v3;
			if (v3) {
				*(uint32_t*)(a1 + 296) = *(uint32_t*)(v3 + 12);
				v4 = *(uint32_t*)(v3 + 16);
				*(uint32_t*)(a1 + 300) = v4;
				*(uint32_t*)(a1 + 304) = v4;
				return 0;
			}
		}
		*(uint32_t*)(a1 + 300) = 0;
	}
	return 0;
}
#endif


//----- (004BD940) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BD940
int sub_4BD940(int a1) {
	void (*v1)(int); // eax

	if (*(uint32_t*)(a1 + 128)) {
		if (*(int*)(a1 + 128) != -1) {
			--*(uint32_t*)(a1 + 128);
		}
		sub_4BDB90((uint32_t*)a1, *(uint32_t**)(a1 + 288));
	} else {
		sub_4BDB90((uint32_t*)a1, 0);
	}
	v1 = *(void (**)(int))(a1 + 140);
	if (v1) {
		v1(a1);
	}
	if (*(uint32_t*)(a1 + 288)) {
		(*(void (**)(int))(*(uint32_t*)(a1 + 172) + 36))(a1);
	}
	return 0;
}
#endif


//----- (004BD9B0) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BD9B0
int sub_4BD9B0(uint32_t* a2) {
	int v1;               // eax
	int (*v2)(uint32_t*); // eax
	int result;           // eax

	a2[72] = 0;
	v1 = a2[31];
	LOBYTE(v1) = v1 & 0xFA;
	a2[31] = v1;
	a2[32] = 0;
	sub_4864A0(a2 + 4);
	v2 = (int (*)(uint32_t*))a2[36];
	if (v2) {
		result = v2(a2);
	} else {
		result = 0;
	}
	return result;
}
#endif


//----- (004BDA60) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BDA60
void sub_4BDA60(void* lpMem) {
	sub_4BDA80((int)lpMem);
	sub_486E90((int)lpMem);
	sub_4BD7A0(lpMem);
}
#endif


//----- (004BDA80) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BDA80
int sub_4BDA80(int a1) {
	int result = 0;      // eax
	int (*result2)(int); // eax

	if (*(uint8_t*)(a1 + 124) & 5) {
		(*(void (**)(int))(*(uint32_t*)(a1 + 172) + 16))(a1);
	}
	result2 = *(int (**)(int))(a1 + 148);
	if (result2) {
		result = (int)result2(a1);
	}
	*(uint32_t*)(a1 + 288) = 0;
	return result;
}
#endif


//----- (004BDB20) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BDB20
int sub_4BDB20(int a1) {
	int result; // eax

	result = a1;
	*(uint32_t*)(a1 + 124) |= 0x10u;
	return result;
}
#endif


//----- (004BDB30) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BDB30
int sub_4BDB30(int a1) {
	int result; // eax

	result = a1;
	*(uint32_t*)(a1 + 124) &= 0xFFFFFFEF;
	return result;
}
#endif


//----- (004BDB40) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BDB40
int sub_4BDB40(int a2) {
	int result; // eax

	if (*(uint8_t*)(a2 + 124) & 5) {
		return -2146500608;
	}
	if (!*(uint32_t*)(a2 + 288)) {
		return -2147024896;
	}
	sub_486520((unsigned int*)(a2 + 16));
	sub_4BD840(a2);
	result = (*(int (**)(int))(*(uint32_t*)(a2 + 172) + 12))(a2);
	if (!result) {
		*(uint32_t*)(a2 + 124) |= 1u;
	}
	return result;
}
#endif


//----- (004BDB90) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BDB90
void sub_4BDB90(uint32_t* a1, uint32_t* a2) {
	int v2;       // eax
	int v3;       // eax
	uint32_t* v4; // eax
	int v5;       // eax

	a1[72] = a2;
	if (a2) {
		v2 = sub_487C80((int)a2);
		a1[73] = v2;
		if (v2) {
			a1[74] = *(uint32_t*)(v2 + 12);
			v3 = *(uint32_t*)(v2 + 16);
			a1[75] = v3;
			a1[76] = v3;
			*a2 = 0;
		} else {
			v4 = (uint32_t*)a1[72];
			a1[74] = *v4;
			v5 = v4[1];
			a1[75] = v5;
			a1[76] = v5;
		}
	}
}
#endif


//----- (004BDC00) --------------------------------------------------------
#if 0 // Native audio owner: src/legacy/audio_source_native.go:sub_4BDC00
int sub_4BDC00(int a1) {
	int result; // eax

	result = a1;
	*(uint32_t*)(a1 + 8) = 0;
	*(uint32_t*)(a1 + 4) = 0;
	return result;
}
#endif


#if 0 // The secondary weapon encoder remains available to the C fixture.
//----- (004BF010) --------------------------------------------------------
int nox_xxx_clientReportSecondaryWeapon_4BF010(nox_drawable* drawable) {
	char v3[3]; // [esp+0h] [ebp-4h]
	v3[0] = -32;
	*(uint16_t*)&v3[1] = nox_xxx_netGetUnitCodeCli_578B00(drawable);
	return nox_xxx_netClientSend2_4E53C0(31, v3, 3, 0, 1);
}

int nox_client_getPlayerDollState_4BF7E0(nox_player_doll_state_t* state) {
	if (!state) {
		return 0;
	}
	memset(state, 0, sizeof(*state));
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	if (!player) {
		return 0;
	}
	state->armor_mask = player->field_0;
	state->weapon_mask = player->field_4;
	state->color_skin = player->color_skin;
	state->color_hair = player->color_hair;
	state->color_mustache = player->color_mustache;
	state->color_goatee = player->color_goatee;
	state->color_beard = player->color_beard;
	state->color_unknown = player->color_unknown;
	state->variant = player->info.isFemale;
	return 1;
}

nox_video_bag_image_t* nox_client_getPlayerDollImage_4BF9F0(uintptr_t table_offset, int layer) {
	return getMemPtr(0x973A20, table_offset + 4 * (uintptr_t)layer);
}

//----- (004BF7E0) --------------------------------------------------------
short sub_4BF7E0(uint32_t* a1) {
	int v1;             // esi
	int v2;             // ebp
	int v3 = 0;         // eax
	int v5;             // ebx
	int v6;             // eax
	int v7;             // eax
	int v8;             // ebx
	int v9;             // eax
	int v10;            // ebx
	int v11;            // eax
	int v13;            // [esp+10h] [ebp-4h]
	uintptr_t v14;      // [esp+18h] [ebp+4h]
	uintptr_t v15;      // [esp+18h] [ebp+4h]
	nox_player_doll_state_t doll;

	v1 = a1[1] + 15;
	v2 = *a1 + 11;
	v13 = a1[1] + 15;
	nox_client_drawSetColor_434460(nox_color_black_2650656);
	nox_client_drawRectFilledOpaque_49CE30(v2, v1, 200, 200);
	if (getMemPtr(0x852978, 8)) {
		if (nox_client_getPlayerDollState_4BF7E0(&doll)) {
			nox_draw_setMaterial_4341D0(1, doll.color_hair);
			nox_draw_setMaterial_4341D0(2, doll.color_goatee);
			nox_draw_setMaterial_4341D0(3, doll.color_unknown);
			nox_draw_setMaterial_4341D0(4, doll.color_beard);
			nox_draw_setMaterial_4341D0(5, doll.color_mustache);
			nox_draw_setMaterial_4341D0(6, doll.color_skin);
			if (doll.color_skin == doll.color_hair) {
				nox_client_drawImageAt_47D2C0(nox_client_getPlayerDollImage_4BF9F0(24, doll.variant), v2, v1);
			} else {
				nox_client_drawImageAt_47D2C0(nox_client_getPlayerDollImage_4BF9F0(16, doll.variant), v2, v1);
			}
			v5 = 0;
			v14 = 32 + 104 * (uintptr_t)doll.variant;
			do {
				if (doll.armor_mask & (1 << v5) && !((1 << v5) & 0x3000000)) {
					v6 = sub_415CD0((char*)(1 << v5));
					sub_4BF9F0(1 << v5, v6, v2, v13, v14, v5, 0);
				}
				++v5;
			} while (v5 < 26);
			if (doll.armor_mask & 2) {
				v7 = sub_415CD0((char*)2);
				sub_4BF9F0(2, v7, v2, v13, v14, 0, 1);
			}
			v8 = 0;
			do {
				if (doll.armor_mask & (1 << v8) && (1 << v8) & 0x3000000) {
					v9 = sub_415CD0((char*)(1 << v8));
					sub_4BF9F0(1 << v8, v9, v2, v13, v14, v8, 0);
				}
				++v8;
			} while (v8 < 26);
			v10 = 0;
			v15 = 256 + 108 * (uintptr_t)doll.variant;
			do {
				v3 = doll.weapon_mask;
				if (v3 & (1 << v10)) {
					v11 = sub_415840((char*)(1 << v10));
					LOWORD(v3) = sub_4BF9F0(1 << v10, v11, v2, v13, v15, v10, 0);
				}
				++v10;
			} while (v10 < 27);
		}
	}
	return v3;
}

//----- (004BF9F0) --------------------------------------------------------
short sub_4BF9F0(int a1, int a2, int a3, int a4, uintptr_t table_offset, int a6, int a7) {
	(void)a1;
	nox_drawable* drawable = sub_461600(a2);
	if (drawable) {
		void* def;
		if (drawable->flags28 & 0x2000000) {
			def = nox_xxx_equipClothFindDefByTT_413270(drawable->field_27);
		} else {
			def = nox_xxx_getProjectileClassById_413250(drawable->field_27);
		}
		if (def) {
			for (int i = 0; i < 6; ++i) {
				uint32_t color = nox_modifier_getColorRGB(def, i + 1);
				nox_draw_setMaterial_4340A0(i + 1, color & 0xff, (color >> 8) & 0xff,
									 (color >> 16) & 0xff);
			}
			for (int i = 0; i < 4; ++i) {
				void* modifier = drawable->item_modifiers[i];
				if (modifier) {
					uint32_t color = nox_modifier_effect_getColorRGB(modifier);
					nox_draw_setMaterial_4340A0(nox_modifier_getColorSlot(def, i), color & 0xff,
									 (color >> 8) & 0xff, (color >> 16) & 0xff);
				}
			}
		}
		if (a7) {
			nox_client_drawImageAt_47D2C0(getMemPtr(0x5D4594, 1319052), a3, a4);
		} else {
			nox_client_drawImageAt_47D2C0(nox_client_getPlayerDollImage_4BF9F0(table_offset, a6), a3, a4);
		}
	}
	return (short)(uintptr_t)drawable;
}
// 4BFA4C: variable 'v10' is possibly undefined
// 4BFA4C: variable 'v11' is possibly undefined

//----- (004BFAD0) --------------------------------------------------------
int sub_4BFAD0() {
	int v0;         // esi
	int v1;         // ebx
	int i;          // edi
	char* v3;       // eax
	const char* v4; // ecx
	int v5;         // ebp
	int v6;         // ebp

	v0 = 0;
	v1 = 0;
	for (i = 0; i < 8; i += 4) {
		v3 = nox_xxx_gLoadImg_42F970((const char*)getMemPtr(0x587000, 180960 + i));
		v4 = (const char*)getMemPtr(0x587000, 180968 + i);
		setMemPtr(0x973A20, 16 + i, v3);
		setMemPtr(0x973A20, 24 + i, nox_xxx_gLoadImg_42F970(v4));
		v5 = 26;
		do {
			setMemPtr(0x973A20, 32 + v1,
					  nox_xxx_gLoadImg_42F970((const char*)getMemPtr(0x587000, 180976 + v1)));
			v1 += 4;
			--v5;
		} while (v5);
		v6 = 27;
		do {
			setMemPtr(0x973A20, 256 + v0,
					  nox_xxx_gLoadImg_42F970((const char*)getMemPtr(0x587000, 181184 + v0)));
			v0 += 4;
			--v6;
		} while (v6);
	}
	setMemPtr(0x5D4594, 1319052, nox_xxx_gLoadImg_42F970("MaleMedievalCloakTop"));
	return 1;
}

#endif

//----- (004BFD30) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_4BFD30() { return dword_5d4594_1319056; }
#endif

#if 0 // Original item amount draw logic for the C layout fixture.
enum {
	NOX_ITEM_AMOUNT_IMAGE_BASE,
	NOX_ITEM_AMOUNT_IMAGE_UP_LIT,
	NOX_ITEM_AMOUNT_IMAGE_DOWN_LIT,
	NOX_ITEM_AMOUNT_IMAGE_YES_PRESSED,
	NOX_ITEM_AMOUNT_IMAGE_NO_PRESSED,
	NOX_ITEM_AMOUNT_IMAGE_BASE_NO_TAG,
	NOX_ITEM_AMOUNT_IMAGE_YES_PRESSED_NO_TAG,
	NOX_ITEM_AMOUNT_IMAGE_NO_PRESSED_NO_TAG,
};

//----- (004C0030) --------------------------------------------------------
int sub_4C0030(nox_window* win, void* draw) {
	(void)draw;
	nox_client_drawRectFilledAlpha_49CF10(0, 0, nox_win_width, nox_win_height);
	unsigned int x;
	unsigned int y;
	nox_gui_getWindowOffs_46AA20(win, &x, &y);
	nox_video_bag_image_t* base =
		nox_gui_itemAmount_images_1319196[dword_5d4594_1319264 ? NOX_ITEM_AMOUNT_IMAGE_BASE
															   : NOX_ITEM_AMOUNT_IMAGE_BASE_NO_TAG];
	nox_client_drawImageAt_47D2C0(base, x, y);
	nox_gui_itemAmount_item_1319256->pos.x = (int)x + (int)dword_587000_183456;
	nox_gui_itemAmount_item_1319256->pos.y = (int)y + (int)dword_587000_183460;
	// The saved viewport is thirteen PE32 words, not a native-width record.
	nox_draw_viewport_t viewport = {
		.x1 = *getMemIntPtr(0x5D4594, 1319108),
		.y1 = *getMemIntPtr(0x5D4594, 1319112),
		.x2 = *getMemIntPtr(0x5D4594, 1319116),
		.y2 = *getMemIntPtr(0x5D4594, 1319120),
		.field_4 = *getMemIntPtr(0x5D4594, 1319124),
		.field_5 = *getMemIntPtr(0x5D4594, 1319128),
		.field_6 = *getMemIntPtr(0x5D4594, 1319132),
		.field_7 = *getMemIntPtr(0x5D4594, 1319136),
		.width = *getMemIntPtr(0x5D4594, 1319140),
		.height = *getMemIntPtr(0x5D4594, 1319144),
		.field_10 = *getMemU32Ptr(0x5D4594, 1319148),
		.field_11 = *getMemU32Ptr(0x5D4594, 1319152),
		.field_12 = *getMemIntPtr(0x5D4594, 1319156),
	};
	nox_gui_itemAmount_item_1319256->draw_func((uint32_t*)&viewport, nox_gui_itemAmount_item_1319256);
	if (nox_xxx_wndGetChildByID_46B0C0(nox_gui_itemAmount_dialog_1319228, 3603)->draw_data.field_0 & 4) {
		nox_client_drawImageAt_47D2C0(nox_gui_itemAmount_images_1319196[NOX_ITEM_AMOUNT_IMAGE_DOWN_LIT], x, y);
	}
	if (nox_xxx_wndGetChildByID_46B0C0(nox_gui_itemAmount_dialog_1319228, 3602)->draw_data.field_0 & 4) {
		nox_client_drawImageAt_47D2C0(nox_gui_itemAmount_images_1319196[NOX_ITEM_AMOUNT_IMAGE_UP_LIT], x, y);
	}
	if (nox_xxx_wndGetChildByID_46B0C0(nox_gui_itemAmount_dialog_1319228, 3604)->draw_data.field_0 & 4) {
		nox_video_bag_image_t* image =
			nox_gui_itemAmount_images_1319196[dword_5d4594_1319264 ? NOX_ITEM_AMOUNT_IMAGE_YES_PRESSED
																   : NOX_ITEM_AMOUNT_IMAGE_YES_PRESSED_NO_TAG];
		nox_client_drawImageAt_47D2C0(image, x, y);
	}
	if (nox_xxx_wndGetChildByID_46B0C0(nox_gui_itemAmount_dialog_1319228, 3605)->draw_data.field_0 & 4) {
		nox_video_bag_image_t* image =
			nox_gui_itemAmount_images_1319196[dword_5d4594_1319264 ? NOX_ITEM_AMOUNT_IMAGE_NO_PRESSED
																   : NOX_ITEM_AMOUNT_IMAGE_NO_PRESSED_NO_TAG];
		nox_client_drawImageAt_47D2C0(image, x, y);
	}
	return 1;
}

//----- (004C01C0) --------------------------------------------------------
#endif

//----- (004C12C0) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_4C12C0() { return dword_5d4594_1320964; }
#endif

#if 0 // Retained for the C player-name fixture.
//----- (004C1320) --------------------------------------------------------
int nox_xxx_netP2PStartTrade_4C1320(const uint8_t* data) {
	nox_playerInfo* v1; // esi
	nox_window* v3; // eax
	nox_window* v4; // eax
	int v5;       // [esp+4h] [ebp-8h]
	int v6;       // [esp+8h] [ebp-4h]

	v1 = (nox_playerInfo*)dword_8531A0_2576;
	if (!dword_8531A0_2576 || !data) {
		return 0;
	}
	if (dword_5d4594_1320964 == 1) {
		return 0;
	}
	dword_5d4594_1320964 = 1;
	sub_4C1410();
	nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1319844), (const wchar2_t*)(data + 2));
	nox_xxx_wnd_46ABB0(dword_5d4594_1320940, 1);
	nox_xxx_showP2PTradeWnd_4C12D0();
	nox_window_get_size(dword_5d4594_1320940, &v6, &v5);
	nox_window_setPos_46A9B0(dword_5d4594_1320940, 198, 193);
	v3 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3702);
	nox_window_call_field_94(v3, 16385, (uintptr_t)v1->name_final, 0);
	v4 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3703);
	nox_window_call_field_94(v4, 16385, (uintptr_t)getMemAt(0x5D4594, 1319844), 0);
	sub_467BB0();
	return 1;
}

#endif

#if 0 // Retained for the C player-name fixture.
//----- (004C1BF0) --------------------------------------------------------
int nox_xxx_prepareP2PTrade_4C1BF0() {
	int result;     // eax
	nox_playerInfo* v1; // esi
	nox_window* v2; // eax
	nox_window* v3; // eax

	result = dword_5d4594_1320964;
	v1 = (nox_playerInfo*)dword_8531A0_2576;
	if (dword_5d4594_1320964) {
		if (dword_8531A0_2576) {
			sub_4C1410();
			v2 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3702);
			nox_window_call_field_94(v2, 16385, (uintptr_t)v1->name_final, 0);
			v3 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3703);
			result = (int)nox_window_call_field_94(v3, 16385, (uintptr_t)getMemAt(0x5D4594, 1319844), 0);
		}
	}
	return result;
}

#endif

//----- (004C4770) --------------------------------------------------------
void nox_xxx_drawObject_4C4770_draw(nox_draw_viewport_t* vp, nox_drawable* dr, void* img) {
	if (!vp || !dr || !img) {
		return;
	}
	int same_team = 0;
	int observer_player = 0;
	if (!dword_5d4594_1321520) {
		dword_5d4594_1321520 = nox_xxx_getTTByNameSpriteMB_44CFC0("Ghost");
	}

	if (dr->flags28 & 4) {
		nox_playerInfo* info = (nox_playerInfo*)nox_common_playerInfoGetByID_417040(dr->field_32);
		if (nox_player_netCode_85319C == dr->field_32) {
			if (info && (info->field_3680 & 1)) {
				observer_player = 1;
			}
		} else if (info && (info->field_3680 & 1)) {
			return;
		}
		nox_object_team_t* local_team = nox_xxx_objGetTeamByNetCode_418C80(nox_player_netCode_85319C);
		nox_object_team_t* object_team = nox_xxx_objGetTeamByNetCode_418C80(dr->field_32);
		if (local_team && object_team &&
			(nox_player_netCode_85319C == dr->field_32 || nox_xxx_servCompareTeams_419150(local_team, object_team))) {
			same_team = 1;
		}
	}

	int world_dx = (int)dr->pos.x - (int)vp->field_4;
	int screen_x = world_dx + (int)vp->x1 - (uint8_t)dr->field_0;
	int world_y = (int)dr->pos.y - (int16_t)dr->z - ((uint8_t*)&dr->field_0)[1] - (int16_t)dr->field_26_1;
	int screen_y = (int)vp->y1 + world_y - (int)vp->field_5;

	if (*getMemU32Ptr(0x587000, 80808) && !(dr->flags30 & 0x40000000) && !(dr->flags30 & 1) &&
		(dr->flags28 & 0x80)) {
		int2 p1 = {(int)vp->x1 + world_dx, (int)vp->y1 + (int)dr->pos.y - (int)vp->field_5};
		int dir = dr->field_74_4;
		int2 p2 = {p1.field_0 + *getMemIntPtr(0x587000, 196184 + 8 * dir),
				   p1.field_4 + *getMemIntPtr(0x587000, 196188 + 8 * dir)};
		int mx = (p1.field_0 + p2.field_0) >> 1;
		int my = (p1.field_4 + p2.field_4) >> 1;
		if (dir < 0x18 && dir && (dir < 8 || dir > 0x10)) {
			if (sub_4C5630(mx - 5, mx - 5, my)) {
				p1.field_0 -= 2;
				p2.field_0 -= 2;
			} else if (sub_4C5630(mx + 5, mx + 5, my)) {
				p1.field_0 += 2;
				p2.field_0 += 2;
			}
		} else if (sub_4C5630(mx, mx, my - 5)) {
			p1.field_4 -= 2;
			p2.field_4 -= 2;
		} else if (sub_4C5630(mx, mx, my + 5)) {
			p1.field_4 += 2;
			p2.field_4 += 2;
		}
		int clipped = 0;
		int clip_x = nox_win_width;
		if (!sub_4C42A0(&p1, &p2, &clipped, &clip_x)) {
			return;
		}
	}

	uint8_t* light;
	if (((dr->flags28 & 0x80000) && (dr->flags30 & 0x1000000)) || (dr->flags30 & 0x40000000)) {
		light = getMemAt(0x587000, 185472);
	} else {
		nox_point pos = {(int)dr->pos.x, (int)dr->pos.y};
		light = (uint8_t*)sub_469920(&pos);
	}

	static nox_drawable* previous_drawable;
	static int previous_z;
	int z = (int16_t)dr->z;
	int restore_clip = z < 0 && previous_z < 0 && previous_drawable == dr;
	int draw_z = z < 0 ? -z : 0;
	previous_drawable = dr;
	previous_z = z;

	if (nox_client_drawable_testBuff_4356C0(dr, 25)) {
		nox_xxx_draw_434600(1);
		nox_draw_setColorMultAndIntensity_433E40(nox_color_blue_2650684);
	} else if ((dr->flags28 & 2) && (dr->flags30 & 0x40000000) && !(dr->flags30 & 0x8020)) {
		nox_xxx_draw_434600(1);
		nox_draw_setColorMultAndIntensity_433E40(nox_color_blue_2650684);
	} else {
		sub_4345F0(1);
		nox_draw_setColorMultAndIntensityRGB_433CD0(light[0], light[4], light[8]);
	}

	if (dr->field_120) {
		double alpha = 1.0 - (double)(uint32_t)(gameFrame() - dr->field_85) / (double)(int)gameFPS();
		if (alpha < 0.0) {
			alpha = 0.001;
		}
		uint8_t fade = (uint8_t)(alpha * 255.0);
		if (dr->field_27 == dword_5d4594_1321520) {
			uint8_t ghost = (uint8_t)sub_4C4EC0(vp, dr);
			if (ghost < fade) {
				fade = ghost;
			}
		} else if (dr->flags30 & 0x4000000) {
			fade >>= 1;
		}
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(fade);
	} else if (dr->field_27 == dword_5d4594_1321520) {
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580((uint8_t)sub_4C4EC0(vp, dr));
	} else if (dr->flags30 & 0x4000000) {
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(0x80u);
	}

	nox_drawable* local = getMemPtr(0x852978, 8);
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	if (nox_client_drawable_testBuff_4356C0(dr, 0) || observer_player ||
		(dr->field_27 == dword_5d4594_1321520 && local && nox_client_drawable_testBuff_4356C0(local, 21))) {
		if (player && (player->field_3680 & 1)) {
			nox_client_drawEnableAlpha_434560(1);
			nox_client_drawSetAlpha_434580(0x80u);
		} else {
			int dx = (int)dr->pos.x - (int32_t)dr->field_8;
			int dy = (int)dr->pos.y - (int32_t)dr->field_9;
			if (abs(dx) < 4 && abs(dy) < 4 && local && !nox_client_drawable_testBuff_4356C0(local, 21) && !same_team) {
				return;
			}
			int distance = 0;
			if (dx || dy) {
				int scale = (int32_t)dr->field_5 - (int32_t)dr->field_10;
				if (!scale) {
					scale = 1;
				}
				distance = nox_double2int(sqrt((double)(dx * dx + dy * dy))) / scale;
			}
			if (observer_player) {
				nox_client_drawEnableAlpha_434560(1);
				nox_client_drawSetAlpha_434580(0x80u);
			} else if (!same_team && local && nox_client_drawable_testBuff_4356C0(local, 21)) {
				nox_xxx_draw_434600(1);
				nox_draw_setColorMultAndIntensity_433E40(dword_8531A0_2572);
				nox_client_drawEnableAlpha_434560(1);
				nox_client_drawSetAlpha_434580(0xFFu);
				goto draw_enchant_overlays;
			} else {
				uint8_t invis_alpha = distance < 8 ? (uint8_t)((distance << 7) / 8) : 0x80u;
				if (!invis_alpha) {
					invis_alpha = 1;
				}
				if (same_team && invis_alpha <= 1 &&
					(((dr->flags28 & 2) && dr->field_69 == 8) || ((dr->flags28 & 4) && !dr->field_69))) {
					nox_xxx_draw_434600(1);
					nox_draw_setColorMultAndIntensity_433E40(dword_8531A0_2572);
					invis_alpha = 0x80u;
				}
				nox_client_drawEnableAlpha_434560(1);
				nox_client_drawSetAlpha_434580(invis_alpha);
			}
		}
	}

draw_enchant_overlays:
	if (!(dr->flags28 & 4) && nox_client_drawable_testBuff_4356C0(dr, 23) &&
		!nox_common_gameFlags_check_40A5C0(2048)) {
		nox_xxx_draw_434600(1);
		nox_draw_setColorMultAndIntensity_433E40(((uint8_t)gameFrame() & 1) ? nox_color_white_2523948 : nox_color_blue_2650684);
	}
	if (!nox_client_drawable_testBuff_4356C0(dr, 23) && nox_client_drawable_testBuff_4356C0(dr, 11)) {
		nox_xxx_draw_434600(1);
		nox_draw_setColorMultAndIntensity_433E40(*getMemIntPtr(0x85B3FC, 956));
	}
	if (restore_clip) {
		nox_xxx_wndDraw_49F7F0();
		sub_49F7C0_def();
	} else {
		sub_47D370(draw_z);
	}
	nox_client_drawImageAt_47D2C0(img, screen_x, screen_y);
	sub_4345F0(0);
	nox_client_drawEnableAlpha_434560(0);
	nox_xxx_draw_434600(0);
	if (restore_clip) {
		sub_49F860();
	}
	((uint16_t*)&dr->field_0)[1] = 1;
	((uint16_t*)&dr->field_1)[0] = *getMemU16Ptr(0x973F18, 88);
	((uint16_t*)&dr->field_1)[1] = *getMemU16Ptr(0x973F18, 76);
	dr->field_2 = img;
}

#if 0 // Superseded by the native-layout implementation above.
static void nox_xxx_drawObjectLegacy_4C4770_draw(nox_draw_viewport_t* vp, nox_drawable* dr, void* img) {
	int* a1 = vp;
	unsigned char* a2 = dr;
	int a3 = img;
	unsigned char* v3;  // esi
	char* v4;           // eax
	uint32_t* v5;       // edi
	uint32_t* v6;       // eax
	int v7;             // ecx
	int v8;             // edi
	int v9;             // ecx
	int v10;            // ebp
	int v11;            // edx
	int v12;            // eax
	int v13;            // ebx
	int v14;            // edx
	int v15;            // ebx
	int v16;            // edi
	int v17;            // ebx
	unsigned char v18;  // al
	unsigned char v19;  // cl
	int v20;            // ecx
	int v21;            // edx
	int v22;            // ecx
	int v23;            // edi
	int v24;            // ebx
	int v25;            // ebx
	int v26;            // edi
	int v27;            // ecx
	unsigned char* v28; // edi
	short v29;          // ax
	int v30;            // ebp
	int v31;            // edx
	int v32;            // eax
	int v33;            // eax
	int v34;            // edx
	int v35;            // ecx
	double v36;         // st7
	double v37;         // st7
	unsigned char v38;  // bl
	unsigned char v39;  // al
	int v40;            // edx
	int v41;            // ecx
	int v42;            // edi
	int v43;            // ebx
	int v44;            // eax
	int v45;            // eax
	int v46;            // eax
	int v47;            // ebp
	int v48;            // eax
	int v49;            // eax
	int v50;            // eax
	int v53;            // [esp+18h] [ebp-30h]
	int v54;            // [esp+1Ch] [ebp-2Ch]
	int2 a1a;           // [esp+20h] [ebp-28h]
	int2 a2a;           // [esp+28h] [ebp-20h]
	int a4;             // [esp+30h] [ebp-18h]
	int a3a;            // [esp+34h] [ebp-14h]
	int v59a;           // [esp+38h] [ebp-10h]
	int v59b;           // [esp+38h] [ebp-10h]
	long long v60;      // [esp+40h] [ebp-8h]
	unsigned char v61;  // [esp+4Ch] [ebp+4h]
	unsigned char v62;  // [esp+4Ch] [ebp+4h]
	unsigned char v63;  // [esp+50h] [ebp+8h]
	unsigned char v64;  // [esp+50h] [ebp+8h]

	v53 = 0;
	v54 = 0;
	if (!dword_5d4594_1321520) {
		dword_5d4594_1321520 = nox_xxx_getTTByNameSpriteMB_44CFC0("Ghost");
	}
	v3 = a2;
	if (a2[112] & 4) {
		v4 = nox_common_playerInfoGetByID_417040(*((uint32_t*)a2 + 32));
		if (nox_player_netCode_85319C == *((uint32_t*)a2 + 32)) {
			if (v4 && v4[3680] & 1) {
				v54 = 1;
			}
		} else if (v4 && v4[3680] & 1) {
			return;
		}
		v5 = nox_xxx_objGetTeamByNetCode_418C80(nox_player_netCode_85319C);
		if (v5) {
			v6 = nox_xxx_objGetTeamByNetCode_418C80(*((uint32_t*)a2 + 32));
			if (v6) {
				if (nox_player_netCode_85319C == *((uint32_t*)a2 + 32) ||
					nox_xxx_servCompareTeams_419150((int)v5, (int)v6)) {
					v53 = 1;
				}
			}
		}
	}
	v7 = *((uint32_t*)a2 + 3);
	a3a = 0;
	v8 = *a1;
	v9 = v7 - a1[4];
	v10 = a1[1];
	v11 = *((short*)a2 + 53);
	v59a = v9 + *a1 - *a2;
	v12 = *((uint32_t*)a2 + 4);
	v13 = *((uint32_t*)a2 + 4) - v11 - a2[1] - *((short*)a2 + 52);
	v14 = a1[5];
	v59b = v10 + v13 - v14;
	a4 = nox_win_width;
	if (*getMemU32Ptr(0x587000, 80808)) {
		v15 = *((uint32_t*)a2 + 30);
		if (!(v15 & 0x40000000) && !(v15 & 1) && (a2[112] & 0x80u) != 0) {
			v16 = v9 + v8;
			a1a.field_0 = v16;
			v17 = v10 + v12 - v14;
			v18 = a2[299];
			v19 = a2[299];
			a1a.field_4 = v17;
			v20 = 8 * v19;
			v21 = v16 + *getMemIntPtr(0x587000, 196184 + v20);
			v22 = v17 + *getMemIntPtr(0x587000, 196188 + v20);
			a2a.field_0 = v21;
			a2a.field_4 = v22;
			if (v18 < 0x18u && v18 && (v18 < 8u || v18 > 0x10u)) {
				v23 = (v21 + v16) >> 1;
				v24 = (v22 + v17) >> 1;
				if (sub_4C5630(v23 - 5, v23 - 5, v24)) {
					a1a.field_0 -= 2;
					a2a.field_0 -= 2;
				} else if (sub_4C5630(v23 + 5, v23 + 5, v24)) {
					a1a.field_0 += 2;
					a2a.field_0 += 2;
				}
			} else {
				v25 = (v22 + v17) >> 1;
				v26 = (v21 + v16) >> 1;
				if (sub_4C5630(v26, v26, v25 - 5)) {
					v27 = a2a.field_4 - 2;
					a1a.field_4 -= 2;
					a2a.field_4 = v27;
				} else if (sub_4C5630(v26, v26, v25 + 5)) {
					v27 = a2a.field_4 + 2;
					a1a.field_4 += 2;
					a2a.field_4 = v27;
				}
			}
			if (!sub_4C42A0(&a1a, &a2a, &a3a, &a4)) {
				return;
			}
		}
	}
	if ((*((uint32_t*)a2 + 28) & 0x80000) && (*((uint32_t*)a2 + 30) & 0x1000000) || (*((uint32_t*)a2 + 30) & 0x40000000)) {
		v28 = getMemAt(0x587000, 185472);
	} else {
		v28 = sub_469920((int2*)(a2 + 12));
	}
	v29 = *((uint16_t*)a2 + 52);
	int a2b = 0;
	if (v29 >= 0) {
		v30 = 0;
	} else {
		if (*getMemIntPtr(0x5D4594, 1321512) < 0 && a2 == *(unsigned char**)getMemAt(0x5D4594, 1321516)) {
			a2b = 1;
		}
		v30 = -v29;
	}
	*getMemU32Ptr(0x5D4594, 1321516) = a2;
	v31 = *((short*)a2 + 52);
	a1a.field_0 = v30;
	*getMemU32Ptr(0x5D4594, 1321512) = v31;
	if (nox_client_drawable_testBuff_4356C0((int)a2, 25)) {
		nox_xxx_draw_434600(1);
		nox_draw_setColorMultAndIntensity_433E40(nox_color_blue_2650684);
	} else if ((a2[112] & 2) && (v32 = *((uint32_t*)a2 + 30), v32 & 0x40000000) && !(v32 & 0x8020)) {
		nox_xxx_draw_434600(1);
		nox_draw_setColorMultAndIntensity_433E40(nox_color_blue_2650684);
	} else {
		sub_4345F0(1);
		LOBYTE(v34) = v28[8];
		v33 = v28[4];
		LOBYTE(v35) = *v28;
		nox_draw_setColorMultAndIntensityRGB_433CD0(v35, v33, v34);
	}
	if (*((uint32_t*)a2 + 120)) {
		v60 = (unsigned int)(gameFrame() - *((uint32_t*)a2 + 85));
		v36 = (double)v60;
		v60 = gameFPS();
		v37 = 1.0 - v36 / (double)(int)gameFPS();
		if (v37 < 0.0) {
			v37 = 0.001;
		}
		v38 = (long long)(v37 * 255.0);
		v63 = (long long)(v37 * 255.0);
		if (*((uint32_t*)v3 + 27) == dword_5d4594_1321520) {
			v39 = sub_4C4EC0(a1, (int)v3);
			if (v39 < v38) {
				v64 = v39;
				nox_client_drawEnableAlpha_434560(1);
				nox_client_drawSetAlpha_434580(v64);
			} else {
				nox_client_drawEnableAlpha_434560(1);
				nox_client_drawSetAlpha_434580(v63);
			}
		} else {
			if (*((uint32_t*)v3 + 30) & 0x4000000) {
				v63 = v38 >> 1;
			}
			nox_client_drawEnableAlpha_434560(1);
			nox_client_drawSetAlpha_434580(v63);
		}
	} else if (*((uint32_t*)a2 + 27) == dword_5d4594_1321520) {
		v61 = sub_4C4EC0(a1, (int)a2);
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(v61);
	} else if (*((uint32_t*)a2 + 30) & 0x4000000) {
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(0x80u);
	}
	if (!(!nox_client_drawable_testBuff_4356C0((int)v3, 0) && !v54 &&
		(*((uint32_t*)v3 + 27) != dword_5d4594_1321520 || !*getMemU32Ptr(0x852978, 8) ||
		 !nox_client_drawable_testBuff_4356C0(*getMemIntPtr(0x852978, 8), 21)))) {
		if (dword_8531A0_2576 && (*(uint8_t*)(dword_8531A0_2576 + 3680) & 1)) {
			nox_client_drawEnableAlpha_434560(1);
			nox_client_drawSetAlpha_434580(0x80u);
		} else {
			v40 = *((uint32_t*)v3 + 8);
			v41 = *((uint32_t*)v3 + 9);
			v42 = *((uint32_t*)v3 + 3) - v40;
			v43 = *((uint32_t*)v3 + 4) - v41;
			v44 = *((uint32_t*)v3 + 3) - v40;
			if (v42 < 0) {
				v44 = v40 - *((uint32_t*)v3 + 3);
			}
			if (v44 < 4) {
				v45 = *((uint32_t*)v3 + 4) - v41;
				if (v43 < 0) {
					v45 = v41 - *((uint32_t*)v3 + 4);
				}
				if (v45 < 4) {
					v46 = *getMemU32Ptr(0x852978, 8);
					if (!*getMemU32Ptr(0x852978, 8)) {
						goto LABEL_82;
					}
					if (!nox_client_drawable_testBuff_4356C0(*getMemIntPtr(0x852978, 8), 21)) {
						LOWORD(v4) = v53;
						if (!v53) {
							return;
						}
					}
				}
			}
			v46 = *getMemU32Ptr(0x852978, 8);
		LABEL_82:
			if (v42 || v43) {
				v47 = *((uint32_t*)v3 + 5) - *((uint32_t*)v3 + 10);
				if (!v47) {
					v47 = 1;
				}
				v48 = nox_double2int(sqrt((double)(v42 * v42 + v43 * v43))) / v47;
				v30 = a1a.field_0;
				v42 = v48;
				v46 = *getMemU32Ptr(0x852978, 8);
			}
			if (!v46) {
				goto LABEL_123;
			}
			if (v54) {
				LOBYTE(v49) = -128;
				v62 = v49;
				goto LABEL_B;
			}
			if (!(v53 || !nox_client_drawable_testBuff_4356C0(v46, 21))) {
				nox_xxx_draw_434600(1);
				nox_draw_setColorMultAndIntensity_433E40(dword_8531A0_2572);
				v62 = -1;
				nox_client_drawEnableAlpha_434560(1);
				nox_client_drawSetAlpha_434580(v62);
				goto LABEL_A;
			}
		LABEL_123:
			if (v42 < 8) {
				v49 = (v42 << 7) / 8;
				v62 = v49;
				if (!(uint8_t)v49) {
					LOBYTE(v49) = 1;
					v62 = v49;
				}
			} else {
				LOBYTE(v49) = -128;
				v62 = v49;
			}
		LABEL_B:
			if (v53 && (unsigned char)v49 <= 1u) {
				if ((v50 = *((uint32_t*)v3 + 28), v50 & 2) && (*((uint32_t*)v3 + 69) == 8) ||
					(v50 & 4) && !*((uint32_t*)v3 + 69)) {
					nox_xxx_draw_434600(1);
					nox_draw_setColorMultAndIntensity_433E40(dword_8531A0_2572);
					v62 = -128;
				}
			}
			nox_client_drawEnableAlpha_434560(1);
			nox_client_drawSetAlpha_434580(v62);
		}
	}
LABEL_A:
	if (!(v3[112] & 4) && nox_client_drawable_testBuff_4356C0((int)v3, 23) && !nox_common_gameFlags_check_40A5C0(2048)) {
		nox_xxx_draw_434600(1);
		if ((unsigned char)gameFrame() & 1) {
			nox_draw_setColorMultAndIntensity_433E40(nox_color_white_2523948);
		} else {
			nox_draw_setColorMultAndIntensity_433E40(nox_color_blue_2650684);
		}
	}
	if (!nox_client_drawable_testBuff_4356C0((int)v3, 23) && nox_client_drawable_testBuff_4356C0((int)v3, 11)) {
		nox_xxx_draw_434600(1);
		nox_draw_setColorMultAndIntensity_433E40(*getMemIntPtr(0x85B3FC, 956));
	}
	if (a2b) {
		nox_xxx_wndDraw_49F7F0();
		sub_49F7C0_def();
	} else {
		sub_47D370(v30);
	}
	nox_client_drawImageAt_47D2C0(a3, v59a, v59b);
	sub_4345F0(0);
	nox_client_drawEnableAlpha_434560(0);
	nox_xxx_draw_434600(0);
	if (a2b) {
		sub_49F860();
	}
	*((uint16_t*)v3 + 1) = 1;
	LOWORD(v4) = *getMemU16Ptr(0x973F18, 88);
	*((uint16_t*)v3 + 2) = *getMemU16Ptr(0x973F18, 88);
	*((uint16_t*)v3 + 3) = *getMemU16Ptr(0x973F18, 76);
	*((uint32_t*)v3 + 2) = a3;
}
#endif
// 4C4AD0: variable 'v35' is possibly undefined
// 4C4AD0: variable 'v34' is possibly undefined

//----- (004C5050) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
void sub_4C5050() { dword_5d4594_1321800 = 0; }
#endif

//----- (004CA540) --------------------------------------------------------
#if 0 // R228: original PE32 body; native registered entry is exported by Go.
int nox_xxx_sprite_4CA540(uint32_t* a1, int a2) {
	uint32_t* v2; // esi
	double v3;    // st7
	double v4;    // st6
	double v5;    // st5
	double v6;    // st4
	int v7;       // eax
	int v8;       // edi
	int v9;       // eax
	float v11;    // [esp+0h] [ebp-14h]
	float v12;    // [esp+0h] [ebp-14h]
	float v13;    // [esp+10h] [ebp-4h]
	float v14;    // [esp+1Ch] [ebp+8h]

	v2 = (uint32_t*)a2;
	v3 = 0.0;
	v4 = *(float*)(a2 + 468);
	v5 = *(float*)(a2 + 472);
	v6 = 0.0;
	v7 = gameFrame() - *(uint32_t*)(a2 + 316) + 1;
	do {
		--v7;
		v13 = -(v5 * *(float*)(a2 + 476));
		v4 = v4 - v4 * *(float*)(a2 + 476);
		v5 = v5 + v13;
		v3 = v3 + v4;
		v6 = v6 + v5;
	} while (v7);
	v14 = v6;
	v11 = (double)(int)v2[81] + v3;
	v8 = nox_float2int(v11);
	v12 = (double)(int)v2[82] + v14;
	v9 = nox_float2int(v12);
	if (v8 > 0 && v9 > 0 && v8 < 5888 && v9 < 5888) {
		nox_xxx_updateSpritePosition_49AA90(v2, v8, v9);
		if (sub_4992B0(*a1 + v2[3] - a1[4], v2[4] + a1[1] - a1[5])) {
			return 1;
		}
	}
	nox_xxx_spriteDeleteStatic_45A4E0_drawable((int)v2);
	return 0;
}

#endif // R228 native registered entry

//----- (004CA650) --------------------------------------------------------
#if 0
int sub_4CA650(int a1, int a2) {
	int v2;             // esi
	int v3;             // eax
	int v4;             // ebx
	int v5;             // edi
	int v6;             // eax
	int v7;             // ecx
	int v8;             // eax
	int v9;             // ebp
	int v10;            // eax
	int v11;            // edi
	unsigned short v12; // dx
	int v13;            // ebp
	int v14;            // eax
	int result;         // eax
	int v16;            // [esp+10h] [ebp-10h]
	int v17;            // [esp+28h] [ebp+8h]

	v2 = a2;
	v3 = *(uint32_t*)(a2 + 16);
	v4 = *(unsigned short*)(a2 + 434) - v3;
	v5 = *(unsigned short*)(a2 + 432) - *(uint32_t*)(a2 + 12);
	v6 = sub_48C6B0(v5, v4);
	v7 = v6;
	v8 = *(unsigned char*)(a2 + 443);
	v17 = v8;
	++v7;
	v9 = *(uint32_t*)(v2 + 12);
	v10 = v5 * v8 / v7;
	v11 = *(uint32_t*)(v2 + 16);
	v16 = v9 + v10;
	v12 = *(uint16_t*)(v2 + 432);
	v13 = v9 - v12;
	v14 = v11 + v4 * v17 / v7;
	if (v7 <= 10 ||
		v13 * (v16 - v12) + (v11 - *(unsigned short*)(v2 + 434)) * (v14 - *(unsigned short*)(v2 + 434)) < 0) {
		nox_xxx_spriteDeleteStatic_45A4E0_drawable(v2);
		result = 0;
	} else {
		nox_xxx_updateSpritePosition_49AA90((uint32_t*)v2, v16, v14);
		result = 1;
	}
	return result;
}
#endif
// 4CA67E: variable 'v6' is possibly undefined

//----- (004CA720) --------------------------------------------------------
#if 0 // R228: original PE32 body; native registered entry is exported by Go.
int sub_4CA720(int a1, int a2) {
	int v2;     // ebx
	int v3;     // ebp
	int v4;     // ecx
	int v5;     // edx
	int v6;     // ebx
	int v7;     // kr00_4
	int v8;     // ebx
	int result; // eax

	v2 = gameFrame() - *(uint32_t*)(a2 + 316);
	if (v2 >= 60 || (v3 = *(unsigned short*)(a2 + 432), abs(*(uint32_t*)(a2 + 12) - (unsigned short)v3) < 10) &&
						abs(*(uint32_t*)(a2 + 16) - *(unsigned short*)(a2 + 434)) < 10) {
		nox_xxx_spriteDeleteStatic_45A4E0_drawable(a2);
		result = 0;
	} else {
		v4 = *(short*)(a2 + 440) - v2 * *(short*)(a2 + 440) / 60;
		v5 = (v2 << 8) / 120;
		if (*(uint8_t*)(a2 + 443)) {
			v6 = v5 + *(unsigned char*)(a2 + 442);
		} else {
			v6 = *(unsigned char*)(a2 + 442) - v5;
		}
		if (v6 < 0) {
			v6 += (unsigned int)(255 - v6) >> 8 << 8;
		}
		if (v6 >= 256) {
			v6 += -256 * ((unsigned int)v6 >> 8);
		}
		v7 = v4 * *getMemIntPtr(0x587000, 192088 + 8 * v6);
		v8 = *(unsigned short*)(a2 + 434) + v4 * *getMemI32Ptr(0x587000, 192092 + 8 * v6) / 16;
		nox_xxx_updateSpritePosition_49AA90((uint32_t*)a2, v3 + v7 / 16, v8);
		*(uint32_t*)(a2 + 32) = v3 + v7 / 16;
		*(uint32_t*)(a2 + 36) = v8;
		result = 1;
	}
	return result;
}

#endif // R228 native registered entry

// Native-width implementation: client/gui.NewProgressBarRaw.
#if 0
//----- (004CAF10) --------------------------------------------------------
nox_window* nox_gui_newProgressBar_4CAF10(int a1, int a2, int a3, int a4, int a5, int a6, uint32_t* a7) {
	int v7;       // eax
	uint32_t* v8; // esi

	v7 = a7[2];
	if (!(v7 & 0x1000)) {
		return 0;
	}
	v8 = nox_window_new(a1, a2 & 0xFFFFFFF7, a3, a4, a5, a6, sub_4CAF80);
	sub_4CAFB0((int)v8);
	if (v8) {
		if (!a7[4]) {
			a7[4] = v8;
		}
		nox_gui_windowCopyDrawData_46AF80((int)v8, a7);
	}
	return v8;
}

//----- (004CAF80) --------------------------------------------------------
int sub_4CAF80(int a1, int a2, int a3, int a4) {
	if (a2 == 16416 && a3 >= 0 && a3 <= 100) {
		*(uint32_t*)(a1 + 32) = a3;
	}
	return 0;
}

//----- (004CAFB0) --------------------------------------------------------
int sub_4CAFB0(int a1) {
	int result; // eax

	result = a1;
	if (a1) {
		if ((signed char)*(uint8_t*)(a1 + 4) >= 0) {
			result = nox_window_set_all_funcs((uint32_t*)a1, 0, sub_4CAFF0, 0);
		} else {
			result = nox_window_set_all_funcs((uint32_t*)a1, 0, sub_4CB1A0, 0);
		}
	}
	return result;
}

//----- (004CAFF0) --------------------------------------------------------
int sub_4CAFF0(uint32_t* a1, uint32_t* a2) {
	int v2;          // ecx
	int v3;          // ebp
	int v4;          // edi
	int v5;          // edi
	int v6;          // edi
	int v7;          // ebp
	int xLeft;       // [esp+10h] [ebp-98h]
	int yTop;        // [esp+14h] [ebp-94h]
	int v11;         // [esp+18h] [ebp-90h]
	int v12;         // [esp+1Ch] [ebp-8Ch]
	int v13;         // [esp+20h] [ebp-88h]
	int v14;         // [esp+24h] [ebp-84h]
	wchar2_t v15[64]; // [esp+28h] [ebp-80h]

	v2 = a2[17];
	v3 = a2[9];
	v14 = a2[7];
	v4 = a2[5];
	v11 = v2;
	nox_client_wndGetPosition_46AA60(a1, &xLeft, &yTop);
	if (v4 != 0x80000000) {
		nox_client_drawSetColor_434460(v4);
		nox_client_drawRectFilledOpaque_49CE30(xLeft, yTop, a1[2], a1[3]);
	}
	if (v3 != 0x80000000) {
		v5 = a1[2] * a1[8] / 100;
		nox_client_drawSetColor_434460(v3);
		nox_client_drawRectFilledOpaque_49CE30(xLeft, yTop, v5, a1[3]);
	}
	if (v11 != 0x80000000) {
		if ((a1[1] & 0x2000) == 0x2000) {
			nox_draw_enableTextSmoothing_43F670(1);
		}
		nox_swprintf(v15, L"%i%%", a1[8]);
		nox_xxx_drawGetStringSize_43F840(a2[50], v15, &v13, &v12, 64);
		v6 = xLeft + a1[2] / 2 - v13 / 2;
		v7 = a1[3] / 2 - v12 / 2 + yTop + 1;
		nox_xxx_drawSetTextColor_434390(v11);
		nox_xxx_drawStringWrap_43FAF0(a2[50], v15, v6, v7, a1[2], 0);
		nox_draw_enableTextSmoothing_43F670(0);
	}
	if (v14 != 0x80000000) {
		nox_client_drawSetColor_434460(v14);
		nox_client_drawBorderLines_49CC70(xLeft, yTop, a1[2], a1[3]);
	}
	return 1;
}

//----- (004CB1A0) --------------------------------------------------------
int sub_4CB1A0(uint32_t* a1, int a2) {
	int v2;       // ebx
	uint32_t* v3; // esi
	int v4;       // edi
	int xLeft;    // [esp+Ch] [ebp-8h]
	int yTop;     // [esp+10h] [ebp-4h]

	v2 = *(uint32_t*)(a2 + 24);
	v3 = a1;
	nox_client_wndGetPosition_46AA60(a1, &xLeft, &yTop);
	nox_xxx_wndDraw_49F7F0();
	v4 = v3[2] * v3[8] / 100;
	nox_client_copyRect_49F6F0(xLeft, yTop, v4, v3[3]);
	if (v2 && v4 > 0) {
		nox_client_drawImageAt_47D2C0(v2, xLeft, yTop);
	}
	sub_49F860();
	return 1;
}
#endif

//----- (004CCD00) --------------------------------------------------------
#if 0 // R228: original PE32 bodies; native registered entries are exported by Go.
int sub_4CCD00(int a1, int a2) {
	unsigned int i; // ecx

	for (i = *(uint32_t*)(a2 + 432); i < gameFrame(); ++i) {
		if (*(float*)(a2 + 436) > 0.0) {
			*(float*)(a2 + 436) = *(float*)(a2 + 440) + *(float*)(a2 + 436);
			*(float*)(a2 + 440) = *(float*)(a2 + 440) - 1.0;
		}
		if (*(float*)(a2 + 436) <= 0.0) {
			*(uint32_t*)(a2 + 436) = 0;
			*(uint32_t*)(a2 + 440) = 0;
		}
	}
	*(uint16_t*)(a2 + 104) = (long long)*(float*)(a2 + 436);
	*(uint8_t*)(a2 + 296) = (long long)*(float*)(a2 + 440);
	*(uint32_t*)(a2 + 432) = gameFrame();
	return 1;
}

//----- (004CCDB0) --------------------------------------------------------
int nox_xxx_updDrawFist_4CCDB0(int a1, int a2) {
	unsigned int i; // edx
	double v3;      // st7
	double v4;      // st7
	double v5;      // st7

	for (i = *(uint32_t*)(a2 + 432); i < gameFrame(); ++i) {
		v3 = *(float*)(a2 + 436) + *(float*)(a2 + 440);
		*(float*)(a2 + 436) = v3;
		if (v3 >= 0.0) {
			*(float*)(a2 + 440) = *(float*)(a2 + 440) - 0.5;
		} else {
			v4 = -*(float*)(a2 + 440) * *(float*)(a2 + 444);
			*(uint32_t*)(a2 + 436) = 0;
			v5 = v4 * 0.1;
			*(float*)(a2 + 440) = v5;
			if (v5 < 2.0) {
				*(uint32_t*)(a2 + 436) = 0;
				*(uint32_t*)(a2 + 440) = 0;
			}
		}
	}
	*(uint16_t*)(a2 + 104) = (long long)*(float*)(a2 + 436);
	*(uint8_t*)(a2 + 296) = (long long)*(float*)(a2 + 440);
	*(uint32_t*)(a2 + 432) = gameFrame();
	return 1;
}

#endif // R228 native registered entry

//----- (004CCE70) --------------------------------------------------------
#if 0 // Original PE32 bodies; registered symbols are exported by Go.
int sub_4CCE70(int a1, uint32_t* a2) {
	if (!a2[120] && !nox_xxx_checkGameFlagPause_413A50()) {
		sub_4CCEA0(a2, 5);
	}
	return 1;
}

//----- (004CD090) --------------------------------------------------------
int sub_4CD090(int a1, uint32_t* a2) {
	if (!a2[120] && !nox_xxx_checkGameFlagPause_413A50()) {
		sub_4CCEA0(a2, 4);
	}
	return 1;
}

//----- (004CD0C0) --------------------------------------------------------
int sub_4CD0C0(int a1, uint32_t* a2) {
	if (!a2[120] && !nox_xxx_checkGameFlagPause_413A50()) {
		sub_4CCEA0(a2, 3);
	}
	return 1;
}

//----- (004CD0F0) --------------------------------------------------------
int sub_4CD0F0(int a1, uint32_t* a2) {
	if (!a2[120] && !nox_xxx_checkGameFlagPause_413A50()) {
		sub_4CCEA0(a2, 2);
	}
	return 1;
}

//----- (004CD120) --------------------------------------------------------
int sub_4CD120(int a1, uint32_t* a2) {
	if (!a2[120] && !nox_xxx_checkGameFlagPause_413A50()) {
		sub_4CCEA0(a2, 1);
	}
	return 1;
}
#endif

// 4CD249: variable 'v11' is possibly undefined
// 4CD25A: variable 'v13' is possibly undefined
// 4CD27E: variable 'v15' is possibly undefined
// 4CD28F: variable 'v16' is possibly undefined

//----- (004CD400) --------------------------------------------------------
#if 0 // Original PE32 body; the registered symbol is exported by Go.
int sub_4CD400(uint32_t* a1, int a2) {
	int v2; // eax

	v2 = dword_5d4594_1522968;
	if (!dword_5d4594_1522968) {
		v2 = nox_xxx_getTTByNameSpriteMB_44CFC0("CharmOrb");
		dword_5d4594_1522968 = v2;
	}
	sub_4CD150(v2, a1, a2, 1);
	sub_4CD150(*(int*)&dword_5d4594_1522968, a1, a2, 0);
	return 1;
}
#endif

// 4CD516: variable 'v5' is possibly undefined
// 4CD527: variable 'v7' is possibly undefined
// 4CD54B: variable 'v9' is possibly undefined
// 4CD55C: variable 'v11' is possibly undefined
