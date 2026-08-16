package server

import "encoding/binary"

const (
	scavengerReportPlayerClass4D8CD0 = uint8(4)
	scavengerReportOpcode4D8CD0      = uint8(85)
	scavengerReportRecipient4D8CD0   = int32(255)
	scavengerReportRemove4D8CD0      = int32(1)
)

type scavengerReportResultKind4D8CD0 uint8

const (
	scavengerReportOwnerResult4D8CD0 scavengerReportResultKind4D8CD0 = iota
	scavengerReportSendResult4D8CD0
)

// scavengerReportResult4D8CD0 separates the original mixed EAX domains. The
// non-Player path returns the owner pointer, while the Player path forwards the
// packet-send result.
type scavengerReportResult4D8CD0[O any] struct {
	kind  scavengerReportResultKind4D8CD0
	owner O
	send  int32
}

// scavengerReportHooks4D8CD0 exposes the exact native-pointer field loads and
// callback boundaries in GAME.EXE 004D8CD0. UpdateData is cached before the
// Player gate and its Player pointer is loaded separately for both counters.
type scavengerReportHooks4D8CD0[O, U, L any] struct {
	loadOwnerArg func() O
	loadClassLow func(O) uint8
	loadUpdate   func(O) U
	unitCode     func(O) uint32
	loadPlayer   func(U) L
	loadCount    func(L) uint16
	loadMaximum  func(L) uint16
	sendPacket   func(int32, [7]byte, O, int32) int32
}

// scavengerReport4D8CD0 preserves GAME.EXE 004D8CD0. The packet is always
// little-endian and deliberately truncates the unit code and both counters to
// their low 16 bits, matching the original MOVW stores.
func scavengerReport4D8CD0[O, U, L any](hooks scavengerReportHooks4D8CD0[O, U, L]) scavengerReportResult4D8CD0[O] {
	owner := hooks.loadOwnerArg()
	classLow := hooks.loadClassLow(owner)
	update := hooks.loadUpdate(owner)
	if classLow&scavengerReportPlayerClass4D8CD0 == 0 {
		return scavengerReportResult4D8CD0[O]{
			kind:  scavengerReportOwnerResult4D8CD0,
			owner: owner,
		}
	}

	var packet [7]byte
	packet[0] = scavengerReportOpcode4D8CD0
	binary.LittleEndian.PutUint16(packet[1:3], uint16(hooks.unitCode(owner)))
	player := hooks.loadPlayer(update)
	binary.LittleEndian.PutUint16(packet[3:5], hooks.loadCount(player))
	player = hooks.loadPlayer(update)
	binary.LittleEndian.PutUint16(packet[5:7], hooks.loadMaximum(player))
	var related O
	return scavengerReportResult4D8CD0[O]{
		kind: scavengerReportSendResult4D8CD0,
		send: hooks.sendPacket(
			scavengerReportRecipient4D8CD0,
			packet,
			related,
			scavengerReportRemove4D8CD0,
		),
	}
}
