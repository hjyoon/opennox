#include "client__draw__glyphdraw.h"
#include "GAME1.h"
#include "GAME1_2.h"
#include "client__draw__animdraw.h"
#include "client__video__draw_common.h"
#include "operators.h"

extern uint32_t dword_8531A0_2572;

//----- (004B9C70) --------------------------------------------------------
int nox_thing_glyph_draw(int* a1, nox_drawable* dr) {
	char v3; // cl
	int64_t v4;  // ecx
	int64_t v5;  // eax
	int64_t v6;  // eax
	int v7;  // esi

	nox_drawable* local_player = getMemPtr(0x852978, 8);
	uintptr_t alpha = (uintptr_t)dr;

	if (!nox_common_gameFlags_check_40A5C0(2) || !local_player) {
		goto LABEL_10;
	}
	if (dr->flags30 & 0x40000000) {
		LOBYTE(alpha) = -1;
		goto LABEL_10;
	}
	if (nox_client_drawable_testBuff_4356C0(local_player, 21)) {
		nox_xxx_draw_434600(1);
		nox_draw_setColorMultAndIntensity_433E40(dword_8531A0_2572);
		v3 = -1;
		LOBYTE(alpha) = v3;
		goto LABEL_10;
	}
	v4 = (int64_t)dr->pos.x - (int64_t)local_player->pos.x;
	v5 = (int64_t)dr->pos.y - (int64_t)local_player->pos.y;
	v6 = v4 * v4 + v5 * v5;
	if (v6 >= 22500) {
		return 1;
	}
	v3 = -56 - 200 * v6 / 22500;
	LOBYTE(alpha) = v3;
LABEL_10:
	nox_client_drawEnableAlpha_434560(1);
	nox_client_drawSetAlpha_434580((unsigned char)alpha);
	v7 = nox_thing_animate_draw(a1, dr);
	nox_client_drawEnableAlpha_434560(0);
	nox_xxx_draw_434600(0);
	return v7;
}
