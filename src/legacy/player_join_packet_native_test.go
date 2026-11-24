package legacy

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
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

func TestJoinRosterNativeCEntry4DE300(t *testing.T) {
	type queued struct {
		op   netmsg.Op
		code uint16
		size int
	}
	for _, tc := range []struct {
		name     string
		headless bool
	}{
		{name: "rendered host"},
		{name: "headless host", headless: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handles.InitHigh()
			t.Cleanup(handles.Release)
			s := server.NewHigh(nil, nil, strman.New())
			t.Cleanup(s.Close)
			preserveImportantPacketState(t)
			setImportantNativeListC(nil, nil)
			setImportantAllocClassC(nil)

			oldServer, oldEngine, oldGame := GetServer, noxflags.GetEngine(), noxflags.GetGame()
			oldType := swapJoinFlagTypeC(444)
			t.Cleanup(func() {
				GetServer = oldServer
				noxflags.ResetEngine()
				noxflags.SetEngine(oldEngine)
				noxflags.ResetGame()
				noxflags.SetGame(oldGame)
				swapJoinFlagTypeC(oldType)
			})
			GetServer = func() Server { return &chatLegacyServer528AC0{srv: s} }
			noxflags.ResetEngine()
			noxflags.ResetGame()
			gameFrameHook = func() uint32 { return 99 }
			if tc.headless {
				noxflags.SetEngine(noxflags.EngineNoRendering)
			}
			for _, i := range []ntype.PlayerInd{1, 4, server.HostPlayerIndex} {
				p := s.Players.ByIndRaw(i)
				p.Active, p.PlayerInd, p.NetCodeVal = 1, byte(i), uint32(0x100+i)
				p.SetField2096("player")
			}
			(*memmap.PtrT[[32]uint16](0x5D4594, 1565524))[4] = 0x4321
			unit, freeUnit := alloc.NewHigh(server.Object{})
			t.Cleanup(freeUnit)
			unit.ObjClass, unit.NetCode = object.ClassPlayer, 0x321
			s.Players.ByIndRaw(1).PlayerUnit = unit
			items, freeItems := alloc.Make([]server.Object{}, 3)
			t.Cleanup(freeItems)
			items[0].ObjClass, items[0].TypeInd = object.ClassWeapon, 333
			items[1].ObjClass, items[1].TypeInd = object.ClassInfoBook|object.ClassArmor, 334
			items[2].ObjClass, items[2].TypeInd = object.ClassFlag, 444
			for i := range items {
				items[i].InvHolder = unit
				if i+1 < len(items) {
					items[i].InvNextItem = &items[i+1]
				}
			}
			unit.InvFirstItem = &items[0]

			if got := callJoinRosterC(4); got != nil {
				t.Fatalf("roster returned %p, want nil", got)
			}
			want := []queued{
				{netmsg.MSG_NEW_PLAYER, 0x101, 129},
				{netmsg.MSG_REPORT_MUNDANE_ARMOR_EQUIP, 0x8321, 7},
				{netmsg.MSG_REPORT_MUNDANE_WEAPON_EQUIP, 0x8321, 7},
			}
			if !tc.headless {
				want = append(want, queued{netmsg.MSG_NEW_PLAYER, 0x11f, 129})
			}
			head, _ := importantNativeListC()
			callJoinInventoryC(4, s.Players.ByIndRaw(4))
			after, _ := importantNativeListC()
			if head != after {
				t.Fatal("player without a unit added inventory packets")
			}
			packet := (*importantPacketC)(head)
			for i := len(want) - 1; i >= 0; i-- {
				exp := want[i]
				if packet == nil {
					t.Fatalf("missing queued packet %d", i)
				}
				payload := unsafe.Slice((*byte)(unsafe.Pointer(&packet.payload[0])), int(packet.payload_size))
				if len(payload) != exp.size || netmsg.Op(payload[0]) != exp.op ||
					binary.LittleEndian.Uint16(payload[1:3]) != exp.code || packet.recipient != 4 ||
					packet.sequence_enabled != 1 || uint16(packet.sequence[4]) != uint16(0x4321+i) {
					t.Fatalf("queued packet %d = % x, expected %v", i, payload, exp)
				}
				packet = packet.next
			}
			if packet != nil {
				t.Fatal("roster added extra packets")
			}
		})
	}
}

func TestGauntletNoticeNativeCEntry4D9E70(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantAllocClassC(nil)
	gameFrameHook = func() uint32 { return 38 }
	if got := callGauntletNoticeC(17); got != 1 {
		t.Fatalf("Gauntlet notice returned %d, want 1", got)
	}
	first, last := importantNativeListC()
	if first == nil || last != first {
		t.Fatalf("expected one notice, got %p and %p", first, last)
	}
	packet := (*importantPacketC)(first)
	data := unsafe.Slice((*byte)(unsafe.Pointer(&packet.payload[0])), int(packet.payload_size))
	if !bytes.Equal(data, []byte{240, 20}) || packet.recipient != 17 ||
		packet.sequence_enabled != 0 || packet.remove_if_disconnected != 1 {
		t.Fatalf("Gauntlet notice differs: % x", data)
	}
}
