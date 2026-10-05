package opennox

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"

	"github.com/opennox/libs/noxnet/netmsg"
)

func consoleServerCommandPacket(id byte, player uint16, command string) []byte {
	if i := strings.IndexByte(command, 0); i >= 0 {
		command = command[:i]
	}
	chars := utf16.Encode([]rune(command))
	// The stock protocol counts UTF-16 units (including the NUL) in one byte.
	if len(chars) > 254 {
		chars = chars[:254]
		if last := chars[len(chars)-1]; last >= 0xD800 && last <= 0xDBFF {
			chars = chars[:len(chars)-1]
		}
	}
	buf := make([]byte, 5+2*(len(chars)+1))
	buf[0], buf[1] = byte(netmsg.MSG_SERVER_CMD), id
	binary.LittleEndian.PutUint16(buf[2:], player)
	if len(chars) != 0 {
		buf[4] = byte(len(chars) + 1)
	}
	for i, char := range chars {
		binary.LittleEndian.PutUint16(buf[5+2*i:], char)
	}
	return buf
}
