package server

import (
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
)

const audioEventPacketListKind501FD0 = uint8(1)

// audioEventPacketHooks501FD0 describes the observable loads and final enqueue
// performed by GAME.EXE 00501FD0. Pointer-shaped values remain native-width
// generic handles; only PE32 scalar values use fixed-width Go types.
type audioEventPacketHooks501FD0[O, U, P, E comparable] struct {
	loadUpdate          func(O) U
	loadEventObject     func(E) O
	loadEventPositionX  func(E) float32
	loadEventSound      func(E) int32
	loadPlayer          func(U) P
	loadPlayerPositionX func(P) float32
	floatToInt          func(float32) int32
	loadWindowWidth     func() int32
	loadPlayerIndex     func(P) uint8
	enqueue             func(uint8, uint8, [4]byte) bool
}

func audioEventPackedWord501FD0(soundID, percentage int32) uint16 {
	return uint16(uint32(uint16(soundID)) | uint32(percentage)<<10)
}

// audioEventSignedDivide501FD0 models IA-32 IDIV, including both fault cases.
// Go defines MinInt32 / -1 to wrap, while IDIV raises a divide exception.
func audioEventSignedDivide501FD0(numerator, denominator int32) int32 {
	if denominator == 0 {
		panic("GAME.EXE 00501FD0 signed division by zero")
	}
	if numerator == math.MinInt32 && denominator == -1 {
		panic("GAME.EXE 00501FD0 signed division overflow")
	}
	return numerator / denominator
}

// sendAudioEventPacket501FD0 preserves GAME.EXE 00501FD0's field-access,
// conversion, arithmetic, reload, and enqueue order. The Player pointer is
// deliberately reloaded after displacement calculation, and both the
// multiply and divide retain IA-32 signed-dword behavior.
func sendAudioEventPacket501FD0[O, U, P, E comparable](
	unit O,
	event E,
	percentage int32,
	hooks audioEventPacketHooks501FD0[O, U, P, E],
) bool {
	update := hooks.loadUpdate(unit)
	eventObject := hooks.loadEventObject(event)

	message := byte(netmsg.MSG_AUDIO_EVENT)
	if unit == eventObject {
		message = byte(netmsg.MSG_AUDIO_PLAYER_EVENT)
	}
	eventX := hooks.loadEventPositionX(event)
	soundID := hooks.loadEventSound(event)
	packed := audioEventPackedWord501FD0(soundID, percentage)

	player := hooks.loadPlayer(update)
	listenerX := hooks.loadPlayerPositionX(player)
	delta := eventX - listenerX
	distance := hooks.floatToInt(delta)
	windowWidth := hooks.loadWindowWidth()
	halfWidth := windowWidth / 2
	numerator := distance * int32(50)
	movement := audioEventSignedDivide501FD0(numerator, halfWidth)

	packet := [4]byte{
		message,
		byte(movement),
		byte(packed),
		byte(packed >> 8),
	}
	player = hooks.loadPlayer(update)
	playerIndex := hooks.loadPlayerIndex(player)
	return hooks.enqueue(playerIndex, audioEventPacketListKind501FD0, packet)
}
