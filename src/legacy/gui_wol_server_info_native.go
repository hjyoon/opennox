package legacy

/*
#include "client__gui__window.h"
#include "client__shell__noxworld.h"

typedef const nox_gui_server_ent_t nox_wol_const_server_t;
*/
import "C"

import (
	"encoding/binary"
	"strconv"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export nox_client_gui_serverInfoBlockCheckExp_439370
func nox_client_gui_serverInfoBlockCheckExp_439370(pos *C.int2, server *C.nox_wol_const_server_t) {
	if server == nil {
		return
	}
	sub_439450(C.int(pos.field_0), C.int(pos.field_4), (*C.uint32_t)(unsafe.Pointer(pos)))
	sub_46B120(wolWindows.serverInfo, nil)
	nox_xxx_wndShowModalMB_46A8C0(wolWindows.serverInfo)
	sub_46C690(wolWindows.serverInfo)
	nox_window_setPos_46A9B0(wolWindows.serverInfo, int32(pos.field_0), int32(pos.field_4))
	nox_client_gui_serverInfoBlock_4394D0(server)
	Set_dword_5d4594_815056(1)
	record := (*Nox_gui_server_ent_t)(unsafe.Pointer(server))
	*memmap.PtrUint16(0x5D4594, 814604) = uint16(record.Port())
	if noxflags.HasEngine(noxflags.EngineNoRendering) {
		nox_xxx_wnd_46ABB0_go(wolWindows.serverInfoButton, 0)
	}
}

func wolInfoAppend(text *wchar2_t, style int) {
	nox_window_call_field_94_fnc(wolWindows.serverDetails, 16397, uintptr(unsafe.Pointer(text)), uintptr(style), nil, 0)
}

func wolInfoSeparator(offset uintptr) {
	wolInfoAppend((*wchar2_t)(memmap.PtrOff(0x587000, offset)), -1)
}

func wolInfoSpellRules(record *Nox_gui_server_ent_t) {
	wolInfoAppend(wolWindowString("DisabledSpells", 836), 14)
	found := false
	for spell := int32(1); spell <= 136; spell++ {
		if !bool(nox_xxx_spellIsValid_424B50(spell)) {
			continue
		}
		mask := binary.LittleEndian.Uint32(record.Field_33_3[4*(spell/32):])
		if mask&(1<<uint32(spell%32)) == 0 && nox_xxx_spellFlags_424A70(spell)&0x7000000 != 0 {
			wolInfoAppend(nox_xxx_spellTitle_424930(spell), 4)
			found = true
		}
	}
	if !found {
		wolInfoAppend(wolWindowString("None", 866), 4)
	}
}

func wolInfoEquipmentRules(mask uint32, count int, title string, titleLine, noneLine int32, lookup func(C.int) *wchar2_t) {
	wolInfoAppend(wolWindowString(title, titleLine), 14)
	found := false
	for i := range count {
		bit := uint32(1) << i
		if mask&bit == 0 {
			if name := lookup(C.int(bit)); name != nil {
				wolInfoAppend(name, -1)
				found = true
			}
		}
	}
	if !found {
		wolInfoAppend(wolWindowString("None", noneLine), 4)
	}
}

//export nox_client_gui_serverInfoBlock_4394D0
func nox_client_gui_serverInfoBlock_4394D0(server *C.nox_wol_const_server_t) {
	if server == nil {
		return
	}
	record := (*Nox_gui_server_ent_t)(unsafe.Pointer(server))
	resolution := int32(record.Field_25_2 & 0x7f)
	buffer, release := alloc.New([512]wchar2_t{})
	defer release()
	nox_window_call_field_94_fnc(wolWindows.serverDetails, 16399, 0, 0, nil, 0)
	wolInfoAppend(wolWindowString("Name", 749), 14)
	name := record.ServerName()
	if name == "" {
		name = wolRowAddress(record)
	}
	wolInfoAppend(wolCopyBytesToWide(buffer[:], name), -1)
	wolInfoSeparator(89396)
	wolInfoAppend(wolWindowString("Ping", 774), 14)
	ping := "--"
	if record.Ping() != 9999 {
		ping = strconv.Itoa(record.Ping())
	}
	wolInfoAppend(wolCopyBytesToWide(buffer[:], ping), -1)
	wolInfoSeparator(89464)
	wolInfoAppend(wolWindowString("GameType", 787), 14)
	wolInfoAppend(nox_gui_wol_gameModeString_43BCB0(C.short(record.Flags())), -1)
	if record.Flags().Has(noxflags.GameModeQuest) {
		wolInfoSeparator(89520)
		wolInfoAppend(wolWindowString("Stage", 794), 14)
		wolInfoAppend(wolCopyBytesToWide(buffer[:], strconv.Itoa(record.QuestLevel())), -1)
	}
	wolInfoSeparator(89580)
	wolInfoAppend(wolWindowString("Map", 801), 14)
	wolInfoAppend(wolCopyBytesToWide(buffer[:], record.MapName()), -1)
	wolInfoSeparator(89636)
	if flags := record.Flags(); flags&(noxflags.GameFlag15|noxflags.GameFlag16) != 0 {
		id, line := "Clan", int32(815)
		if flags.Has(noxflags.GameFlag15) {
			id, line = "Individual", 811
		}
		wolInfoAppend(wolWindowString(id, line), 6)
		wolInfoAppend(wolWindowString("Ladder", 817), 6)
	}
	wolInfoSeparator(89788)
	wolInfoAppend(wolWindowString("Occupancy", 822), 14)
	players := strconv.Itoa(record.Players()) + "/" + strconv.Itoa(record.MaxPlayers()) + "\n"
	wolInfoAppend(wolCopyBytesToWide(buffer[:], players), -1)
	if !record.Flags().Has(noxflags.GameOnline) {
		return
	}
	wolInfoSeparator(89860)
	wolInfoAppend(wolWindowString("Resolution", 831), 14)
	wolInfoAppend(get_video_mode_string(resolution), -1)
	wolInfoSeparator(89916)
	wolInfoSpellRules(record)
	wolInfoSeparator(90024)
	wolInfoEquipmentRules(binary.LittleEndian.Uint32(record.Field_38_3[:]), 27, "DisabledWeapons", 872, 899, sub_4159F0)
	wolInfoSeparator(90132)
	wolInfoEquipmentRules(binary.LittleEndian.Uint32(record.Field_39_3[:]), 26, "DisabledArmor", 905, 920, sub_415E80)
}
