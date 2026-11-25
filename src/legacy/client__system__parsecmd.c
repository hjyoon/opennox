#include "client__system__parsecmd.h"

#include "client__gui__guicon.h"
#include "client__gui__guimsg.h"
#include "client__gui__servopts__general.h"
#include "client__gui__servopts__guiserv.h"
#include "common__log.h"
#include "common__strman.h"
#include "common__system__settings.h"
#include "server__script__script.h"
#include "common__system__team.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5_2.h"
#include "client__video__draw_common.h"
#include "common__magic__speltree.h"
#include "player_level_set_4ef410.h"

extern uint32_t nox_server_connectionType_3596;
extern uint32_t dword_5d4594_2650652;
extern uint32_t nox_player_netCode_85319C;

unsigned int nox_client_consoleIsServer_823684 = 0;

//----- (00440A20) --------------------------------------------------------
void sub_440A20(wchar2_t* a1, ...) {
	va_list va; // [esp+8h] [ebp+8h]

	va_start(va, a1);
	nox_vswprintf((wchar2_t*)getMemAt(0x5D4594, 822660), a1, va);
	nox_xxx_printCentered_445490((wchar2_t*)getMemAt(0x5D4594, 822660));
}
