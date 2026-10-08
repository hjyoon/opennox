package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/sound"
)

type audioEventHooks48EA70 struct {
	connected func() bool
	play      func(sound.ID, int, int) bool
}

// The original A6/A7 packets are four bytes: opcode, signed pan, and a
// little-endian word containing a ten-bit sound ID and six-bit half-volume.
// Admission and distance attenuation remain server-side; both client paths
// use the initialized native FX bank instead of the dormant PE32 allocator.
func handleAudioEventNative48EA70(op netmsg.Op, data []byte, hooks audioEventHooks48EA70) int {
	if op != netmsg.MSG_AUDIO_EVENT && op != netmsg.MSG_AUDIO_PLAYER_EVENT {
		return -1
	}
	const size = 4
	if len(data) < size {
		return -1
	}
	if !hooks.connected() {
		return size
	}
	pan := int(int8(data[1]))
	packed := binary.LittleEndian.Uint16(data[2:size])
	volume := int((packed >> 9) & 0x7e)
	id := sound.ID(packed & 0x3ff)
	hooks.play(id, volume, pan)
	return size
}

func (c *Client) handleAudioEventPacketNative48EA70(op netmsg.Op, data []byte) int {
	return handleAudioEventNative48EA70(op, data, audioEventHooks48EA70{
		connected: nox_client_isConnected,
		play: func(id sound.ID, volume, pan int) bool {
			played := nativeAudioFX.playPanned(id, volume, pan)
			if audioEffectsDebug {
				audioEffectsLog.Printf("network %s sound=%s volume=%d pan=%d submitted=%t", op, id, volume, pan, played)
			}
			return played
		},
	})
}
