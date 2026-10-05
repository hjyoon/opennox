package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func nox_xxx_netServerCmd_440950(id byte, cmd string) {
	buf := consoleServerCommandPacket(id, uint16(legacy.ClientPlayerNetCode()), cmd)
	nox_xxx_netClientSend2_4E53C0(server.HostPlayerIndex, buf, nil, 1)
}

func nox_xxx_serverHandleClientConsole_443E90(pl *server.Player, a2 byte, cmd string) {
	legacy.Nox_xxx_serverHandleClientConsole_443E90(pl, a2, cmd)
}

func nox_xxx_cmdSayDo_46A4B0(text string, a2 int) {
	legacy.Nox_xxx_cmdSayDo_46A4B0(text, a2)
}

func nox_console_sendSysOpPass_4409D0(pass string) {
	buf := make([]byte, 21)
	buf[0] = byte(netmsg.MSG_SYSOP_PW)
	alloc.StrCopy16B(buf[1:17], pass)
	binary.LittleEndian.PutUint16(buf[17:], 0)
	binary.LittleEndian.PutUint16(buf[19:], uint16(legacy.ClientPlayerNetCode()))
	nox_xxx_netClientSend2_4E53C0(server.HostPlayerIndex, buf, nil, 1)
}
