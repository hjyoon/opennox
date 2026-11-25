#include "client__draw__flagdraw.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME3_1.h"
#include "client__draw__weapondraw.h"

extern uint32_t dword_8531A0_2572;

//----- (004B9500) --------------------------------------------------------
int nox_thing_flag_draw(int* a1, nox_drawable* dr) {
	nox_draw_viewport_t* vp = (nox_draw_viewport_t*)a1;
	nox_thing_weapon_animate_draw(a1, dr);
	if (nox_common_gameFlags_check_40A5C0(128) && (dr->flags30 & 0x1000000)) {
		int x = (int)vp->x1 + (int)dr->pos.x - (int)vp->field_4;
		int y = (int)dr->pos.y + (int)vp->y1 - (int)(int16_t)dr->z -
				nox_float2int(dr->field_25) - (int)vp->field_5;
		nox_team_t* team = sub_418A80(sub_4B94E0(dr));
		if (team) {
			int text_width = 0;
			nox_xxx_drawSetTextColor_434390(dword_8531A0_2572);
			nox_xxx_drawGetStringSize_43F840(0, team->name, &text_width, 0, 0);
			nox_xxx_drawString_43F6E0(0, team->name, text_width / -2 + x, y);
		}
	}
	return 1;
}
