package server

type scavengerReportNativeDeps4D8CD0 struct {
	unitCode   func(*Object) uint32
	sendPacket func(int32, [7]byte, *Object, int32) int32
}

func scavengerReportNative4D8CD0(
	owner *Object,
	deps scavengerReportNativeDeps4D8CD0,
) scavengerReportResult4D8CD0[*Object] {
	return scavengerReport4D8CD0(scavengerReportHooks4D8CD0[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		unitCode: deps.unitCode,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadCount: func(player *Player) uint16 {
			return uint16(player.Field2152)
		},
		loadMaximum: func(player *Player) uint16 {
			return uint16(player.Field2156)
		},
		sendPacket: deps.sendPacket,
	})
}

// ScavengerHuntReport4D8CD0 emits the original seven-byte report.
func (s *Server) ScavengerHuntReport4D8CD0(owner *Object) {
	ScavengerHuntReportResult4D8CD0(owner, s.NetSendPacketXxx0)
}

// ScavengerHuntReportResult4D8CD0 returns the non-player owner or the signed
// packet-send result with a nil owner.
func ScavengerHuntReportResult4D8CD0(owner *Object, send func(int, []byte, *Object, int) int) (*Object, int32) {
	result := scavengerReportNative4D8CD0(owner, scavengerReportNativeDeps4D8CD0{
		unitCode: func(obj *Object) uint32 {
			return uint32(ObjectNetCode(obj))
		},
		sendPacket: func(recipient int32, packet [7]byte, related *Object, remove int32) int32 {
			return int32(send(int(recipient), packet[:], related, int(remove)))
		},
	})
	return result.owner, result.send
}
