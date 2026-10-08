package opennox

import (
	"encoding/binary"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestAudioEventNative48EA70PackedWords(t *testing.T) {
	for _, op := range []netmsg.Op{netmsg.MSG_AUDIO_EVENT, netmsg.MSG_AUDIO_PLAYER_EVENT} {
		for packed := 0; packed <= 0xffff; packed++ {
			packet := []byte{byte(op), 193, byte(packed), byte(packed >> 8), 0xaa, 0xbb}
			calls := 0
			got := handleAudioEventNative48EA70(op, packet, audioEventHooks48EA70{
				connected: func() bool { calls++; return true },
				play: func(id sound.ID, volume, pan int) bool {
					if calls != 1 || int(id) != packed%1024 || volume != 2*(packed/1024) || pan != -63 {
						t.Fatalf("%s word=%#04x: calls=%d id=%d volume=%d pan=%d", op, packed, calls, id, volume, pan)
					}
					calls++
					return false // Failed or muted playback must still consume four bytes.
				},
			})
			if got != 4 || calls != 2 {
				t.Fatalf("%s word=%#04x: consumed=%d calls=%d", op, packed, got, calls)
			}
		}
	}
}

func TestAudioEventNative48EA70SignedPan(t *testing.T) {
	for _, op := range []netmsg.Op{netmsg.MSG_AUDIO_EVENT, netmsg.MSG_AUDIO_PLAYER_EVENT} {
		for pan := -128; pan <= 127; pan++ {
			packet := []byte{byte(op), byte(pan), 0, 0}
			binary.LittleEndian.PutUint16(packet[2:], uint16(sound.SoundWalkOnStone)|50<<10)
			calls := 0
			if got := handleAudioEventNative48EA70(op, packet, audioEventHooks48EA70{
				connected: func() bool { return true },
				play: func(id sound.ID, volume, gotPan int) bool {
					calls++
					if id != sound.SoundWalkOnStone || volume != 100 || gotPan != pan {
						t.Fatalf("%s wire pan=%d: sound=%v volume=%d signed pan=%d", op, pan, id, volume, gotPan)
					}
					return true
				},
			}); got != 4 || calls != 1 {
				t.Fatalf("%s wire pan=%d: consumed=%d playback calls=%d", op, pan, got, calls)
			}
		}
	}
}

func TestAudioEventNative48EA70Guards(t *testing.T) {
	for _, op := range []netmsg.Op{netmsg.MSG_AUDIO_EVENT, netmsg.MSG_AUDIO_PLAYER_EVENT} {
		for size := 0; size < 4; size++ {
			got := handleAudioEventNative48EA70(op, make([]byte, size), audioEventHooks48EA70{
				connected: func() bool { t.Fatal("short packet loaded connection state"); return true },
				play:      func(sound.ID, int, int) bool { t.Fatal("short packet played sound"); return true },
			})
			if got != -1 {
				t.Fatalf("%s short length=%d: consumed=%d, want -1", op, size, got)
			}
		}
		calls := 0
		got := handleAudioEventNative48EA70(op, []byte{byte(op), 128, 255, 255, 99}, audioEventHooks48EA70{
			connected: func() bool { calls++; return false },
			play:      func(sound.ID, int, int) bool { t.Fatal("disconnected packet played sound"); return true },
		})
		if got != 4 || calls != 1 {
			t.Fatalf("%s disconnected: consumed=%d connection loads=%d", op, got, calls)
		}
	}
	if got := handleAudioEventNative48EA70(netmsg.MSG_FX_RICOCHET, make([]byte, 4), audioEventHooks48EA70{}); got != -1 {
		t.Fatalf("unrelated opcode consumed %d bytes", got)
	}
}

func FuzzAudioEventNative48EA70(f *testing.F) {
	f.Add(byte(netmsg.MSG_AUDIO_EVENT), []byte{0xa6, 0xff, 0xff, 0xff})
	f.Add(byte(netmsg.MSG_AUDIO_PLAYER_EVENT), []byte{0xa7})
	f.Add(byte(0), []byte{})
	f.Fuzz(func(t *testing.T, opcode byte, data []byte) {
		calls := 0
		op := netmsg.Op(opcode)
		n := handleAudioEventNative48EA70(op, data, audioEventHooks48EA70{
			connected: func() bool { return true },
			play: func(id sound.ID, volume, pan int) bool {
				calls++
				if id < 0 || id > 1023 || volume < 0 || volume > 126 || volume%2 != 0 || pan < -128 || pan > 127 {
					t.Fatalf("invalid decoded sound=%d volume=%d pan=%d", id, volume, pan)
				}
				return true
			},
		})
		valid := (op == netmsg.MSG_AUDIO_EVENT || op == netmsg.MSG_AUDIO_PLAYER_EVENT) && len(data) >= 4
		if valid && (n != 4 || calls != 1) || !valid && (n != -1 || calls != 0) {
			t.Fatalf("opcode=%#x length=%d: consumed=%d calls=%d valid=%t", opcode, len(data), n, calls, valid)
		}
	})
}
