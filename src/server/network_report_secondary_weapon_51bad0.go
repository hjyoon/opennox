package server

const networkReportSecondaryWeaponPacketSize51BAD0 = 3

// networkReportSecondaryWeaponHooks51BAD0 exposes the ordered packet reads and
// calls in the MSG_REPORT_SECONDARY_WEAPON branch at GAME.EXE
// 0051C2E4..0051C36B. Object identities remain native-width handles.
type networkReportSecondaryWeaponHooks51BAD0[O comparable] struct {
	loadWireCode      func() uint16
	dynamicUnitCode   func(uint16) uint32
	netDebug          func() bool
	testHighBit       func(uint16)
	objectFromNetCode func(uint32) O
	report            func(O, O)
}

// networkReportSecondaryWeapon51BAD0 preserves the original three-byte
// packet branch. A zero code or an unresolved nonzero code both clear the
// selected weapon by reporting a zero item.
func networkReportSecondaryWeapon51BAD0[O comparable](
	owner O,
	hooks networkReportSecondaryWeaponHooks51BAD0[O],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}
	var item O
	if wireCode != 0 {
		item = hooks.objectFromNetCode(code)
	}
	hooks.report(owner, item)
	return networkReportSecondaryWeaponPacketSize51BAD0
}
