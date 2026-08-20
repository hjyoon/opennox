package server

const currentHPReportOpcode4D8620 = byte(65)

type currentHPReportHooks4D8620[O, H any] struct {
	loadObjectArg  func() O
	loadHealth     func(O) H
	getUnitNetCode func(O) uint32
	loadCurrent    func(H) uint16
	loadRecipient  func() int32
	sendReliable   func(int32, [4]byte, O, int32) int32
}

// currentHPReport4D8620 preserves GAME.EXE 004D8620. The first HealthData
// load is only a nil gate. Unit-code lookup is followed by a live HealthData
// reload whose Cur access remains unguarded. The recipient is delayed until
// after Cur has been read and the whole signed send result escapes.
func currentHPReport4D8620[O, H comparable](hooks currentHPReportHooks4D8620[O, H]) int32 {
	obj := hooks.loadObjectArg()
	health := hooks.loadHealth(obj)
	var nilHealth H
	if health == nilHealth {
		return 0
	}

	var packet [4]byte
	packet[0] = currentHPReportOpcode4D8620
	netCode := uint16(hooks.getUnitNetCode(obj))
	packet[1] = byte(netCode)
	packet[2] = byte(netCode >> 8)

	health = hooks.loadHealth(obj)
	current := hooks.loadCurrent(health)
	recipient := hooks.loadRecipient()
	packet[3] = byte(current >> 1)

	var nilObject O
	return hooks.sendReliable(recipient, packet, nilObject, 1)
}
