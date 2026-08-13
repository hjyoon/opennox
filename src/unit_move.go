package opennox

import (
	"encoding/binary"
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
)

// unitMoveTruncLow32 reproduces the low EAX half returned by the x87 helper at
// 00566DCC. That helper truncates a float to a signed 64-bit integer and the
// caller at 004E7010 compares only the low 32 bits. Invalid x87 conversions
// produce the integer-indefinite value 0x8000000000000000, whose low half is
// zero.
func unitMoveTruncLow32(v float32) uint32 {
	f := float64(v)
	const (
		minInt64          = -9223372036854775808.0
		maxInt64Exclusive = 9223372036854775808.0
	)
	if math.IsNaN(f) || f < minInt64 || f >= maxInt64Exclusive {
		return 0
	}
	return uint32(int64(f))
}

func unitMoveSameIntegerPosition4E7010(old, next types.Pointf) bool {
	return unitMoveTruncLow32(old.X) == unitMoveTruncLow32(next.X) &&
		unitMoveTruncLow32(old.Y) == unitMoveTruncLow32(next.Y)
}

type unitMoveOnline4E7010Hooks[U, P any] struct {
	frame        func() uint32
	setMoveFrame func(U, uint32)
	player       func(U) P
	playerIndex  func(P) uint8
	markPlayer   func(uint8)
	sendPacket   func(uint8, [5]byte)
}

// unitMoveOnline4E7010 preserves the observable load/call order of the online
// player block in GAME.EXE. In particular, the player pointer is loaded after
// the first frame call and reloaded after marking the player and reading the
// second frame. The update-data pointer itself remains cached throughout.
func unitMoveOnline4E7010[U, P any](ud U, h unitMoveOnline4E7010Hooks[U, P]) {
	frame := h.frame()
	h.setMoveFrame(ud, frame)
	player := h.player(ud)
	h.markPlayer(h.playerIndex(player))

	frame = h.frame()
	var buf [5]byte
	buf[0] = byte(netmsg.MSG_FORGET_DRAWABLES)
	binary.LittleEndian.PutUint32(buf[1:], frame)
	player = h.player(ud)
	h.sendPacket(h.playerIndex(player), buf)
}
