package opennox

import (
	"image"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"
)

func rayFXTestPacket48EA70(op netmsg.Op) []byte {
	return []byte{byte(op), 0x34, 0x12, 0x78, 0x56, 0xbc, 0x9a, 0xf0, 0xde}
}

func TestDecodeRayFXState48EA70(t *testing.T) {
	for _, op := range []netmsg.Op{
		netmsg.MSG_FX_PLASMA,
		netmsg.MSG_FX_LIGHTNING,
		netmsg.MSG_FX_ENERGY_BOLT,
		netmsg.MSG_FX_CHAIN_LIGHTNING_BOLT,
		netmsg.MSG_FX_DRAIN_MANA,
		netmsg.MSG_FX_CHARM,
		netmsg.MSG_FX_GREATER_HEAL,
	} {
		t.Run(op.String(), func(t *testing.T) {
			data := append(rayFXTestPacket48EA70(op), 0xaa, 0xbb)
			state, ok := decodeRayFXState48EA70(op, data)
			if !ok {
				t.Fatal("packet was rejected")
			}
			if state.Op != op || state.From != image.Pt(0x1234, 0x5678) || state.To != image.Pt(0x9abc, 0xdef0) {
				t.Fatalf("state = %+v", state)
			}
			if got := state.Packet; got != [9]byte{byte(op), 0x34, 0x12, 0x78, 0x56, 0xbc, 0x9a, 0xf0, 0xde} {
				t.Fatalf("fixed packet = %x", got)
			}
			data[1] = 0
			if state.Packet[1] != 0x34 {
				t.Fatal("fixed packet aliases input storage")
			}
		})
	}
}

func TestHandleRayFXNative48EA70(t *testing.T) {
	for _, tc := range []struct {
		op                netmsg.Op
		lightning, plasma int
	}{
		{netmsg.MSG_FX_PLASMA, 0, 1},
		{netmsg.MSG_FX_LIGHTNING, 1, 0},
		{netmsg.MSG_FX_ENERGY_BOLT, 0, 0},
		{netmsg.MSG_FX_CHAIN_LIGHTNING_BOLT, 1, 0},
		{netmsg.MSG_FX_DRAIN_MANA, 0, 0},
		{netmsg.MSG_FX_CHARM, 0, 0},
		{netmsg.MSG_FX_GREATER_HEAL, 0, 0},
	} {
		t.Run(tc.op.String(), func(t *testing.T) {
			var draws, lightning, plasma int
			got := handleRayFXNative48EA70(tc.op, rayFXTestPacket48EA70(tc.op), rayFXHooks48EA70{
				connected: func() bool { return true },
				drawRay: func(packet [9]byte) {
					draws++
					if packet[0] != byte(tc.op) {
						t.Fatalf("draw opcode = %#x", packet[0])
					}
				},
				lightningSparks: func(from, to image.Point) {
					lightning++
					if from != image.Pt(0x1234, 0x5678) || to != image.Pt(0x9abc, 0xdef0) {
						t.Fatalf("lightning endpoints = %v -> %v", from, to)
					}
				},
				plasmaEndSparks: func(to image.Point) {
					plasma++
					if to != image.Pt(0x9abc, 0xdef0) {
						t.Fatalf("plasma endpoint = %v", to)
					}
				},
			})
			if got != rayFXPacketSize48EA70 || draws != 1 || lightning != tc.lightning || plasma != tc.plasma {
				t.Fatalf("result/draw/lightning/plasma = %d/%d/%d/%d", got, draws, lightning, plasma)
			}
		})
	}
}

func TestHandleRayFXNative48EA70Gates(t *testing.T) {
	packet := rayFXTestPacket48EA70(netmsg.MSG_FX_CHARM)
	for n := 0; n < rayFXPacketSize48EA70; n++ {
		if got := handleRayFXNative48EA70(netmsg.MSG_FX_CHARM, packet[:n], rayFXHooks48EA70{}); got != -1 {
			t.Fatalf("length %d = %d", n, got)
		}
	}
	if got := handleRayFXNative48EA70(netmsg.MSG_FX_MAGIC, packet, rayFXHooks48EA70{}); got != -1 {
		t.Fatalf("unsupported opcode = %d", got)
	}
	if got := handleRayFXNative48EA70(netmsg.MSG_FX_DRAIN_MANA, packet, rayFXHooks48EA70{}); got != -1 {
		t.Fatalf("mismatched opcode = %d", got)
	}

	called := false
	if got := handleRayFXNative48EA70(netmsg.MSG_FX_CHARM, packet, rayFXHooks48EA70{
		connected: func() bool { return false },
		drawRay:   func([9]byte) { called = true },
	}); got != rayFXPacketSize48EA70 || called {
		t.Fatalf("disconnected result/call = %d/%t", got, called)
	}
}
