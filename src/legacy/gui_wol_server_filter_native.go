package legacy

/*
#include "GAME2_2.h"
#include "client__gui__window.h"

typedef const void* nox_wol_filter_const_data;
extern nox_window* dword_5d4594_1193380;
extern nox_window* dword_5d4594_1193384;
*/
import "C"

import (
	"encoding/binary"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type wolServerFilter struct {
	checkLatency   uint32
	excludeClosed  uint32
	excludePrivate uint32
	questLevel     uint32
	maxLatency     uint32
	matchRules     uint32
	reserved       [4]uint32
	matchVersion   uint32
}

func wolServerFilterMode(index C.int) *uint32 {
	return memmap.PtrUint32(0x5D4594, uintptr(1193372+4*index))
}

func wolServerFilterRecord(index C.int) *wolServerFilter {
	return memmap.PtrT[wolServerFilter](0x5D4594, uintptr(1193388+44*index))
}

func wolFilterSelected(win *nox_window) uint32 {
	return uint32(win.draw_data.field_0 >> 2 & 1)
}

//export sub_489870
func sub_489870() C.int {
	win := C.dword_5d4594_1193380
	if *wolServerFilterMode(0) == 2 {
		filter := wolServerFilterRecord(0)
		child := func(id int32) *nox_window { return nox_xxx_wndGetChildByID_46B0C0(win, id) }
		filter.checkLatency = wolFilterSelected(child(10028))
		pingText := nox_window_call_field_94_fnc(child(10031), 16413, 0, 0, nil, 0)
		filter.maxLatency = uint32(parseLegacyWideDecimal((*wchar2_t)(unsafe.Pointer(pingText))))
		filter.excludeClosed = wolFilterSelected(child(10029))
		filter.excludePrivate = wolFilterSelected(child(10030))
		filter.questLevel = 0
		if wolFilterSelected(child(10015)) != 0 {
			filter.questLevel = 0x82 - wolFilterSelected(child(10016))
		}
		filter.matchRules = wolFilterSelected(child(10014))
		filter.matchVersion = wolFilterSelected(child(10018))
	}
	return C.int(nox_window_set_hidden(win, 1))
}

func wolServerMatchesRules(record *Nox_gui_server_ent_t) bool {
	var rules server.Settings2
	copy(rules.Field0[:], record.MapNameBuf[:])
	readMapRuleDefaults57A1E0(&rules, nil, nil, ruleSourceMap|ruleSourceInternet, uint16(record.Flags()))
	for i, value := range rules.Field24.Vals {
		if value != binary.LittleEndian.Uint32(record.Field_33_3[i*4:]) {
			return false
		}
	}
	return rules.Field44 == record.Field_38_3 && rules.Field48 == binary.LittleEndian.Uint32(record.Field_39_3[:])
}

//export nox_xxx_checkSomeFlagsOnJoin_4899C0
func nox_xxx_checkSomeFlagsOnJoin_4899C0(raw *C.nox_gui_server_ent_t) C.int {
	record := (*Nox_gui_server_ent_t)(unsafe.Pointer(raw))
	const currentVersion = C.NOX_CLIENT_VERS_CODE
	switch *wolServerFilterMode(0) {
	case 1:
		status := server.ServerFlags(record.StatusVal)
		return C.int(bool2int(!status.Has(server.ServerClosed|server.ServerPrivate) && record.VersionVal == currentVersion))
	case 2:
		filter := wolServerFilterRecord(0)
		ping := uint32(record.PingVal)
		if filter.checkLatency != 0 && ping > filter.maxLatency && ping != 9999 {
			return 0
		}
		status := server.ServerFlags(record.StatusVal)
		if filter.excludeClosed != 0 && status.Has(server.ServerClosed) ||
			filter.excludePrivate != 0 && status.Has(server.ServerPrivate) {
			return 0
		}
		if stage := record.Field_25_2; int8(stage) < 0 && filter.questLevel > uint32(stage&0x7f) {
			return 0
		}
		if filter.matchRules != 0 && !wolServerMatchesRules(record) {
			return 0
		}
		return C.int(bool2int(filter.matchVersion == 0 || record.VersionVal == currentVersion))
	default:
		return 1
	}
}

//export sub_489B80
func sub_489B80(parent *nox_window) *nox_window {
	win := nox_new_window_from_file(internCStr("filter.wnd"), C.nox_xxx_windowMplayFilterProc_489E70)
	C.dword_5d4594_1193380 = win
	if win == nil {
		return nil
	}
	child := func(id int32) *nox_window { return nox_xxx_wndGetChildByID_46B0C0(win, id) }
	options := child(10012)
	C.dword_5d4594_1193384 = options
	filter := wolServerFilterRecord(0)
	sub_46B120(win, parent)
	sub_46B120(options, win)
	nox_xxx_wndSetProc_46B2C0(options, C.nox_xxx_windowMplayFilterProc_489E70)
	if filter.checkLatency != 0 {
		child(10028).draw_data.field_0 |= 4
	}
	text, free := CWString(strconv.FormatInt(int64(int32(filter.maxLatency)), 10))
	nox_window_call_field_94_fnc(child(10031), 16414, uintptr(unsafe.Pointer(text)), ^uintptr(0), nil, 0)
	free()
	for _, field := range []struct {
		value uint32
		id    int32
	}{
		{filter.excludeClosed, 10029},
		{filter.excludePrivate, 10030},
		{filter.questLevel, 10015},
	} {
		if field.value != 0 {
			child(field.id).draw_data.field_0 |= 4
		}
	}
	radio := int32(10016)
	if filter.questLevel&2 != 0 {
		radio = 10017
	}
	selected := child(radio)
	selected.draw_data.field_0 |= 4
	for _, field := range []struct {
		value uint32
		id    int32
	}{
		{filter.matchRules, 10014},
		{filter.matchVersion, 10018},
	} {
		if field.value != 0 {
			selected = child(field.id)
			selected.draw_data.field_0 |= 4
		}
	}
	switch *wolServerFilterMode(0) {
	case 0:
		selected = child(10024)
		nox_window_set_hidden(options, 1)
	case 1:
		selected = child(10025)
		nox_window_set_hidden(options, 1)
	case 2:
		selected = child(10026)
		sub_489DC0()
	}
	selected.draw_data.field_0 |= 4
	return win
}

//export sub_489DC0
func sub_489DC0() {
	options := C.dword_5d4594_1193384
	nox_window_set_hidden(options, 0)
	ping := nox_xxx_wndGetChildByID_46B0C0(options, 10031)
	nox_xxx_wnd_46ABB0_go(ping, int32(wolFilterSelected(nox_xxx_wndGetChildByID_46B0C0(options, 10028))))
	C.sub_46AD20(options, 10016, 10017, C.int(wolFilterSelected(nox_xxx_wndGetChildByID_46B0C0(options, 10015))))
}

//export nox_xxx_windowMplayFilterProc_489E70
func nox_xxx_windowMplayFilterProc_489E70(win *nox_window, event C.int, eventArg, unused C.uintptr_t) C.int {
	if event == 23 {
		return 1
	}
	if event != 16391 {
		return 0
	}
	control := (*nox_window)(unsafe.Pointer(uintptr(eventArg)))
	id := nox_xxx_wndGetID_46B0A0(control)
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	root, options := C.dword_5d4594_1193380, C.dword_5d4594_1193384
	switch id {
	case 10015:
		C.sub_46AD20(root, 10016, 10017, C.int(1-wolFilterSelected(control)))
	case 10024, 10025:
		*wolServerFilterMode(0) = uint32(id - 10024)
		nox_window_set_hidden(options, 1)
	case 10026:
		*wolServerFilterMode(0) = 2
		sub_489DC0()
	case 10028:
		ping := nox_xxx_wndGetChildByID_46B0C0(root, 10031)
		nox_xxx_wnd_46ABB0_go(ping, int32(1-wolFilterSelected(control)))
	}
	return 0
}

//export sub_489FB0
func sub_489FB0() C.int {
	win := C.dword_5d4594_1193380
	if win == nil {
		return 0
	}
	sub_489870()
	nox_xxx_wndClearCaptureMain_46ADE0(win)
	ret := nox_xxx_windowDestroyMB_46C4E0(win)
	C.dword_5d4594_1193380 = nil
	return C.int(ret)
}

//export sub_489FF0
func sub_489FF0(index, mode C.int, setup C.nox_wol_filter_const_data) C.int {
	*wolServerFilterMode(index) = uint32(mode)
	*wolServerFilterRecord(index) = *(*wolServerFilter)(unsafe.Pointer(setup))
	return 11 * index
}
