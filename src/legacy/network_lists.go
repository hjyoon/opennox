package legacy

/*
#include "defs.h"
extern unsigned int dword_5d4594_2650652;
int nox_xxx_netOnPacketRecvCli_48EA70(int a1, unsigned char* data, int sz);
int sub_48D660();
int sub_4DF9B0(void* a1, void* a2, void* a3, int a4);
void nox_xxx_netImportant_4E5770(uint8_t player_index, int message_kind);
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

var (
	GetNetPlayerBufSize         func() int
	Nox_netlist_addToMsgListSrv func(ind ntype.PlayerInd, buf []byte) bool
	importantNetSendHook        = func(playerIndex uint8, messageKind int, data []byte) bool {
		ind := ntype.PlayerInd(playerIndex)
		kind := netlist.Kind(messageKind)
		if playerIndex == 31 {
			return GetServer().S().NetList.AddToMsgListCli(ind, kind, data)
		}
		return GetServer().S().NetList.ClientSend0(ind, kind, data, GetNetPlayerBufSize())
	}
)

//export nox_netlist_addToMsgListCli_40EBC0
func nox_netlist_addToMsgListCli_40EBC0(ind1, ind2 int, buf *C.uchar, sz int) int {
	return bool2int(GetServer().S().NetList.AddToMsgListCli(ntype.PlayerInd(ind1), netlist.Kind(ind2), unsafe.Slice((*byte)(unsafe.Pointer(buf)), sz)))
}

//export nox_netlist_clientSendWrap_40ECA0
func nox_netlist_clientSendWrap_40ECA0(ind1, ind2 int, buf *C.uchar, sz int) int {
	return bool2int(GetServer().S().NetList.ClientSend0(ntype.PlayerInd(ind1), netlist.Kind(ind2), unsafe.Slice((*byte)(unsafe.Pointer(buf)), sz), GetNetPlayerBufSize()))
}

//export nox_server_importantSend_4E5770
func nox_server_importantSend_4E5770(playerIndex C.uint8_t, messageKind C.int, data *C.uint8_t, size C.uint32_t) C.int {
	buf := unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))
	return C.int(bool2int(importantNetSendHook(uint8(playerIndex), int(messageKind), buf)))
}

//export nox_netlist_addToMsgListSrv_40EF40
func nox_netlist_addToMsgListSrv_40EF40(ind int, buf *C.uchar, sz int) C.bool {
	return C.bool(Nox_netlist_addToMsgListSrv(ntype.PlayerInd(ind), unsafe.Slice((*byte)(unsafe.Pointer(buf)), sz)))
}

func Nox_xxx_netImportant_4E5770(a1 byte, a2 int) {
	C.nox_xxx_netImportant_4E5770(C.uint8_t(a1), C.int(a2))
}
