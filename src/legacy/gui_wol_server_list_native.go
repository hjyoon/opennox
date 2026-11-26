package legacy

/*
#include <stddef.h>
#include <stdlib.h>
#include "GAME2_3.h"
#include "client__shell__noxworld.h"
#include "cgo_const.h"
#include "noxstring.h"

typedef const nox_gui_server_ent_t nox_wol_const_server_t;

extern nox_list_item_t nox_gui_wol_servers_list;

enum { nox_wol_record_offset = offsetof(nox_gui_server_node_t, server) };
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func wolServerNode(record unsafe.Pointer) *C.nox_gui_server_node_t {
	return (*C.nox_gui_server_node_t)(unsafe.Add(record, -int(C.nox_wol_record_offset)))
}

func wolServerRecord(item *C.nox_list_item_t) *Nox_gui_server_ent_t {
	return (*Nox_gui_server_ent_t)(unsafe.Add(unsafe.Pointer(item), C.nox_wol_record_offset))
}

func wolServerHead() *C.nox_list_item_t { return &C.nox_gui_wol_servers_list }

func wolServerSortMode() *C.uint32_t { return &legacyAddressableWords.nox_wol_servers_sorting_166704 }

func ensureWOLServerList() {
	head := wolServerHead()
	if head.field_0 == nil || head.field_1 == nil {
		nox_common_list_clear_425760(head)
	}
}

//export nox_wol_server_row_get
func nox_wol_server_row_get(record *C.nox_wol_const_server_t) *nox_window {
	if record == nil {
		return nil
	}
	return wolServerNode(unsafe.Pointer(record)).row
}

//export nox_wol_server_row_set
func nox_wol_server_row_set(record *C.nox_gui_server_ent_t, row *nox_window) {
	if record != nil {
		wolServerNode(unsafe.Pointer(record)).row = row
	}
}

//export sub_49FFA0
func sub_49FFA0(destroy C.int) *C.nox_list_item_t {
	head := wolServerHead()
	initialized := memmap.PtrUint32(0x5D4594, 1305808)
	if *initialized == 0 {
		nox_common_list_clear_425760(head)
	} else {
		ensureWOLServerList()
	}
	for item := nox_common_list_getFirstSafe_425890(head); item != nil; {
		next := nox_common_list_getNextSafe_4258A0(item)
		node := (*C.nox_gui_server_node_t)(unsafe.Pointer(item))
		nox_common_list_remove_425920(unsafe.Pointer(item))
		if destroy != 0 && node.row != nil {
			nox_xxx_windowDestroyMB_46C4E0(node.row)
		}
		C.free(unsafe.Pointer(node))
		item = next
	}
	*initialized = 1
	return nil
}

//export sub_4A0020
func sub_4A0020() *C.nox_list_item_t { return wolServerHead() }

func assignWOLSortKey(record *Nox_gui_server_ent_t, mode uint32) {
	switch mode {
	case 2:
		record.Sort_key = int32(record.PlayersVal)
	case 3:
		record.Sort_key = 32 - int32(record.PlayersVal)
	case 6:
		record.Sort_key = record.PingVal
	case 7:
		record.Sort_key = 1000 - record.PingVal
	case 8:
		record.Sort_key = int32(record.StatusVal & 0x30)
	case 9:
		record.Sort_key = 48 - int32(record.StatusVal&0x30)
	}
}

func wolServerSortsBefore(a, b *Nox_gui_server_ent_t, mode uint32) bool {
	switch mode {
	case 0, 1:
		cmp := C.nox_strcmpi((*C.char)(unsafe.Pointer(&a.ServerNameBuf[0])), (*C.char)(unsafe.Pointer(&b.ServerNameBuf[0])))
		if mode == 0 {
			return cmp <= 0
		}
		return cmp >= 0
	case 4, 5:
		first := nox_gui_wol_gameModeString_43BCB0(C.short(a.Flags()))
		second := nox_gui_wol_gameModeString_43BCB0(C.short(b.Flags()))
		cmp := C.nox_wcscmp(first, second)
		if mode == 4 {
			return cmp <= 0
		}
		return cmp >= 0
	case 2, 3, 6, 7, 8, 9:
		return a.Sort_key <= b.Sort_key
	}
	return false
}

func insertWOLServer(node *C.nox_list_item_t) C.int {
	record := wolServerRecord(node)
	mode := uint32(*wolServerSortMode())
	assignWOLSortKey(record, mode)
	head := wolServerHead()
	index := C.int(0)
	for item := nox_common_list_getFirstSafe_425890(head); item != nil; item = nox_common_list_getNextSafe_4258A0(item) {
		if wolServerSortsBefore(record, wolServerRecord(item), mode) {
			nox_common_list_append_4258E0(item, node)
			return index
		}
		index++
	}
	nox_common_list_append_4258E0(head, node)
	return index
}

//export nox_wol_servers_addResult_4A0030
func nox_wol_servers_addResult_4A0030(record *C.nox_gui_server_ent_t) C.int {
	if record == nil {
		return -1
	}
	ensureWOLServerList()
	node := (*C.nox_gui_server_node_t)(C.calloc(1, C.sizeof_nox_gui_server_node_t))
	if node == nil {
		return -1
	}
	dst := unsafe.Slice((*byte)(unsafe.Pointer(wolServerRecord(&node.list))), C.sizeof_nox_gui_server_ent_t)
	src := unsafe.Slice((*byte)(unsafe.Pointer(record)), len(dst))
	copy(dst, src)
	return insertWOLServer(&node.list)
}

//export nox_wol_servers_sortBtnHandler_4A0290
func nox_wol_servers_sortBtnHandler_4A0290(id C.int) {
	column := id - 10047
	if column < 0 || column >= 5 {
		return
	}
	mode := C.uint32_t(2 * column)
	if *wolServerSortMode() == mode {
		mode++
	}
	*wolServerSortMode() = mode
}

//export sub_4A0360
func sub_4A0360() *C.nox_list_item_t {
	for item := nox_common_list_getFirstSafe_425890(wolServerHead()); item != nil; item = nox_common_list_getNextSafe_4258A0(item) {
		nox_gui_wol_newServerLine_43B7C0(wolServerRecord(item).C())
	}
	return nil
}

//export sub_4A0390
func sub_4A0390() *C.nox_list_item_t {
	var items []*C.nox_list_item_t
	for item := nox_common_list_getFirstSafe_425890(wolServerHead()); item != nil; {
		next := nox_common_list_getNextSafe_4258A0(item)
		nox_common_list_remove_425920(unsafe.Pointer(item))
		items = append(items, item)
		item = next
	}
	for _, item := range items {
		insertWOLServer(item)
	}
	return sub_4A0360()
}

//export sub_4A0410
func sub_4A0410(address *C.nox_cgo_const_char, port C.short) C.int {
	ip := GoStringP(unsafe.Pointer(address))
	for item := nox_common_list_getFirstSafe_425890(wolServerHead()); item != nil; item = nox_common_list_getNextSafe_4258A0(item) {
		server := wolServerRecord(item)
		if server.Addr() == ip && server.Port() == int(port) {
			return 0
		}
	}
	return 1
}

//export sub_4A0490
func sub_4A0490(id C.int) *C.nox_gui_server_ent_t {
	for item := nox_common_list_getFirstSafe_425890(wolServerHead()); item != nil; item = nox_common_list_getNextSafe_4258A0(item) {
		if record := wolServerRecord(item); record.Field_9 == uint32(id) {
			return record.C()
		}
	}
	return nil
}

//export sub_4A04C0
func sub_4A04C0(index C.int) *C.nox_gui_server_ent_t {
	for item := nox_common_list_getFirstSafe_425890(wolServerHead()); item != nil; item = nox_common_list_getNextSafe_4258A0(item) {
		if index == 0 {
			return wolServerRecord(item).C()
		}
		index--
	}
	return nil
}
