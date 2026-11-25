package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guirank.h"

typedef const wchar2_t rank_const_text_t;
extern unsigned int dword_587000_145664, dword_587000_145668;
extern uint32_t dword_587000_145672;
extern uint32_t dword_5d4594_1090120;
extern nox_window* dword_5d4594_1090100;
*/
import "C"

import (
	"fmt"
	"math"
	"runtime"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//export sub_46F080
func sub_46F080(win *C.nox_window, draw *C.nox_window_data) C.int {
	if win == nil || draw == nil {
		return 0
	}
	if !noxflags.HasGame(noxflags.GameFlag4) && C.dword_587000_145668 != 6 {
		next := C.dword_587000_145668
		if next == 0 {
			next = 6
		}
		C.dword_5d4594_1090120 = next - 1
		sub_4703F0()
		C.dword_587000_145668 = 6
		if C.dword_5d4594_1090120 == 0 {
			return 1
		}
	}
	x, y := meterWindowPosition(win)
	limit := clientRankGameLimit()
	if win.flags&0x80 != 0 {
		nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(draw.bg_image), int32(x), int32(y))
	} else if draw.bg_color != 0x80000000 {
		nox_client_drawRectFilledAlpha_49CF10(int32(x), int32(y), int32(win.width), int32(win.height))
	}
	lastFrame := memmap.PtrUint32(0x5D4594, 1090124)
	if C.dword_587000_145664 != 0 || gameFrame() > *lastFrame+gameFPS() {
		set_dword_5d4594_3799468(1)
		*lastFrame = gameFrame()
		C.dword_587000_145672 = math.MaxUint32
		sub_46DB80()
		sub_46DCC0()
		C.dword_587000_145664 = 0
		clientRankFillDisplay(int(C.dword_5d4594_1090120), limit)
		C.sub_46FC50()
		C.sub_46FD80()
	}
	if int32(C.dword_587000_145672) >= 0 {
		sub_46FAE0()
	}
	return 1
}

func clientRankGameLimit() int32 {
	if noxflags.HasGame(noxflags.GameHost) {
		return int32(uint16(nox_xxx_servGamedataGet_40A020(C.short(noxflags.GetGame()))))
	}
	data := nox_xxx_cliGamedataGet_416590(0)
	return int32(*(*uint16)(unsafe.Add(unsafe.Pointer(data), 54)))
}

func clientRankFillDisplay(mode int, limit int32) {
	local := uint32(ClientPlayerNetCode())
	var team *server.Team
	if member := nox_xxx_objGetTeamByNetCode_418C80(C.int(local)); member != nil {
		team = GetServer().S().Teams.ByID(server.TeamID(member.id))
	}
	player := Get_dword_8531A0_2576()
	showRank := player != nil && !clientRankObserver(player)
	teamCount := memmap.Uint8(0x5D4594, 1090116)
	if teamCount != 0 && (mode == 2 || mode == 3) {
		clientRankDisplayTeams(teamCount)
	}
	secondPagePadding := clientRankListRows(1090060)
	playerCount := memmap.Uint8(0x5D4594, 1090117)
	if playerCount != 0 && (mode == 2 || mode == 4 || mode == 5) {
		sub_46F8F0(0, 0)
		if mode == 4 {
			playerCount = min(playerCount, 3)
		}
		for i := range playerCount {
			page := C.int(i >> 4)
			if i == 16 {
				sub_46F8F0(1, C.int(secondPagePadding))
			}
			clientRankDisplayPlayer(mode, limit, int(page), clientRankPlayerAt(i), local)
		}
	} else if mode == 1 {
		sub_46FFD0()
	}

	title := clientRankStockText(147724)
	switch mode {
	case 1:
		title = clientRankLocalized("Noxworld.c:Quest", 1487)
	case 2:
		title = clientRankLocalized("TeamPlayerRank", 1475)
	case 3:
		title = clientRankLocalized("Teams", 1519)
		showRank = showRank || team != nil
	case 4:
		title = clientRankLocalized("Top3", 1479)
	case 5:
		title = clientRankLocalized("WolRank", 1483)
	}
	label := memmap.PtrT[[256]uint16](0x5D4594, 1086692)
	if !noxflags.HasGame(noxflags.GameHost) || !noxflags.HasEngine(noxflags.EngineNoRendering) {
		switch mode {
		case 1:
			text := GoWString(clientRankLocalized("Noxworld.c:Stage", 1499))
			alloc.StrCopyZero16(label[:], fmt.Sprintf("%s %d", text, nox_gui_getQuestStage_450B10()))
		case 3:
			var position uint8
			if team != nil {
				position = uint8(sub_46FF70(C.int(team.Lessons)))
			}
			text := GoWString(clientRankLocalized("yourteamrank", 1525))
			alloc.StrCopyZero16(label[:], fmt.Sprintf("%s %d / %d", text, position, memmap.Uint8(0x5D4594, 1090116)))
		case 2, 4, 5:
			text := GoWString(clientRankLocalized("yourrank", 1501))
			alloc.StrCopyZero16(label[:], fmt.Sprintf("%s %d / %d", text, uint8(sub_46FEE0()), memmap.Uint8(0x5D4594, 1090118)))
		}
	}
	labelWindow := C.dword_5d4594_1090100
	if showRank {
		if wndIsShown_nox_xxx_wndIsShown_46ACC0(labelWindow) != 0 {
			nox_window_set_hidden(labelWindow, 0)
		}
		clientRankSetText(labelWindow, (*wchar2_t)(&label[0]))
	} else if wndIsShown_nox_xxx_wndIsShown_46ACC0(labelWindow) == 0 {
		nox_window_set_hidden(labelWindow, 1)
	}
	clientRankSetText(clientRankWindow(1090104), title)
}

func clientRankListRows(offset int) uint16 {
	if list := clientRankList(offset); list != nil {
		return uint16(list.field_11_1)
	}
	return 0
}

func clientRankSetText(win *C.nox_window, text *wchar2_t) uintptr {
	result := nox_window_call_field_94_fnc(win, 16385, uintptr(unsafe.Pointer(text)), 0, internCStr("client_rank_display_native.go"), 0)
	runtime.KeepAlive(text)
	return result
}

func clientRankDisplayTeams(count byte) {
	sub_46DC60(clientRankWindow(1090060), 9, (*C.rank_const_text_t)(clientRankLocalized("team", 1338)))
	sub_46DC60(clientRankWindow(1090068), 9, (*C.rank_const_text_t)(clientRankStockText(146512)))
	sub_46DC60(clientRankWindow(1090076), 9, (*C.rank_const_text_t)(clientRankStockText(146516)))
	sub_46DC60(clientRankWindow(1090084), 9, (*C.rank_const_text_t)(clientRankLocalized("score", 1341)))
	sub_46DC60(clientRankWindow(1090092), 9, (*C.rank_const_text_t)(clientRankStockText(146564)))
	color := C.uchar(9)
	for i := range count {
		row := clientRankTeamAt(i)
		color = sub_46FEB0(C.uchar(i))
		clientRankWriteWideCell(1090060, color, (*wchar2_t)(&row.Name[0]))
		clientRankWriteWideCell(1090068, color, clientRankStockText(146576))
		clientRankWriteWideCell(1090076, color, clientRankStockText(146580))
		clientRankWriteNumberCell(1090084, color, row.Score, "")
		clientRankWriteWideCell(1090092, color, clientRankStockText(146592))
	}
	for i, offset := range [5]uintptr{146596, 146600, 146604, 146608, 146612} {
		clientRankWriteWideCell(1090060+8*i, color, clientRankStockText(offset))
	}
}

func clientRankDisplayPlayer(mode int, limit int32, page int, row *clientRankPlayerRow, local uint32) {
	start := 1090060 + 4*page
	overLimit := noxflags.HasGame(noxflags.GameModeElimination) && limit > 0 && int32(row.Score) >= limit
	color := C.uchar(9)
	if row.Status&1 == 0 || row.Status&0x20 != 0 || overLimit {
		dim := row.Status&0x20 != 0 || overLimit
		if row.TeamID == -1 {
			color = 3
			if dim {
				color = 2
			}
		} else {
			color = sub_46FEB0(sub_46FE60(C.int(row.TeamID)))
			if dim {
				color -= 2
			}
		}
	}
	if row.NetCode == local {
		C.dword_587000_145672 = C.uint32_t(clientRankListRows(start))
		*memmap.PtrUint32(0x5D4594, 1088996) = uint32(page)
	}
	clientRankWriteWideCell(start, color, (*wchar2_t)(&row.Name[0]))
	clientRankWriteWideCell(start+16, color, C.nox_rank_class_name(C.uint(row.Class)))
	if mode != 5 || row.Score > 0 {
		clientRankWriteNumberCell(start+24, color, int32(row.Score), "")
	} else {
		clientRankWriteWideCell(start+24, color, clientRankStockText(146640))
	}
	clientRankWriteNumberCell(start+32, color, int32(row.Ping), "")
	if mode == 5 {
		if player := asPlayerS(nox_common_playerInfoGetByID_417040(int32(row.NetCode))); player != nil {
			buf := memmap.PtrT[[512]uint16](0x5D4594, 1089000)
			name := player.Field2096Buf[:alloc.StrLenS(player.Field2096Buf[:])]
			buf[0] = '('
			for i, ch := range name {
				buf[i+1] = uint16(ch)
			}
			buf[len(name)+1], buf[len(name)+2] = ')', 0
			sub_46DC00(clientRankWindow(start+8), color, (*C.rank_const_text_t)(&buf[0]))
		}
	} else {
		var markerColor C.uint8_t
		marker := sub_46FB50(C.int(row.Objective), &markerColor)
		sub_46DC60(clientRankWindow(start+8), C.uchar(markerColor), (*C.rank_const_text_t)(marker))
	}
}
