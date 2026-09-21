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

// GAME.EXE 004FB0B0 is restored by spell_result_4fb0b0.go. The original
// indexed a packed PE32 char-pointer table, which cannot be dereferenced as a
// native char** on 64-bit hosts.
