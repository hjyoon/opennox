package legacy

/*
#include <math.h>
#include "defs.h"
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
#include "client__gui__window.h"
#include "client__gui__guiquit.h"
#include "client__video__draw_common.h"
#include "input.h"
#include "operators.h"
extern nox_window *dword_5d4594_1309812, *dword_5d4594_1309820, *dword_5d4594_1309824;
extern nox_window *dword_5d4594_1309828, *dword_5d4594_1309832, *dword_5d4594_1309836;
extern uint32_t nox_server_sendMotd_108752, nox_server_connectionType_3596, dword_5d4594_2650652;
extern void *dword_587000_127004, *dword_587000_122852, *dword_587000_93164;
extern uint32_t dword_5d4594_1313532, dword_5d4594_1313536, dword_5d4594_1313540;
extern uint32_t dword_5d4594_1313564, dword_5d4594_1313740, dword_5d4594_1313788;
extern uint64_t qword_581450_9552;
extern int nox_win_width;
int nox_gui_menu_proc_ext(int);
void nox_video_setMenuOptions(nox_window*);
void nox_common_writecfgfile(char*);
void sub_4AA650(void);
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

var game3MapFactory = game3MapRuntime

func game3MapLoadError()               { C.nox_xxx_spriteLoadError_4356E0() }
func game3MapFloatInt(v float32) int32 { return int32(C.nox_float2int(C.float(v))) }
func game3MapTOC(v uint16) uint16      { return uint16(C.nox_xxx_objectTOCgetTT_42C2B0(C.ushort(v))) }
func game3MapClass(v uint16, kind int) bool {
	switch kind {
	case 0:
		return C.sub_44D060(C.int(v)) != 0
	case 1:
		return C.sub_44D040(C.int(v)) != 0
	default:
		return C.sub_44D090(C.int(v)) != 0
	}
}
func game3AttachWall(d *client.Drawable, x, y int32) {
	C.sub_410390((*C.nox_drawable)(d.C()), C.int(x), C.int(y))
}

//export sub_4ABDA0
func sub_4ABDA0(id C.int, version, mapVersion C.short, count *C.uint32_t) *C.nox_drawable {
	return (*C.nox_drawable)(game3LegacySpriteLoad(int32(id), int16(version), int16(mapVersion), (*uint32)(unsafe.Pointer(count)), game3MapFactory()).C())
}

//export nox_xxx_spriteLoadFromMap_4AC020
func nox_xxx_spriteLoadFromMap_4AC020(id C.int, version C.short, count *C.uint32_t) *C.nox_drawable {
	return (*C.nox_drawable)(game3SpriteLoad(int32(id), int16(version), (*uint32)(unsafe.Pointer(count)), game3MapFactory()).C())
}

//export nox_client_mapSpecialRWObjectData_4AC610
func nox_client_mapSpecialRWObjectData_4AC610() C.int {
	return C.int(game3SpecialObjectData(game3MapFactory()))
}

//export nox_xxx_clientLoadSomeObject_4AC6E0
func nox_xxx_clientLoadSomeObject_4AC6E0(id C.ushort) C.int {
	return C.int(game3ClientLoadSome(uint16(id), game3MapFactory()))
}

//export sub_4AC7B0
func sub_4AC7B0(id C.int) C.int { return C.int(game3LoadObject(int32(id), 0, game3MapFactory())) }

//export nox_xxx_colorLightClientLoad_4AC980
func nox_xxx_colorLightClientLoad_4AC980(id C.int) C.int {
	return C.int(game3LoadObject(int32(id), 1, game3MapFactory()))
}

//export nox_xxx_cliLoadTeamBase_4ACE00
func nox_xxx_cliLoadTeamBase_4ACE00(id C.int) C.int {
	return C.int(game3LoadObject(int32(id), 2, game3MapFactory()))
}

//export sub_4ACEF0
func sub_4ACEF0(id C.int) C.int { return C.int(game3LoadObject(int32(id), 3, game3MapFactory())) }

//export sub_4AD040
func sub_4AD040(id C.int) C.int { return C.int(game3LoadObject(int32(id), 4, game3MapFactory())) }

//export sub_4AD570
func sub_4AD570(win *C.nox_window, draw *C.nox_window_data) C.int {
	if C.dword_5d4594_2650652 != 1 || C.dword_5d4594_1309812 == nil {
		return 1
	}
	list := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1309812, 10317)
	p := GetClient().GetMousePos()
	if list != nil && C.wndIsShown_nox_xxx_wndIsShown_46ACC0(list) == 0 && !bool(C.nox_xxx_wndPointInWnd_46AAB0(list, C.int(p.X), C.int(p.Y))) {
		C.nox_xxx_wndClearCaptureMain_46ADE0(list)
		C.nox_window_set_hidden(list, 1)
	}
	return 1
}

//export nox_xxx_windowServerOptionsGeneralProc_4AD5D0
func nox_xxx_windowServerOptionsGeneralProc_4AD5D0(win *C.nox_window, event C.int, sender *C.nox_window, arg C.uintptr_t) C.int {
	if event == 16391 {
		if sender == nil {
			return 0
		}
		id := C.nox_xxx_wndGetID_46B0A0(sender)
		C.nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		switch id {
		case 10301:
			if C.nox_server_doPlayersAutoRespawn_40A5F0() != 0 {
				C.nox_xxx_ruleSetNoRespawn_40A5E0(0)
			} else {
				C.nox_xxx_ruleSetNoRespawn_40A5E0(1)
			}
		case 10302:
			C.nox_server_sendMotd_108752 ^= 1
		case 10304:
			if C.sub_4D0D70() != 0 {
				C.sub_4D0D90(0)
			} else {
				C.sub_4D0D90(1)
			}
		case 10305:
			C.sub_409EF0(2)
		case 10306:
			C.sub_409EF0(0x2000)
			if C.sub_409F40(0x2000) == 0 {
				C.sub_4D7EA0()
			}
		case 10316:
			p := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1309812, 10317)
			if p != nil {
				C.nox_window_set_hidden(p, 0)
				C.nox_xxx_wndShowModalMB_46A8C0(p)
				C.nox_xxx_wndSetCaptureMain_46ADC0(p)
			}
		case 10319:
			C.sub_4BDFD0()
		}
		return 0
	}
	if event == 16393 {
		C.nox_xxx_rateUpdate_40A6D0(C.int(4 - int32(arg)))
		return 0
	}
	if event != 16400 || sender == nil || C.dword_5d4594_1309812 == nil || C.nox_xxx_wndGetID_46B0A0(sender) != 10317 {
		return 0
	}
	data := (*C.nox_scrollListBox_data)(sender.widget_data)
	selected := int32(C.nox_window_call_field_94_fnc(sender, 16404, 0, 0, internCStr("game3_map_options_services.go"), 0))
	if data == nil || selected < 0 || selected >= int32(data.field_11_0) {
		return 0
	}
	typ := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1309812, 10316)
	rate := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1309812, 10312)
	if typ == nil || rate == nil {
		return 0
	}
	text := C.nox_window_call_field_94_fnc(sender, 16406, arg, 0, internCStr("game3_map_options_services.go"), 0)
	if text == 0 {
		return 0
	}
	C.nox_window_call_field_94_fnc(typ, 16385, text, ^C.uintptr_t(0), internCStr("game3_map_options_services.go"), 0)
	C.nox_server_connectionType_3596 = C.uint32_t(selected + 1)
	C.nox_xxx_rateUpdate_40A6D0(C.sub_40A710(C.int(selected + 1)))
	C.nox_window_call_field_94_fnc(rate, 16394, C.uintptr_t(4-C.nox_xxx_rateGet_40A6C0()), 0, internCStr("game3_map_options_services.go"), 0)
	C.nox_window_set_hidden(sender, 1)
	C.nox_xxx_wndClearCaptureMain_46ADE0(sender)
	return 0
}

//export sub_4AD820
func sub_4AD820() C.int {
	if C.dword_5d4594_1309812 != nil {
		C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1309812)
	}
	C.dword_5d4594_1309812 = nil
	return C.sub_4BE610()
}

//export sub_4AD9B0
func sub_4AD9B0(skip C.int) C.int {
	if C.wndIsShown_nox_xxx_wndIsShown_46ACC0(C.dword_5d4594_1309820) != 0 {
		return 0
	}
	C.sub_413A00(0)
	C.sub_44D8F0()
	if skip == 0 {
		C.nox_video_applyOptions()
		C.nox_common_writecfgfile(internCStr("nox.cfg"))
	}
	C.nox_window_set_hidden(C.dword_5d4594_1309820, 1)
	C.nox_window_set_hidden(C.dword_5d4594_1309824, 1)
	if skip == 0 {
		C.sub_445C40()
	}
	return 1
}

//export sub_4ADA40
func sub_4ADA40() C.int {
	C.sub_413A00(1)
	C.nox_xxx_wndShowModalMB_46A8C0(C.dword_5d4594_1309820)
	C.nox_xxx_wndShowModalMB_46A8C0(C.dword_5d4594_1309824)
	C.nox_client_advVideoOptsLoad_4CB330()
	cut := C.nox_video_getCutSize_4766D0()
	id := C.int(311)
	if cut > 69 {
		id = 312
		if cut > 79 {
			id = 313
			if cut > 89 {
				id = 314
			}
		}
	}
	if w := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1309820, id); w != nil {
		C.nox_window_call_field_94_fnc(w, 16392, 0, 0, internCStr("game3_map_options_services.go"), 0)
	}
	return C.nox_draw_setTabWidth_43FE20(15)
}

func game3IngameOptionsSlider(w *C.nox_window, state unsafe.Pointer, image, lit *C.nox_video_bag_image_t) bool {
	if w == nil || w.field_100 == nil || state == nil {
		return false
	}
	w.field_100.width = 24
	w.field_100.height = 20
	C.sub_4B5700(w, nil, nil, unsafe.Pointer(image), unsafe.Pointer(lit), unsafe.Pointer(lit))
	C.nox_window_call_field_94_fnc(w, 16395, 0, 0x4000, internCStr("game3_map_options_services.go"), 0)
	C.nox_window_call_field_94_fnc(w, 16394, C.uintptr_t(*(*uint32)(unsafe.Add(state, 4))>>16), 0, internCStr("game3_map_options_services.go"), 0)
	return true
}
func game3IngameOptionsChecked(w *C.nox_window, on C.int) {
	if w == nil {
		return
	}
	if on == 1 {
		w.draw_data.field_0 |= 4
	} else {
		w.draw_data.field_0 &^= 4
	}
}
func game3IngameOptionsToggle(w *C.nox_window) {
	if w != nil {
		C.nox_window_call_field_93(w, 21, 28, 2)
	}
}

//export nox_game_initOptionsInGame_4ADAD0
func nox_game_initOptionsInGame_4ADAD0() C.int {
	image := C.nox_xxx_gLoadImg_42F970(internCStr("OptionsVolumeSlider"))
	lit := C.nox_xxx_gLoadImg_42F970(internCStr("OptionsVolumeSliderLit"))
	C.dword_5d4594_1309820 = C.nox_new_window_from_file(internCStr("Options.wnd"), unsafe.Pointer(C.nox_xxx_windowOptionsProc_4ADF30))
	root := C.dword_5d4594_1309820
	if root == nil || C.nox_client_advVideoOpts_New_4CB590(root) == 0 {
		return 0
	}
	if !game3IngameOptionsSlider(C.nox_xxx_wndGetChildByID_46B0C0(root, 351), C.dword_587000_127004, image, lit) {
		return 0
	}
	C.dword_5d4594_1309828 = C.nox_xxx_wndGetChildByID_46B0C0(root, 361)
	game3IngameOptionsChecked(C.dword_5d4594_1309828, C.sub_453070())
	if !game3IngameOptionsSlider(C.nox_xxx_wndGetChildByID_46B0C0(root, 352), C.dword_587000_122852, image, lit) {
		return 0
	}
	C.dword_5d4594_1309836 = C.nox_xxx_wndGetChildByID_46B0C0(root, 362)
	game3IngameOptionsChecked(C.dword_5d4594_1309836, C.sub_44D990())
	if !game3IngameOptionsSlider(C.nox_xxx_wndGetChildByID_46B0C0(root, 353), C.dword_587000_93164, image, lit) {
		return 0
	}
	C.dword_5d4594_1309832 = C.nox_xxx_wndGetChildByID_46B0C0(root, 363)
	game3IngameOptionsChecked(C.dword_5d4594_1309832, C.sub_43DC30())
	C.nox_video_setMenuOptions(root)
	id := C.int(334)
	if C.nox_video_getFullScreen() != 0 {
		id = 333
	}
	if w := C.nox_xxx_wndGetChildByID_46B0C0(root, id); w != nil {
		C.nox_window_call_field_94_fnc(w, 16392, 1, 0, internCStr("game3_map_options_services.go"), 0)
	}
	for id := C.int(320); id <= 332; id++ {
		if w := C.nox_xxx_wndGetChildByID_46B0C0(root, id); w != nil && w.draw_data.field_0&4 == 0 {
			C.nox_xxx_wndClearFlag_46AD80(w, 8)
		}
	}
	C.dword_5d4594_1309824 = C.nox_window_new(nil, 32, 0, 0, 1, 1, nil)
	back := C.dword_5d4594_1309824
	if back == nil {
		return 0
	}
	x := memmap.Int32(0x587000, 174072)
	off := uintptr(174080)
	for x != -1 {
		w := C.nox_window_new(back, 0, C.int(x), C.int(memmap.Int32(0x587000, off-4)), C.int(memmap.Int32(0x587000, off)), C.int(memmap.Int32(0x587000, off+4)), nil)
		if w == nil {
			return 0
		}
		C.nox_xxx_wndSetDrawFn_46B340(w, unsafe.Pointer(C.sub_4ADEF0))
		x = memmap.Int32(0x587000, off+8)
		off += 16
	}
	left := (int(C.nox_win_width) - int(root.width)) / 2
	C.nox_window_setPos_46A9B0(root, C.int(left), 0)
	C.nox_window_setPos_46A9B0(back, C.int(left), 0)
	C.nox_window_set_hidden(C.nox_xxx_wndGetChildByID_46B0C0(root, 371), 0)
	C.nox_window_set_hidden(root, 1)
	C.nox_window_set_hidden(back, 1)
	return 1
}

//export sub_4ADEF0
func sub_4ADEF0(win *C.nox_window, draw *C.nox_window_data) C.int {
	w := asWindow(win)
	p := w.GlobalPos()
	s := w.Size()
	GetClient().R2().DrawRectFilledAlpha(p.X, p.Y, s.X, s.Y)
	return 1
}

//export nox_xxx_windowOptionsProc_4ADF30
func nox_xxx_windowOptionsProc_4ADF30(win *C.nox_window, event C.int, sender *C.nox_window, arg C.uintptr_t) C.int {
	switch event {
	case 16389:
		C.nox_xxx_clientPlaySoundSpecial_452D80(920, 100)
		return 1
	case 16391:
		id := C.nox_xxx_wndGetID_46B0A0(sender)
		if id >= 380 {
			return C.nox_gui_menu_proc_ext(id)
		}
		switch id {
		case 311:
			C.nox_draw_setCutSize_476700(65, 0)
		case 312:
			C.nox_draw_setCutSize_476700(75, 0)
		case 313:
			C.nox_draw_setCutSize_476700(85, 0)
		case 314:
			C.nox_draw_setCutSize_476700(100, 0)
		case 341:
			sub_4AD9B0(0)
			C.sub_445C40()
			C.sub_4C4260()
		case 361:
			if C.sub_453070() == 1 {
				C.sub_453050()
			} else {
				C.nox_xxx____setargv_9_453060()
			}
		case 362:
			if C.sub_44D990() == 1 {
				C.sub_44D960()
			} else {
				C.sub_44D970()
			}
		case 363:
			if C.sub_43DC30() == 1 {
				C.sub_43DC00()
			} else {
				C.sub_43DC10()
				C.sub_486320(C.dword_587000_93164, C.int(*(*uint32)(unsafe.Add(C.dword_587000_93164, 4))>>16))
			}
		case 371:
			sub_4AD9B0(0)
		}
		C.nox_xxx_clientPlaySoundSpecial_452D80(921, 100)
		return 1
	case 16393, 16396:
		id := C.nox_xxx_wndGetID_46B0A0(sender)
		if event == 16396 && id != 351 && id != 352 {
			return 0
		}
		var active *C.nox_window
		if event == 16393 {
			active = C.nox_client_getWin1064916_46C720()
		}
		value := C.int(int32(arg))
		switch id {
		case 351:
			C.sub_486320(C.dword_587000_127004, value)
			if event == 16393 && active != nil && active.parent == sender {
				return 0
			}
			if arg != 0 {
				if C.sub_453070() == 0 {
					game3IngameOptionsToggle(C.dword_5d4594_1309828)
				}
				C.nox_xxx_clientPlaySoundSpecial_452D80(768, 100)
			} else if C.sub_453070() == 1 {
				game3IngameOptionsToggle(C.dword_5d4594_1309828)
			}
		case 352:
			C.sub_486320(C.dword_587000_122852, value)
			if arg != 0 {
				if C.sub_44D990() == 0 {
					game3IngameOptionsToggle(C.dword_5d4594_1309836)
				}
				C.sub_4AA650()
			} else if C.sub_44D990() == 1 {
				game3IngameOptionsToggle(C.dword_5d4594_1309836)
			}
		case 353:
			C.sub_486320(C.dword_587000_93164, value)
			if arg != 0 {
				if C.sub_43DC30() == 0 {
					game3IngameOptionsToggle(C.dword_5d4594_1309832)
				}
			} else if C.sub_43DC30() == 1 {
				game3IngameOptionsToggle(C.dword_5d4594_1309832)
			}
		case 316:
			C.nox_video_setGammaSlider(value)
		case 318:
			C.nox_input_setSensitivity(C.powf(10, C.float(float32(value)/50-1)))
		}
		return 0
	}
	return 0
}

//export sub_4AE3B0
func sub_4AE3B0() C.int {
	v := C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1309820)
	C.dword_5d4594_1309820 = nil
	return v
}

//export sub_4AE3D0
func sub_4AE3D0() C.int {
	if C.wndIsShown_nox_xxx_wndIsShown_46ACC0(C.dword_5d4594_1309820) != 0 {
		return C.sub_4C4280()
	}
	return 1
}

//export sub_4AEE30
func sub_4AEE30() C.longlong {
	var out int64
	scale := math.Float64frombits(uint64(C.qword_581450_9552))
	for i := 0; i < 320; i++ {
		angle := float64(i+192) * (*memmap.PtrFloat64(0x581450, 9768)) * (*memmap.PtrFloat64(0x581450, 9760))
		out = int64(float64(C.sin(C.double(angle))) * scale)
		*memmap.PtrUint32(0x5D4594, uintptr(1309840+4*i)) = uint32(out)
	}
	return C.longlong(out)
}
