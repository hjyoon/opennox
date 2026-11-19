package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
)

// SendTeamBaseModifierReport4D92D0 sends all four modifier slots. The caller
// supplies the init record it read before resolving the TeamBase type.
func (s *Server) SendTeamBaseModifierReport4D92D0(recipient int, obj *Object, attrs *ModifierInitData, teamBaseType int32) int32 {
	const eligible = object.ClassFlag | object.ClassWeapon | object.ClassArmor | object.ClassWand
	if !obj.Class().HasAny(eligible) && int32(obj.TypeInd) != teamBaseType {
		return teamBaseType
	}
	packet := [7]byte{byte(netmsg.MSG_REPORT_MODIFIER), 0, 0, 0xff, 0xff, 0xff, 0xff}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(ObjectNetCode(obj)))
	if attrs != nil {
		for i, mod := range attrs.Modifiers {
			if mod != nil {
				packet[i+3] = byte(mod.Index())
			}
		}
	}
	return int32(s.NetSendPacketXxx0(recipient, packet[:], nil, 1))
}
