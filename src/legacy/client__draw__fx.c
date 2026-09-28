#include <math.h>

#include "client__draw__fx.h"
#include "common__random.h"

#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME3_1.h"
#include "operators.h"

//----- (00499490) --------------------------------------------------------
void nox_client_orb_set_payload_499490(nox_drawable* dr, const uint16_t* destination, uint8_t radius,
									   uint8_t fade, uint8_t mode) {
	if (!dr || !destination) {
		return;
	}
	uint8_t* payload = (uint8_t*)&dr->union_u32[0];
	memcpy(payload, destination, 2 * sizeof(*destination));
	payload[11] = radius;
	payload[12] = fade;
	payload[13] = mode;
	payload[14] = mode;
}

void sub_499490(int a1, uint16_t* a2, int a3, int a4, char a5, char a6) {
	nox_drawable* dr =
		nox_xxx_spriteLoadAdd_45A360_drawable(a1, a3 + (unsigned short)a2[2], a4 + (unsigned short)a2[3]);
	if (dr) {
		nox_client_orb_set_payload_499490(
			dr, a2, (uint8_t)a5,
			(uint8_t)nox_common_randomIntMinMax_415FF0(3, 10, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 138),
			(uint8_t)a6);
		nox_xxx_sprite_45A110_drawable(dr);
	}
}

//----- (00499520) --------------------------------------------------------
void nox_client_mana_bomb_orb_set_payload_499520(nox_drawable* dr, const int16_t* path, uint8_t angle,
											 uint8_t reverse, uint8_t fade, uint8_t mode) {
	if (!dr || !path) {
		return;
	}
	int dx = (int)path[2] - (int)path[0];
	int dy = (int)path[3] - (int)path[1];
	uint16_t distance = (uint16_t)sqrt((double)dx * dx + (double)dy * dy);
	uint8_t* payload = (uint8_t*)&dr->union_u32[0];
	memcpy(payload, path, 4 * sizeof(*path));
	memcpy(payload + 8, &distance, sizeof(distance));
	payload[10] = angle;
	payload[11] = reverse;
	payload[12] = fade;
	payload[13] = mode;
	payload[14] = mode;
	dr->field_116 = (void*)sub_4CA720;
	dr->field_127 = (dr->field_127 & 0xFFFF0000u) | angle;
}

void sub_499520(int a1, short* a2, short a3, char a4, char a5) {
	if (!a2) {
		return;
	}
	nox_drawable* dr = nox_xxx_spriteLoadAdd_45A360_drawable(a1, a2[2], a2[3]);
	if (dr) {
		nox_client_mana_bomb_orb_set_payload_499520(
			dr, a2, (uint8_t)a3, (uint8_t)a4,
			(uint8_t)nox_common_randomIntMinMax_415FF0(3, 10, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 182),
			(uint8_t)a5);
		nox_xxx_sprite_45A110_drawable(dr);
	}
}

//----- (00499610) --------------------------------------------------------
void nox_client_ballistic_fx_set_state_499610(nox_drawable* dr, uint32_t speed, uint32_t start_frame,
										  uint32_t end_frame, uint8_t angle) {
	if (!dr) {
		return;
	}
	dr->union_u32[0] = (uint32_t)dr->pos.x << 12;
	dr->union_u32[1] = (uint32_t)dr->pos.y << 12;
	dr->field_74_4 = angle;
	dr->union_u32[2] = speed;
	dr->union_u32[3] = start_frame;
	dr->union_u32[4] = end_frame;
}

int nox_xxx_makePointFxCli_499610(int a1, int a2, int a3, int a4, int a5, int a6) {
	int result = a2;
	if (a2 > 0) {
		do {
			nox_drawable* dr = nox_xxx_spriteLoadAdd_45A360_drawable(a1, a5, a6);
			if (dr) {
				uint32_t frame = gameFrame();
				uint8_t angle = (uint8_t)nox_common_randomIntMinMax_415FF0(
					0, 255, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 227);
				uint32_t speed = (uint32_t)nox_common_randomIntMinMax_415FF0(
					1, a3, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 230);
				uint32_t end_frame = frame + (uint32_t)nox_common_randomIntMinMax_415FF0(
					a4, 64, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 233);
				nox_client_ballistic_fx_set_state_499610(
					dr, speed, frame, end_frame, angle);
				dr->z = 0;
				dr->vel_z =
					nox_common_randomIntMinMax_415FF0(2, 10, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 239);
				nox_xxx_sprite_45A110_drawable(dr);
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
void nox_client_falling_spark_set_payload_499950(nox_drawable* dr, const int2* origin, uint16_t z,
											int8_t velocity_z, uint8_t fade) {
	if (!dr || !origin) {
		return;
	}
	dr->z = z;
	dr->field_26_1 = 0;
	dr->vel_z = velocity_z;
	uint8_t* payload = (uint8_t*)&dr->union_u32[0];
	memcpy(payload, origin, sizeof(*origin));
	memcpy(payload + 8, &z, sizeof(z));
	payload[10] = fade;
}

nox_drawable* sub_499950(int a1, int2* a2, int2* a3, unsigned short a4, char a5) {
	if (!a2 || !a3) {
		return NULL;
	}
	nox_drawable* dr = nox_xxx_spriteLoadAdd_45A360_drawable(a1, a2->field_0, a2->field_4);
	if (dr) {
		nox_client_falling_spark_set_payload_499950(
			dr, a3, a4, a5,
			(uint8_t)nox_common_randomIntMinMax_415FF0(3, 10, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 406));
		nox_xxx_sprite_45A110_drawable(dr);
	}
	return dr;
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
	int v8;     // [esp+1Ch] [ebp+Ch]

	v3 = 2400 * a3 / 255 + 200;
	result = 84215050 * a3;
	v5 = 10 * a3 / 255 + 5;
	if (180 * a3 / 255 + 10 > 0) {
		v8 = 180 * a3 / 255 + 10;
		do {
			nox_drawable* dr = nox_xxx_spriteLoadAdd_45A360_drawable(a2, a1->field_0, a1->field_4);
			if (dr) {
				uint32_t frame = gameFrame();
				uint8_t angle = (uint8_t)nox_common_randomIntMinMax_415FF0(
					0, 255, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 822);
				uint32_t speed = (uint32_t)nox_common_randomIntMinMax_415FF0(
					1, v3, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 825);
				uint32_t end_frame = frame + (uint32_t)nox_common_randomIntMinMax_415FF0(
					v5, 96, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 828);
				nox_client_ballistic_fx_set_state_499610(
					dr, speed, frame, end_frame, angle);
				dr->z =
					nox_common_randomIntMinMax_415FF0(5, 15, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 834);
				dr->field_26_1 = 0;
				dr->vel_z =
					nox_common_randomIntMinMax_415FF0(0, 8, "C:\\NoxPost\\src\\client\\Draw\\Fx.c", 836);
				nox_xxx_sprite_45A110_drawable(dr);
			}
			result = --v8;
		} while (v8);
	}
	return result;
}
