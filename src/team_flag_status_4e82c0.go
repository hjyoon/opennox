package opennox

import "unsafe"

const (
	teamFlagStatusBaseOffset4E82C0 uintptr = 1567740
	teamFlagStatusRecordSize4E82C0 uintptr = 6
)

// teamFlagStatusRecord4E82C0 is one six-byte team record beginning at the
// original address 0x00753190. GAME.EXE leaves byte +3 untouched.
type teamFlagStatusRecord4E82C0 struct {
	TeamID         uint8
	FlagIndex      uint8
	Status         uint8
	Reserved       uint8
	CarrierNetCode uint16
}

var (
	_ = [1]struct{}{}[6-unsafe.Sizeof(teamFlagStatusRecord4E82C0{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(teamFlagStatusRecord4E82C0{}.TeamID)]
	_ = [1]struct{}{}[1-unsafe.Offsetof(teamFlagStatusRecord4E82C0{}.FlagIndex)]
	_ = [1]struct{}{}[2-unsafe.Offsetof(teamFlagStatusRecord4E82C0{}.Status)]
	_ = [1]struct{}{}[3-unsafe.Offsetof(teamFlagStatusRecord4E82C0{}.Reserved)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(teamFlagStatusRecord4E82C0{}.CarrierNetCode)]
)

type teamFlagStatusHooks4E82C0 struct {
	storeTeamID         func(uint8)
	storeFlagIndex      func(uint8)
	storeStatus         func(uint8)
	storeCarrierNetCode func(uint16)
	send                func(int32, uint8, uint8, uint8, uint16) int32
}

func teamFlagStatusRecordOffset4E82C0(teamID uint8) uintptr {
	return teamFlagStatusBaseOffset4E82C0 + uintptr(teamID)*teamFlagStatusRecordSize4E82C0
}

// teamFlagStatus4E82C0 preserves GAME.EXE 004E82C0. The inputs are already
// narrowed at the original byte/word boundary. Stores occur in original
// memory order, the reserved byte is not touched, and only then is the exact
// tuple broadcast to recipient 255. The downstream 32-bit result is forwarded
// without normalization.
func teamFlagStatus4E82C0(
	teamID, status, flagIndex uint8,
	carrierNetCode uint16,
	hooks teamFlagStatusHooks4E82C0,
) int32 {
	hooks.storeTeamID(teamID)
	hooks.storeFlagIndex(flagIndex)
	hooks.storeStatus(status)
	hooks.storeCarrierNetCode(carrierNetCode)
	return hooks.send(255, teamID, status, flagIndex, carrierNetCode)
}

func teamFlagStatusNative4E82C0(
	record *teamFlagStatusRecord4E82C0,
	teamID, status, flagIndex uint8,
	carrierNetCode uint16,
	send func(int32, uint8, uint8, uint8, uint16) int32,
) int32 {
	return teamFlagStatus4E82C0(teamID, status, flagIndex, carrierNetCode, teamFlagStatusHooks4E82C0{
		storeTeamID: func(value uint8) {
			record.TeamID = value
		},
		storeFlagIndex: func(value uint8) {
			record.FlagIndex = value
		},
		storeStatus: func(value uint8) {
			record.Status = value
		},
		storeCarrierNetCode: func(value uint16) {
			record.CarrierNetCode = value
		},
		send: send,
	})
}
