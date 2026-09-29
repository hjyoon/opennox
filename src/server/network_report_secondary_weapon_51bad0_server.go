package server

import "encoding/binary"

// NetworkReportSecondaryWeaponPacketSize51BAD0 is the exact
// MSG_REPORT_SECONDARY_WEAPON packet width.
const NetworkReportSecondaryWeaponPacketSize51BAD0 = networkReportSecondaryWeaponPacketSize51BAD0

// NetworkReportSecondaryWeaponRuntime51BAD0 supplies the engine-debug
// observation and native-width secondary-weapon report dispatch.
type NetworkReportSecondaryWeaponRuntime51BAD0 struct {
	NetDebug    func() bool
	TestHighBit func(uint16)
	Report      func(*Object, *Object)
}

// NetworkReportSecondaryWeapon51BAD0 binds the packet decoder to native
// Object pointers on every supported architecture.
func (s *Server) NetworkReportSecondaryWeapon51BAD0(
	owner *Object,
	packet *[NetworkReportSecondaryWeaponPacketSize51BAD0]byte,
	runtime NetworkReportSecondaryWeaponRuntime51BAD0,
) int32 {
	return networkReportSecondaryWeapon51BAD0(owner, networkReportSecondaryWeaponHooks51BAD0[*Object]{
		loadWireCode: func() uint16 {
			return binary.LittleEndian.Uint16(packet[1:3])
		},
		dynamicUnitCode:   s.packetDynamicUnitCode578B40,
		netDebug:          runtime.NetDebug,
		testHighBit:       runtime.TestHighBit,
		objectFromNetCode: s.ObjectFromNetCode4ECCB0,
		report:            runtime.Report,
	})
}
