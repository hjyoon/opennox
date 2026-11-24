package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestNewPlayerPacketNativeCEntry4DDA90(t *testing.T) {
	player, free := alloc.NewHigh(server.Player{})
	t.Cleanup(free)
	*player = server.Player{
		NetCodeVal: 0x9234, Lessons: -1234, Field2140: 0x89abcdef,
		ArmorEquip: 0x11223344, WeaponEquip: 0x55667788,
		Field2152: 0x1aa, Field2156: 0x2bb, Field3680: 0xffffffff,
	}
	info := make([]byte, server.PlayerInfoWireSize)
	for i := range info {
		info[i] = byte(13*i + 7)
	}
	if err := player.Info().UnmarshalBinary(info); err != nil {
		t.Fatal(err)
	}
	for _, profile := range []string{"", "online", "ABCDEFGHI", "ABCDEFGHIJK"} {
		for _, state := range []byte{2, 3} {
			player.Field3676 = state
			player.SetField2096(profile)
			got := bytes.Repeat([]byte{0xcc}, 132)
			want := bytes.Clone(got)
			copy(want[:3], []byte{45, 0x34, 0x92})
			copy(want[3:100], info)
			copy(want[100:119], []byte{
				0x2e, 0xfb, 0xef, 0xcd, 0x44, 0x33, 0x22, 0x11, 0x88, 0x77, 0x66, 0x55,
				0x23, 0x04, 0, 0, 0xaa, 0xbb, byte(state - 2),
			})
			copy(want[119:], profile)
			want[119+len(profile)] = 0
			callNewPlayerPacketC(&got[0], player)
			if !bytes.Equal(got, want) {
				t.Fatalf("profile %q, state %d: got % x, want % x", profile, state, got, want)
			}
		}
	}
}

func TestGUISettingsPacketNativeCEntry4DD9B0(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	preserveImportantPacketState(t)
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	setImportantNativeListC(nil, nil)
	setImportantAllocClassC(nil)
	seq := memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	seq[6] = 0x1234
	gameFrameHook = func() uint32 { return 87 }

	var settings [58]byte
	for i := range settings {
		settings[i] = byte(17*i + 5)
	}
	want := append([]byte{177, 0xfe}, settings[:]...)
	if got := callGUISettingsPacketC(0xfe, &settings[0], 6); got != 1 {
		t.Fatalf("enqueue returned %d, want 1", got)
	}
	first, last := importantNativeListC()
	if first == nil || last != first {
		t.Fatalf("expected one important packet, got %p and %p", first, last)
	}
	packet := (*importantPacketC)(first)
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&packet.payload[0])), int(packet.payload_size))
	if !bytes.Equal(payload, want) || packet.recipient != 6 || packet.sequence_enabled != 1 ||
		packet.remove_if_disconnected != 0 || uint32(packet.created_frame) != 87 ||
		uint16(packet.sequence[6]) != 0x1234 || seq[6] != 0x1235 {
		t.Fatalf("packet metadata or payload differed: % x", payload)
	}
	clear(settings[:])
	if !bytes.Equal(payload, want) {
		t.Fatal("queued packet aliases caller settings")
	}
	if got := callGUISettingsPacketC(0, &settings[0], 32); got != 0 {
		t.Fatalf("invalid recipient returned %d, want 0", got)
	}
	newFirst, newLast := importantNativeListC()
	if newFirst != first || newLast != last {
		t.Fatal("invalid recipient changed the queue")
	}
}
