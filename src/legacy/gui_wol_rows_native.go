package legacy

/*
#include "client__gui__window.h"
#include "client__shell__noxworld.h"

extern nox_window* nox_wol_wnd_world_814980;
extern nox_window* nox_wol_wnd_gameList_815012;
extern nox_window *dword_5d4594_814984, *dword_5d4594_814988;
*/
import "C"

import (
	"strconv"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func wolCopyBytesToWide(dst []wchar2_t, src string) *wchar2_t {
	for i := range len(src) {
		dst[i] = wchar2_t(src[i])
	}
	dst[len(src)] = 0
	return &dst[0]
}

func wolAppendListText(win *nox_window, text *wchar2_t) {
	nox_window_call_field_94_fnc(win, 16397, uintptr(unsafe.Pointer(text)), 4, nil, 0)
}

func wolRowAddress(record *Nox_gui_server_ent_t) string {
	return record.Addr() + ":" + strconv.Itoa(record.Port())
}

func wolAppendLANRow(record *Nox_gui_server_ent_t) {
	position := (*wchar2_t)(memmap.PtrOff(0x587000, 91164))
	wolAppendListText(C.nox_wol_wnd_gameList_815012, position)

	columns := wolGameColumns()
	buffer, release := alloc.New([128]wchar2_t{})
	defer release()
	text := &buffer[0]
	label := record.ServerName()
	if record.ServerNameBuf[0] == memmap.Uint8(0x5D4594, 815120) {
		label = wolRowAddress(record)
	}
	wolCopyBytesToWide(buffer[:], label)
	sub_43BC10(text, 100)
	wolAppendListText(columns[0], text)

	population := strconv.Itoa(record.Players()) + "/" + strconv.Itoa(record.MaxPlayers())
	wolAppendListText(columns[1], wolCopyBytesToWide(buffer[:], population))

	mode := nox_gui_wol_gameModeString_43BCB0(C.short(record.Flags()))
	if record.Flags().Has(noxflags.GameModeQuest) {
		dst := (*wchar2_t)(memmap.PtrOff(0x5D4594, 814772))
		nox_wcscpy(dst, mode)
		nox_wcscat(dst, internWStr(" "+strconv.Itoa(record.QuestLevel())))
		mode = dst
	}
	wolAppendListText(columns[2], mode)

	ping := "--"
	if record.Ping() != 9999 {
		ping = strconv.Itoa(record.Ping())
	}
	wolAppendListText(columns[3], wolCopyBytesToWide(buffer[:], ping))

	*text = 0
	if record.Status()&0x20 != 0 {
		nox_wcscat(text, wolWindowString("Noxworld.wnd:private", 3486))
	}
	if record.Status()&0x10 != 0 {
		if *text != 0 {
			nox_wcscat(text, internWStr("+"))
		}
		nox_wcscat(text, wolWindowString("Noxworld.wnd:closed", 3494))
	}
	if *text == 0 {
		id, line := "Open", int32(3505)
		if record.Players() >= record.MaxPlayers() {
			id, line = "Full", 3501
		}
		nox_wcscat(text, wolWindowString(id, line))
	}
	wolAppendListText(columns[4], text)
}

func wolCreateMapRow(record *Nox_gui_server_ent_t) *nox_window {
	draw := C.nox_window_data{style: 257, win: C.nox_wol_wnd_world_814980}
	region := Get_dword_587000_87412()
	if region == -1 {
		region = int(sub_437860(C.int(record.Field_11_0), C.int(record.Field_11_2)))
		offset := uintptr(87528 + 8*region)
		record.Field_11_0 = (record.Field_11_0 - memmap.Int16(0x587000, offset)) >> 1
		record.Field_11_2 = (record.Field_11_2 - memmap.Int16(0x587000, offset+2)) >> 1
		parent := nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_814988, int32(region+10054))
		return nox_gui_newButtonOrCheckbox_4A91A0(parent, 1185,
			C.int(record.Field_11_0)-5, C.int(record.Field_11_2)-5, 10, 10, &draw)
	}
	offset := uintptr(87528 + 8*region)
	record.Field_11_0 -= memmap.Int16(0x587000, offset)
	record.Field_11_2 -= memmap.Int16(0x587000, offset+2)
	return nox_gui_newButtonOrCheckbox_4A91A0(C.dword_5d4594_814984, 1192,
		C.int(record.Field_11_0)-10, C.int(record.Field_11_2)-10, 20, 20, &draw)
}

//export nox_gui_wol_newServerLine_43B7C0
func nox_gui_wol_newServerLine_43B7C0(raw *C.nox_gui_server_ent_t) {
	record := (*Nox_gui_server_ent_t)(unsafe.Pointer(raw))
	if Get_dword_587000_87408() != 0 {
		wolAppendLANRow(record)
		return
	}
	if nox_wol_server_row_get(raw) != nil {
		sub_437320(raw)
		return
	}
	row := wolCreateMapRow(record)
	if row == nil {
		return
	}
	nox_wol_server_row_set(raw, row)
	sub_437320(raw)
	label := record.ServerName()
	if label == "" {
		label = wolRowAddress(record)
	}
	suffix := " -- ms"
	if record.Ping() != 9999 {
		suffix = " " + strconv.Itoa(record.Ping()) + "ms"
	}
	var tip [128]wchar2_t
	nox_xxx_wndWddSetTooltip_46B000(&row.draw_data, wolCopyBytesToWide(tip[:], label+suffix))
	nox_xxx_wndSetProc_46B2C0_go(row, C.nox_xxx_windowMultiplayerSub_439E70)
	nox_xxx_wndSetID_46B080(row, int32(record.Field_9)+10070)
}
