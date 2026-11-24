package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
)

func GameSettingsPacket(frame, version, gameFlags, serverFlags uint32, playerLimit, modeLimit, timed byte) [20]byte {
	packet := [20]byte{byte(netmsg.MSG_GAME_SETTINGS)}
	binary.LittleEndian.PutUint32(packet[1:5], frame)
	binary.LittleEndian.PutUint32(packet[5:9], version)
	binary.LittleEndian.PutUint32(packet[9:13], gameFlags&0x7fff0)
	binary.LittleEndian.PutUint32(packet[13:17], serverFlags)
	packet[17], packet[18], packet[19] = playerLimit, modeLimit, timed
	return packet
}

func GameSettingsDetailsPacket(name string, settings *Settings2) [49]byte {
	packet := [49]byte{byte(netmsg.MSG_GAME_SETTINGS_2)}
	copy(packet[1:16], name)
	for i, value := range settings.Field24.Vals {
		binary.LittleEndian.PutUint32(packet[17+4*i:], value)
	}
	copy(packet[37:41], settings.Field44[:])
	binary.LittleEndian.PutUint32(packet[41:45], settings.Field48)
	return packet
}
