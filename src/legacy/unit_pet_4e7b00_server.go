package legacy

import (
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
	packet := server.ObjectTypePacket(netmsg.MSG_REPORT_MONITOR_CREATURE, obj)
	s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
	return netReportTotalHealthNative4D85C0(s, playerInd, obj)
}

func netUnmonitorCreatureNative4D92A0(s *server.Server, playerInd byte, obj *server.Object) int {
	packet := creatureUnmonitorPacket(obj)
	return s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
}

func netReportTotalHealthNative4D85C0(s *server.Server, playerInd byte, obj *server.Object) int {
	packet, ok := server.CreatureTotalHealthPacket(obj)
	if !ok {
		return 0
	}
	return s.NetSendPacketXxx1(int(playerInd), packet[:], nil, 1)
}

// NetReportTotalHealthNative4D85C0 exposes the native-width implementation to
// the restored player-stat reporting path.
func NetReportTotalHealthNative4D85C0(s *server.Server, playerInd byte, obj *server.Object) int {
	return netReportTotalHealthNative4D85C0(s, playerInd, obj)
}
