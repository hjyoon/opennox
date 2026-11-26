#ifndef OPENNOX_GAME2_SERVER_BRIDGE_H
#define OPENNOX_GAME2_SERVER_BRIDGE_H
#include <stdlib.h>
#include <stdint.h>
#include <math.h>
#include <stdio.h>
#include <time.h>
#include "client/audio/ail/compat_mss.h"
#include "quest_preview_44e110.h"
#include "client__gui__gadgets__listbox.h"
#include "client__gui__guibook.h"
#include "client__gui__guibrief.h"
#include "client__gui__guicon.h"
#include "client__gui__guisave.h"
#include "client__gui__guispell.h"
#include "client__gui__servopts__general.h"
#include "client__shell__mainmenu.h"
#include "common__gamemech__pausefx.h"
#include "common__net_list.h"
#include "common__crypt.h"
#include "client__audio__audevent.h"
#include "client__system__ctrlevnt.h"
#include "input_common.h"
#include <string.h>
#include "operators.h"
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
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME5_2.h"
#include "common__strman.h"
#include "common__system__team.h"
#include "common__system__settings.h"
#include "client__gui__window.h"
#include "client__drawable__drawable.h"
#include "client__gui__servopts__access.h"
#include "client__gui__servopts__playrlst.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__gui_ctf.h"
#include "client__gui__gamewin__gamewin.h"
#include "client__video__draw_common.h"
#include "common__magic__speltree.h"
#include "audio_ambient_native_45a9b0.h"
void sub_4537F0(void);
extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_5d4594_1045420;
extern uint32_t dword_5d4594_831076;
extern uint32_t dword_5d4594_831084;
extern uint32_t dword_5d4594_830864;
extern void* dword_587000_127004;
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern int nox_win_width;
extern int nox_win_height;
extern uint32_t dword_8531A0_2572;
extern void nox_xxx_clientPlaySoundSpecial_native_452D80(int a1, int a2);
// Keep the source location used by the C window dispatch macro.
static uintptr_t game2_server_event(nox_window* w,int e,uintptr_t a,uintptr_t b) {
 return nox_window_call_field_94(w,e,a,b);
}
#endif
