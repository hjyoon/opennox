#include <math.h>

#include "client__draw__fx.h"
#include "common__random.h"

#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME3_1.h"
#include "operators.h"

//----- (00499490) --------------------------------------------------------
void sub_499490(int a1, uint16_t* a2, int a3, int a4, char a5, char a6) {
	uint32_t* result; // eax
	uint32_t* v7;     // esi

	result = nox_xxx_spriteLoadAdd_45A360_drawable(a1, a3 + (unsigned short)a2[2], a4 + (unsigned short)a2[3]);
	v7 = result;
	if (result) {
		*((uint16_t*)result + 216) = *a2;
		*((uint16_t*)result + 217) = a2[1];
		*((uint8_t*)result + 443) = a5;
		*((uint8_t*)result + 444) = nox_common_randomIntMinMax_415FF0(3, 10, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 138);
		*((uint8_t*)v7 + 446) = a6;
		*((uint8_t*)v7 + 445) = a6;
		nox_xxx_sprite_45A110_drawable(v7);
	}
}

//----- (00499520) --------------------------------------------------------
void sub_499520(int a1, short* a2, short a3, char a4, char a5) {
	int v5;           // ebp
	int v6;           // edi
	uint32_t* result; // eax
	uint32_t* v8;     // esi

	v6 = a2[2];
	v5 = a2[3];
	result = nox_xxx_spriteLoadAdd_45A360_drawable(a1, v6, v5);
	v8 = result;
	if (result) {
		*((uint16_t*)result + 216) = a2[0];
		*((uint16_t*)result + 217) = a2[1];
		*((uint16_t*)result + 218) = v6;
		*((uint16_t*)result + 219) = v5;
		*((uint8_t*)result + 442) = a3;
		*((uint16_t*)result + 220) = (long long)sqrt((double)((v6 - a2[0]) * (v6 - a2[0]) + (v5 - a2[1]) * (v5 - a2[1])));
		*((uint8_t*)result + 443) = a4;
		*((uint8_t*)result + 444) = nox_common_randomIntMinMax_415FF0(3, 10, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 182);
		*((uint8_t*)v8 + 446) = a5;
		*((uint8_t*)v8 + 445) = a5;
		v8[116] = sub_4CA720;
		*((uint16_t*)v8 + 254) = a3;
		nox_xxx_sprite_45A110_drawable(v8);
	}
}

//----- (00499610) --------------------------------------------------------
int nox_xxx_makePointFxCli_499610(int a1, int a2, int a3, int a4, int a5, int a6) {
	int result; // eax
	int v7;     // eax
	int v8;     // esi

	result = a2;
	if (a2 > 0) {
		do {
			v7 = nox_xxx_spriteLoadAdd_45A360_drawable(a1, a5, a6);
			v8 = v7;
			if (v7) {
				*(uint32_t*)(v7 + 432) = *(uint32_t*)(v7 + 12) << 12;
				*(uint32_t*)(v7 + 436) = *(uint32_t*)(v7 + 16) << 12;
				*(uint8_t*)(v7 + 299) =
					nox_common_randomIntMinMax_415FF0(0, 255, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 227);
				*(uint32_t*)(v8 + 440) =
					nox_common_randomIntMinMax_415FF0(1, a3, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 230);
				*(uint32_t*)(v8 + 448) =
					gameFrame() +
					nox_common_randomIntMinMax_415FF0(a4, 64, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 233);
				*(uint32_t*)(v8 + 444) = gameFrame();
				*(uint16_t*)(v8 + 104) = 0;
				*(uint8_t*)(v8 + 296) =
					nox_common_randomIntMinMax_415FF0(2, 10, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 239);
				nox_xxx_sprite_45A110_drawable((uint32_t*)v8);
			}
			result = --a2;
		} while (a2);
	}
	return result;
}

//----- (00499710) --------------------------------------------------------
int nox_xxx_drawEnergyBolt_499710(int a1, int a2, short a3, int a4) {
	for (int i = 0; i < 2; ++i) {
		nox_drawable* dr = nox_xxx_spriteLoadAdd_45A360_drawable(a4, a1, a2);
		if (!dr) {
			continue;
		}
		dr->union_u32[0] = (uint32_t)a1 << 12;
		dr->union_u32[1] = (uint32_t)a2 << 12;
		dr->field_74_4 = nox_common_randomIntMinMax_415FF0(0, 255, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 268);
		dr->union_u32[2] = nox_common_randomIntMinMax_415FF0(1, 3000, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 271);
		dr->union_u32[4] = gameFrame() + nox_common_randomIntMinMax_415FF0(
			5, 20, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 274);
		dr->union_u32[3] = gameFrame();
		dr->z = a3 + nox_common_randomIntMinMax_415FF0(0, 20, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 279);
		dr->vel_z = nox_common_randomIntMinMax_415FF0(0, 4, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 280);
		nox_xxx_sprite_45A110_drawable(dr);
	}
	return 0;
}

//----- (00499950) --------------------------------------------------------
int sub_499950(int a1, int2* a2, int2* a3, unsigned short a4, char a5) {
	int result; // eax
	int v6;     // esi

	result = nox_xxx_spriteLoadAdd_45A360_drawable(a1, a2->field_0, a2->field_4);
	v6 = result;
	if (result) {
		*(uint32_t*)(result + 104) = a4;
		*(uint8_t*)(result + 296) = a5;
		*(uint16_t*)(result + 440) = a4;
		*(uint8_t*)(result + 442) =
			nox_common_randomIntMinMax_415FF0(3, 10, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 406);
		*(int2*)(v6 + 432) = *a3;
		nox_xxx_sprite_45A110_drawable((uint32_t*)v6);
	}
	return result;
}

//----- (004999D0) --------------------------------------------------------
int nox_xxx_makeLightningParticles_4999D0(int a1, int2* a2, int2* a3) {
	int dx = a3->field_0 - a2->field_0;
	int dy = a3->field_4 - a2->field_4;
	int length = (int)sqrt((double)(dx * dx + dy * dy));
	if (length <= 0) {
		return length;
	}
	int at = nox_common_randomIntMinMax_415FF0(0, length, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 437);
	int step = at;
	while (at <= length) {
		int x = a2->field_0 + dx * at / length;
		int y = a2->field_4 + dy * at / length;
		nox_drawable* dr = nox_xxx_spriteLoadAdd_45A360_drawable(a1, x, y);
		if (dr) {
			dr->union_u32[0] = (uint32_t)x << 12;
			dr->union_u32[1] = (uint32_t)y << 12;
			dr->field_74_4 = nox_common_randomIntMinMax_415FF0(0, 255, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 458);
			dr->union_u32[2] = nox_common_randomIntMinMax_415FF0(1, 3000, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 461);
			dr->union_u32[4] = gameFrame() + nox_common_randomIntMinMax_415FF0(5, 20, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 464);
			dr->union_u32[3] = gameFrame();
			dr->z = nox_common_randomIntMinMax_415FF0(15, 30, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 471);
			dr->vel_z = nox_common_randomIntMinMax_415FF0(-4, 4, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 472);
			nox_xxx_sprite_45A110_drawable(dr);
		}
		step = nox_common_randomIntMinMax_415FF0(8, 100, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 439);
		at += step;
	}
	return step;
}

//----- (00499E70) --------------------------------------------------------
int nox_xxx_draw_499E70(int a1, int a2, int a3, int a4, int a5, int a6, int a7) {
	int v7;     // edi
	int v8;     // esi
	char v9;    // bl
	int v10;    // ebp
	int v11;    // eax
	int result; // eax
	int v13;    // [esp+10h] [ebp-4h]

	v13 = 100;
	do {
		if (a6 == 2) {
			v7 = a3;
			v8 = a2 + nox_common_randomIntMinMax_415FF0(0, a4, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 555);
		} else {
			v8 = a2;
			v7 = a3 + nox_common_randomIntMinMax_415FF0(0, a5, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 563);
		}
		v9 = nox_common_randomIntMinMax_415FF0(2, 5, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 568);
		v10 = nox_common_randomIntMinMax_415FF0(-40, -20, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 571);
		if (a7 == 1) {
			v11 = nox_common_randomIntMinMax_415FF0(-20, 0, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 573);
		} else {
			v11 = nox_common_randomIntMinMax_415FF0(0, 20, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 575);
		}
		nox_client_newScreenParticle_431540(a1, v8, v7, v11, v10, 1, v9, 0, 0, 1);
		result = --v13;
	} while (v13);
	return result;
}

//----- (0049A150) --------------------------------------------------------
int sub_49A150(int2* a1, int a2, unsigned char a3) {
	int v3;     // ebx
	int result; // eax
	int v5;     // ebp
	int v6;     // eax
	int v7;     // esi
	int v8;     // [esp+1Ch] [ebp+Ch]

	v3 = 2400 * a3 / 255 + 200;
	result = 84215050 * a3;
	v5 = 10 * a3 / 255 + 5;
	if (180 * a3 / 255 + 10 > 0) {
		v8 = 180 * a3 / 255 + 10;
		do {
			v6 = nox_xxx_spriteLoadAdd_45A360_drawable(a2, a1->field_0, a1->field_4);
			v7 = v6;
			if (v6) {
				if (v6 != -432) {
					*(uint32_t*)(v6 + 432) = *(uint32_t*)(v6 + 12) << 12;
					*(uint32_t*)(v6 + 436) = *(uint32_t*)(v6 + 16) << 12;
					*(uint8_t*)(v6 + 299) =
						nox_common_randomIntMinMax_415FF0(0, 255, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 822);
					*(uint32_t*)(v7 + 440) =
						nox_common_randomIntMinMax_415FF0(1, v3, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 825);
					*(uint32_t*)(v7 + 448) =
						gameFrame() +
						nox_common_randomIntMinMax_415FF0(v5, 96, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 828);
					*(uint32_t*)(v7 + 444) = gameFrame();
				}
				*(uint16_t*)(v7 + 104) =
					nox_common_randomIntMinMax_415FF0(5, 15, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 834);
				*(uint16_t*)(v7 + 106) = 0;
				*(uint8_t*)(v7 + 296) =
					nox_common_randomIntMinMax_415FF0(0, 8, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 836);
				nox_xxx_sprite_45A110_drawable((uint32_t*)v7);
			}
			result = --v8;
		} while (v8);
	}
	return result;
}
