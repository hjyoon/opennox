package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
)

func BuildPlayerStatsReport4D8990(unit *Object, netCode func(*Object) int) ([14]byte, bool) {
	var packet [14]byte
	if unit == nil || !unit.ObjClass.Has(object.ClassPlayer) || unit.UpdateData == nil || unit.HealthData == nil {
		return packet, false
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	if update.Player == nil {
		return packet, false
	}
	player := update.Player
	info := player.Info()
	packet[0] = byte(netmsg.MSG_REPORT_STATS)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(netCode(unit)))
	binary.LittleEndian.PutUint16(packet[3:5], unit.HealthData.Max)
	binary.LittleEndian.PutUint16(packet[5:7], update.ManaMax)
	binary.LittleEndian.PutUint16(packet[7:9], unit.CarryCapacity)
	binary.LittleEndian.PutUint16(packet[9:11], uint16(info.Field2235()))
	binary.LittleEndian.PutUint16(packet[11:13], uint16(info.Field2239()))
	packet[13] = player.Level
	return packet, true
}
