#include "client__gui__guimsg.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_3.h"
#include "GAME5_2.h"
#include "client__gui__window.h"
#include "common__magic__speltree.h"
#include "common__strman.h"
#include "operators.h"

//----- (004D7F90) --------------------------------------------------------
int nox_xxx_netSendSpellAward_4D7F90(int a1, int a2, char a3, int a4) {
	int result; // eax
	int v5;     // eax

	result = a1;
	if (*(uint8_t*)(a1 + 8) & 4) {
		v5 = *(uint32_t*)(a1 + 748);
		LOBYTE(a1) = 111;
		BYTE1(a1) = a2;
		BYTE2(a1) = *(uint8_t*)(*(uint32_t*)(v5 + 276) + 4 * a2 + 3696);
		HIBYTE(a1) = a3;
		if (a4) {
			HIBYTE(a1) = a3 | 0x80;
		}
		result = nox_xxx_netSendPacket1_4E5390(*(unsigned char*)(*(uint32_t*)(v5 + 276) + 2064), &a1, 4, 0, 1);
	}
	return result;
}

//----- (004FB0B0) --------------------------------------------------------
void nox_xxx_abilGetError_4FB0B0_magic_plyrspel(int a1) {
	wchar2_t* v1; // eax

	v1 = nox_strman_loadString_40F1D0(*(char**)getMemAt(0x587000, 216380 + 4 * a1), 0,
									  "C:\\NoxPost\\src\\Server\\Magic\\plyrspel.c", 86);
	nox_xxx_printCentered_445490(v1);
}
