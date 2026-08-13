package opennox

import "unsafe"

// gameBallStatusRecord4E8290 is the four-byte record at original address
// 0x0075318C. GAME.EXE leaves the byte between State and NetCode untouched.
type gameBallStatusRecord4E8290 struct {
	State    uint8
	Reserved uint8
	NetCode  uint16
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(gameBallStatusRecord4E8290{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(gameBallStatusRecord4E8290{}.State)]
	_ = [1]struct{}{}[1-unsafe.Offsetof(gameBallStatusRecord4E8290{}.Reserved)]
	_ = [1]struct{}{}[2-unsafe.Offsetof(gameBallStatusRecord4E8290{}.NetCode)]
)

type gameBallStatusHooks4E8290 struct {
	storeState   func(uint8)
	storeNetCode func(uint16)
	send         func(int32, uint8, uint16) int32
}

// gameBallStatus4E8290 preserves GAME.EXE 004E8290. Both input values are
// already narrowed at the original char/short boundary. The state byte is
// stored before the net code, the intervening record byte is not touched, and
// only then is the exact pair broadcast to recipient 255. The downstream
// 32-bit return value is forwarded without normalization.
func gameBallStatus4E8290(state uint8, netCode uint16, hooks gameBallStatusHooks4E8290) int32 {
	hooks.storeState(state)
	hooks.storeNetCode(netCode)
	return hooks.send(255, state, netCode)
}

func gameBallStatusNative4E8290(
	record *gameBallStatusRecord4E8290,
	state uint8,
	netCode uint16,
	send func(int32, uint8, uint16) int32,
) int32 {
	return gameBallStatus4E8290(state, netCode, gameBallStatusHooks4E8290{
		storeState: func(value uint8) {
			record.State = value
		},
		storeNetCode: func(value uint16) {
			record.NetCode = value
		},
		send: send,
	})
}
