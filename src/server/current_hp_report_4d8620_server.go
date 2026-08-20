package server

type currentHPReportNativeDeps4D8620 struct {
	getUnitNetCode func(*Object) uint32
	sendReliable   func(int32, [4]byte, *Object, int32) int32
}

func currentHPReportNative4D8620(
	recipient int32,
	obj *Object,
	deps currentHPReportNativeDeps4D8620,
) int32 {
	return currentHPReport4D8620(currentHPReportHooks4D8620[*Object, *HealthData]{
		loadObjectArg: func() *Object {
			return obj
		},
		loadHealth: func(obj *Object) *HealthData {
			return obj.HealthData
		},
		getUnitNetCode: deps.getUnitNetCode,
		loadCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		loadRecipient: func() int32 {
			return recipient
		},
		sendReliable: deps.sendReliable,
	})
}

func currentHPReportServerDeps4D8620(s *Server) currentHPReportNativeDeps4D8620 {
	return currentHPReportNativeDeps4D8620{
		getUnitNetCode: func(obj *Object) uint32 {
			return uint32(s.GetUnitNetCode(obj))
		},
		sendReliable: func(recipient int32, packet [4]byte, related *Object, remove int32) int32 {
			return int32(s.NetSendPacketXxx1(int(recipient), packet[:], related, int(remove)))
		},
	}
}

// CurrentHPReport4D8620 sends the original four-byte reliable health packet
// using native-width Object and HealthData pointers.
func (s *Server) CurrentHPReport4D8620(recipient int32, obj *Object) int32 {
	return currentHPReportNative4D8620(recipient, obj, currentHPReportServerDeps4D8620(s))
}
