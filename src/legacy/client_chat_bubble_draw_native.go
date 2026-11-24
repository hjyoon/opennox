package legacy

/*
#include "GAME2_3.h"

extern uint32_t nox_color_white_2523948;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_48D990
func sub_48D990(view *nox_draw_viewport_t) {
	fontHeight := nox_xxx_guiFontHeightMB_43F320(nil)
	sub_437260()
	defer sub_437290()
	sub_48DCF0(view)
	for bubble := chatBubbleHead; bubble != nil; bubble = bubble.next {
		if bubble.visible != 0 {
			drawChatBubble(bubble, fontHeight)
		}
	}
}

func drawChatBubble(bubble *C.nox_chat_bubble, fontHeight int32) {
	var playerName *wchar2_t
	textColor := int32(C.nox_color_white_2523948)
	if drawable := asDrawable(bubble.drawable); drawable != nil && drawable.ObjClass&4 != 0 {
		team := teamObjectByNetCode418C80(drawable.NetCode32)
		if info := nox_common_playerInfoGetByID_417040(int32(bubble.net_code)); info != nil {
			playerName = (*wchar2_t)(unsafe.Pointer(&info.name_final[0]))
		}
		if team != nil {
			if resolved := nox_xxx_getTeamByID_418AB0(int32(team.ID)); resolved != nil {
				textColor = int32(nox_xxx_materialGetTeamColor_418D50(resolved))
			}
		}
	}
	x, y := int32(bubble.x), int32(bubble.y)
	w, h := int32(bubble.width), int32(bubble.height)
	outline := memmap.Int32(0x852978, 4)
	nox_client_drawSetColor_434460(outline)
	drawChatBubbleBorder(x, y, w, h, fontHeight, bubble.draw_tail != 0)
	x, y = x-1, y-1
	nox_client_drawSetColor_434460(memmap.Int32(0x85B3FC, 956))
	drawChatBubbleBorder(x, y, w, h, fontHeight, bubble.draw_tail != 0)

	nox_xxx_drawSetTextColor_434390(int32(C.nox_color_white_2523948))
	nox_xxx_drawSetColor_4343E0(outline)
	nox_xxx_drawStringWrapHL_43FD00(nil, &bubble.text[0], x, y, 128, 0)
	if playerName != nil {
		nox_xxx_drawSetTextColor_434390(textColor)
		nox_xxx_drawStringWrapHL_43FD00(nil, playerName, x, y-fontHeight-1, 128, 0)
	}
}

func drawChatBubbleBorder(x, y, width, height, fontHeight int32, tail bool) {
	line := func(x0, y0, x1, y1 int32) {
		nox_client_drawAddPoint_49F500(x0, y0)
		nox_client_drawAddPoint_49F500(x1, y1)
		nox_client_drawLineFromPoints_49E4B0()
	}
	half := fontHeight / 2
	right, bottom := x+width, y+height
	outerBottom := bottom + half
	line(x, y-half, x-half, y)
	line(x, y-half, right, y-half)
	line(right+half, y, right, y-half)
	line(right+half, y, right+half, bottom)
	line(right, outerBottom, right+half, bottom)
	if tail {
		middle := x + width/2
		line(right, outerBottom, middle+half, outerBottom)
		line(middle, outerBottom+fontHeight, middle+half, outerBottom)
		line(middle, outerBottom+fontHeight, middle-half, outerBottom)
		line(x, outerBottom, middle-half, outerBottom)
	} else {
		line(x, outerBottom, right, outerBottom)
	}
	line(x, outerBottom, x-half, bottom)
	line(x-half, y, x-half, bottom)
}
