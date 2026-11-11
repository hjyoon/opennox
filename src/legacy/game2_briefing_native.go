package legacy

/*
#include <math.h>
#include "audio_ambient_native_45a9b0.h"
#include <stdio.h>
#include <time.h>
#include "client/audio/ail/compat_mss.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "quest_preview_44e110.h"
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
#include "client__drawable__drawable.h"
#include "common__system__settings.h"
#include "common__system__team.h"
#include "client__gui__gadgets__listbox.h"
#include "client__gui__gamewin__gamewin.h"
#include "client__gui__gui_ctf.h"
#include "client__gui__guibook.h"
#include "client__gui__guibrief.h"
#include "client__gui__guicon.h"
#include "client__gui__guisave.h"
#include "client__gui__guispell.h"
#include "client__gui__servopts__access.h"
#include "client__gui__servopts__general.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__servopts__playrlst.h"
#include "client__gui__window.h"
#include "client__shell__mainmenu.h"
#include "common__gamemech__pausefx.h"
#include "common__net_list.h"
#include "common__strman.h"
#include "common__crypt.h"
#include "operators.h"
#include "client__audio__audevent.h"
#include "client__video__draw_common.h"
#include "client__system__ctrlevnt.h"
#include "common__magic__speltree.h"
#include "input_common.h"
#include <string.h>
#include "game2_quickbar_native.h"
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
extern void* dword_5d4594_1046360;
extern uint32_t dword_587000_126996;
extern uint32_t dword_5d4594_1045428;
extern uint32_t dword_5d4594_831244;
extern uint32_t dword_5d4594_831076;
extern nox_window* dword_5d4594_1049524;
extern uint32_t dword_5d4594_1047524;
extern void* dword_5d4594_1046356;
extern uint32_t dword_5d4594_1046864;
extern uint32_t dword_5d4594_831256;
extern uint32_t dword_5d4594_1049536;
extern nox_window* dword_5d4594_1046944;
extern uint32_t dword_5d4594_1045424;
extern void* dword_5d4594_1049696;
extern nox_window* dword_5d4594_1046540;
extern uint32_t dword_5d4594_831276;
extern uint32_t dword_5d4594_1046648;
extern uint32_t dword_5d4594_1046640;
extern nox_window* dword_5d4594_1045692;
extern uint32_t dword_5d4594_831084;
extern nox_window* dword_5d4594_1046956;
extern uint32_t dword_5d4594_1047936;
extern float2 obj_5d4594_1046620;
extern nox_window* dword_5d4594_1046536;
extern uint32_t dword_5d4594_1045436;
extern uint32_t dword_5d4594_831220;
extern uint32_t dword_5d4594_1045460;
extern uint32_t dword_587000_122848;
extern uint32_t dword_5d4594_1046932;
extern nox_window* dword_5d4594_1046528;
extern nox_window* dword_5d4594_1046948;
extern uint32_t dword_5d4594_830864;
extern uint32_t dword_5d4594_1047528;
extern uint32_t dword_5d4594_1046636;
extern uintptr_t dword_5d4594_832520;
extern uintptr_t dword_5d4594_832500;
extern uintptr_t dword_5d4594_832528;
extern nox_window* dword_5d4594_1045688;
extern uintptr_t dword_5d4594_832524;
extern uintptr_t dword_5d4594_832512;
extern uintptr_t dword_5d4594_832496;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132136;
extern uintptr_t dword_5d4594_832516;
extern nox_window* dword_5d4594_1046520;
extern nox_window* dword_5d4594_1049532;
extern uintptr_t dword_5d4594_832508;
extern uintptr_t dword_5d4594_832504;
extern uint32_t dword_5d4594_831224;
extern nox_things_imageRef_t* dword_5d4594_1046928;
extern nox_window* dword_5d4594_1045508;
extern nox_window* dword_5d4594_1046504;
extern uintptr_t dword_5d4594_832492;
extern uint32_t dword_5d4594_1047512;
extern uintptr_t dword_5d4594_832532;
extern uint32_t dword_5d4594_1046656;
extern uint32_t dword_5d4594_1049484;
extern nox_window* dword_5d4594_1046516;
extern uintptr_t dword_5d4594_832536;
extern nox_window* dword_5d4594_1046952;
extern void* dword_587000_81128;
extern void* dword_587000_122852;
extern nox_window* dword_5d4594_1046532;
extern uint32_t dword_5d4594_1049496;
extern uint32_t dword_5d4594_1047932;
extern nox_window* dword_5d4594_1049512;
extern nox_window* nox_wnd_briefing_831232;
extern uint32_t dword_5d4594_1045432;
extern nox_things_imageRef_t* dword_5d4594_1046924;
extern void* dword_587000_127004;
extern nox_window* dword_5d4594_1045468;
extern nox_window* dword_5d4594_1046508;
extern nox_window* dword_5d4594_1045480;
extern uint32_t dword_5d4594_1047520;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132132;
extern nox_window* dword_5d4594_1045484;
extern nox_window* dword_5d4594_1045464;
extern nox_window* dword_5d4594_1049520;
extern uint32_t dword_5d4594_1046936;
extern uint32_t dword_5d4594_1047540;
extern nox_window* dword_5d4594_831236;
extern nox_window* dword_5d4594_1046496;
extern uint32_t dword_5d4594_1046872;
extern unsigned int nox_gameDisableMapDraw_5d4594_2650672;
extern nox_window* dword_5d4594_1045636;
extern nox_window* dword_5d4594_1049508;
extern uintptr_t dword_5d4594_1047516;
extern nox_window* dword_5d4594_1049500;
extern nox_window* dword_5d4594_1045604;
extern nox_video_bag_image_t* nox_ctf_flag_team_border;
extern uint32_t dword_5d4594_1046868;
extern nox_window* dword_5d4594_1049504;
extern uintptr_t dword_5d4594_832484;
extern nox_window* dword_5d4594_1045684;
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern nox_window* dword_5d4594_1046492;
extern int nox_win_width;
extern int nox_win_height;
extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_black_2650656;
extern uint32_t nox_color_violet_2598268;
extern uint32_t nox_color_red_2589776;
extern uint32_t nox_color_green_2614268;
extern uint32_t nox_color_cyan_2649820;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_orange_2614256;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t dword_8531A0_2572;
extern void nox_xxx_clientPlaySoundSpecial_native_452D80(int a1, int a2);
extern nox_window* nox_win_unk1;
extern nox_window* nox_wnd_briefing_831232;
extern unsigned int nox_gameDisableMapDraw_5d4594_2650672;
extern uint32_t dword_5d4594_831092;
int nox_client_getBriefDuration(void);
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"unsafe"
)

type game2BriefingHooks struct {
	Thing        func(int32) unsafe.Pointer
	Image        func(string) unsafe.Pointer
	DeleteSprite func(unsafe.Pointer) int32
}

var game2BriefingTestHooks *game2BriefingHooks

func g2bThing(id C.int) *C.nox_thing {
	if h := game2BriefingTestHooks; h != nil && h.Thing != nil {
		return (*C.nox_thing)(h.Thing(int32(id)))
	}
	return C.nox_get_thing(id)
}
func g2bImage(name string) *C.nox_video_bag_image_t {
	if h := game2BriefingTestHooks; h != nil && h.Image != nil {
		return (*C.nox_video_bag_image_t)(h.Image(name))
	}
	return C.nox_xxx_gLoadImg_42F970(internCStr(name))
}
func g2bDelete(p unsafe.Pointer) int32 {
	if h := game2BriefingTestHooks; h != nil && h.DeleteSprite != nil {
		return h.DeleteSprite(p)
	}
	return int32(C.nox_xxx_spriteDelete_45A4B0((*C.nox_drawable)(p)))
}
func g2bPosition() *float32 { return (*float32)(unsafe.Pointer(&C.dword_5d4594_831276)) }
func g2bEvent(win *C.nox_window, event C.int, arg C.uintptr_t) C.uintptr_t {
	return C.nox_window_call_field_94_fnc(win, event, arg, 0, internCStr("GAME2.c"), 0)
}

//export sub_44D040
func sub_44D040(id C.int) C.int {
	p := g2bThing(id)
	if p == nil {
		return 0
	}
	return C.int((p.pri_class >> 22) & 1)
}

//export sub_44D060
func sub_44D060(id C.int) C.int {
	p := g2bThing(id)
	if p != nil && p.pri_class&0x400000 != 0 && p.sub_class&0x18 != 0 {
		return 1
	}
	return 0
}

//export sub_44D090
func sub_44D090(id C.int) C.int {
	p := g2bThing(id)
	if p != nil && p.pri_class&0x20400000 != 0 {
		return 1
	}
	return 0
}

//export sub_44D960
func sub_44D960() { C.dword_587000_122848 = 0 }

//export sub_44D970
func sub_44D970() C.int {
	v := C.dword_5d4594_831092
	if v != 0 {
		C.dword_587000_122848 = 1
	}
	return C.int(v)
}

//export sub_44D990
func sub_44D990() C.int { return C.int(C.dword_587000_122848) }

//export sub_44E110
func sub_44E110() *C.uint32_t { return nox_client_questPreview_native_44E110() }

//export nox_xxx_playGMCAPsmth_44E3E0
func nox_xxx_playGMCAPsmth_44E3E0() C.int {
	result := C.int(C.dword_5d4594_831224)
	*memmap.PtrUint32(0x5D4594, 831248) = 1
	if result != 0 {
		result = C.nox_xxx_playDialogFile_44D900((*C.uchar)(unsafe.Pointer(C.dword_5d4594_831240)), 100)
		C.dword_5d4594_831244 = 1
	}
	return result
}

//export sub_44E560
func sub_44E560() *C.nox_window {
	*memmap.PtrUint32(0x5D4594, 831284) = uint32((C.nox_win_width - 640) / 2)
	*memmap.PtrUint32(0x5D4594, 831288) = uint32((C.nox_win_height - 480) / 2)
	result := C.nox_window_new(nil, 56, 0, 0, C.nox_win_width, C.nox_win_height, (*[0]byte)(unsafe.Pointer(C.nox_xxx_wndProc_44E6E0)))
	C.dword_5d4594_831236 = result
	if result != nil {
		C.nox_xxx_wndSetWindowProc_46B300(result, unsafe.Pointer(C.nox_client_wndQuestBriefProc_44E630))
		result.draw_data.bg_color = C.nox_color_black_2650656
		C.nox_wnd_briefing_831232 = C.nox_new_window_from_file(internCStr("Briefing.wnd"), nil)
		result = C.nox_wnd_briefing_831232
		if result != nil {
			C.sub_46B120(result, C.dword_5d4594_831236)
			C.nox_window_setPos_46A9B0(result, C.int(memmap.Int32(0x5D4594, 831284)), C.int(memmap.Int32(0x5D4594, 831288)))
			child := C.nox_xxx_wndGetChildByID_46B0C0(result, 1010)
			C.nox_xxx_wndSetDrawFn_46B340(child, unsafe.Pointer(C.sub_44E6F0))
			C.sub_44E410()
			result = C.dword_5d4594_831236
		}
	}
	return result
}

//export nox_client_wndQuestBriefProc_44E630
func nox_client_wndQuestBriefProc_44E630(unused1 *C.nox_window, event C.int, unused2 C.uintptr_t, unused3 C.uintptr_t) C.int {
	if memmap.Uint32(0x5D4594, 831248) != 0 && event != 18 && event != 17 && C.sub_41D1B0() != 1 {
		C.sub_450580()
		if C.dword_5d4594_831220 != 0 {
			if C.dword_5d4594_831220 == 255 {
				C.nox_gameDisableMapDraw_5d4594_2650672 = 0
				duration := C.nox_client_getIntroScreenDuration_44E3B0()
				C.nox_video_fadeInScreen_44DAB0(duration, 1, unsafe.Pointer(C.sub_44E320))
				C.nox_gameDisableMapDraw_5d4594_2650672 = 1
			}
		} else {
			C.nox_xxx_wndClearCaptureMain_46ADE0(C.dword_5d4594_831236)
			C.nox_xxx_windowFocus_46B500(nil)
			C.dword_5d4594_831256 = 1
			C.nox_savegame_sub_46D580()
		}
		C.nox_xxx_wndSetWindowProc_46B300(C.dword_5d4594_831236, nil)
	}
	return 1
}

//export nox_xxx_wndProc_44E6E0
func nox_xxx_wndProc_44E6E0(unused1 *C.nox_window, event C.int, unused2 C.uintptr_t, unused3 C.uintptr_t) C.int {
	if event == 23 {
		return 1
	}
	return 0
}

//export sub_44E6F0
func sub_44E6F0(win *C.nox_window, draw *C.nox_window_data) C.int {
	*g2bPosition() = float32(float64(*g2bPosition()) - float64(sub_44E8B0()))
	y := C.nox_float2int(C.float(*g2bPosition()))
	C.nox_window_setPos_46A9B0(win, 0, y)
	C.nox_client_copyRect_49F6F0(C.int(memmap.Int32(0x5D4594, 831284)), C.int(memmap.Int32(0x5D4594, 831288)), 640, 480)
	flags := memmap.Uint8(0x5D4594, 832472)
	if flags&1 != 0 {
		C.sub_44E8E0(win, draw)
	} else if flags&2 != 0 {
		C.sub_44F0F0(win, draw)
	} else if flags&4 != 0 {
		C.sub_44F300(win, draw)
	} else {
		C.nox_xxx_wndStaticDrawNoImage_488D00(win, draw)
	}
	C.nox_client_copyRect_49F6F0(0, 0, C.nox_getBackbufWidth(), C.nox_getBackbufHeight())
	if int64(*g2bPosition()) <= int64(memmap.Int32(0x5D4594, 831280)) && C.dword_5d4594_831244 == 1 && C.sub_44D930() == 0 && uint64(C.nox_platform_get_ticks())-memmap.Uint64(0x5D4594, 831292) > uint64(C.nox_client_getBriefDuration()) && (memmap.Uint32(0x5D4594, 832488) == 1 || flags&5 == 0) {
		if flags&2 != 0 && C.dword_5d4594_832480 != 0 {
			g2qSound(582, 100)
		}
		C.sub_450580()
		if C.dword_5d4594_831220 != 0 {
			C.nox_gameDisableMapDraw_5d4594_2650672 = 0
			duration := C.nox_client_getIntroScreenDuration_44E3B0()
			C.nox_video_fadeInScreen_44DAB0(duration, 1, unsafe.Pointer(C.sub_44E320))
			C.nox_gameDisableMapDraw_5d4594_2650672 = 1
		} else if C.dword_5d4594_831256 == 0 {
			C.nox_xxx_wndClearCaptureMain_46ADE0(C.dword_5d4594_831236)
			C.dword_5d4594_831256 = 1
			C.nox_savegame_sub_46D580()
			C.nox_xxx_wndSetWindowProc_46B300(C.dword_5d4594_831236, nil)
		}
		C.nox_xxx_wndSetDrawFn_46B340(win, unsafe.Pointer(C.sub_44E8D0))
	}
	return 1
}

//export sub_44E8B0
func sub_44E8B0() C.double {
	if C.dword_5d4594_831220 == 255 {
		return 1
	}
	return 0
}

//export sub_44E8D0
func sub_44E8D0(unused1 *C.nox_window, unused2 *C.nox_window_data) C.int { return 1 }

//export nox_client_lockScreenBriefing_450160
func nox_client_lockScreenBriefing_450160(screen, mode C.int, flags C.char) C.int {
	player := g2qPlayer()
	C.dword_5d4594_831260 = 1
	C.dword_5d4594_831244 = 1
	*memmap.PtrUint32(0x5D4594, 832488) = 0
	*memmap.PtrUint32(0x5D4594, 832472) = 0
	C.dword_5d4594_831256 = 0
	*memmap.PtrUint8(0x5D4594, 831252) = byte(screen + 1)
	class := mode
	if player != nil {
		class = C.int(game2PlayerClass(player))
	}
	outer := C.dword_5d4594_831236
	C.nox_xxx_wndShowModalMB_46A8C0(outer)
	C.nox_xxx_wndSetCaptureMain_46ADC0(outer)
	C.nox_xxx_windowFocus_46B500(outer)
	C.nox_xxx_wndSetWindowProc_46B300(outer, unsafe.Pointer(C.nox_client_wndQuestBriefProc_44E630))
	textWin := C.nox_xxx_wndGetChildByID_46B0C0(C.nox_wnd_briefing_831232, 1010)
	C.nox_xxx_wndSetDrawFn_46B340(textWin, unsafe.Pointer(C.sub_44E6F0))
	C.dword_5d4594_831220 = C.uint32_t(mode)
	sound := C.int(0)
	switch screen {
	case 255:
		C.nox_xxx_wndSetIcon_46AE60(C.nox_wnd_briefing_831232, nil)
		g2bEvent(textWin, 16385, C.uintptr_t(uintptr(*memmap.PtrPtr(0x5D4594, 831268))))
		C.nox_xxx_wndClearFlag_46AD80(textWin, 0x2000)
		C.dword_5d4594_831240 = C.nox_xxx_GetEndgameDialog_578D80()
		sound = 24
		C.dword_5d4594_831220 = 255
	case 254:
		C.nox_server_currentMapGetFilename_409B30()
		C.nox_xxx_wndSetIcon_46AE60(C.nox_wnd_briefing_831232, g2bImage("GauntletStartMines"))
		g2bEvent(textWin, 16385, C.uintptr_t(uintptr(memmap.PtrOff(0x5D4594, 832540))))
		if byte(flags)&1 != 0 {
			*memmap.PtrUint32(0x5D4594, 832472) = 1
			C.nox_xxx_wndSetIcon_46AE60(C.nox_wnd_briefing_831232, g2bImage("MenuSystemBG"))
		} else if byte(flags)&4 != 0 {
			*memmap.PtrUint32(0x5D4594, 832472) = 4
			C.nox_xxx_wndSetIcon_46AE60(C.nox_wnd_briefing_831232, g2bImage("GauntletInstructionBackground"))
		} else {
			*memmap.PtrUint32(0x5D4594, 832472) = 2
			C.nox_xxx_wndSetIcon_46AE60(C.nox_wnd_briefing_831232, (*C.nox_video_bag_image_t)(*memmap.PtrPtr(0x5D4594, 832460)))
			if text := *memmap.PtrPtr(0x5D4594, 832464); text != nil {
				g2bEvent(textWin, 16385, C.uintptr_t(uintptr(text)))
			}
		}
		C.nox_client_resetScreenParticles_431510()
		C.nox_xxx_bookHideMB_45ACA0(1)
		C.sub_446780()
		C.dword_5d4594_831240 = nil
		C.dword_5d4594_831244 = 1
		C.dword_5d4594_831224 = 0
	default:
		offset := uintptr(32 * (screen + class + 10*class))
		base := uintptr(831300)
		if mode == 0 {
			base = 831316
		}
		C.nox_xxx_wndSetIcon_46AE60(C.nox_wnd_briefing_831232, (*C.nox_video_bag_image_t)(*memmap.PtrPtr(0x5D4594, base+offset)))
		g2bEvent(textWin, 16385, C.uintptr_t(uintptr(*memmap.PtrPtr(0x5D4594, base+4+offset))))
		C.dword_5d4594_831240 = (*C.char)(*memmap.PtrPtr(0x5D4594, base+8+offset))
		sound = C.int(memmap.Uint32(0x5D4594, base+12+offset))
	}
	text := *(**C.wchar2_t)(textWin.widget_data)
	height := (*C.int)(memmap.PtrOff(0x5D4594, 831280))
	C.nox_xxx_drawGetStringSize_43F840(textWin.draw_data.font, text, nil, height, 480)
	C.sub_46AB20(textWin, 640, *height)
	if screen == 255 {
		C.dword_5d4594_831276 = 1140457472
		*height = -20 - *height
	} else {
		*height = (480 - C.nox_xxx_guiFontHeightMB_43F320(textWin.draw_data.font) - *height) / 2
		*g2bPosition() = float32(*height)
	}
	C.sub_431290()
	C.dword_5d4594_831224 = 0
	if mode != 0 || screen == 255 {
		if screen != 254 {
			C.dword_5d4594_831224 = 1
		}
	} else if screen != 254 {
		if player.field_4408_4576[screen] != 0 {
			C.nox_xxx_wndClearCaptureMain_46ADE0(outer)
			C.dword_5d4594_831256 = 1
			C.nox_savegame_sub_46D580()
			C.nox_xxx_wndSetWindowProc_46B300(outer, nil)
		} else {
			C.dword_5d4594_831224 = 1
			player.field_4408_4576[screen] = 1
		}
	}
	C.sub_43DD70(sound, 50)
	*memmap.PtrUint64(0x5D4594, 831292) = uint64(C.nox_platform_get_ticks())
	*memmap.PtrUint32(0x5D4594, 831248) = 0
	duration := C.nox_client_getIntroScreenDuration_44E3B0()
	C.nox_video_fadeOutScreen_44DB30(duration, 1, unsafe.Pointer(C.nox_xxx_playGMCAPsmth_44E3E0))
	if C.dword_5d4594_831224 != 0 {
		C.dword_5d4594_831244 = 0
	}
	return C.int(C.dword_5d4594_831224)
}

//export sub_4505E0
func sub_4505E0() C.int {
	if win := C.dword_5d4594_831236; win != nil {
		g2qDestroy(win)
		C.dword_5d4594_831236 = nil
		C.nox_wnd_briefing_831232 = nil
	}
	C.dword_5d4594_831260 = 0
	C.dword_5d4594_832484 = 0
	slots := questPreviewSlots44E110()
	result := C.int(0)
	for _, index := range []int{3, 0, 2, 1, 4, 5, 6, 7, 8, 9, 10, 11} {
		slot := slots[index]
		if *slot != nil {
			v := g2bDelete(unsafe.Pointer(*slot))
			if index == 11 {
				result = C.int(v)
			}
		}
		*slot = nil
	}
	return result
}

//export sub_450750
func sub_450750() C.uchar { return C.uchar(memmap.Uint8(0x5D4594, 831252)) }

//export sub_450760
func sub_450760(value C.char) C.char { *memmap.PtrUint8(0x5D4594, 831252) = byte(value); return value }

//export nox_game2_quest_score_compare_native
func nox_game2_quest_score_compare_native(a, b unsafe.Pointer) C.int {
	x := (*C.nox_quest_stats_row_450770)(a).score
	y := (*C.nox_quest_stats_row_450770)(b).score
	if x == y {
		return 0
	}
	if x < y {
		return 1
	}
	return -1
}

//export sub_450AD0
func sub_450AD0(image *C.nox_video_bag_image_t) *C.nox_video_bag_image_t {
	if image == nil {
		image = g2bImage("WarriorChapterBegin8")
	}
	*memmap.PtrPtr(0x5D4594, 832460) = unsafe.Pointer(image)
	return image
}

//export sub_450AF0
func sub_450AF0(text *C.wchar2_t) *C.wchar2_t {
	*memmap.PtrPtr(0x5D4594, 832464) = unsafe.Pointer(text)
	return text
}

//export nox_gui_setQuestStage_450B00
func nox_gui_setQuestStage_450B00(value C.int) { *memmap.PtrUint32(0x5D4594, 832468) = uint32(value) }

//export nox_gui_getQuestStage_450B10
func nox_gui_getQuestStage_450B10() C.int { return C.int(memmap.Uint32(0x5D4594, 832468)) }
