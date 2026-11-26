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
extern uint32_t dword_5d4594_1046852;
extern uint32_t dword_5d4594_1047532;
extern void* dword_5d4594_1049692;
extern uint32_t dword_5d4594_1046652;
extern uint32_t dword_5d4594_1047536;
extern uint32_t dword_5d4594_1045420;
extern char* dword_5d4594_831240;
extern uint32_t dword_5d4594_1045640;
extern uint32_t dword_5d4594_831260;
extern uint32_t dword_587000_129656;
extern uint32_t dword_5d4594_832480;
extern uint32_t dword_587000_126996;
extern uint32_t dword_5d4594_1045428;
extern uint32_t dword_5d4594_831244;
extern uint32_t dword_5d4594_831076;
extern uint32_t dword_5d4594_1047524;
extern uint32_t dword_5d4594_1046864;
extern uint32_t dword_5d4594_831256;
extern uint32_t dword_5d4594_1049536;
extern uint32_t dword_5d4594_1045424;
extern void* dword_5d4594_1049696;
extern uint32_t dword_5d4594_831276;
extern uint32_t dword_5d4594_1046648;
extern uint32_t dword_5d4594_1046640;
extern uint32_t dword_5d4594_831084;
extern uint32_t dword_5d4594_1047936;
extern float2 obj_5d4594_1046620;
extern uint32_t dword_5d4594_1045436;
extern uint32_t dword_5d4594_831220;
extern uint32_t dword_5d4594_1045460;
extern uint32_t dword_587000_122848;
extern uint32_t dword_5d4594_1046932;
extern uint32_t dword_5d4594_830864;
extern uint32_t dword_5d4594_1047528;
extern uint32_t dword_5d4594_1046636;
extern uintptr_t dword_5d4594_832520;
extern uintptr_t dword_5d4594_832500;
extern uintptr_t dword_5d4594_832528;
extern uintptr_t dword_5d4594_832524;
extern uintptr_t dword_5d4594_832512;
extern uintptr_t dword_5d4594_832496;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132136;
extern uintptr_t dword_5d4594_832516;
extern uintptr_t dword_5d4594_832508;
extern uintptr_t dword_5d4594_832504;
extern uint32_t dword_5d4594_831224;
extern uintptr_t dword_5d4594_832492;
extern uint32_t dword_5d4594_1047512;
extern uintptr_t dword_5d4594_832532;
extern uint32_t dword_5d4594_1046656;
extern uint32_t dword_5d4594_1049484;
extern uintptr_t dword_5d4594_832536;
extern void* dword_587000_81128;
extern void* dword_587000_122852;
extern uint32_t dword_5d4594_1049496;
extern uint32_t dword_5d4594_1047932;
extern nox_window* nox_wnd_briefing_831232;
extern uint32_t dword_5d4594_1045432;
extern void* dword_587000_127004;
extern uint32_t dword_5d4594_1047520;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132132;
extern uint32_t dword_5d4594_1046936;
extern uint32_t dword_5d4594_1047540;
extern nox_window* dword_5d4594_831236;
extern uint32_t dword_5d4594_1046872;
extern unsigned int nox_gameDisableMapDraw_5d4594_2650672;
extern uintptr_t dword_5d4594_1047516;
extern nox_video_bag_image_t* nox_ctf_flag_team_border;
extern uint32_t dword_5d4594_1046868;
extern uintptr_t dword_5d4594_832484;
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
