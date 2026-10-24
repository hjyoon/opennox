package server

import (
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// Exercise the existing production entry and real player roster before adding
// an injectable boundary. Keeping a nonzero marker and maximum fade suppresses
// audio/network effects without suppressing original sampling or its stores.
func TestPlayerCamperNative4D7CC0QuietSampling(t *testing.T) {
	for _, tc := range []struct {
		name               string
		position, previous types.Pointf
		wantStamp          uint32
	}{
		{"stationary-binary32", types.Ptf(123.5, -45.25), types.Ptf(123.5, -45.25), 100},
		{"moved-binary32", types.Ptf(200, -90), types.Ptf(123.5, -45.25), 7},
		{"signed-zero", types.Ptf(math.Float32frombits(0x80000000), 0), types.Ptf(0, math.Float32frombits(0x80000000)), 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newQuestPlayerStateFixture4D79C0(1)
			s, unit, ud := f.server, &f.units[0], &f.updates[0]
			s.SetFrame(100)
			s.Players.Camper.init(s)
			s.Players.Camper.LastCheck = 50
			s.Players.Camper.StartTime = 10
			s.Players.Camper.FadeTime = math.MaxUint32
			s.Players.Camper.RadiusSq = 900
			unit.PosVec = tc.position
			ud.Field62, ud.Field63 = math.Float32bits(tc.previous.X), math.Float32bits(tc.previous.Y)
			ud.Field64, ud.IsCamping = 7, 1
			assertQuestPlayerStatePointerAbove32Bits4D79C0(t, "camper unit", unsafe.Pointer(unit))
			assertQuestPlayerStatePointerAbove32Bits4D79C0(t, "camper update", unsafe.Pointer(ud))
			assertQuestPlayerStatePointerAbove32Bits4D79C0(t, "camper player", unsafe.Pointer(ud.Player))
			s.Players.Camper.Update()
			if ud.Field62 != math.Float32bits(tc.position.X) || ud.Field63 != math.Float32bits(tc.position.Y) || ud.Field64 != tc.wantStamp || ud.IsCamping != 1 || s.Players.Camper.LastCheck != 100 {
				t.Fatalf("cache=%08x/%08x want=%08x/%08x stamp=%d/%d marker=%d last=%d", ud.Field62, ud.Field63, math.Float32bits(tc.position.X), math.Float32bits(tc.position.Y), ud.Field64, tc.wantStamp, ud.IsCamping, s.Players.Camper.LastCheck)
			}
		})
	}
}

// The production adapter must retain the real native player iteration and
// marker packet endpoints. CTF without flag possession suppresses audio while
// leaving the original On marker transition active.
func TestPlayerCamperNative4D7CC0MarkerPackets(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCTF)
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	for _, on := range []bool{true, false} {
		t.Run(map[bool]string{true: "on", false: "off"}[on], func(t *testing.T) {
			f := newQuestPlayerStateFixture4D79C0(2)
			s := f.server
			s.SetFrame(100)
			s.Players.Camper.init(s)
			s.Players.Camper.LastCheck = 100
			s.Players.Camper.StartTime = 10
			s.Players.Camper.FadeTime = 5
			for i := range f.updates {
				f.updates[i].Field64 = 100
				if !on {
					f.updates[i].Field64 = 0
					f.updates[i].IsCamping = 1
				}
			}
			calls := 0
			s.NetSendPacketXxx = func(recipient int, packet []byte, related *Object, remove, sequence int) int {
				unitIndex := calls / 2
				recipientIndex := calls % 2
				if unitIndex >= 2 || recipient != recipientIndex || related != nil || remove != 1 || sequence != 0 || len(packet) != 7 {
					t.Fatal("native marker send", recipient, packet, remove, sequence)
				}
				marker := byte(1)
				before := uint32(0)
				if !on {
					marker = 2
					before = 1
				}
				if packet[0] != byte(netmsg.MSG_INTERESTING_ID) || binary.LittleEndian.Uint16(packet[1:3]) != uint16(f.units[unitIndex].NetCode) || binary.LittleEndian.Uint16(packet[3:5]) != f.units[unitIndex].TypeInd || packet[5] != marker || packet[6] != 2 || f.updates[unitIndex].IsCamping != before {
					t.Fatal("native marker payload/store order", packet, f.updates[unitIndex].IsCamping)
				}
				calls++
				return 1
			}
			s.Players.Camper.Update()
			want := uint32(0)
			if on {
				want = 1
			}
			if calls != 4 || f.updates[0].IsCamping != want || f.updates[1].IsCamping != want {
				t.Fatal("native roster/marker result", calls, f.updates[0].IsCamping, f.updates[1].IsCamping)
			}
		})
	}
}
