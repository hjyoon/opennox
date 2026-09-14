#include <string.h>

#include "client__draw__harpoondraw.h"
#include "GAME2.h"
#include "GAME3.h"
#include "GAME5_2.h"
#include "client__draw__slavedraw.h"
#include "client__video__draw_common.h"

//----- (004B64A0) --------------------------------------------------------
int nox_thing_harpoon_draw(int* a1, nox_drawable* dr) { return nox_thing_slave_draw(a1, dr); }

//----- (004B61F0) --------------------------------------------------------
int nox_thing_harpoon_rope_draw(int* a1, nox_drawable* dr) {
	nox_draw_viewport_t* vp = (nox_draw_viewport_t*)a1;
	const uint8_t* ray = (const uint8_t*)&dr->union_u32[0];
	uint32_t sourceCode, targetCode;
	memcpy(&sourceCode, ray + 5, sizeof(sourceCode));
	memcpy(&targetCode, ray + 9, sizeof(targetCode));
	int2 a1a, a2a;
	if (!ray[0]) {
		a1a.field_0 = (int)vp->x1 + (uint16_t)sourceCode - (int)vp->field_4;
		a1a.field_4 = (int)vp->y1 + (uint16_t)(sourceCode >> 16) - (int)vp->field_5 - 20;
		a2a.field_0 = (int)vp->x1 + (uint16_t)targetCode - (int)vp->field_4;
		a2a.field_4 = (int)vp->y1 + (uint16_t)(targetCode >> 16) - (int)vp->field_5 - 20;
	} else {
		uint16_t fromCode = (uint16_t)sourceCode;
		uint16_t toCode = (uint16_t)targetCode;
		nox_drawable* from = nox_xxx_netTestHighBit_578B70(fromCode)
			? nox_xxx_netSpriteByCodeStatic_45A720(fromCode & 0x7fff)
			: nox_xxx_netSpriteByCodeDynamic_45A6F0(fromCode);
		nox_drawable* to = nox_xxx_netTestHighBit_578B70(toCode)
			? nox_xxx_netSpriteByCodeStatic_45A720(toCode & 0x7fff)
			: nox_xxx_netSpriteByCodeDynamic_45A6F0(toCode);
		if (!from || !to) {
			return 1;
		}
		a1a.field_0 = (int)vp->x1 + (int)from->pos.x - (int)vp->field_4;
		a1a.field_4 = (int)vp->y1 + (int)from->pos.y - (int)vp->field_5;
		a2a.field_0 = (int)vp->x1 + (int)to->pos.x - (int)vp->field_4;
		a2a.field_4 = (int)vp->y1 + (int)to->pos.y - (int)vp->field_5 - (int16_t)to->field_26_1 - (int16_t)to->z - 8;
		a1a.field_0 += *getMemU32Ptr(0x587000, 175864 + 8 * from->field_74_2);
		a1a.field_4 += *getMemU32Ptr(0x587000, 175868 + 8 * from->field_74_2);
	}
	*getMemU32Ptr(0x5D4594, 1312492) = nox_color_rgb_4344A0(144, 104, 64);
	*getMemU32Ptr(0x5D4594, 1312496) = nox_color_rgb_4344A0(24, 16, 0);
	sub_4B63B0(&a1a, &a2a);
	return 1;
}
