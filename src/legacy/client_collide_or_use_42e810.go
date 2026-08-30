package legacy

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

const (
	clientCollideOrUseOpcode42E810 = byte(0x7B)
	clientStaticUnitMask578B00     = uint32(0x20400000)
)

func clientWireUnitCode578B00(dr *client.Drawable) uint16 {
	if dr == nil {
		return 0
	}
	code := dr.NetCode32
	if code >= 0x8000 {
		return 0
	}
	if uint32(dr.Class())&clientStaticUnitMask578B00 != 0 {
		code |= 0x8000
	}
	return uint16(code)
}

func clientCollideOrUsePacket42E810(dr *client.Drawable, playerStatus uint32) ([3]byte, bool) {
	var packet [3]byte
	if dr == nil || playerStatus&3 != 0 {
		return packet, false
	}
	code := clientWireUnitCode578B00(dr)
	packet[0] = clientCollideOrUseOpcode42E810
	packet[1] = byte(code)
	packet[2] = byte(code >> 8)
	return packet, true
}

func Nox_xxx_clientCollideOrUse_42E810(dr *client.Drawable) {
	if dr == nil {
		return
	}
	var playerStatus uint32
	if pl := Get_dword_8531A0_2576(); pl != nil {
		playerStatus = pl.Field3680
	}
	packet, ok := clientCollideOrUsePacket42E810(dr, playerStatus)
	if !ok {
		return
	}
	GetServer().S().NetList.AddToMsgListCli(ntype.PlayerInd(31), netlist.Kind0, packet[:])
}
