package opennox

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

func requireObjectStatusHighAddress48EA70(t *testing.T, ptr unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", ptr)
	}
}

func objectStatusTestPacket48EA70(op netmsg.Op, size int) []byte {
	packet := make([]byte, size)
	packet[0] = byte(op)
	return packet
}

func TestHandleObjectStatusNative48EA70RejectsShortPackets(t *testing.T) {
	ops := []netmsg.Op{
		netmsg.MSG_REPORT_X_STATUS,
		netmsg.MSG_REPORT_PLAYER_STATUS,
		netmsg.MSG_REPORT_MODIFIER,
		netmsg.MSG_REPORT_STAT_MODIFIER,
		netmsg.MSG_REPORT_ANIMATION_FRAME,
	}
	for _, op := range ops {
		size, ok := objectStatusPacketSize48EA70(op)
		if !ok {
			t.Fatalf("missing size for opcode %d", op)
		}
		for n := 0; n < size; n++ {
			if got := handleObjectStatusNative48EA70(op, make([]byte, n), objectStatusPacketHooks48EA70{}); got != -1 {
				t.Fatalf("opcode %d accepted %d-byte packet: %d", op, n, got)
			}
		}
	}
	if got := handleObjectStatusNative48EA70(netmsg.MSG_PING, make([]byte, 32), objectStatusPacketHooks48EA70{}); got != -1 {
		t.Fatalf("unsupported opcode result = %d, want -1", got)
	}
	if got := handleNPCReportNative48EA70(make([]byte, npcReportPacketSize48EA70-1), npcReportPacketHooks48EA70{}); got != -1 {
		t.Fatalf("short NPC packet result = %d, want -1", got)
	}
	if got := handleClientStatusNative48EA70(make([]byte, clientStatusReportPacketSize48EA70-1), clientStatusPacketHooks48EA70{}); got != -1 {
		t.Fatalf("short client-status packet result = %d, want -1", got)
	}
}

func TestHandleObjectXStatusNative48EA70PreservesHighAddressAndNamespace(t *testing.T) {
	dr := new(client.Drawable)
	requireObjectStatusHighAddress48EA70(t, unsafe.Pointer(dr))
	dr.ObjClass = object.ClassMonsterGenerator | object.ClassLight | object.ClassMonster
	dr.ObjFlags = object.FlagFlicker | object.FlagActive
	dr.Flags70Val = 0x20

	packet := objectStatusTestPacket48EA70(netmsg.MSG_REPORT_X_STATUS, objectXStatusPacketSize48EA70)
	binary.LittleEndian.PutUint16(packet[1:3], 0x9234)
	binary.LittleEndian.PutUint32(packet[3:7], 0x80000c55)
	wantPacket := append([]byte(nil), packet...)
	var codes []uint16
	prepared := 0
	hooks := objectStatusPacketHooks48EA70{
		connected: func() bool { return true },
		byNetCode: func(code uint16) *client.Drawable {
			codes = append(codes, code)
			return dr
		},
		prepareGenerator: func(got *client.Drawable) {
			prepared++
			if got != dr || got.Flags70Val != 0x80000c55 {
				t.Fatalf("prepare callback got drawable %p flags %#x", got, got.Flags70Val)
			}
			got.UnionEffect().Field_108 = 77
		},
	}
	if got := handleObjectStatusNative48EA70(netmsg.MSG_REPORT_X_STATUS, packet, hooks); got != objectXStatusPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d, want %d", got, objectXStatusPacketSize48EA70)
	}
	if !reflect.DeepEqual(codes, []uint16{0x9234}) {
		t.Fatalf("lookup codes = %#v, want [0x9234]", codes)
	}
	if prepared != 1 || dr.UnionEffect().Field_108 != 77 {
		t.Fatalf("prepare result = calls %d, field %d", prepared, dr.UnionEffect().Field_108)
	}
	if dr.Flags70Val != 0x80000c55 {
		t.Fatalf("flags70 = %#x, want 0x80000c55", dr.Flags70Val)
	}
	if dr.ObjClass.Has(object.ClassLight) ||
		!dr.ObjClass.Has(object.ClassMonsterGenerator) ||
		!dr.ObjClass.Has(object.ClassMonster) {
		t.Fatalf("class after terminal status = %#x", dr.ObjClass)
	}
	if dr.ObjFlags.Has(object.FlagFlicker) || !dr.ObjFlags.Has(object.FlagActive) {
		t.Fatalf("object flags after terminal status = %#x", dr.ObjFlags)
	}
	if !reflect.DeepEqual(packet, wantPacket) {
		t.Fatalf("packet mutated: got %x, want %x", packet, wantPacket)
	}
}

func TestPrepareMonsterGeneratorStatus48EA70(t *testing.T) {
	dr := new(client.Drawable)
	data := new([33]byte)
	data[27] = 7
	data[32] = 0xff
	dr.DrawData = unsafe.Pointer(&data[0])
	prepareMonsterGeneratorStatus48EA70(dr)
	if got := dr.UnionEffect().Field_108; got != 7*256 {
		t.Fatalf("prepared frame span = %d, want %d", got, 7*256)
	}
	prepareMonsterGeneratorStatus48EA70(nil)
	prepareMonsterGeneratorStatus48EA70(&client.Drawable{})
}

func TestHandlePlayerStatusNative48EA70UsesNativeDrawable(t *testing.T) {
	dr := new(client.Drawable)
	requireObjectStatusHighAddress48EA70(t, unsafe.Pointer(dr))
	dr.ObjFlags = object.FlagActive
	packet := objectStatusTestPacket48EA70(netmsg.MSG_REPORT_PLAYER_STATUS, playerStatusPacketSize48EA70)
	binary.LittleEndian.PutUint32(packet[1:5], uint32(object.FlagEnabled|object.FlagSelected))
	hooks := objectStatusPacketHooks48EA70{
		connected:     func() bool { return true },
		localDrawable: func() *client.Drawable { return dr },
	}
	if got := handleObjectStatusNative48EA70(netmsg.MSG_REPORT_PLAYER_STATUS, packet, hooks); got != playerStatusPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d, want %d", got, playerStatusPacketSize48EA70)
	}
	if want := object.FlagEnabled | object.FlagSelected; dr.ObjFlags != want {
		t.Fatalf("local drawable flags = %#x, want %#x", dr.ObjFlags, want)
	}
}

func TestHandleObjectModifierNative48EA70PreservesPointers(t *testing.T) {
	dr := new(client.Drawable)
	requireObjectStatusHighAddress48EA70(t, unsafe.Pointer(dr))
	dr.Field_113 = 0xaabbccdd
	mods := [4]*server.ModifierEff{new(server.ModifierEff), nil, new(server.ModifierEff), new(server.ModifierEff)}
	for _, mod := range mods {
		if mod != nil {
			requireObjectStatusHighAddress48EA70(t, mod.C())
		}
	}
	packet := []byte{byte(netmsg.MSG_REPORT_MODIFIER), 0x34, 0x92, 7, 0xff, 9, 10}
	wantPacket := append([]byte(nil), packet...)
	var ids []byte
	hooks := objectStatusPacketHooks48EA70{
		connected: func() bool { return true },
		byNetCode: func(code uint16) *client.Drawable {
			if code != 0x9234 {
				t.Fatalf("lookup code = %#x, want 0x9234", code)
			}
			return dr
		},
		modifierByID: func(id byte) *server.ModifierEff {
			ids = append(ids, id)
			switch id {
			case 7:
				return mods[0]
			case 9:
				return mods[2]
			case 10:
				return mods[3]
			default:
				return nil
			}
		},
	}
	if got := handleObjectStatusNative48EA70(netmsg.MSG_REPORT_MODIFIER, packet, hooks); got != objectModifierPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d, want %d", got, objectModifierPacketSize48EA70)
	}
	item := drawableItemModifiers48EA70(dr)
	if item.Field_108 != mods[0].C() || item.Field_109 != nil || item.Field_110 != mods[2].C() || item.Field_111 != mods[3].C() {
		t.Fatalf("modifier pointers = [%p %p %p %p]", item.Field_108, item.Field_109, item.Field_110, item.Field_111)
	}
	if item.Field_112_0 != -1 || item.Field_112_2 != -1 || dr.Field_113 != 0xaabbccdd {
		t.Fatalf("modifier cache/adjacent field = %d/%d/%#x", item.Field_112_0, item.Field_112_2, dr.Field_113)
	}
	if !reflect.DeepEqual(ids, []byte{7, 0xff, 9, 10}) {
		t.Fatalf("modifier IDs = %v", ids)
	}
	if !reflect.DeepEqual(packet, wantPacket) {
		t.Fatalf("packet mutated: got %x, want %x", packet, wantPacket)
	}
}

func TestHandleStatModifierAndAnimationFrameNative48EA70(t *testing.T) {
	stat := objectStatusTestPacket48EA70(netmsg.MSG_REPORT_STAT_MODIFIER, statModifierPacketSize48EA70)
	binary.LittleEndian.PutUint16(stat[1:3], 0x9234)
	binary.LittleEndian.PutUint32(stat[3:7], math.Float32bits(-12.5))
	stat[7] = 19
	var statIndex byte
	var statValue float32
	hooks := objectStatusPacketHooks48EA70{
		connected:    func() bool { return true },
		localNetCode: func() uint32 { return 0x1234 },
		setStatModifier: func(index byte, value float32) {
			statIndex, statValue = index, value
		},
	}
	if got := handleObjectStatusNative48EA70(netmsg.MSG_REPORT_STAT_MODIFIER, stat, hooks); got != statModifierPacketSize48EA70 {
		t.Fatalf("stat consumed bytes = %d", got)
	}
	if statIndex != 19 || statValue != -12.5 {
		t.Fatalf("stat update = %d/%v, want 19/-12.5", statIndex, statValue)
	}

	dr := new(client.Drawable)
	requireObjectStatusHighAddress48EA70(t, unsafe.Pointer(dr))
	dr.AnimFrameSlave = 7
	frame := objectStatusTestPacket48EA70(netmsg.MSG_REPORT_ANIMATION_FRAME, animationFramePacketSize48EA70)
	binary.LittleEndian.PutUint16(frame[1:3], 0x9234)
	binary.LittleEndian.PutUint32(frame[3:7], 0x10203040)
	hooks.byNetCode = func(code uint16) *client.Drawable {
		if code != 0x9234 {
			t.Fatalf("frame lookup code = %#x", code)
		}
		return dr
	}
	if got := handleObjectStatusNative48EA70(netmsg.MSG_REPORT_ANIMATION_FRAME, frame, hooks); got != animationFramePacketSize48EA70 {
		t.Fatalf("frame consumed bytes = %d", got)
	}
	if dr.Field_78 != 7 || dr.AnimFrameSlave != 0x10203040 {
		t.Fatalf("animation frames = previous %#x, current %#x", dr.Field_78, dr.AnimFrameSlave)
	}
}

func TestHandleObjectStatusNative48EA70DisconnectedConsumesWithoutMutation(t *testing.T) {
	hooks := objectStatusPacketHooks48EA70{
		connected: func() bool { return false },
		byNetCode: func(uint16) *client.Drawable {
			t.Fatal("drawable lookup while disconnected")
			return nil
		},
		localDrawable: func() *client.Drawable {
			t.Fatal("local drawable lookup while disconnected")
			return nil
		},
		modifierByID: func(byte) *server.ModifierEff {
			t.Fatal("modifier lookup while disconnected")
			return nil
		},
		setStatModifier: func(byte, float32) { t.Fatal("stat update while disconnected") },
	}
	for _, op := range []netmsg.Op{
		netmsg.MSG_REPORT_X_STATUS,
		netmsg.MSG_REPORT_PLAYER_STATUS,
		netmsg.MSG_REPORT_MODIFIER,
		netmsg.MSG_REPORT_STAT_MODIFIER,
		netmsg.MSG_REPORT_ANIMATION_FRAME,
	} {
		size, _ := objectStatusPacketSize48EA70(op)
		if got := handleObjectStatusNative48EA70(op, objectStatusTestPacket48EA70(op, size), hooks); got != size {
			t.Fatalf("opcode %d consumed %d bytes, want %d", op, got, size)
		}
	}
	if got := handleNPCReportNative48EA70(make([]byte, npcReportPacketSize48EA70), npcReportPacketHooks48EA70{
		connected: func() bool { return false },
		byID: func(int) *server.NPC {
			t.Fatal("NPC lookup while disconnected")
			return nil
		},
	}); got != npcReportPacketSize48EA70 {
		t.Fatalf("NPC consumed bytes = %d", got)
	}
}

func TestHandleNPCReportNative48EA70ResetsAndColorsHighAddressRecord(t *testing.T) {
	npc := new(server.NPC)
	requireObjectStatusHighAddress48EA70(t, npc.C())
	npc.WeaponEquip = 0xffffffff
	packet := objectStatusTestPacket48EA70(netmsg.MSG_REPORT_NPC, npcReportPacketSize48EA70)
	binary.LittleEndian.PutUint16(packet[1:3], 0x9234)
	for i := 0; i < 18; i++ {
		packet[3+i] = byte(i + 1)
	}
	wantPacket := append([]byte(nil), packet...)
	var events []string
	hooks := npcReportPacketHooks48EA70{
		connected: func() bool { return true },
		byID: func(id int) *server.NPC {
			events = append(events, "lookup")
			if id != 0x1234 {
				t.Fatalf("NPC ID = %#x, want 0x1234", id)
			}
			return npc
		},
		newNPC: func(int) *server.NPC {
			t.Fatal("allocated an existing NPC")
			return nil
		},
		resetNPC: func(got *server.NPC, id int) {
			events = append(events, "reset")
			*got = server.NPC{LiveVal: 1, IDVal: int32(id)}
		},
		color: func(r, g, b byte) uint32 {
			events = append(events, "color")
			return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
		},
	}
	if got := handleNPCReportNative48EA70(packet, hooks); got != npcReportPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d", got)
	}
	if npc.IDVal != 0x1234 || npc.LiveVal != 1 || npc.Field1312 != 1 || npc.WeaponEquip != 0 {
		t.Fatalf("NPC identity/state = live %d id %#x high %d weapon %#x", npc.LiveVal, npc.IDVal, npc.Field1312, npc.WeaponEquip)
	}
	for i, got := range npc.Color8 {
		off := 3 * i
		want := uint32(packet[3+off])<<16 | uint32(packet[4+off])<<8 | uint32(packet[5+off])
		if got != want {
			t.Fatalf("color %d = %#x, want %#x", i, got, want)
		}
	}
	wantEvents := []string{"lookup", "reset", "color", "color", "color", "color", "color", "color"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if !reflect.DeepEqual(packet, wantPacket) {
		t.Fatalf("packet mutated: got %x, want %x", packet, wantPacket)
	}
}

func TestHandleNPCReportNative48EA70AllocatesMissingRecord(t *testing.T) {
	npc := new(server.NPC)
	packet := objectStatusTestPacket48EA70(netmsg.MSG_REPORT_NPC, npcReportPacketSize48EA70)
	binary.LittleEndian.PutUint16(packet[1:3], 0x4321)
	created := 0
	if got := handleNPCReportNative48EA70(packet, npcReportPacketHooks48EA70{
		connected: func() bool { return true },
		byID:      func(int) *server.NPC { return nil },
		newNPC: func(id int) *server.NPC {
			created++
			*npc = server.NPC{LiveVal: 1, IDVal: int32(id)}
			return npc
		},
	}); got != npcReportPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d", got)
	}
	if created != 1 || npc.IDVal != 0x4321 || npc.Field1312 != 0 {
		t.Fatalf("created NPC = calls %d id %#x high %d", created, npc.IDVal, npc.Field1312)
	}
}

func TestHandleClientStatusNative48EA70AppliesMaskBeforeGUI(t *testing.T) {
	player := &server.Player{NetCodeVal: 0x1234, Field3680: 0x80000423}
	requireObjectStatusHighAddress48EA70(t, player.C())
	packet := objectStatusTestPacket48EA70(netmsg.MSG_REPORT_CLIENT_STATUS, clientStatusReportPacketSize48EA70)
	binary.LittleEndian.PutUint16(packet[1:3], 0x1234)
	binary.LittleEndian.PutUint32(packet[3:7], 0xdeadbeef)
	wantPacket := append([]byte(nil), packet...)
	wantMask := uint32(0xdeadbeef) & clientStatusReportMask48EA70
	wantStatus := uint32(0x80000423)&^clientStatusReportClearMask48EA70 | wantMask
	var events []string
	poisoned := -1
	hooks := clientStatusPacketHooks48EA70{
		playerByID: func(id int) *server.Player {
			events = append(events, "lookup")
			if id != 0x1234 {
				t.Fatalf("player ID = %#x", id)
			}
			return player
		},
		host: func() bool { return false },
		unsetStatus: func(got *server.Player, mask uint32) {
			events = append(events, "unset")
			if got != player || mask != clientStatusReportClearMask48EA70 {
				t.Fatalf("unset args = %p/%#x", got, mask)
			}
			got.Field3680 &^= mask
		},
		needStatus: func(got *server.Player, mask uint32) {
			events = append(events, "need")
			if got != player || mask != wantMask {
				t.Fatalf("need args = %p/%#x, want mask %#x", got, mask, wantMask)
			}
			got.Field3680 |= mask
		},
		renderingDisabled: func() bool { return false },
		localNetCode:      func() uint32 { return 0x1234 },
		onClientStatus: func(status uint32) {
			events = append(events, "gui")
			if status != wantStatus {
				t.Fatalf("GUI status = %#x, want %#x", status, wantStatus)
			}
		},
		setPoisoned: func(value int) {
			events = append(events, "poison")
			poisoned = value
		},
	}
	if got := handleClientStatusNative48EA70(packet, hooks); got != clientStatusReportPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d", got)
	}
	if player.Field3680 != wantStatus || poisoned != int((wantStatus>>10)&1) {
		t.Fatalf("final status/poison = %#x/%d, want %#x/%d", player.Field3680, poisoned, wantStatus, (wantStatus>>10)&1)
	}
	if want := []string{"lookup", "unset", "need", "gui", "poison"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if !reflect.DeepEqual(packet, wantPacket) {
		t.Fatalf("packet mutated: got %x, want %x", packet, wantPacket)
	}
}

func TestHandleClientStatusNative48EA70HostAndRenderGates(t *testing.T) {
	player := &server.Player{Field3680: 0x401}
	packet := objectStatusTestPacket48EA70(netmsg.MSG_REPORT_CLIENT_STATUS, clientStatusReportPacketSize48EA70)
	binary.LittleEndian.PutUint16(packet[1:3], 7)
	hooks := clientStatusPacketHooks48EA70{
		playerByID: func(int) *server.Player { return player },
		host:       func() bool { return true },
		unsetStatus: func(*server.Player, uint32) {
			t.Fatal("host packet cleared status")
		},
		needStatus: func(*server.Player, uint32) {
			t.Fatal("host packet set status")
		},
		renderingDisabled: func() bool { return true },
		localNetCode:      func() uint32 { return 7 },
		onClientStatus:    func(uint32) { t.Fatal("render-disabled packet updated GUI") },
		setPoisoned:       func(int) { t.Fatal("render-disabled packet updated poison meter") },
	}
	if got := handleClientStatusNative48EA70(packet, hooks); got != clientStatusReportPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d", got)
	}
	if player.Field3680 != 0x401 {
		t.Fatalf("host status changed to %#x", player.Field3680)
	}

	lookedUp := 0
	if got := handleClientStatusNative48EA70(packet, clientStatusPacketHooks48EA70{
		playerByID: func(int) *server.Player {
			lookedUp++
			return nil
		},
		host: func() bool { t.Fatal("host check after missing player"); return false },
	}); got != clientStatusReportPacketSize48EA70 || lookedUp != 1 {
		t.Fatalf("missing player result/lookups = %d/%d", got, lookedUp)
	}
}
