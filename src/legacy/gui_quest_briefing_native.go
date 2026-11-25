package legacy

/*
#include "client__gui__guibrief.h"
#include "noxstring.h"

extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_587000_122956;
extern uint32_t nox_xxx_aSpellphoneme_3_587000_123008;
extern int nox_win_width;
extern int nox_win_height;

static inline void nox_briefing_number(wchar2_t* dst, const wchar2_t* format, int value) {
	nox_swprintf(dst, format, value);
}
static inline void nox_briefing_literal(wchar2_t* dst, const wchar2_t* format) {
	nox_swprintf(dst, format);
}
static inline void nox_briefing_pair(wchar2_t* dst, const wchar2_t* left, const wchar2_t* right) {
	nox_swprintf(dst, L"%s - %s", left, right);
}
static inline void nox_briefing_player_line(wchar2_t* dst, int rank, const wchar2_t* name) {
	nox_swprintf(dst, L"%d) %s", rank, name);
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const briefingStringsFile = "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c"

func briefingText(name string, line int32) *C.wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(name), nil, internCStr(briefingStringsFile), line)
}

func loadChapterBriefing(image string, line int32, offset uintptr) {
	*memmap.PtrPtr(0x5D4594, offset) = unsafe.Pointer(g2bImage(image))
	var dialog *C.char
	key := internCStr("Briefing:" + image)
	text := nox_strman_loadString_40F1D0(key, &dialog, internCStr(briefingStringsFile), line)
	*memmap.PtrPtr(0x5D4594, offset+4) = unsafe.Pointer(text)
	*memmap.PtrPtr(0x5D4594, offset+8) = unsafe.Pointer(dialog)
}

//export sub_44E410
func sub_44E410() *C.wchar2_t {
	for chapter := 1; chapter <= 11; chapter++ {
		for class := range 3 {
			className := GoString((*C.char)(*memmap.PtrPtr(0x587000, 122944+4*uintptr(class))))
			sound := uint32(chapter + 2)
			switch chapter {
			case 1:
				sound = 1
			case 2:
				sound = [...]uint32{4, 2, 3}[class]
			}
			offset := 831300 + 32*uintptr(11*class+chapter-1)
			loadChapterBriefing(fmt.Sprintf("%sChapterBegin%d", className, chapter), 1221, offset)
			*memmap.PtrUint32(0x5D4594, offset+12) = sound
			loadChapterBriefing(fmt.Sprintf("%sChapterLoss%d", className, chapter), 1227, offset+16)
			*memmap.PtrUint32(0x5D4594, offset+28) = sound
		}
	}
	*memmap.PtrPtr(0x5D4594, 831264) = unsafe.Pointer(g2bImage("CreditsImage"))
	var dialog *C.char
	text := nox_strman_loadString_40F1D0(internCStr("Nox:Credits"), &dialog, internCStr(briefingStringsFile), 1233)
	*memmap.PtrPtr(0x5D4594, 831268) = unsafe.Pointer(text)
	*memmap.PtrPtr(0x5D4594, 831272) = unsafe.Pointer(dialog)
	return text
}

func briefingMeasure(font unsafe.Pointer, text *C.wchar2_t) (int32, int32) {
	var width, height C.int
	nox_xxx_drawGetStringSize_43F840(font, text, &width, &height, 0)
	return int32(width), int32(height)
}

func briefingContinue(font unsafe.Pointer, blink *C.uint32_t, y, line int32) C.int {
	result := C.int(gameFrame() / 30)
	if gameFrame()%30 != 0 {
		if *blink != 1 {
			return result
		}
	} else {
		if *blink == 1 {
			*blink = C.uint32_t(gameFrame() % 30)
			return 1
		}
		*blink = 1
	}
	text := briefingText("GeneralPrint:QuestSplash12", line)
	var width C.int
	nox_xxx_drawGetStringSize_43F840(font, text, &width, nil, 0)
	x := (int32(C.nox_win_width)-C.NOX_DEFAULT_WIDTH)/2 + C.NOX_DEFAULT_WIDTH/2
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	return C.int(nox_xxx_drawString_43F6E0(font, text, x-int32(width)/2, y))
}

func drawQuestScoreRow(font unsafe.Pointer, row *questWinRow450770, rank int, x, y, nameWidth, width, lineHeight, spacing int32) {
	var text [256]C.wchar2_t
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_orange_2614256()))
	name := (*C.wchar2_t)(unsafe.Pointer(&row.Player.NameFinal[0]))
	C.nox_briefing_player_line(&text[0], C.int(rank), name)
	length := alloc.StrLenN(&text[0], len(text))
	measured, _ := briefingMeasure(font, &text[0])
	for measured >= nameWidth && length > 5 {
		length--
		text[length] = 0
		measured, _ = briefingMeasure(font, &text[0])
	}
	nox_xxx_drawStringWrap_43FAF0(font, &text[0], x, y, width, lineHeight)

	format := internWStr(" %d")
	labelWidth := int32(*questWinWidthSlot450770())
	y += spacing + spacing/2
	fields := [...]struct {
		key   string
		line  int32
		value int32
		color uint32
	}{
		{"GUIBrief.c:GeneratorsDestroyed", 529, int32(row.Generators), Get_nox_color_green_2614268()},
		{"GUIBrief.c:numSecretsFound", 541, int32(row.Secrets), Get_nox_color_green_2614268()},
		{"GUIBrief.c:Kills", 553, int32(row.Kills), Get_nox_color_green_2614268()},
		{"GUIBrief.c:TotalScore", 565, int32(row.Score), Get_nox_color_blue_2650684()},
	}
	for _, field := range fields {
		nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
		label := briefingText(field.key, field.line)
		nox_xxx_drawStringWrap_43FAF0(font, label, x, y, labelWidth, lineHeight)
		C.nox_briefing_number(&text[0], format, C.int(field.value))
		nox_xxx_drawSetTextColor_434390(int32(field.color))
		nox_xxx_drawStringWrap_43FAF0(font, &text[0], x+labelWidth, y, width-labelWidth, lineHeight)
		y += spacing
	}
}

//export sub_44E8E0
func sub_44E8E0(_ *C.nox_window, draw *C.nox_window_data) C.int {
	width, height := int32(C.nox_win_width), int32(C.nox_win_height)
	originX := (width - C.NOX_DEFAULT_WIDTH) / 2
	originY := (height - C.NOX_DEFAULT_HEIGHT) / 2
	font := draw.font
	var title, details, text [256]C.wchar2_t

	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	C.nox_wcscpy(&title[0], briefingText("GUIBrief.c:GauntletStatTitle", 449))
	C.nox_wcscpy(&details[0], briefingText("Noxworld.c:Stage", 451))
	C.nox_wcscat(&details[0], internWStr(" "))
	nox_itow(C.int(memmap.Int32(0x5D4594, 831228)), &details[C.nox_wcslen(&details[0])], 10)
	C.nox_briefing_pair(&text[0], &title[0], &details[0])
	tw, lineHeight := briefingMeasure(font, &text[0])
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	nox_xxx_drawString_43F6E0(font, &text[0], width/2-tw/2, height/2-240+lineHeight)

	firstX := memmap.Int32(0x587000, 122960)
	nextX := memmap.Int32(0x587000, 122968)
	rowWidth := nextX - memmap.Int32(0x587000, 122964) - 8
	nameWidth := nextX - firstX - 16
	spacing := int32(nox_float2int(C.float(float64(lineHeight) * 1.5)))
	rows := questWinRows450770()
	var players, mySecrets, friendsSecrets int
	for i, row := range rows {
		if row.Player == nil {
			continue
		}
		players++
		if C.uintptr_t(uintptr(unsafe.Pointer(row.Player))) == C.dword_8531A0_2576 {
			mySecrets = int(row.CoopSecrets)
		} else {
			friendsSecrets += int(row.CoopSecrets)
		}
		offset := uintptr(122960 + 8*i)
		x := width/2 - 320 + memmap.Int32(0x587000, offset)
		y := height/2 - 240 + memmap.Int32(0x587000, offset+4)
		drawQuestScoreRow(font, &row, i+1, x, y, nameWidth, rowWidth, lineHeight, spacing)
	}

	total := C.int(memmap.Uint32(0x5D4594, 832356))
	C.nox_briefing_number(&title[0], briefingText("GeneralPrint:SecretsTotal", 583), total)
	tw, th := briefingMeasure(font, &title[0])
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_orange_2614256()))
	nox_xxx_drawString_43F6E0(font, &title[0], originX+320-tw/2, originY+3*(150-th))

	if mySecrets != 0 {
		C.nox_briefing_number(&title[0], briefingText("GeneralPrint:SecretsFound", 593), C.int(mySecrets))
	} else {
		C.nox_briefing_literal(&title[0], briefingText("GeneralPrint:SecretsNoneFound", 596))
	}
	if players <= 1 {
		C.nox_wcscpy(&text[0], &title[0])
	} else {
		if friendsSecrets != 0 {
			if format := briefingText("GeneralPrint:SecretsFoundByFriends", 606); format != nil {
				C.nox_briefing_number(&details[0], format, C.int(friendsSecrets))
			}
		} else {
			C.nox_wcscpy(&details[0], briefingText("GeneralPrint:SecretsNoneFoundByFriends", 612))
		}
		C.nox_briefing_pair(&text[0], &title[0], &details[0])
	}
	tw, th = briefingMeasure(font, &text[0])
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_orange_2614256()))
	nox_xxx_drawString_43F6E0(font, &text[0], originX+320-tw/2, originY+2*(225-th))
	return briefingContinue(font, &C.dword_587000_122956, originY+450, 642)
}

//export sub_44F0F0
func sub_44F0F0(_ *C.nox_window, draw *C.nox_window_data) C.int {
	width, height := int32(C.nox_win_width), int32(C.nox_win_height)
	font := draw.font
	var format, stage [256]C.wchar2_t
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	C.nox_wcscpy(&format[0], briefingText("Noxworld.c:Stage", 668))
	C.nox_wcscat(&format[0], internWStr(" %d"))
	C.nox_briefing_number(&stage[0], &format[0], nox_gui_getQuestStage_450B10())
	tw, th := briefingMeasure(font, &stage[0])
	nox_xxx_drawString_43F6E0(font, &stage[0], width/2-tw/2, height/2+3*(th-80))
	if message := (*C.wchar2_t)(*memmap.PtrPtr(0x5D4594, 832464)); message != nil {
		tw, th = briefingMeasure(font, message)
		nox_xxx_drawString_43F6E0(font, message, width/2-tw/2, height/2+3*(80-th))
	}
	return briefingContinue(font, &C.nox_xxx_aSpellphoneme_3_587000_123008, (height-C.NOX_DEFAULT_HEIGHT)/2+462, 714)
}
