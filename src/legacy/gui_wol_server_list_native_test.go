package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestWOLServerListKeepsPackedRecordsAndNativeRows(t *testing.T) {
	head := wolServerHead()
	initialized := memmap.PtrUint32(0x5D4594, 1305808)
	savedHead, savedInit, savedSort := *head, *initialized, *wolServerSortMode()
	nox_common_list_clear_425760(head)
	*initialized, *wolServerSortMode() = 1, 0
	t.Cleanup(func() {
		sub_49FFA0(0)
		*head, *initialized, *wolServerSortMode() = savedHead, savedInit, savedSort
	})

	makeRecord := func(id uint32, name string) *Nox_gui_server_ent_t {
		srv := &Nox_gui_server_ent_t{Field_9: id, PingVal: 777, PlayersVal: 9, Field_42: 0xba}
		copy(srv.ServerNameBuf[:], name)
		srv.SetAddr("127.0.0.1")
		srv.SetPort(18590)
		return srv
	}
	first := makeRecord(1, "Arena")
	second := makeRecord(2, "arena")
	if got := nox_wol_servers_addResult_4A0030(first.C()); got != 0 {
		t.Fatalf("first insertion index = %d", got)
	}
	if got := nox_wol_servers_addResult_4A0030(second.C()); got != 0 {
		t.Fatalf("equal names must insert before the existing row, got index %d", got)
	}
	entry := sub_4A0490(1)
	if entry == nil || entry != sub_4A04C0(1) || sub_4A0490(2) != sub_4A04C0(0) {
		t.Fatal("record lookup lost native identity or order")
	}
	rows := func(p unsafe.Pointer) []byte { return unsafe.Slice((*byte)(p), 169) }
	if !bytes.Equal(rows(unsafe.Pointer(first)), rows(unsafe.Pointer(entry))) {
		t.Fatal("insertion changed a byte in the packed server payload")
	}
	row, free := alloc.NewHigh(nox_window{})
	t.Cleanup(free)
	nox_wol_server_row_set(entry, row)
	if wolServerNode(unsafe.Pointer(entry)).row != row || !bytes.Equal(rows(unsafe.Pointer(first)), rows(unsafe.Pointer(entry))) {
		t.Fatal("row identity overlapped the packed server payload")
	}
	if sub_4A0410(internCStr("127.0.0.1"), 18590) != 0 || sub_4A0410(internCStr("127.0.0.1"), 18591) != 1 {
		t.Fatal("address/port lookup missed the inserted records")
	}
}
