#include <math.h>
#include <string.h>

#include "client__draw__lightning.h"
#include "common__random.h"

#include "GAME1.h"
#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME2_3.h"
#include "GAME5_2.h"
#include "client__draw__fx.h"
#include "client__gui__window.h"
#include "client__video__draw_common.h"
#include "operators.h"

extern uint32_t dword_5d4594_1316452;
extern uint32_t dword_5d4594_1316448;
extern uint32_t dword_5d4594_1316456;
extern uint32_t nox_xxx_lightningSteps_587000_178216;
extern uint32_t dword_5d4594_1316484;
extern uint32_t dword_5d4594_1316472;
extern uint32_t dword_5d4594_1316436;
extern uint32_t dword_5d4594_1316476;
extern uint32_t dword_5d4594_1316492;

static uint32_t nox_lightningPackPoint(int2 point) {
	return (uint32_t)(uint16_t)point.field_0 | ((uint32_t)(uint16_t)point.field_4 << 16);
}

//----- (004BB070) --------------------------------------------------------
void nox_client_drawResetPoints_49F5A0();
int nox_xxx_drawLightningStep_4BB070(uint32_t a1, uint32_t a2) {
	int v2;           // eax
	bool v3;          // zf
	int v4;           // esi
	int v5;           // edi
	int v6;           // ebx
	int v7;           // ebp
	unsigned char v8; // al
	int v9;           // ecx
	int v10;          // eax
	int v11;          // ecx
	int v12;          // eax
	int v14;          // edi
	int v15;          // esi
	int v16;          // eax
	int v17;          // [esp+10h] [ebp-14h]
	int v18;          // [esp+10h] [ebp-14h]
	int v19;          // [esp+14h] [ebp-10h]
	int v20;          // [esp+14h] [ebp-10h]
	int v21;          // [esp+18h] [ebp-Ch]
	int v22;          // [esp+18h] [ebp-Ch]
	int v23;          // [esp+20h] [ebp-4h]
	int v24;          // [esp+20h] [ebp-4h]
	int v25;          // [esp+28h] [ebp+4h]
	int v26;          // [esp+28h] [ebp+4h]
	int v27;          // [esp+2Ch] [ebp+8h]
	int v28;          // [esp+2Ch] [ebp+8h]

	v2 = dword_5d4594_1316492 + 1;
	v3 = dword_5d4594_1316492 + 1 == dword_5d4594_1316448;
	v23 = 0;
	++dword_5d4594_1316492;
	if (!v3) {
		v14 = 1;
		v15 = 1;
		if (v2 > 1) {
			v16 = v2 - 1;
			do {
				v14 *= *getMemU32Ptr(0x587000, 178212) + 9;
				v15 *= 10;
				--v16;
			} while (v16);
		}
		LOWORD(v24) = v15 *
						  nox_common_randomIntMinMax_415FF0(-dword_5d4594_1316476, *(int*)&dword_5d4594_1316476,
															"C:\\NoxPost\\src\\Client\\Draw\\lightning.c", 193) /
						  v14 +
					  (((short)a1 + (short)a2) >> 1);
		HIWORD(v24) = v15 *
						  nox_common_randomIntMinMax_415FF0(-dword_5d4594_1316476, *(int*)&dword_5d4594_1316476,
															"C:\\NoxPost\\src\\Client\\Draw\\lightning.c", 196) /
						  v14 +
					  ((SHIWORD(a1) + SHIWORD(a2)) >> 1);
		nox_xxx_drawLightningStep_4BB070(a1, v24);
		nox_xxx_drawLightningStep_4BB070(v24, a2);
		return --dword_5d4594_1316492;
	}
	if (*getMemU32Ptr(0x5D4594, 1316508)) {
		v23 = getMemByte(0x5D4594, 1316420) + 48;
		nox_draw_set54RGB32_434040(*getMemIntPtr(0x5D4594, 1316440));
		sub_434080(12);
		v4 = SHIWORD(a1);
		v5 = (short)a1;
		nox_client_drawAddPoint_49F500((short)a1, SHIWORD(a1));
		v6 = SHIWORD(a2);
		v7 = (short)a2;
		nox_client_drawAddPoint_49F500((short)a2, SHIWORD(a2));
		sub_49E4F0(v23);
		v23 = 1;
	} else {
		nox_draw_set54RGB32_434040(*(int*)&dword_5d4594_1316472);
		sub_434080(3);
		v4 = SHIWORD(a1);
		v5 = (short)a1;
		nox_client_drawAddPoint_49F500((short)a1, SHIWORD(a1));
		v6 = SHIWORD(a2);
		v7 = (short)a2;
		nox_client_drawAddPoint_49F500((short)a2, SHIWORD(a2));
		sub_49E4F0(32);
	}
	nox_client_drawSetColor_434460(*(int*)&dword_5d4594_1316472);
	nox_client_drawAddPoint_49F500(v5, v4);
	nox_client_drawAddPoint_49F500(v7, v6);
	nox_client_drawLineFromPoints_49E4B0();
	if (!v23) {
		return --dword_5d4594_1316492;
	}
	v8 = getMemByte(0x5D4594, 1316420);
	v25 = 1;
	if ((getMemByte(0x5D4594, 1316420) & 0xFEu) > 2) {
		v19 = v7 + 1;
		v21 = v5 + 1;
		v17 = v4 + 1;
		v27 = v6 + 1;
		do {
			nox_client_drawResetPoints_49F5A0();
			v9 = v5 - v7;
			if (v5 - v7 < 0) {
				v9 = v7 - v5;
			}
			v10 = v4 - v6;
			if (v4 - v6 < 0) {
				v10 = v6 - v4;
			}
			if (v9 <= v10) {
				nox_client_drawAddPoint_49F500(v19, v6);
				nox_client_drawAddPoint_49F500(v21, v4);
			} else {
				nox_client_drawAddPoint_49F500(v7, v27);
				nox_client_drawAddPoint_49F500(v5, v17);
			}
			nox_client_drawLineFromPoints_49E4B0();
			++v27;
			++v17;
			++v21;
			v8 = getMemByte(0x5D4594, 1316420);
			++v19;
			++v25;
		} while (v25 < (getMemByte(0x5D4594, 1316420) >> 1));
	}
	v26 = 1;
	if ((v8 & 0xFEu) <= 2) {
		return --dword_5d4594_1316492;
	}
	v20 = v7 - 1;
	v18 = v5 - 1;
	v22 = v4 - 1;
	v28 = v6 - 1;
	do {
		nox_client_drawResetPoints_49F5A0();
		v11 = v5 - v7;
		if (v5 - v7 < 0) {
			v11 = v7 - v5;
		}
		v12 = v4 - v6;
		if (v4 - v6 < 0) {
			v12 = v6 - v4;
		}
		if (v11 <= v12) {
			nox_client_drawAddPoint_49F500(v20, v6);
			nox_client_drawAddPoint_49F500(v18, v4);
		} else {
			nox_client_drawAddPoint_49F500(v7, v28);
			nox_client_drawAddPoint_49F500(v5, v22);
		}
		nox_client_drawLineFromPoints_49E4B0();
		--v28;
		--v22;
		--v18;
		++v26;
		--v20;
	} while (v26 < (getMemByte(0x5D4594, 1316420) >> 1));
	return --dword_5d4594_1316492;
}

//----- (004BAE60) --------------------------------------------------------
int nox_xxx_lightningProc2_4BAE60(int2* a1, int2* a2, int a3, short* a4, int a5, int a6, int a7) {
	int2* v7;      // ebx
	int2* v8;      // ebp
	int v9;        // esi
	int v10;       // eax
	int v11;       // eax
	long long v12; // rax
	int v13;       // eax
	int v14;       // edi
	int v15;       // ebx
	int v16;       // eax
	int v17;       // edx
	int v18;       // esi
	int v19;       // ecx
	int result;    // eax

	v7 = a2;
	v8 = a1;
	if (a1->field_0 - a2->field_0 >= 0) {
		v9 = a1->field_0 - a2->field_0;
	} else {
		v9 = a2->field_0 - a1->field_0;
	}
	v10 = a2->field_4;
	if (a1->field_4 - v10 >= 0) {
		v11 = a1->field_4 - v10;
	} else {
		v11 = v10 - a1->field_4;
	}
	v12 = (long long)sqrt((double)(v9 * v9 + v11 * v11));
	if ((int)v12 >= 512) {
		dword_5d4594_1316476 = *getMemU32Ptr(0x587000, 178204);
		dword_5d4594_1316448 = nox_xxx_lightningSteps_587000_178216;
	} else {
		*(uint32_t*)&dword_5d4594_1316476 =
			*getMemU32Ptr(0x587000, 178208) +
			(int)v12 * (*getMemU32Ptr(0x587000, 178204) - *getMemU32Ptr(0x587000, 178208)) / 512;
		bool v13p = 0;
		if ((int)v12 < 64) {
			v13 = *(uint32_t*)&nox_xxx_lightningSteps_587000_178216 - 3;
		} else if ((int)v12 < 128) {
			v13 = *(uint32_t*)&nox_xxx_lightningSteps_587000_178216 - 2;
		} else if ((int)v12 < 256) {
			v13 = *(uint32_t*)&nox_xxx_lightningSteps_587000_178216 - 1;
		} else {
			v13 = *(uint32_t*)&nox_xxx_lightningSteps_587000_178216;
			v13p = 1;
		}
		if (v13 >= 1 || v13p) {
			*(uint32_t*)&dword_5d4594_1316448 = v13;
		} else {
			*(uint32_t*)&dword_5d4594_1316448 = 1;
		}
	}
	*getMemU32Ptr(0x5D4594, 1316532) = a3;
	if (a3 == 1 || a3 == 3) {
		*getMemU16Ptr(0x5D4594, 1316432) = *a4;
		*getMemU16Ptr(0x5D4594, 1316434) = a4[2];
		v14 = a1->field_0;
		v15 = a2->field_0;
		v16 = a1->field_0 - a2->field_0;
		if (v16 < 0) {
			v16 = v15 - v14;
		}
		v17 = a1->field_4;
		v18 = a2->field_4;
		v19 = a1->field_4 - v18;
		if (v19 < 0) {
			v19 = v18 - v17;
		}
		if (v16 <= v19) {
			*getMemU32Ptr(0x5D4594, 1316500) = (v17 >= v18) + 2;
		} else {
			*getMemU32Ptr(0x5D4594, 1316500) = v14 >= v15;
		}
		v7 = a2;
	}
	// The PE32 routine stored each endpoint in one 32-bit stack slot. Using
	// pointer-typed decompiler temporaries here corrupts the packed value on
	// native-width ABIs, so preserve the original two signed 16-bit halves
	// explicitly.
	uint32_t packedStart = nox_lightningPackPoint(*v8);
	uint32_t packedEnd = nox_lightningPackPoint(*v7);
	if (a6) {
		dword_5d4594_1316492 = 1;
		dword_5d4594_1316472 = dword_5d4594_1316456;
		*getMemU32Ptr(0x5D4594, 1316508) = 0;
		nox_xxx_drawLightningStep_4BB070(packedStart, packedEnd);
		dword_5d4594_1316492 = 1;
		dword_5d4594_1316472 = dword_5d4594_1316452;
		nox_xxx_drawLightningStep_4BB070(packedStart, packedEnd);
	}
	result = a7;
	if (a7) {
		dword_5d4594_1316492 = 1;
		dword_5d4594_1316472 = dword_5d4594_1316436;
		*getMemU32Ptr(0x5D4594, 1316440) = dword_5d4594_1316484;
		*getMemU32Ptr(0x5D4594, 1316508) = 1;
		result = nox_xxx_drawLightningStep_4BB070(packedStart, packedEnd);
	}
	return result;
}

// A ray's PE32 wire payload is unaligned, but the union and the referenced
// drawables have native-width layouts. Resolve both endpoint forms centrally.
static bool nox_lightningRayEndpoints(nox_drawable* dr, int2* fromPos, int2* toPos) {
	const uint8_t* ray = (const uint8_t*)&dr->union_u32[0];
	uint32_t sourceCode, targetCode;
	memcpy(&sourceCode, ray + 5, sizeof(sourceCode));
	memcpy(&targetCode, ray + 9, sizeof(targetCode));
	if (!ray[0]) {
		fromPos->field_0 = (uint16_t)sourceCode;
		fromPos->field_4 = (uint16_t)(sourceCode >> 16);
		toPos->field_0 = (uint16_t)targetCode;
		toPos->field_4 = (uint16_t)(targetCode >> 16);
		return true;
	}
	uint16_t fromCode = (uint16_t)sourceCode;
	uint16_t toCode = (uint16_t)targetCode;
	nox_drawable* from = nox_xxx_netTestHighBit_578B70(fromCode)
		? nox_xxx_netSpriteByCodeStatic_45A720(fromCode & 0x7fff)
		: nox_xxx_netSpriteByCodeDynamic_45A6F0(fromCode);
	nox_drawable* to = nox_xxx_netTestHighBit_578B70(toCode)
		? nox_xxx_netSpriteByCodeStatic_45A720(toCode & 0x7fff)
		: nox_xxx_netSpriteByCodeDynamic_45A6F0(toCode);
	if (!from || !to) {
		return false;
	}
	fromPos->field_0 = (int)from->pos.x;
	fromPos->field_4 = (int)from->pos.y;
	toPos->field_0 = (int)to->pos.x;
	toPos->field_4 = (int)to->pos.y;
	return true;
}

static void nox_lightningRayScreen(int* view, int2 fromPos, int2 toPos, int2* fromScreen, int2* toScreen) {
	nox_draw_viewport_t* vp = (nox_draw_viewport_t*)view;
	fromScreen->field_0 = (int)vp->x1 + fromPos.field_0 - (int)vp->field_4;
	fromScreen->field_4 = (int)vp->y1 + fromPos.field_4 - (int)vp->field_5 - 20;
	toScreen->field_0 = (int)vp->x1 + toPos.field_0 - (int)vp->field_4;
	toScreen->field_4 = (int)vp->y1 + toPos.field_4 - (int)vp->field_5 - 20;
}

//----- (004BAC80) --------------------------------------------------------
int nox_thing_lightning_draw(int* a1, nox_drawable* dr) {
	int2 a1a, a2a, fromPos, toPos;
	if (!nox_lightningRayEndpoints(dr, &fromPos, &toPos)) {
		return 1;
	}
	nox_lightningRayScreen(a1, fromPos, toPos, &a1a, &a2a);
	dword_5d4594_1316452 = *getMemU32Ptr(0x5D4594, 1316428);
	dword_5d4594_1316436 = *getMemU32Ptr(0x5D4594, 1316464);
	dword_5d4594_1316456 = *getMemU32Ptr(0x5D4594, 1316424);
	dword_5d4594_1316484 = *getMemU32Ptr(0x5D4594, 1316488);
	*getMemU8Ptr(0x5D4594, 1316420) = 1;
	nox_xxx_lightningProc2_4BAE60(&a1a, &a2a, 2, 0, 1, 1, 1);
	if (!nox_xxx_checkGameFlagPause_413A50()) {
		nox_xxx_makeLightningParticles_4999D0(*getMemIntPtr(0x5D4594, 1316520), &fromPos, &toPos);
	}
	return 1;
}

//----- (004BB3F0) --------------------------------------------------------
int nox_thing_chain_lightning_bolt_draw(int* a1, nox_drawable* dr) {
	int2 a1a, a2a, fromPos, toPos;
	if (!nox_lightningRayEndpoints(dr, &fromPos, &toPos)) {
		return 1;
	}
	nox_lightningRayScreen(a1, fromPos, toPos, &a1a, &a2a);
	dword_5d4594_1316452 = *getMemU32Ptr(0x5D4594, 1316428);
	dword_5d4594_1316436 = *getMemU32Ptr(0x5D4594, 1316464);
	dword_5d4594_1316456 = *getMemU32Ptr(0x5D4594, 1316424);
	dword_5d4594_1316484 = *getMemU32Ptr(0x5D4594, 1316488);
	*getMemU8Ptr(0x5D4594, 1316420) = 1;
	nox_xxx_lightningProc2_4BAE60(&a1a, &a2a, 2, 0, 1, 1, 1);
	if (!nox_xxx_checkGameFlagPause_413A50()) {
		nox_xxx_makeLightningParticles_4999D0(*getMemIntPtr(0x5D4594, 1316520), &fromPos, &toPos);
	}
	return 1;
}

//----- (004BB5D0) --------------------------------------------------------
int nox_thing_energy_bolt_draw(int* a1, nox_drawable* dr) {
	const uint8_t* ray = (const uint8_t*)&dr->union_u32[0];
	int2 a1a, a2a, fromPos, toPos;
	if (!nox_lightningRayEndpoints(dr, &fromPos, &toPos)) {
		return 1;
	}
	nox_lightningRayScreen(a1, fromPos, toPos, &a1a, &a2a);
	*getMemU8Ptr(0x5D4594, 1316420) = 2 * ((int8_t)ray[1] + 127);
	dword_5d4594_1316436 = *getMemU32Ptr(0x5D4594, 1316496);
	dword_5d4594_1316484 = *getMemU32Ptr(0x5D4594, 1316468);
	nox_xxx_lightningProc2_4BAE60(&a1a, &a2a, 2, 0, 0, 0, 1);
	if (!nox_xxx_checkGameFlagPause_413A50()) {
		nox_xxx_makeLightningParticles_4999D0(*getMemIntPtr(0x5D4594, 1316524), &fromPos, &toPos);
	}
	return 1;
}

//----- (004BB7B0) --------------------------------------------------------
int nox_thing_green_bolt_draw(int* a1, nox_drawable* dr) {
	uint8_t* ray = (uint8_t*)&dr->union_u32[0];
	if (!ray[0]) {
		uint32_t ticks;
		memcpy(&ticks, ray + 1, sizeof(ticks));
		if (ticks) {
			--ticks;
			memcpy(ray + 1, &ticks, sizeof(ticks));
			if (!ticks) {
				nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr);
				return 0;
			}
		}
	}
	int2 a1a, a2a, fromPos, toPos;
	if (!nox_lightningRayEndpoints(dr, &fromPos, &toPos)) {
		return 1;
	}
	nox_lightningRayScreen(a1, fromPos, toPos, &a1a, &a2a);
	dword_5d4594_1316452 = *getMemU32Ptr(0x5D4594, 1316444);
	dword_5d4594_1316436 = *getMemU32Ptr(0x5D4594, 1316504);
	dword_5d4594_1316456 = *getMemU32Ptr(0x5D4594, 1316460);
	dword_5d4594_1316484 = *getMemU32Ptr(0x5D4594, 1316480);
	*getMemU8Ptr(0x5D4594, 1316420) = 1;
	nox_xxx_lightningProc2_4BAE60(&a1a, &a2a, 2, 0, 1, 1, 1);
	if (!nox_xxx_checkGameFlagPause_413A50()) {
		nox_xxx_makeLightningParticles_4999D0(*getMemIntPtr(0x5D4594, 1316528), &fromPos, &toPos);
	}
	return 1;
}
