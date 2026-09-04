package legacy

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

const clientTalkOpcode42E7B0 = uint16(0x01D0)

func clientTalkPacket42E7B0(dr *client.Drawable, playerStatus uint32, inventoryState, quitMenuState int) ([4]byte, bool) {
	var packet [4]byte
	if dr == nil || playerStatus&3 != 0 || inventoryState == 1 || quitMenuState == 1 {
		return packet, false
	}
	code := uint16(dr.NetCode32)
	packet[0] = byte(clientTalkOpcode42E7B0 & 0xFF)
	packet[1] = byte(clientTalkOpcode42E7B0 >> 8)
	packet[2] = byte(code)
	packet[3] = byte(code >> 8)
	return packet, true
}

func clientTalk42E7B0(dr *client.Drawable, playerStatus uint32, inventoryState, quitMenuState int) {
	packet, ok := clientTalkPacket42E7B0(dr, playerStatus, inventoryState, quitMenuState)
	if !ok {
		return
	}
	GetServer().S().NetList.AddToMsgListCli(ntype.PlayerInd(31), netlist.Kind0, packet[:])
}

func Nox_xxx_clientTalk_42E7B0(dr *client.Drawable) {
	var playerStatus uint32
	if pl := Get_dword_8531A0_2576(); pl != nil {
		playerStatus = pl.Field3680
	}
	clientTalk42E7B0(dr, playerStatus, Sub_478030(), Nox_gui_xxx_check_446360())
}
