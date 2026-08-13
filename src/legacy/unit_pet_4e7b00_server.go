package legacy

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

type unitPetNativeDeps4E7B00 struct {
	monitor    func(byte, *server.Object)
	mark       func(byte, *server.Object, uint32)
	setOwner   func(*server.Object, *server.Object)
	unmonitor  func(byte, *server.Object)
	unmark     func(byte, *server.Object, uint32)
	clearOwner func(*server.Object)
}

func unitBecomePetNative4E7B00(owner, pet *server.Object, deps unitPetNativeDeps4E7B00) {
	unitBecomePet4E7B00(owner, pet, unitPetNativeHooks4E7B00(deps))
}

func unitBecomeEnemyNative4E7B60(owner, pet *server.Object, deps unitPetNativeDeps4E7B00) {
	unitBecomeEnemy4E7B60(owner, pet, unitPetNativeHooks4E7B00(deps))
}

func unitPetNativeHooks4E7B00(deps unitPetNativeDeps4E7B00) unitPetHooks4E7B00[
	*server.Object,
	*server.PlayerUpdateData,
	*server.Player,
] {
	return unitPetHooks4E7B00[*server.Object, *server.PlayerUpdateData, *server.Player]{
		subclass: func(obj *server.Object) uint32 {
			return uint32(obj.ObjSubClass)
		},
		setSubclass: func(obj *server.Object, subclass uint32) {
			obj.ObjSubClass = object.SubClass(subclass)
		},
		updateData: func(obj *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(obj.UpdateData)
		},
		player: func(update *server.PlayerUpdateData) *server.Player {
			return update.Player
		},
		playerInd: func(player *server.Player) byte {
			return player.PlayerInd
		},
		monitor:    deps.monitor,
		mark:       deps.mark,
		setOwner:   deps.setOwner,
		unmonitor:  deps.unmonitor,
		unmark:     deps.unmark,
		clearOwner: deps.clearOwner,
	}
}

func unitPetRuntimeDeps4E7B00(s *server.Server) unitPetNativeDeps4E7B00 {
	return unitPetNativeDeps4E7B00{
		monitor: func(playerInd byte, obj *server.Object) {
			netMonitorCreatureNative4D9250(s, playerInd, obj)
		},
		mark: func(playerInd byte, obj *server.Object, flags uint32) {
			s.Players.Nox_xxx_netMarkMinimapObject_417190(ntype.PlayerInd(playerInd), obj, flags)
		},
		setOwner: func(owner, obj *server.Object) {
			s.ObjSetOwner(owner, obj)
		},
		unmonitor: func(playerInd byte, obj *server.Object) {
			netUnmonitorCreatureNative4D92A0(s, playerInd, obj)
		},
		unmark: func(playerInd byte, obj *server.Object, flags uint32) {
			s.Players.Nox_xxx_netUnmarkMinimapObj_417300(ntype.PlayerInd(playerInd), obj, flags)
		},
		clearOwner: func(obj *server.Object) {
			s.ObjClearOwner(obj)
		},
	}
}

func netMonitorCreatureNative4D9250(s *server.Server, playerInd byte, obj *server.Object) int {
	var packet [5]byte
	packet[0] = byte(netmsg.MSG_REPORT_MONITOR_CREATURE)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	binary.LittleEndian.PutUint16(packet[3:], obj.TypeInd)
	s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
	return netReportTotalHealthNative4D85C0(s, playerInd, obj)
}

func netUnmonitorCreatureNative4D92A0(s *server.Server, playerInd byte, obj *server.Object) int {
	var packet [3]byte
	packet[0] = byte(netmsg.MSG_REPORT_UNMONITOR_CREATURE)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	return s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
}

func netReportTotalHealthNative4D85C0(s *server.Server, playerInd byte, obj *server.Object) int {
	if obj.HealthData == nil {
		return 0
	}
	var packet [7]byte
	packet[0] = byte(netmsg.MSG_REPORT_TOTAL_HEALTH)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(obj)))
	health := obj.HealthData
	binary.LittleEndian.PutUint16(packet[3:], health.Cur)
	binary.LittleEndian.PutUint16(packet[5:], health.Max)
	return s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
}
