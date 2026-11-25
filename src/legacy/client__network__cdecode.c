#include "game2_packet_bridge.h"
#include <math.h>

#include "client__network__cdecode.h"

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
#include "GAME3_3.h"
#include "GAME5.h"
#include "GAME5_2.h"
#include "client__draw__fx.h"
#include "client__gui__guiinv.h"
#include "client__gui__guishop.h"
#include "client__gui__window.h"
#include "client__video__draw_common.h"
#include "common__magic__speltree.h"
#include "common__net_list.h"
#include "common__random.h"
#include "common__strman.h"
#include "common__system__team.h"
#include "defs.h"
#include "operators.h"

#include "client__draw__animdraw.h"
#include "client__draw__drawrays.h"
#include "client__draw__fx.h"
#include "client__draw__parse__parse.h"
#include "client__gui__chathelp.h"
#include "client__gui__gui_ctf.h"
#include "client__gui__guibrief.h"
#include "client__gui__guicon.h"
#include "client__gui__guigen.h"
#include "client__gui__guiggovr.h"
#include "client__gui__guiinv.h"
#include "client__gui__guijourn.h"
#include "client__gui__guimsg.h"
#include "client__gui__guishop.h"
#include "client__gui__guitrade.h"
#include "client__gui__servopts__playrlst.h"
#include "client__network__inform.h"
#include "client__shell__noxworld.h"
extern uint32_t dword_5d4594_1200832;
extern uint32_t dword_5d4594_1200776;
extern uint32_t dword_5d4594_1200796;
extern uint32_t dword_5d4594_1200768;
extern uintptr_t dword_8531A0_2576;
extern uint32_t nox_server_sanctuaryHelp_54276;
extern uint32_t nox_client_gui_flag_1556112;
extern uint32_t dword_5d4594_1200804;
extern uint32_t nox_gameDisableMapDraw_5d4594_2650672;
extern uint32_t nox_player_netCode_85319C;

int nox_client_getFadeDuration();

void nox_client_onClientStatusA(int v);
void sub_456140(unsigned char a1);

uint32_t nox_client_fadeObjects_80836 = 1;

//----- (0048EA70) --------------------------------------------------------
void sub_41CC00(char* s);
void* nox_new_npc(int id);
void nox_init_npc(void* ptr, int id);
void nox_npc_set_328(int id, int a2);
void nox_xxx_cliDestroyObj_45A9A0(nox_drawable* a1);
int nox_xxx_netOnPacketRecvCli_48EA70_switch(int a1, int op, unsigned char* data, int sz) {
	unsigned char* old = data;
	long long v5;           // rax
	int k;                  // ecx
	int v16;                // eax
	int v17;                // esi
	int v18;                // eax
	int v19;                // esi
	int v20;                // eax
	int v21;                // esi
	int v22;                // esi
	int v24;                // eax
	int v25;                // esi
	int v26;                // esi
	uint32_t* v27;          // esi
	unsigned char v28;      // al
	int v29;                // esi
	unsigned char v30;      // bl
	unsigned int v31;       // eax
	int v32;                // ecx
	int v33;                // eax
	nox_playerInfo* j;      // esi
	int v42;                // eax
	char* v43;              // eax
	int v44;                // eax
	int v45;                // esi
	int v46;                // esi
	int v47;                // eax
	unsigned int v48;       // eax
	int v49;                // eax
	int v50;                // esi
	int v51;                // esi
	int v52;                // eax
	unsigned int v53;       // eax
	int v54;                // eax
	int v55;                // esi
	int v56;                // eax
	int v57;                // eax
	int v58;                // esi
	int v59;                // eax
	int v60;                // esi
	int v62;                // eax
	int v63;                // esi
	int v64;                // eax
	int v65;                // esi
	char v66;               // cl
	int v69;                // eax
	int v70;                // esi
	int v71;                // eax
	int v72;                // esi
	uint32_t* v73;          // esi
	int v74;                // eax
	int v75;                // eax
	unsigned int v76;       // eax
	int v77;                // eax
	int v78;                // esi
	int v79;                // edi
	int v80;                // esi
	int* v81;               // ebx
	int v82;                // eax
	int v83;                // esi
	int v84;                // eax
	int v85;                // esi
	int v88;                // eax
	int v89;                // esi
	int v98;                // eax
	nox_playerInfo* v99;    // esi
	int v100;               // ebx
	int v101;               // eax
	char* v108;             // ebx
	uint32_t* v111;         // esi
	uint32_t* v112;         // eax
	int v113;               // ecx
	uint32_t* v114;         // ebx
	int v115;               // eax
	int v116;               // edx
	unsigned int v117;      // ecx
	unsigned short v118;    // ax
	char* v119;             // esi
	unsigned int v129;      // eax
	int v130;               // esi
	int v132;               // eax
	int v133;               // esi
	int v134;               // esi
	int v135;               // ebx
	int v178;               // eax
	int v179;               // esi
	int v180;               // eax
	int v181;               // esi
	unsigned int v184;      // eax
	unsigned int v185;      // esi
	unsigned short v186;    // ax
	unsigned short v187;    // cx
	uint32_t* v188;         // eax
	int v189;               // eax
	int v190;               // esi
	unsigned short v191;    // ax
	unsigned short v192;    // cx
	unsigned int v193;      // eax
	unsigned int v194;      // esi
	uint32_t* v195;         // eax
	int v196;               // eax
	int v197;               // esi
	int v198;               // eax
	int v199;               // ebx
	int v200;               // eax
	wchar2_t* v201;          // eax
	int v202;               // edi
	wchar2_t* v203;          // eax
	wchar2_t* v204;          // eax
	wchar2_t* v205;          // eax
	char* v206;             // eax
	char* v207;             // esi
	int v208;               // eax
	wchar2_t* v213;            // eax
	int v214;                 // eax
	int v215;                 // esi
	char* v216;               // ebx
	wchar2_t* v217;            // eax
	uint32_t* v218;             // eax
	int v219;                 // edi
	wchar2_t* v220;            // eax
	unsigned int v221;        // eax
	int v222;                 // eax
	int v223;                 // esi
	int v224;                 // eax
	int v225;                 // eax
	int v226;                 // eax
	int v227;                 // esi
	int v228;                 // ebx
	int v229;                 // ecx
	int v230;                 // edi
	int v231;                 // ebx
	uint32_t* v232;             // esi
	char v233;                // al
	int v234;                 // esi
	int v235;                 // eax
	int v236;                 // eax
	int v237;                 // esi
	int v239;                 // ebx
	int v240;                 // edi
	int v241;                 // esi
	int v242;                 // edi
	int v243;                 // edi
	int v244;                 // ebx
	int v245;                 // eax
	uint32_t* v246;             // esi
	char v247;                // al
	short v248;             // ax
	int v249;               // esi
	bool v250;              // zf
	int v251;               // ecx
	unsigned int v252;      // ebx
	int v253;               // edi
	int v254;               // esi
	int v255;               // eax
	int v256;               // esi
	int v257;               // eax
	int v258;               // ecx
	int v259;               // eax
	int v260;               // edi
	int v261;               // eax
	int v262;               // esi
	int v263;               // eax
	int v264;               // eax
	int v265;               // eax
	int v266;               // eax
	int v267;               // eax
	int v268;               // eax
	int v269;               // eax
	int v270;               // eax
	wchar2_t* v271;          // eax
	char* v272;             // esi
	uint32_t* v273;         // edi
	unsigned int v274;      // eax
	unsigned int v275;      // esi
	int v276;               // ebx
	int v277;               // esi
	uint32_t* v278;         // ebx
	int v279;               // esi
	wchar2_t* v280;          // esi
	unsigned int v281;      // eax
	unsigned int v282;      // esi
	int v283;               // eax
	int v284;               // eax
	unsigned short v285;    // ax
	unsigned short v286;    // bx
	int v287;               // ebx
	char* v288;             // edi
	uint8_t* v289;          // esi
	wchar2_t* v290;          // eax
	nox_playerInfo* v296;   // esi
	int v297;               // eax
	int v298;               // esi
	char v315;              // [esp-8h] [ebp-1758h]
	unsigned char v316;     // [esp-8h] [ebp-1758h]
	int v317;               // [esp-4h] [ebp-1754h]
	int v318;               // [esp-4h] [ebp-1754h]
	unsigned short v319;    // [esp-4h] [ebp-1754h]
	int v320;               // [esp+0h] [ebp-1750h]
	int v321;               // [esp+0h] [ebp-1750h]
	int v322;               // [esp+0h] [ebp-1750h]
	char v323;              // [esp+0h] [ebp-1750h]
	size_t v324;            // [esp+0h] [ebp-1750h]
	char v325;              // [esp+0h] [ebp-1750h]
	char v326;              // [esp+0h] [ebp-1750h]
	int v327;               // [esp+0h] [ebp-1750h]
	float v328;             // [esp+4h] [ebp-174Ch]
	int v329;               // [esp+4h] [ebp-174Ch]
	int v331;               // [esp+4h] [ebp-174Ch]
	float v332;             // [esp+4h] [ebp-174Ch]
	int v337;               // [esp+4h] [ebp-174Ch]
	int v338;               // [esp+4h] [ebp-174Ch]
	int v339;               // [esp+4h] [ebp-174Ch]
	float v340;             // [esp+4h] [ebp-174Ch]
	int v341;               // [esp+4h] [ebp-174Ch]
	char v342;              // [esp+4h] [ebp-174Ch]
	char* v343;             // [esp+4h] [ebp-174Ch]
	size_t v344;            // [esp+4h] [ebp-174Ch]
	int v345;               // [esp+4h] [ebp-174Ch]
	char v346;              // [esp+4h] [ebp-174Ch]
	char v347;              // [esp+4h] [ebp-174Ch]
	unsigned int v356;      // [esp+18h] [ebp-1738h]
	int v357;               // [esp+18h] [ebp-1738h]
	int* v358;              // [esp+18h] [ebp-1738h]
	int v360;               // [esp+20h] [ebp-1730h]
	int i;                  // [esp+24h] [ebp-172Ch]
	int v362;               // [esp+28h] [ebp-1728h]
	char v365[6];           // [esp+34h] [ebp-171Ch]
	int v367;               // [esp+44h] [ebp-170Ch]
	int v368;               // [esp+48h] [ebp-1708h]
	int v371;               // [esp+54h] [ebp-16FCh]
	int v374;               // [esp+60h] [ebp-16F0h]
	int2 v377;              // [esp+6Ch] [ebp-16E4h]
	int2 v378;              // [esp+74h] [ebp-16DCh]
	int2 v379;              // [esp+7Ch] [ebp-16D4h]
	int v382[8];            // [esp+A0h] [ebp-16B0h]
	wchar2_t v386[20];       // [esp+F0h] [ebp-1660h]
	wchar2_t v387[20];       // [esp+118h] [ebp-1638h]
	wchar2_t v397[100];      // [esp+8C0h] [ebp-E90h]

	if (op == 0x9e && (!data || sz < 7)) {
		return -1;
	}
	unsigned char* end = data + sz;

	switch (op) {
	case 47: // Native object messages share the existing typed Go receiver.
		return nox_game2_decode_native_object(a1, op, data, sz);
#if 0 // Exact former packed-pointer object-message case.
	case 47: // MSG_SIMPLE_OBJ
		nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		if (nox_client_isConnected_43C700()) {
			if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ADDITIONAL_NETWORK_TEST)) {
				nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
			}
			LODWORD(v5) = nox_xxx_spriteCreate_48E970(*(unsigned short*)(data + 3), *(unsigned short*)(data + 1),
													  *(unsigned short*)(data + 5), *(unsigned short*)(data + 7));
			if ((uint32_t)v5) {
				k = gameFrame();
				*(uint32_t*)(v5 + 288) = gameFrame();
			}
		}
		return 9;
#endif // native object-message dependency
	case 48: // Native object messages share the existing typed Go receiver.
		return nox_game2_decode_native_object(a1, op, data, sz);
#if 0 // Exact former packed-pointer object-message case.
	case 48: // MSG_COMPLEX_OBJ
		nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		if (!nox_client_isConnected_43C700()) {
			return 11;
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ADDITIONAL_NETWORK_TEST)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		if (*(uint16_t*)(data + 3) || *(uint16_t*)(data + 1)) {
			LODWORD(v5) = nox_xxx_spriteCreate_48E970(*(unsigned short*)(data + 3), *(unsigned short*)(data + 1),
													  *(unsigned short*)(data + 5), *(unsigned short*)(data + 7));
			v29 = v5;
			if ((uint32_t)v5) {
				*(uint32_t*)(v5 + 288) = gameFrame();
				nox_xxx_spriteSetFrameMB_45AB80(v5, *(unsigned char*)(data + 10));
				LOBYTE(v5) = (*(uint8_t*)(data + 9) >> 4) & 7;
				*(uint8_t*)(v29 + 297) = v5;
				if ((unsigned char)v5 > 3u) {
					*(uint8_t*)(v29 + 297) = v5 + 1;
				}
				v30 = *(uint8_t*)(data + 9) & 0xF;
				*(uint8_t*)(data + 9) = v30;
				if (*(uint32_t*)(v29 + 276) != v30) {
					*(uint32_t*)(v29 + 316) = gameFrame();
					HIDWORD(v5) = *(unsigned char*)(data + 9);
					*(uint32_t*)(v29 + 276) = HIDWORD(v5);
				}
			}
			k = nox_player_netCode_85319C;
			if (*(unsigned short*)(data + 1) == nox_player_netCode_85319C && sub_416120(9u)) {
				nox_xxx_cliUpdateCameraPos_435600(*(unsigned short*)(data + 5), *(unsigned short*)(data + 7));
			}
		} else {
			nox_xxx_cliUpdateCameraPos_435600(*(unsigned short*)(data + 5), *(unsigned short*)(data + 7));
			nox_xxx_setKeybTimeout_4160D0(9);
		}
		return 11;
#endif // native object-message dependency
	case 50: { // MSG_OBJECT_OUT_OF_SIGHT
		v49 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v50 = v49;
		if (!nox_client_isConnected_43C700()) {
			return 3;
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ADDITIONAL_NETWORK_TEST)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		nox_drawable* drawable = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
								 ? nox_xxx_netSpriteByCodeStatic_45A720(v50)
								 : nox_xxx_netSpriteByCodeDynamic_45A6F0(v50);
		if (!drawable) {
			return 3;
		}
		if (drawable->draw_func == nox_thing_animate_draw) {
			nox_animate_draw_data_t* draw_data = drawable->field_76;
			if (draw_data && draw_data->animation_kind == 1) {
				return 3;
			}
		}
		if (drawable == getMemPtr(0x852978, 8)) {
			return 3;
		}
		v53 = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		if (v53) {
			nox_xxx_cliDestroyObj_45A9A0(drawable);
		} else {
			nox_xxx_spriteDeleteStatic_45A4E0_drawable(drawable);
		}
		return 3;
	}
	case 51: { // MSG_OBJECT_IN_SHADOWS
		v54 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v55 = v54;
		if (!nox_client_isConnected_43C700()) {
			return 3;
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		nox_drawable* drawable = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
								 ? nox_xxx_netSpriteByCodeStatic_45A720(v55)
								 : nox_xxx_netSpriteByCodeDynamic_45A6F0(v55);
		if (!drawable) {
			return 3;
		}
		k = 1;
		drawable->field_121 = 1;
		drawable->field_120 = 1;
		drawable->field_122 = 1;
		nox_drawable* local_player = getMemPtr(0x852978, 8);
		if (nox_client_fadeObjects_80836) {
			if (drawable != local_player) {
				nox_xxx_spriteTransparentDecay_49B950(drawable, (int)gameFPS());
			}
		} else {
			nox_animate_draw_data_t* draw_data = drawable->field_76;
			if ((drawable->draw_func != nox_thing_animate_draw || !draw_data || draw_data->animation_kind != 1) &&
				drawable != local_player) {
				v53 = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
				if (v53) {
					nox_xxx_cliDestroyObj_45A9A0(drawable);
				} else {
					nox_xxx_spriteDeleteStatic_45A4E0_drawable(drawable);
				}
				return 3;
			}
		}
		return 3;
	}
	case 52: // MSG_OBJECT_FRIEND_ADD
		v57 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v58 = v57;
		if (nox_client_isConnected_43C700()) {
			nox_xxx_cliAddObjFriend_4959F0(v58);
		}
		return 3;
	case 53: // MSG_OBJECT_FRIEND_REMOVE
		v59 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v60 = v59;
		if (nox_client_isConnected_43C700()) {
			sub_495A20(v60);
		}
		return 3;
	case 54: // MSG_RESET_FRIENDS
		if (nox_client_isConnected_43C700()) {
			sub_4959B0();
		}
		return 1;
	case 55: { // MSG_ENABLE_OBJECT
		v62 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v63 = v62;
		if (nox_client_isConnected_43C700()) {
			if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
				nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
			}
			nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							   ? nox_xxx_netSpriteByCodeStatic_45A720(v63)
							   : nox_xxx_netSpriteByCodeDynamic_45A6F0(v63);
			if (dr) {
				dr->flags30 |= 0x1000000u;
			}
		}
		return 3;
	}
	case 56: { // MSG_DISABLE_OBJECT
		v64 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v65 = v64;
		if (nox_client_isConnected_43C700()) {
			if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
				nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
			}
			nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							   ? nox_xxx_netSpriteByCodeStatic_45A720(v65)
							   : nox_xxx_netSpriteByCodeDynamic_45A6F0(v65);
			if (dr) {
				if (dr->flags28 & 0x40000) {
					dr->draw_func = NULL;
				}
				dr->flags30 &= 0xFEFFFFFF;
			}
		}
		return 3;
	}
	case 57: { // MSG_DRAW_FRAME
		v16 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v17 = v16;
		if (!nox_client_isConnected_43C700()) {
			return 4;
		}
		nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
						   ? nox_xxx_netSpriteByCodeStatic_45A720(v17)
						   : nox_xxx_netSpriteByCodeDynamic_45A6F0(v17);
		if (!dr) {
			return 4;
		}
		nox_xxx_spriteSetFrameMB_45AB80_drawable(dr, *(unsigned char*)(data + 3));
		return 4;
	}
	case 59: { // MSG_OPEN_WALL
		if (nox_client_isConnected_43C700()) {
			if (!nox_common_gameFlags_check_40A5C0(1)) {
				uint8_t* wall_ptr = (uint8_t*)sub_410550(*(uint16_t*)(data + 1));
				if (wall_ptr) {
					if (wall_ptr[4] & 4) {
						nox_secret_wall_t* secret = (nox_secret_wall_t*)nox_server_wallData(wall_ptr);
						if (secret) {
							secret->open_delay = 23;
							secret->state = 3;
						}
					}
				}
			}
		}
		return 3;
	}
	case 60: { // MSG_CLOSE_WALL
		if (nox_client_isConnected_43C700()) {
			if (!nox_common_gameFlags_check_40A5C0(1)) {
				uint8_t* wall_ptr = (uint8_t*)sub_410550(*(uint16_t*)(data + 1));
				if (wall_ptr) {
					if (wall_ptr[4] & 4) {
						nox_secret_wall_t* secret = (nox_secret_wall_t*)nox_server_wallData(wall_ptr);
						if (secret) {
							secret->open_delay = 0;
							secret->state = 1;
						}
					}
				}
			}
		}
		return 3;
	}
	case 61: // MSG_CHANGE_OR_ADD_WALL_MAGIC
		if (!nox_client_isConnected_43C700()) {
			return 6;
		}
		LODWORD(v5) = nox_server_getWallAtGrid_410580(*(unsigned char*)(data + 4), *(unsigned char*)(data + 5));
		if (!(uint32_t)v5) {
			LODWORD(v5) = nox_xxx_wallCreateAt_410250(*(unsigned char*)(data + 4), *(unsigned char*)(data + 5));
			if (!(uint32_t)v5) {
				return 6;
			}
		}
		v66 = *(uint8_t*)(data + 1);
		data += 6;
		*(uint8_t*)(v5 + 1) = v66;
		*(uint8_t*)v5 = *(uint8_t*)(data - 4);
		*(uint8_t*)(v5 + 2) = *(uint8_t*)(data - 3);
		return 6;
	case 62: // MSG_REMOVE_WALL_MAGIC
		if (nox_client_isConnected_43C700()) {
			LODWORD(v5) = nox_server_getWallAtGrid_410580(*(unsigned char*)(data + 1), *(unsigned char*)(data + 2));
			if ((uint32_t)v5) {
				nox_xxx_mapDelWallAtPt_410430(*(unsigned char*)(v5 + 5), *(unsigned char*)(v5 + 6));
			}
		}
		return 3;
	case 65: // MSG_REPORT_HEALTH
	case 66: // MSG_REPORT_HEALTH_DELTA
	case 67: // MSG_REPORT_PLAYER_HEALTH
	case 68: // MSG_REPORT_ITEM_HEALTH
	case 69: // MSG_REPORT_MANA
	case 71: // MSG_REPORT_STAMINA
		return nox_client_decodeInventoryReport(op, data, sz);
	case 72: // MSG_REPORT_STATS
		v98 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v99 = (nox_playerInfo*)dword_8531A0_2576;
		v100 = v98;
		if (nox_client_isConnected_43C700()) {
			if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
				nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
			}
			if (v100 == nox_player_netCode_85319C) {
				v101 = nox_common_gameFlags_check_40A5C0(1);
				if (!v101 && v99) {
					nox_playerInfo* player = v99;
					player->info.field_2247 = *(unsigned short*)(data + 3);
					player->info.field_2243 = *(unsigned short*)(data + 5);
					LOWORD(v101) = *(uint16_t*)(data + 9);
					v332 = (double)v101;
					player->info.field_2235 = nox_float2int(v332);
					player->info.field_2239 = *(unsigned short*)(data + 11);
					player->field_3652 = (player->field_3652 & 0xFFFF0000u) | *(uint16_t*)(data + 7);
					player->field_3684 = *(uint8_t*)(data + 13);
				}
				nox_xxx_j_inventoryNameSignInit_467460();
			}
		}
		return 14;
	case 73: // MSG_REPORT_ARMOR_VALUE
	case 74: // MSG_REPORT_GOLD
	case 75: // MSG_REPORT_PICKUP
	case 76: // MSG_REPORT_MODIFIABLE_PICKUP
	case 77: // MSG_REPORT_DROP
		return nox_client_decodeInventoryReport(op, data, sz);
	case 78: // MSG_REPORT_LESSON
		return nox_client_decodeScorePacket(op, data, sz);
	case 79: // MSG_REPORT_MUNDANE_ARMOR_EQUIP
	case 80: // MSG_REPORT_MUNDANE_WEAPON_EQUIP
		if (!nox_client_isConnected_43C700()) {
			return 7;
		}
		LOWORD(v5) = *(uint16_t*)(data + 1);
		v367 = -1;
		*(uint16_t*)(data + 1) = v5 & 0x7FFF;
		if ((((unsigned int)v5 >> 15) & 1) == 1) {
			nox_xxx_clientEquipWeaponArmor_417AA0(*(uint8_t*)data, *(unsigned short*)(data + 1),
												  *(uint32_t*)(data + 3), (const uint8_t*)&v367);
		} else {
			nox_xxx_clientEquip_49A3D0(*(uint8_t*)data, *(unsigned short*)(data + 1), *(uint32_t*)(data + 3),
									   (const uint8_t*)&v367);
		}
		return 7;
	case 81: // MSG_REPORT_MODIFIABLE_WEAPON_EQUIP
	case 82: // MSG_REPORT_MODIFIABLE_ARMOR_EQUIP
		LODWORD(v5) = nox_client_isConnected_43C700();
		if (!(uint32_t)v5) {
			goto LABEL_437;
		}
		LOWORD(v5) = *(uint16_t*)(data + 1);
		v116 = *(uint32_t*)(data + 3);
		v117 = v5;
		v118 = v5 & 0x7FFF;
		*(uint16_t*)(data + 1) = v118;
		if (((v117 >> 15) & 1) == 1) {
			nox_xxx_clientEquipWeaponArmor_417AA0(*(uint8_t*)data, v118, v116, data + 7);
		} else {
			nox_xxx_clientEquip_49A3D0(*(uint8_t*)data, v118, v116, data + 7);
		}
	LABEL_437:
		return 11;
	case 83: // MSG_REPORT_ARMOR_DEQUIP
	case 84: // MSG_REPORT_WEAPON_DEQUIP
		if (!nox_client_isConnected_43C700()) {
			return 7;
		}
		sub_417B80(*(uint8_t*)data, *(unsigned short*)(data + 1), *(uint32_t*)(data + 3));
		return 7;
	case 85: // MSG_REPORT_TREASURE_COUNT
	case 86: // MSG_REPORT_FLAG_BALL_WINNER
	case 87: // MSG_REPORT_FLAG_WINNER
	case 88: // MSG_REPORT_DEATHMATCH_WINNER
	case 89: // MSG_REPORT_DEATHMATCH_TEAM_WINNER
		return nox_client_decodeScorePacket(op, data, sz);
	case 90: { // MSG_REPORT_ENCHANTMENT
		v132 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v133 = v132;
		if (!nox_client_isConnected_43C700()) {
			return 7;
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		nox_drawable* drawable = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
								 ? nox_xxx_netSpriteByCodeStatic_45A720(v133)
								 : nox_xxx_netSpriteByCodeDynamic_45A6F0(v133);
		if (!drawable) {
			return 7;
		}
		int had_light = nox_client_drawable_testBuff_4356C0(drawable, 15);
		drawable->buffs = *(uint32_t*)(data + 3);
		nox_drawable* local_player = getMemPtr(0x852978, 8);
		if (drawable == local_player) {
			sub_467410(*(uint32_t*)(data + 3));
		}
		if (had_light != 1 || nox_client_drawable_testBuff_4356C0(drawable, 15) ||
			(drawable == local_player && (sub_467430() & 8))) {
			return 7;
		}
		nox_thing* v136 = nox_get_thing(drawable->field_27);
		nox_drawable_change_light_intensity_native(drawable, v136->light_intensity);
		return 7;
	}
	case 91: // MSG_REPORT_ITEM_ENCHANTMENT
		if (nox_client_isConnected_43C700()) {
			LOBYTE(v129) = sub_467430();
			v130 = (v129 >> 3) & 1;
			sub_467420(*(uint8_t*)(data + 1));
			nox_drawable* local_player = getMemPtr(0x852978, 8);
			if (v130 == 1 && !(sub_467430() & 8) && local_player &&
				!nox_client_drawable_testBuff_4356C0(local_player, 15)) {
				nox_thing* v131 = nox_get_thing(local_player->field_27);
				nox_drawable_change_light_intensity_native(local_player, v131->light_intensity);
			}
		}
		return 2;
	case 92: { // MSG_REPORT_LIGHT_COLOR
		int code = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		if (!nox_client_isConnected_43C700()) {
			return 6;
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							 ? nox_xxx_netSpriteByCodeStatic_45A720(code)
							 : nox_xxx_netSpriteByCodeDynamic_45A6F0(code);
		if (dr) {
			nox_drawable_change_light_color_native(dr, *(unsigned char*)(data + 3),
										 *(unsigned char*)(data + 4), *(unsigned char*)(data + 5));
		}
		return 6;
	}
	case 93: { // MSG_REPORT_LIGHT_INTENSITY
		int code = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		if (!nox_client_isConnected_43C700()) {
			return 7;
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							 ? nox_xxx_netSpriteByCodeStatic_45A720(code)
							 : nox_xxx_netSpriteByCodeDynamic_45A6F0(code);
		if (dr) {
			nox_drawable_change_light_intensity_native(dr, *(float*)(data + 3));
		}
		return 7;
	}
	case 94: // MSG_REPORT_Z_PLUS
		v178 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v179 = v178;
		if (nox_client_isConnected_43C700()) {
			LODWORD(v5) = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							  ? nox_xxx_netSpriteByCodeStatic_45A720(v179)
							  : nox_xxx_netSpriteByCodeDynamic_45A6F0(v179);
			if ((uint32_t)v5) {
				WORD2(v5) = *(unsigned char*)(data + 3);
				*(uint16_t*)(v5 + 104) = WORD2(v5);
			}
		}
		return 4;
	case 95: // MSG_REPORT_Z_MINUS
		v180 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v181 = v180;
		if (nox_client_isConnected_43C700()) {
			LODWORD(v5) = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							  ? nox_xxx_netSpriteByCodeStatic_45A720(v181)
							  : nox_xxx_netSpriteByCodeDynamic_45A6F0(v181);
			if ((uint32_t)v5) {
				WORD2(v5) = *(unsigned char*)(data + 3);
				HIDWORD(v5) = -HIDWORD(v5);
				*(uint16_t*)(v5 + 104) = WORD2(v5);
			}
		}
		return 4;
	case 96: // MSG_REPORT_EQUIP
	case 97: // MSG_REPORT_DEQUIP
	case 100: // MSG_REPORT_CHARGES
		return nox_client_decodeInventoryReport(op, data, sz);
	case 101: { // MSG_REPORT_X_STATUS
		v71 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v72 = v71;
		if (nox_client_isConnected_43C700()) {
			if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
				nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
			}
			nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							   ? nox_xxx_netSpriteByCodeStatic_45A720(v72)
							   : nox_xxx_netSpriteByCodeDynamic_45A6F0(v72);
			if (dr) {
				v74 = dr->flags70;
				dr->flags70 = *(uint32_t*)(data + 3);
				k = dr->flags28;
				if (k & 0x20000) {
					if (!(v74 & 0x400) && (dr->flags70 & 0x400) == 0x400) {
						sub_4BC720(dr);
					}
					v75 = dr->flags70;
					if (v75 & 0x800) {
						dr->flags28 &= 0xFFF7FFFF;
						dr->flags30 &= 0xDFFFFFFF;
					}
				}
			}
		}
		return 7;
	}
	case 102: // MSG_REPORT_PLAYER_STATUS
		if (nox_client_isConnected_43C700()) {
			nox_drawable* local = getMemPtr(0x852978, 8);
			if (local) {
				local->flags30 = *(uint32_t*)(data + 1);
			}
		}
		return 5;
	case 103: { // MSG_REPORT_MODIFIER
		v77 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v78 = v77;
		if (nox_client_isConnected_43C700()) {
			if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
				nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
			}
			nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							   ? nox_xxx_netSpriteByCodeStatic_45A720(v78)
							   : nox_xxx_netSpriteByCodeDynamic_45A6F0(v78);
			if (dr) {
				for (int i = 0; i < 4; i++) {
					dr->item_modifiers[i] = nox_xxx_modifGetDescById_413330(data[i + 3]);
				}
				dr->item_field_112_0 = -1;
				dr->item_field_112_2 = -1;
			}
		}
		return 7;
	}
	case 104: // MSG_REPORT_STAT_MODIFIER
		v82 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v83 = v82;
		if (!nox_client_isConnected_43C700()) {
			return 8;
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		if (v83 == nox_player_netCode_85319C) {
			LOBYTE(k) = *(uint8_t*)(data + 7);
			sub_467470(k, *(float*)(data + 3));
		}
		return 8;
	case 105: // MSG_REPORT_NPC
		if (nox_client_isConnected_43C700()) {
			v285 = *(uint16_t*)(data + 1);
			v286 = v285;
			v285 &= 0x7FFFu;
			*(uint16_t*)(data + 1) = v285;
			v287 = v286 >> 15;
			v288 = nox_npc_by_id(v285);
			if (v288) {
				nox_init_npc(v288, *(unsigned short*)(data + 1));
			} else {
				v288 = nox_new_npc(*(unsigned short*)(data + 1));
			}
			if (v288) {
				k = (int)(v288 + 8);
				v289 = (uint8_t*)(data + 4);
				v358 = (int*)(v288 + 8);
				v360 = 6;
				do {
					BYTE4(v5) = v289[1];
					LOBYTE(v5) = *v289;
					LOBYTE(k) = *(v289 - 1);
					v289 += 3;
					*v358 = nox_color_rgb_4344A0(k, v5, SHIDWORD(v5));
					k = (int)(v358 + 1);
					LODWORD(v5) = v360 - 1;
					v250 = v360 == 1;
					++v358;
					--v360;
				} while (!v250);
				*((uint32_t*)v288 + 328) = v287;
			}
		}
		return 21;
	case 106: // MSG_REPORT_CLIENT_STATUS
		v119 = nox_common_playerInfoGetByID_417040(*(unsigned short*)(data + 1));
		if (!v119) {
			return 7;
		}
		nox_playerInfo* status_player = (nox_playerInfo*)v119;
		if (!nox_common_gameFlags_check_40A5C0(1)) {
			nox_xxx_playerUnsetStatus_417530(status_player, 1059);
			nox_xxx_netNeedTimestampStatus_4174F0(status_player, *(uint32_t*)(data + 3) & 0x423);
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) ||
			(HIDWORD(v5) = *(unsigned short*)(data + 1), WORD2(v5) != nox_player_netCode_85319C)) {
			return 7;
		}
		nox_client_onClientStatusA(status_player->field_3680);
		sub_470C40((status_player->field_3680 >> 10) & 1);
		return 7;
	case 107: { // MSG_REPORT_ANIMATION_FRAME
		v69 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v70 = v69;
		if (!nox_client_isConnected_43C700()) {
			return 7;
		}
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ENABLE_NET_DEBUG)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
						   ? nox_xxx_netSpriteByCodeStatic_45A720(v70)
						   : nox_xxx_netSpriteByCodeDynamic_45A6F0(v70);
		if (!dr) {
			return 7;
		}
		nox_xxx_spriteSetFrameMB_45AB80_drawable(dr, *(uint32_t*)(data + 3));
		return 7;
	}
	case 108: // MSG_REPORT_ACQUIRE_CREATURE
		v184 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v185 = v184;
		if (!nox_client_isConnected_43C700()) {
			return 5;
		}
		v186 = *(uint16_t*)(data + 3);
		v187 = v186;
		v186 &= 0x7FFFu;
		*(uint16_t*)(data + 3) = v186;
		nox_xxx_cliSummonCreat_4C2E50(*(unsigned short*)(data + 1), v186, v187 & 0x8000);
		if (nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))) {
			v188 = nox_xxx_netSpriteByCodeStatic_45A720(v185);
		} else {
			v188 = nox_xxx_netSpriteByCodeDynamic_45A6F0(v185);
		}
		if (v188 || (v188 = nox_xxx_spriteCreate_48E970(*(unsigned short*)(data + 3), v185, 0, 0)) != 0) {
			sub_459DD0((int)v188, 1);
		}
		sub_495060(*(unsigned short*)(data + 1), 0, 0);
		return 5;
	case 109: // MSG_REPORT_LOSE_CREATURE
		v189 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v190 = v189;
		if (!nox_client_isConnected_43C700()) {
			return 3;
		}
		v191 = *(uint16_t*)(data + 1);
		v192 = v191;
		v191 &= 0x7FFFu;
		*(uint16_t*)(data + 1) = v191;
		nox_xxx_cliSummonOnDieOrBanish_4C3140(v191, (void*)(v192 & 0x8000));
		sub_4950C0(*(unsigned short*)(data + 1));
		v337 = v190;
		if (!nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))) {
			LODWORD(v5) = nox_xxx_netSpriteByCodeDynamic_45A6F0(v337);
			if ((uint32_t)v5) {
				nox_xxx_cliRemoveHealthbar_459E30(v5, 1);
			}
			return 3;
		}
		LODWORD(v5) = nox_xxx_netSpriteByCodeStatic_45A720(v190);
		if ((uint32_t)v5) {
			nox_xxx_cliRemoveHealthbar_459E30(v5, 1);
		}
		return 3;
	case 110: // MSG_REPORT_EXPERIENCE
	case 111: // MSG_REPORT_SPELL_AWARD
		return nox_client_decodeProgressPacket(op, data, sz);
	case 125: // MSG_FX_PLASMA
	case 140: // MSG_FX_LIGHTNING
	case 141: // MSG_FX_ENERGY_BOLT
	case 142: // MSG_FX_CHAIN_LIGHTNING_BOLT
		if (!nox_client_isConnected_43C700()) {
			goto LABEL_732;
		}
		nox_xxx_netDrawRays_49BDD0((unsigned char*)data);
		HIWORD(k) = 0;
		v379.field_0 = *(unsigned short*)(data + 1);
		v379.field_4 = *(unsigned short*)(data + 3);
		LOWORD(k) = *(uint16_t*)(data + 5);
		HIWORD(v5) = 0;
		v377.field_0 = (unsigned short)k;
		WORD2(v5) = *(uint16_t*)(data + 7);
		v377.field_4 = WORD2(v5);
		LOBYTE(v5) = *(uint8_t*)data;
		if (*(uint8_t*)data == 140 || (uint8_t)v5 == 142) {
			nox_xxx_makeLightningParticles_4999D0(*(int*)&dword_5d4594_1200776, &v379, &v377);
		} else if ((uint8_t)v5 == 125) {
			nox_xxx_drawEnergyBolt_499710(*(unsigned short*)(data + 5), *(unsigned short*)(data + 7), 10,
										  *(int*)&dword_5d4594_1200776);
		}
	LABEL_732:
		return 9;
	case 0x7Eu: // MSG_FX_SUMMON
		if (nox_client_isConnected_43C700()) {
			nox_xxx_netHandleSummonPacket_4B7C40(*(uint16_t*)(data + 5), (unsigned short*)(data + 1),
												 *(uint16_t*)(data + 7), *(uint8_t*)(data + 9),
												 *(uint16_t*)(data + 10));
		}
		return 12;
	case 0x7Fu: // MSG_FX_SUMMON_CANCEL
		if (nox_client_isConnected_43C700()) {
			sub_4B7EE0(*(uint16_t*)(data + 1));
		}
		return 3;
	case 0x80u: // MSG_FX_SHIELD
		if (nox_client_isConnected_43C700()) {
			nox_xxx_fxShield_4B8090(*(unsigned short*)(data + 1), *(unsigned char*)(data + 3));
		}
		return 4;
	case 0x81u: // MSG_FX_BLUE_SPARKS
		if (!nox_client_isConnected_43C700()) {
			return 5;
		}
		nox_xxx_makePointFxCli_499610(*(int*)&dword_5d4594_1200776, 50, 1000, 30, *(short*)(data + 1),
									  *(short*)(data + 3));
		return 5;
	case 0x82u: // MSG_FX_YELLOW_SPARKS
		if (!nox_client_isConnected_43C700()) {
			return 5;
		}
		nox_xxx_makePointFxCli_499610(*getMemIntPtr(0x5D4594, 1200780), 25, 500, 25, *(short*)(data + 1),
									  *(short*)(data + 3));
		return 5;
	case 0x83u: // MSG_FX_CYAN_SPARKS
		if (!nox_client_isConnected_43C700()) {
			return 5;
		}
		nox_xxx_makePointFxCli_499610(*getMemIntPtr(0x5D4594, 1200784), 25, 500, 25, *(short*)(data + 1),
									  *(short*)(data + 3));
		return 5;
	case 0x84u: // MSG_FX_VIOLET_SPARKS
		if (nox_client_isConnected_43C700()) {
			nox_xxx_makePointFxCli_499610(*(int*)&dword_5d4594_1200796, 25, 500, 25, *(short*)(data + 1),
										  *(short*)(data + 3));
		}
		return 5;
	case 0x85u: // MSG_FX_EXPLOSION
	case 0x86u: // MSG_FX_LESSER_EXPLOSION
	case 0x87u: // MSG_FX_COUNTERSPELL_EXPLOSION
	case 0x88u: // MSG_FX_THIN_EXPLOSION
	case 0x89u: // MSG_FX_TELEPORT
	case 0x8Au: // MSG_FX_SMOKE_BLAST
	case 0x8Bu: // MSG_FX_DAMAGE_POOF
	case 0xA3u: // MSG_FX_MANA_BOMB_CANCEL
		if (nox_client_isConnected_43C700()) {
			k = 0;
			switch (op) {
			case 133:
				v257 = *getMemU32Ptr(0x5D4594, 1200872);
				if (!*getMemU32Ptr(0x5D4594, 1200872)) {
					v257 = nox_xxx_getTTByNameSpriteMB_44CFC0("FireBoom");
					*getMemU32Ptr(0x5D4594, 1200872) = v257;
				}
				v258 = *(short*)(data + 3);
				LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(v257, *(short*)(data + 1), v258);
				if ((uint32_t)v5) {
					nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
				}
				return 5;
			case 134:
				v257 = *getMemU32Ptr(0x5D4594, 1200876);
				if (!*getMemU32Ptr(0x5D4594, 1200876)) {
					v257 = nox_xxx_getTTByNameSpriteMB_44CFC0("MediumFireBoom");
					*getMemU32Ptr(0x5D4594, 1200876) = v257;
				}
				v258 = *(short*)(data + 3);
				LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(v257, *(short*)(data + 1), v258);
				if ((uint32_t)v5) {
					nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
				}
				return 5;
			case 135:
				v257 = *getMemU32Ptr(0x5D4594, 1200880);
				if (!*getMemU32Ptr(0x5D4594, 1200880)) {
					v257 = nox_xxx_getTTByNameSpriteMB_44CFC0("CounterspellBoom");
					*getMemU32Ptr(0x5D4594, 1200880) = v257;
				}
				v258 = *(short*)(data + 3);
				LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(v257, *(short*)(data + 1), v258);
				if ((uint32_t)v5) {
					nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
				}
				return 5;
			case 136:
				v257 = *getMemU32Ptr(0x5D4594, 1200884);
				if (!*getMemU32Ptr(0x5D4594, 1200884)) {
					v257 = nox_xxx_getTTByNameSpriteMB_44CFC0("ThinFireBoom");
					*getMemU32Ptr(0x5D4594, 1200884) = v257;
				}
				v258 = *(short*)(data + 3);
				LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(v257, *(short*)(data + 1), v258);
				if ((uint32_t)v5) {
					nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
				}
				return 5;
			case 137:
				v257 = *getMemU32Ptr(0x5D4594, 1200888);
				if (!*getMemU32Ptr(0x5D4594, 1200888)) {
					v257 = nox_xxx_getTTByNameSpriteMB_44CFC0("TeleportPoof");
					*getMemU32Ptr(0x5D4594, 1200888) = v257;
				}
				v258 = *(short*)(data + 3) + 2;
				LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(v257, *(short*)(data + 1), v258);
				if ((uint32_t)v5) {
					nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
				}
				return 5;
			case 138:
				if (!*getMemU32Ptr(0x5D4594, 1200900)) {
					*getMemU32Ptr(0x5D4594, 1200900) = nox_xxx_getTTByNameSpriteMB_44CFC0("Smoke");
					*getMemU32Ptr(0x5D4594, 1200896) = nox_xxx_getTTByNameSpriteMB_44CFC0("Puff");
				}
				v259 = nox_xxx_spriteLoadAdd_45A360_drawable(*getMemIntPtr(0x5D4594, 1200900), *(short*)(data + 1),
															 *(short*)(data + 3));
				if (v259) {
					*(uint16_t*)(v259 + 104) = 20;
					nox_xxx_sprite_45A110_drawable((uint32_t*)v259);
				}
				v260 = 6;
				do {
					v341 = *(short*)(data + 3) + nox_common_randomIntMinMax_415FF0(
													 -15, 15, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5711);
					v261 = nox_common_randomIntMinMax_415FF0(-15, 15,
															 "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5710);
					v262 = nox_xxx_spriteLoadAdd_45A360_drawable(*getMemIntPtr(0x5D4594, 1200896),
																 *(short*)(data + 1) + v261, v341);
					if (v262) {
						*(uint16_t*)(v262 + 104) = nox_common_randomIntMinMax_415FF0(
							5, 25, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5712);
						nox_xxx_sprite_45A110_drawable((uint32_t*)v262);
					}
					--v260;
				} while (v260);
				return 5;
			case 139:
				v257 = *getMemU32Ptr(0x5D4594, 1200892);
				if (!*getMemU32Ptr(0x5D4594, 1200892)) {
					v257 = nox_xxx_getTTByNameSpriteMB_44CFC0("DamagePoof");
					*getMemU32Ptr(0x5D4594, 1200892) = v257;
				}
				v258 = *(short*)(data + 3) + 2;
				LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(v257, *(short*)(data + 1), v258);
				if ((uint32_t)v5) {
					nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
				}
				return 5;
			case 163:
				v340 = nox_xxx_gamedataGetFloat_419D40("ManaBombOutRadius");
				v252 = nox_float2int(v340);
				v357 = 150;
				v253 = v252 >> 2;
				for (i = v252 >> 2;; v253 = i) {
					v254 = v253 + nox_common_randomIntMinMax_415FF0(
									  0, v252, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5536);
					if (v254 > (int)v252) {
						v254 = v252;
					}
					v255 = nox_common_randomIntMinMax_415FF0(0, 255, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c",
															 5541);
					LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(
						*getMemIntPtr(0x5D4594, 1200784),
						*(short*)(data + 1) + v254 * *getMemIntPtr(0x587000, 192088 + 8 * v255) / 16,
						*(short*)(data + 3) + v254 * *getMemIntPtr(0x587000, 192092 + 8 * v255) / 16);
					v256 = v5;
					if ((uint32_t)v5) {
						*(uint32_t*)(v5 + 432) = *(uint32_t*)(v5 + 12) << 12;
						*(uint32_t*)(v5 + 436) = *(uint32_t*)(v5 + 16) << 12;
						*(uint8_t*)(v5 + 299) = 0;
						*(uint32_t*)(v5 + 440) = 0;
						*(uint32_t*)(v5 + 448) = gameFrame() +
												 nox_common_randomIntMinMax_415FF0(
													 30, 40, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5562);
						*(uint32_t*)(v256 + 444) = gameFrame();
						*(uint16_t*)(v256 + 104) = 0;
						*(uint8_t*)(v256 + 296) = nox_common_randomIntMinMax_415FF0(
							4, 10, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5568);
						nox_xxx_sprite_45A110_drawable((uint32_t*)v256);
					}
					if (!--v357) {
						break;
					}
				}
				return 5;
			default:
				return 5;
			}
		} else {
			return 5;
		}
		return -1;
	case 0x8Fu: // MSG_FX_DRAIN_MANA
	case 0x90u: // MSG_FX_CHARM
	case 0x91u: // MSG_FX_GREATER_HEAL
		if (nox_client_isConnected_43C700()) {
			nox_xxx_netDrawRays_49BDD0((unsigned char*)data);
		}
		return 9;
	case 0x93u: // MSG_FX_SPARK_EXPLOSION
		if (!*getMemU32Ptr(0x5D4594, 1200852)) {
			*getMemU32Ptr(0x5D4594, 1200852) = nox_xxx_getTTByNameSpriteMB_44CFC0("Spark");
			*getMemU32Ptr(0x5D4594, 1200856) = nox_xxx_getTTByNameSpriteMB_44CFC0("MediumFireBoom");
			*getMemU32Ptr(0x5D4594, 1197380) = nox_xxx_getTTByNameSpriteMB_44CFC0("FireBoom");
		}
		if (!nox_client_isConnected_43C700()) {
			return 6;
		}
		v378.field_0 = *(short*)(data + 1);
		v378.field_4 = *(short*)(data + 3);
		sub_49A150(&v378, *getMemIntPtr(0x5D4594, 1200852), *(uint8_t*)(data + 5));
		v235 = *getMemU32Ptr(0x5D4594, 1197380);
		if (*(uint8_t*)(data + 5) <= 0xAAu) {
			v235 = *getMemU32Ptr(0x5D4594, 1200856);
		}
		LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(v235, *(short*)(data + 1), *(short*)(data + 3));
		if ((uint32_t)v5) {
			nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
		}
		return 6;
	case 0x95u: // MSG_FX_SENTRY_RAY
		if (nox_client_isConnected_43C700()) {
			sub_4C5020(data);
			if (nox_common_randomIntMinMax_415FF0(0, 100, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 4987) <
				25) {
				nox_drawable* local_player = getMemPtr(0x852978, 8);
				if (local_player) {
					v225 = *(unsigned short*)(data + 5) - (int)local_player->pos.x;
					v226 = nox_double2int(
						sqrt((double)(v225 * v225 + (*(unsigned short*)(data + 7) -
												 (int)local_player->pos.y) *
												(*(unsigned short*)(data + 7) -
												 (int)local_player->pos.y))));
					if (v226 < 600) {
						nox_xxx_clientPlaySoundSpecial_452D80(297, 100 * (600 - v226) / 600);
					}
				}
			}
			if (!nox_xxx_checkGameFlagPause_413A50()) {
				v227 = *(unsigned short*)(data + 5) - *(unsigned short*)(data + 1);
				v228 = *(unsigned short*)(data + 7) - *(unsigned short*)(data + 3);
				v229 = nox_double2int(sqrt((double)(v227 * v227 + v228 * v228)));
				if (!v229) {
					v229 = 1;
				}
				v230 = *(unsigned short*)(data + 5) - 4 * v227 / v229;
				v231 = *(unsigned short*)(data + 7) - 4 * v228 / v229;
				LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(*(int*)&dword_5d4594_1200796, v230, v231);
				v362 = v5;
				if ((uint32_t)v5) {
					v232 = (uint32_t*)(v5 + 432);
					if ((uint32_t)v5 != (uint32_t)-432) {
						*(uint32_t*)(v5 + 432) = v230 << 12;
						*(uint32_t*)(v5 + 436) = v231 << 12;
						v233 = nox_common_randomIntMinMax_415FF0(
							0, 255, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5024);
						*(uint8_t*)(v362 + 299) = v233;
						v232[2] = nox_common_randomIntMinMax_415FF0(
							1, 1500, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5027);
						v232[4] = gameFrame() +
								  nox_common_randomIntMinMax_415FF0(
									  5, 20, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5030);
						v232[3] = gameFrame();
					}
					v234 = v362;
					*(uint16_t*)(v362 + 104) = 22;
					*(uint8_t*)(v234 + 296) = nox_common_randomIntMinMax_415FF0(
						-4, 4, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5037);
					nox_xxx_sprite_45A110_drawable((uint32_t*)v234);
				}
			}
		}
		return 9;
	case 0x96u: // MSG_FX_RICOCHET
		if (nox_client_isConnected_43C700()) {
			if (!*getMemU32Ptr(0x5D4594, 1200860)) {
				*getMemU32Ptr(0x5D4594, 1200860) = nox_xxx_getTTByNameSpriteMB_44CFC0("BlueSpark");
			}
			v239 = 5;
			do {
				LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(*getMemIntPtr(0x5D4594, 1200860),
																	*(short*)(data + 1), *(short*)(data + 3));
				v240 = v5;
				if ((uint32_t)v5) {
					if ((uint32_t)v5 != (uint32_t)-432) {
						*(uint32_t*)(v5 + 432) = *(uint32_t*)(v5 + 12) << 12;
						*(uint32_t*)(v5 + 436) = *(uint32_t*)(v5 + 16) << 12;
						*(uint8_t*)(v5 + 299) = nox_common_randomIntMinMax_415FF0(
							0, 255, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5338);
						*(uint32_t*)(v240 + 440) = nox_common_randomIntMinMax_415FF0(
							1333, 4000, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5341);
						*(uint32_t*)(v240 + 448) = gameFrame() +
												   nox_common_randomIntMinMax_415FF0(
													   5, 20, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5344);
						*(uint32_t*)(v240 + 444) = gameFrame();
					}
					*(uint16_t*)(v240 + 104) = 20;
					*(uint8_t*)(v240 + 296) = nox_common_randomIntMinMax_415FF0(
						-5, 5, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 5351);
					nox_xxx_sprite_45A110_drawable((uint32_t*)v240);
				}
				--v239;
			} while (v239);
		}
		return 5;
	case 0x98u: // MSG_FX_GREEN_BOLT
		if (!*getMemU32Ptr(0x5D4594, 1200844)) {
			*getMemU32Ptr(0x5D4594, 1200844) = nox_xxx_getTTByNameSpriteMB_44CFC0("GreenZap");
		}
		if (nox_client_isConnected_43C700()) {
			LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(
							  *getMemIntPtr(0x5D4594, 1200844),
							  *(unsigned short*)(data + 1) +
								  (*(unsigned short*)(data + 5) - *(unsigned short*)(data + 1)) / 2,
							  *(unsigned short*)(data + 3) +
								  (*(unsigned short*)(data + 7) - *(unsigned short*)(data + 3)) / 2) +
						  432;
			*(uint8_t*)v5 = 0;
			*(uint32_t*)(v5 + 5) = *(uint32_t*)(data + 1);
			HIDWORD(v5) = *(uint32_t*)(data + 5);
			*(uint32_t*)(v5 + 9) = HIDWORD(v5);
			k = *(unsigned short*)(data + 9);
			*(uint32_t*)(v5 + 1) = k;
		}
		return 11;
	case 0x9Au: // MSG_FX_WHITE_FLASH
		if (nox_client_isConnected_43C700()) {
			nox_xxx____setargv_11_473920();
		}
		return 5;
	case 0x9Bu: // MSG_FX_GENERATING_MAP
	case 0x9Cu: // MSG_FX_ASSEMBLING_MAP
	case 0x9Du: // MSG_FX_POPULATING_MAP
		return nox_client_decodeProgressPacket(op, data, sz);
	case 0x9Eu: // MSG_FX_DURATION_SPELL
		switch (data[1]) {
		case 1u:
		case 2u:
		case 3u:
		case 4u:
		case 5u:
		case 6u:
		case 7u:
			if (nox_client_isConnected_43C700()) {
				nox_xxx_clientAddRayEffect_49C160(data);
			}
			return 7;
		case 8u:
		case 9u:
		case 0xAu:
		case 0xBu:
		case 0xCu:
		case 0xDu:
		case 0xEu:
			if (!nox_client_isConnected_43C700()) {
				return 7;
			}
			nox_xxx_clientRemoveRayEffect_49C450(data);
			return 7;
		default:
			return -1;
		}
	case 0x9Fu: // MSG_FX_DELTAZ_SPELL_START
		v297 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v298 = v297;
		if (nox_client_isConnected_43C700()) {
			LODWORD(v5) = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							  ? nox_xxx_netSpriteByCodeStatic_45A720(v298)
							  : nox_xxx_netSpriteByCodeDynamic_45A6F0(v298);
			if ((uint32_t)v5) {
				LODWORD(v5) = v5 + 432;
				*(uint32_t*)v5 = gameFrame();
				i = *(char*)(data + 4);
				*(float*)(v5 + 8) = (double)i;
				k = *(unsigned char*)(data + 3);
				i = *(unsigned char*)(data + 3);
				*(float*)(v5 + 4) = (double)i;
				HIDWORD(v5) = *(unsigned char*)(data + 5);
				i = *(unsigned char*)(data + 5);
				*(float*)(v5 + 12) = (double)i;
			}
		}
		return 6;
	case 0xA0u: // MSG_FX_TURN_UNDEAD
		if (!nox_client_isConnected_43C700()) {
			return 5;
		}
		nox_xxx_fxDrawTurnUndead_499880((short*)(data + 1));
		return 5;
	case 0xA1u: // MSG_FX_ARROW_TRAP
		if (!nox_client_isConnected_43C700()) {
			return 6;
		}
		if (!*getMemU32Ptr(0x5D4594, 1200864)) {
			*getMemU32Ptr(0x5D4594, 1200864) = nox_xxx_getTTByNameSpriteMB_44CFC0("ArrowTrap1Smoke");
			*getMemU32Ptr(0x5D4594, 1200868) = nox_xxx_getTTByNameSpriteMB_44CFC0("ArrowTrap2Smoke");
		}
		v251 = *(short*)(data + 1);
		v339 = *(short*)(data + 3);
		if (*(uint8_t*)(data + 5) == 1) {
			LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(*getMemIntPtr(0x5D4594, 1200864), v251 + 15, v339);
			if ((uint32_t)v5) {
				nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
				return 6;
			}
		} else {
			LODWORD(v5) = nox_xxx_spriteLoadAdd_45A360_drawable(*getMemIntPtr(0x5D4594, 1200868), v251 - 3, v339);
			if ((uint32_t)v5) {
				nox_xxx_sprite_45A110_drawable((uint32_t*)v5);
			}
			return 6;
		}
		return 6;
	case 0xA2u: // MSG_FX_VAMPIRISM
		if (nox_client_isConnected_43C700()) {
			if (!*getMemU32Ptr(0x5D4594, 1200848)) {
				*getMemU32Ptr(0x5D4594, 1200848) = nox_xxx_getTTByNameSpriteMB_44CFC0("HealOrb");
			}
			v221 = *(unsigned short*)(data + 9);
			k = v221 & 0xFFFC;
			if (k <= 28) {
				v222 = v221 >> 2;
			} else {
				v222 = 7;
			}
			v223 = v222 + 1;
			if (v222 + 1 > 0) {
				do {
					LOBYTE(v368) = nox_common_randomIntMinMax_415FF0(
						6, 12, "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 4957);
					v323 = v368;
					v318 = nox_common_randomIntMinMax_415FF0(-20, 20,
															 "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 4961);
					v224 = nox_common_randomIntMinMax_415FF0(-20, 20,
															 "C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 4960);
					sub_499490(*getMemIntPtr(0x5D4594, 1200848), (uint16_t*)(data + 1), v224, v318, v323, 0);
					--v223;
				} while (v223);
			}
		}
		return 11;
	case 0xA4u: // MSG_UPDATE_STREAM
		// Handled by handleUpdateStreamPacketNative494A60 before legacy dispatch.
		return 0;
	case 0xA6u: // MSG_AUDIO_EVENT
	case 0xA7u: // MSG_AUDIO_PLAYER_EVENT
		if (!nox_client_isConnected_43C700()) {
			goto LABEL_651;
		}
		v338 = *(char*)(data + 1);
		v321 = (*(unsigned short*)(data + 2) >> 9) & 0x7E;
		v317 = *(uint16_t*)(data + 2) & 0x3FF;
		if (*(char*)data == -89) {
			sub_452E10(v317, v321, v338);
			return 4;
		}
		sub_452DC0(v317, v321, v338);
	LABEL_651:
		return 4;
	case 169: // MSG_INFORM
		return nox_client_handlePacketInformSized_4C9BF0(data, sz);
	case 171: // MSG_IMPORTANT_ACK
	case 0xAEu: // MSG_OUTGOING_CLIENT
		return nox_client_decodeControlPacket(a1, op, data, sz);
	case 0xAFu: // MSG_GAME_SETTINGS
	case 0xB0u: // MSG_GAME_SETTINGS_2
	case 0xB1u: // MSG_UPDATE_GUI_GAME_SETTINGS
		return nox_client_decodeSettingsPacket(op, data, sz);
	case 0xB2u: { // MSG_DOOR_ANGLE
		int code = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		if (nox_client_isConnected_43C700()) {
			nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							 ? nox_xxx_netSpriteByCodeStatic_45A720(code)
							 : nox_xxx_netSpriteByCodeDynamic_45A6F0(code);
			if (dr) {
				dr->field_74_4 = *(uint8_t*)(data + 3);
			}
		}
		return 4;
	}
	case 0xB3u: { // MSG_OBELISK_CHARGE
		int code = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		if (nox_client_isConnected_43C700()) {
			nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							 ? nox_xxx_netSpriteByCodeStatic_45A720(code)
							 : nox_xxx_netSpriteByCodeDynamic_45A6F0(code);
			if (dr) {
				nox_xxx_spriteSetActiveMB_45A990_drawable(dr);
				v328 = (double)(16 * *(unsigned char*)(data + 3) / 10);
				nox_drawable_change_light_intensity_native(dr, v328);
				nox_xxx_spriteSetFrameMB_45AB80_drawable(dr, 8 * *(unsigned char*)(data + 3) / 50);
				if (dr->field_77 == 8) {
					dr->field_77 = 7;
				}
			}
		}
		return 4;
	}
	case 0xB4u: { // MSG_PENTAGRAM_ACTIVATE
		int code = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		if (nox_client_isConnected_43C700()) {
			nox_drawable* dr = nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))
							 ? nox_xxx_netSpriteByCodeStatic_45A720(code)
							 : nox_xxx_netSpriteByCodeDynamic_45A6F0(code);
			if (dr) {
				if (*(uint8_t*)(data + 3)) {
					dr->flags28 |= 0x80000u;
					nox_drawable_change_light_intensity_native(dr, 41.958);
				} else {
					dr->flags28 &= 0xFFF7FFFF;
					nox_drawable_change_light_intensity_native(dr, 0.0);
				}
				nox_xxx_spriteSetFrameMB_45AB80_drawable(dr, *(unsigned char*)(data + 3));
				dr->field_72 = gameFrame();
			}
		}
		return 4;
	}
	case 0xB5u: // MSG_CLIENT_PREDICT_LINEAR
		if (nox_client_isConnected_43C700() &&
			(v329 = *(unsigned short*)(data + 7), v320 = *(unsigned short*)(data + 5),
			 v31 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1)),
			 LODWORD(v5) = nox_xxx_spriteCreate_48E970(*(unsigned short*)(data + 3), v31, v320, v329),
			 (uint32_t)v5)) {
			*(uint16_t*)(v5 + 508) = *(uint16_t*)(data + 9);
			*(float*)(v5 + 468) = (double)*(char*)(data + 12) * 0.0625;
			*(float*)(v5 + 472) = (double)*(char*)(data + 13) * 0.0625;
			*(float*)(v5 + 476) = (double)*(char*)(data + 11) * 0.0625;
			*(uint32_t*)(v5 + 316) = gameFrame();
			*(uint32_t*)(v5 + 324) = *(unsigned short*)(data + 5);
			v32 = *(unsigned short*)(data + 7);
			*(uint32_t*)(v5 + 460) = nox_xxx_sprite_4CA540;
			*(uint32_t*)(v5 + 328) = v32;
			nox_xxx_spriteToList_49BC80_drawable((uint32_t*)v5);
		}
		return 14;
	case 0xBDu: // MSG_SYSOP_RESULT
		if (nox_client_isConnected_43C700()) {
			if (*(uint8_t*)(data + 1) == 1) {
				v290 = nox_strman_loadString_40F1D0("sysopAccess", 0,
													"C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 6351);
			} else {
				v290 = nox_strman_loadString_40F1D0("invalidPass", 0,
													"C:\\NoxPost\\src\\Client\\Network\\cdecode.c", 6356);
			}
			nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, v290);
		}
		return 2;
	case 0xC3u: // Native object messages share the existing typed Go receiver.
		return nox_game2_decode_native_object(a1, op, data, sz);
#if 0 // Exact former packed-pointer object-message case.
	case 0xC3u: // MSG_PLAYER_OBJ
		nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_ADDITIONAL_NETWORK_TEST)) {
			nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1));
		}
		if (*(uint16_t*)(data + 3) || *(uint16_t*)(data + 1)) {
			v27 = nox_xxx_spriteCreate_48E970(*(unsigned short*)(data + 3), *(unsigned short*)(data + 1),
											  *(unsigned short*)(data + 5), *(unsigned short*)(data + 7));
			if (v27) {
				v27[72] = gameFrame();
				nox_xxx_spriteSetFrameMB_45AB80((int)v27, *(unsigned char*)(data + 10));
				v28 = (*(uint8_t*)(data + 9) >> 4) & 7;
				*((uint8_t*)v27 + 297) = v28;
				if (v28 > 3u) {
					*((uint8_t*)v27 + 297) = v28 + 1;
				}
				if (v27[69] != *(unsigned char*)(data + 11)) {
					v27[79] = gameFrame();
					k = *(unsigned char*)(data + 11);
					v27[69] = k;
				}
			}
			HIDWORD(v5) = *(unsigned short*)(data + 1);
			if (WORD2(v5) == nox_player_netCode_85319C && sub_416120(8u)) {
				nox_xxx_cliUpdateCameraPos_435600(*(unsigned short*)(data + 5), *(unsigned short*)(data + 7));
			}
			return 12;
		}
		nox_xxx_cliUpdateCameraPos_435600(*(unsigned short*)(data + 5), *(unsigned short*)(data + 7));
		nox_xxx_setKeybTimeout_4160D0(8);
		return 12;
#endif // native object-message dependency
	case 0xC4u: // MSG_TEAM_MSG
		switch (*(unsigned char*)(data + 1)) {
		case 0u: {
			const int packet_len = 2 * *(unsigned char*)(data + 15) + 18;
			if (!nox_client_isConnected_43C700()) {
				return packet_len;
			}
			memcpy(v386, (const void*)(data + 18), 2 * *(unsigned char*)(data + 15));
			v386[*(unsigned char*)(data + 15)] = 0;
			if (*(uint8_t*)(data + 17)) {
				v271 = nox_server_teamTitle_418C20(*(unsigned char*)(data + 16));
				nox_swprintf(v386, v271);
			}
			nox_team_t* team = nox_xxx_getTeamByID_418AB0(*(uint32_t*)(data + 2));
			if (!team) {
				team = nox_xxx_teamCreate_4186D0(*(uint8_t*)(data + 2));
				if (!team) {
					return packet_len;
				}
			}
			sub_418800(team->name, v386, 0);
			sub_418830(team, *(uint32_t*)(data + 6));
			nox_xxx_netChangeTeamID_419090(team, *(uint32_t*)(data + 10));
			team->def_ind = *(uint8_t*)(data + 16);
			sub_457230(v386);
			if (!(*(uint8_t*)(data + 14) & 1)) {
				return packet_len;
			}
			nox_object_team_t* value = nox_xxx_objGetTeamByNetCode_418C80(nox_player_netCode_85319C);
			if (!value) {
				return packet_len;
			}
			if (nox_common_gameFlags_check_40A5C0(1)) {
				nox_xxx_createAtImpl_4191D0(team->field_57, value, 1, nox_player_netCode_85319C, 1);
				return packet_len;
			}
			sub_419900(team, value, nox_player_netCode_85319C);
			return packet_len;
		}
		case 1u: {
			unsigned int code = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 6));
			if (nox_client_isConnected_43C700()) {
				nox_drawable* drawable;
				if (nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 6))) {
					drawable = nox_xxx_netSpriteByCodeStatic_45A720(code);
				} else {
					drawable = nox_xxx_netSpriteByCodeDynamic_45A6F0(code);
				}
				if (!drawable) {
					drawable = nox_xxx_spriteCreate_48E970(*(unsigned short*)(data + 8), code, 0, 0);
				}
				if (drawable) {
					nox_team_t* team = nox_xxx_getTeamByID_418AB0(*(uint32_t*)(data + 2));
					if (team) {
						nox_object_team_t* value = (nox_object_team_t*)&drawable->field_6;
						nox_xxx_createAtImpl_4191D0(team->field_57, value, 0, *(unsigned short*)(data + 6), 0);
						sub_4571A0(*(unsigned short*)(data + 6), team->field_57);
					}
				}
			}
			return 10;
		}
		case 2u: {
			if (nox_client_isConnected_43C700()) {
				nox_object_team_t* value = nox_xxx_objGetTeamByNetCode_418C80(*(uint32_t*)(data + 2));
				if (value) {
					nox_xxx_netChangeTeamMb_419570(value, *(uint32_t*)(data + 2));
					sub_4571A0(*(uint32_t*)(data + 2), 0);
				}
			}
			return 6;
		}
		case 3u: {
			if (nox_client_isConnected_43C700()) {
				nox_object_team_t* value = nox_xxx_objGetTeamByNetCode_418C80(*(unsigned short*)(data + 6));
				if (value) {
					nox_team_t* team = nox_xxx_getTeamByID_418AB0(*(uint32_t*)(data + 2));
					if (team) {
						if (sub_4196D0(value, team, *(unsigned short*)(data + 6), 0)) {
							sub_4571A0(*(unsigned short*)(data + 6), team->field_57);
						}
					}
				}
			}
			return 10;
		}
		case 4u: {
			if (nox_client_isConnected_43C700()) {
				nox_team_t* team = nox_xxx_getTeamByID_418AB0(*(uint32_t*)(data + 2));
				if (team) {
					nox_xxx_teamRenameMB_418CD0(team, (wchar2_t*)(data + 6));
				}
			}
			return 46;
		}
		case 5u: {
			if (nox_client_isConnected_43C700()) {
				nox_team_t* team = nox_xxx_getTeamByID_418AB0(*(uint32_t*)(data + 2));
				if (team) {
					sub_418D80(team);
				}
			}
			return 6;
		}
		case 6u: {
			if (nox_client_isConnected_43C700()) {
				nox_team_t* team = nox_xxx_getTeamByID_418AB0(*(uint32_t*)(data + 2));
				if (team) {
					nox_wcsncpy(v387, team->name, 256);
					sub_418F20(team, 0);
					sub_456EA0(v387);
				}
			}
			return 6;
		}
		case 7u:
			if (nox_client_isConnected_43C700()) {
				nox_server_teamsZzz_419030(0);
				sub_456FA0();
			}
			return 2;
		case 8u: {
			if (nox_client_isConnected_43C700()) {
				nox_team_t* team = nox_xxx_getTeamByID_418AB0(*(uint32_t*)(data + 2));
				if (team) {
					nox_xxx_netChangeTeamID_419090(team, *(uint32_t*)(data + 6));
				}
			}
			return 10;
		}
		case 9u:
			if (nox_client_isConnected_43C700()) {
				nox_server_teamsResetYyy_417D00();
			}
			return 2;
		case 0xCu: {
			if (nox_client_isConnected_43C700()) {
				nox_playerInfo* player = nox_common_playerInfoGetByID_417040(*(unsigned short*)(data + 2));
				if (player) {
					player->field_2282 = (player->field_2282 & 0xFF00u) | *(uint8_t*)(data + 4);
				}
			}
			return 5;
		}
		}
		return -1;
	case 0xC5u: // MSG_KICK_NOTIFICATION
	case 0xC6u: // MSG_TIMEOUT_NOTIFICATION
		return nox_client_decodeControlPacket(a1, op, data, sz);
	case 0xC9u: // MSG_TRADE
		return nox_client_decodeTradePacket(data, sz);
	case 0xCAu: // MSG_CHAT_KILL
	case 0xCBu: // MSG_MESSAGES_KILL
	case 0xCCu: // MSG_SEQ_IMPORTANT
		return nox_client_decodeControlPacket(a1, op, data, sz);
	case 0xCDu: // MSG_REPORT_ABILITY_AWARD
	case 0xCEu: // MSG_REPORT_ABILITY_STATE
	case 0xCFu: // MSG_REPORT_ACTIVE_ABILITIES
	case 0xD1u: // MSG_REPORT_GUIDE_AWARD
		return nox_client_decodeProgressPacket(op, data, sz);
	case 0xD2u: // MSG_INTERESTING_ID
		v281 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v282 = v281;
		if (!nox_client_isConnected_43C700()) {
			return 7;
		}
		if (nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))) {
			v283 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
			LODWORD(v5) = nox_xxx_netSpriteByCodeStatic_45A720(v283);
		} else {
			v284 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
			LODWORD(v5) = nox_xxx_netSpriteByCodeDynamic_45A6F0(v284);
		}
		if (*(uint8_t*)(data + 5) == 1) {
			if ((uint32_t)v5 ||
				(LODWORD(v5) = nox_xxx_spriteCreate_48E970(*(unsigned short*)(data + 3), v282, 0, 0),
				 (uint32_t)v5)) {
				sub_459DD0(v5, *(uint8_t*)(data + 6));
				return 7;
			}
		} else if ((uint32_t)v5) {
			nox_xxx_cliRemoveHealthbar_459E30(v5, *(uint8_t*)(data + 6));
		}
		return 7;
	case 0xD3u: // MSG_TIMER_STATUS
		return nox_client_decodeControlPacket(a1, op, data, sz);
	case 0xD5u: // MSG_JOURNAL_MSG
		v296 = (nox_playerInfo*)dword_8531A0_2576;
		switch (*(uint8_t*)(data + 1)) {
		case 1:
			if (nox_client_isConnected_43C700()) {
				if (v296) {
					nox_xxx_journalEntryAdd_427490(v296, (char*)(data + 2), *(uint16_t*)(data + 66));
				}
				nox_xxx_cliBuildJournalString_469BC0();
			}
			return 68;
		case 2:
			if (nox_client_isConnected_43C700()) {
				if (v296) {
					nox_xxx_journalEntryRemove_427590(v296, (char*)(data + 2));
				}
				nox_xxx_cliBuildJournalString_469BC0();
			}
			return 68;
		case 3:
			if (nox_client_isConnected_43C700() && v296) {
				nox_xxx_journalUpdateEntry_4276B0(v296, (char*)(data + 2), *(uint16_t*)(data + 66));
			}
			return 68;
		}
		return -1;
	case 0xD6u: // MSG_CHAPTER_END
		return nox_client_decodeProgressPacket(op, data, sz);
	case 0xD7u: // MSG_REPORT_ALL_LATENCY
		return nox_client_decodeControlPacket(a1, op, data, sz);
	case 0xD8u: // MSG_REPORT_FLAG_STATUS
		if (nox_client_isConnected_43C700()) {
			sub_455D80(*(uint8_t*)(data + 2), *(uint8_t*)(data + 1));
			sub_4705F0(*(uint8_t*)(data + 1), *(uint8_t*)(data + 3), *(uint16_t*)(data + 4));
		}
		return 6;
	case 0xD9u: // MSG_REPORT_BALL_STATUS
		if (nox_client_isConnected_43C700()) {
			sub_456140(*(uint8_t*)(data + 1));
			sub_470650(*(uint8_t*)(data + 1), *(uint16_t*)(data + 2));
		}
		return 4;
	case 0xDAu: // MSG_REPORT_OBJECT_POISON
		return nox_client_decodeInventoryReport(op, data, sz);
	case 0xDBu: // MSG_REPORT_MONITOR_CREATURE
		v193 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v194 = v193;
		if (nox_client_isConnected_43C700()) {
			if (nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))) {
				v195 = nox_xxx_netSpriteByCodeStatic_45A720(v194);
			} else {
				v195 = nox_xxx_netSpriteByCodeDynamic_45A6F0(v194);
			}
			if (v195 || (v195 = nox_xxx_spriteCreate_48E970(*(unsigned short*)(data + 3), v194, 0, 0)) != 0) {
				sub_459DD0((int)v195, 1);
			}
			sub_495060(v194, 0, 0);
		}
		return 5;
	case 0xDCu: // MSG_REPORT_UNMONITOR_CREATURE
		v196 = nox_xxx_netClearHighBit_578B30(*(uint16_t*)(data + 1));
		v197 = v196;
		if (!nox_client_isConnected_43C700()) {
			return 3;
		}
		sub_4950C0(*(unsigned short*)(data + 1));
		v337 = v197;
		if (nox_xxx_netTestHighBit_578B70(*(unsigned short*)(data + 1))) {
			LODWORD(v5) = nox_xxx_netSpriteByCodeStatic_45A720(v197);
		} else {
			LODWORD(v5) = nox_xxx_netSpriteByCodeDynamic_45A6F0(v337);
		}
		if ((uint32_t)v5) {
			nox_xxx_cliRemoveHealthbar_459E30(v5, 1);
		}
		return 3;
	case 0xDDu: // MSG_REPORT_TOTAL_HEALTH
	case 0xDEu: // MSG_REPORT_TOTAL_MANA
		return nox_client_decodeInventoryReport(op, data, sz);
	case 0xDFu: // MSG_REPORT_SPELL_STAT
		return nox_client_decodeProgressPacket(op, data, sz);
	case 0xE0u: // MSG_REPORT_SECONDARY_WEAPON
	case 0xE1u: // MSG_REPORT_LAST_QUIVER
	case 0xE2u: // MSG_INFO_BOOK_DATA
		return nox_client_decodeInventoryReport(op, data, sz);
	case 0xE5u: // MSG_MUSIC_EVENT
	case 0xE6u: // MSG_MUSIC_PUSH_EVENT
	case 0xE7u: // MSG_MUSIC_POP_EVENT
		return nox_client_decodeControlPacket(a1, op, data, sz);
	case 0xE8u: // MSG_PLAYER_DIED
		if (nox_client_isConnected_43C700()) {
			LODWORD(v5) = nox_common_playerInfoGetByID_417040(*(unsigned short*)(data + 1));
			v111 = (uint32_t*)v5;
			if ((uint32_t)v5) {
				if ((uint32_t)v5 == dword_8531A0_2576) {
					sub_4BFE40();
					sub_478000();
				}
				v112 = nox_xxx_netSpriteByCodeDynamic_45A6F0(*(unsigned short*)(data + 1));
				if (v112) {
					nox_xxx_cliRemoveHealthbar_459E30((int)v112, 3);
				}
				if (!nox_common_gameFlags_check_40A5C0(4096)) {
					v111[1] = 0;
					LODWORD(v5) = v111 + 582;
					v113 = 27;
					do {
						v114 = (uint32_t*)v5;
						LODWORD(v5) = v5 + 24;
						*(uint32_t*)(v5 - 4) = 0;
						v5 = (unsigned int)v5;
						*v114 = 0;
						--v113;
						v114[1] = 0;
						v114[2] = 0;
						v114[3] = 0;
					} while (v113);
					for (k = 0; k < 624; k += 24) {
						v115 = *(uint32_t*)((char*)v111 + k + 2972);
						if (!(v115 & 0xC0D)) {
							HIDWORD(v5) = ~v115 & *v111;
							*v111 = HIDWORD(v5);
							*(uint32_t*)((char*)v111 + k + 2972) = 0;
							*(uint32_t*)((char*)v111 + k + 2976) = 0;
							*(uint32_t*)((char*)v111 + k + 2980) = 0;
							*(uint32_t*)((char*)v111 + k + 2984) = 0;
							*(uint32_t*)((char*)v111 + k + 2988) = 0;
						}
					}
				}
			}
		}
		return 3;
	case 0xE9u: // MSG_PLAYER_RESPAWN
		if (nox_client_isConnected_43C700()) {
			if (nox_common_gameFlags_check_40A5C0(0x2000)) {
				k = *(unsigned short*)(data + 1);
				if ((unsigned short)k == nox_player_netCode_85319C) {
					sub_45A670(*(uint32_t*)(data + 3));
				}
			}
			if (*(uint8_t*)(data + 8) & 1) {
				nox_xxx_cliPlayerRespawn_417680(
					nox_common_playerInfoGetByID_417040(*(unsigned short*)(data + 1)), *(uint8_t*)(data + 7));
			}
		}
		return 9;
	case 0xEAu: // MSG_FORGET_DRAWABLES
		if (nox_common_gameFlags_check_40A5C0(0x2000)) {
			sub_45A670(*(uint32_t*)(data + 1));
			return 5;
		}
		return 5;
	case 0xEBu: // MSG_RESET_ABILITIES
		return nox_client_decodeProgressPacket(op, data, sz);
	case 0xEDu: // MSG_REPORT_CREATURE_CMD
		return nox_client_decodeProgressPacket(op, data, sz);
	case 0xEEu: // MSG_VOTE
		return nox_client_decodeControlPacket(a1, op, data, sz);
	case 0xF0u: // MSG_GAUNTLET
		return nox_client_decodeQuestPacket(data, sz);
	default:
		return -1;
	}
}
