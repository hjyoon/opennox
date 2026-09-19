package legacy

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

const clientTradeOpcode42E850 = uint16(0x15C9)

func clientTradePacket42E850(dr *client.Drawable, playerStatus uint32, npcDialogState, quitMenuState int) ([4]byte, bool) {
	var packet [4]byte
	if dr == nil || playerStatus&3 != 0 || npcDialogState == 1 || quitMenuState == 1 {
		return packet, false
	}
	code := clientWireUnitCode578B00(dr)
	packet[0] = byte(clientTradeOpcode42E850 & 0xFF)
	packet[1] = byte(clientTradeOpcode42E850 >> 8)
	packet[2] = byte(code)
	packet[3] = byte(code >> 8)
	return packet, true
}

func clientTrade42E850(dr *client.Drawable, playerStatus uint32, npcDialogState, quitMenuState int) {
	packet, ok := clientTradePacket42E850(dr, playerStatus, npcDialogState, quitMenuState)
	if !ok {
		return
	}
	GetServer().S().NetList.AddToMsgListCli(ntype.PlayerInd(31), netlist.Kind0, packet[:])
}

func Nox_xxx_clientTrade_42E850(dr *client.Drawable) {
	var playerStatus uint32
	if pl := Get_dword_8531A0_2576(); pl != nil {
		// The legacy C implementation read the original Win32 byte offset 3680
		// directly. Player contains native-width pointers, so that offset does
		// not address Field3680 on 64-bit builds.
		playerStatus = pl.Field3680
	}
	clientTrade42E850(dr, playerStatus, Sub_47A260(), Nox_gui_xxx_check_446360())
}
