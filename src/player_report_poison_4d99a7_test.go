package opennox

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func playerPoisonReportFixture4D99A7(t *testing.T) (*Server, *server.Object, *server.PlayerUpdateData, *server.Player) {
	t.Helper()
	old := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(old) })
	native := server.New(nil, nil, strman.New())
	t.Cleanup(native.Close)
	unit, freeUnit := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	player, freePlayer := alloc.New(server.Player{})
	t.Cleanup(freeUnit)
	t.Cleanup(freeUpdate)
	t.Cleanup(freePlayer)
	unit.ObjClass, unit.UpdateData = object.ClassPlayer, unsafe.Pointer(update)
	update.Player = player
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(player)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatal("poison report fixture must exceed 4 GiB")
			}
		}
	}
	return &Server{Server: native}, unit, update, player
}

func TestPlayerReportPoisonNative4D99A7AllBytePairs(t *testing.T) {
	s, unit, _, player := playerPoisonReportFixture4D99A7(t)
	player.PlayerInd = 31
	unit.Poison540 = 0xa7 // Independent +540 field must not supply the +440 report.
	unit.Field542 = 1000
	var packet []byte
	sends := 0
	s.Server.NetSendPacketXxx = func(index int, buf []byte, related *server.Object, remove, sequence int) int {
		if index != 31 || related != nil || remove != 1 || sequence != 0 {
			t.Fatal("poison report transport arguments")
		}
		sends++
		packet = append([]byte(nil), buf...)
		return math.MinInt32 // A failed send still acknowledges the live byte.
	}
	for previous := 0; previous < 256; previous++ {
		for current := 0; current < 256; current++ {
			player.Field2172, unit.Field110 = byte(previous), 0x8abcde00|uint32(current)
			packet, sends = nil, 0
			s.playerReportSelfNative4D9900(unit)
			wantSends := 0
			var want []byte
			if previous != current {
				wantSends, want = 1, []byte{91, byte(current)}
			}
			if sends != wantSends || !reflect.DeepEqual(packet, want) || player.Field2172 != byte(current) ||
				unit.Field110 != 0x8abcde00|uint32(current) || unit.Poison540 != 0xa7 || unit.Field542 != 1000 {
				t.Fatalf("previous=%d current=%d sends=%d want=%d packet=%x want=%x cache=%d poison=%d timer=%d",
					previous, current, sends, wantSends, packet, want, player.Field2172, unit.Poison540, unit.Field542)
			}
			packet, sends = nil, 0
			s.playerReportSelfNative4D9900(unit)
			if sends != 0 {
				t.Fatalf("unchanged poison %d repeated packet=%x", current, packet)
			}
		}
	}
}

func TestPlayerReportPoisonNative4D99A7UnsignedRecipients(t *testing.T) {
	s, unit, _, player := playerPoisonReportFixture4D99A7(t)
	var packets [][]byte
	var recipient int
	s.Server.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		if index != recipient || related != nil || remove != 1 || sequence != 0 {
			t.Fatalf("recipient=%d want=%d related=%p remove=%d sequence=%d", index, recipient, related, remove, sequence)
		}
		packets = append(packets, append([]byte(nil), packet...))
		return 0
	}
	for recipient = 0; recipient < 256; recipient++ {
		player.PlayerInd, player.Field2172, unit.Field110 = byte(recipient), 0, 0xabcde7
		packets = nil
		s.playerReportSelfNative4D9900(unit)
		if !reflect.DeepEqual(packets, [][]byte{{91, 0xe7}}) || player.Field2172 != 0xe7 {
			t.Fatalf("recipient=%d packets=%x cache=%d", recipient, packets, player.Field2172)
		}
	}
}

func TestPlayerReportPoisonNative4D99A7ReportsOnlyTheLowItemEnchantmentByte(t *testing.T) {
	s, unit, _, player := playerPoisonReportFixture4D99A7(t)
	var packets [][]byte
	s.Server.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		packets = append(packets, append([]byte(nil), packet...))
		return 0
	}
	unit.Poison540, unit.Field110 = 2, 0x12345600
	s.playerReportSelfNative4D9900(unit)
	if len(packets) != 0 || player.Field2172 != 0 {
		t.Fatalf("poison or high mask bits reported as item enchantment: packets=%x cache=%d", packets, player.Field2172)
	}
	unit.Field110 = 0xffffffff
	s.playerReportSelfNative4D9900(unit)
	if !reflect.DeepEqual(packets, [][]byte{{91, 255}}) || player.Field2172 != 255 || unit.Field110 != 0xffffffff || unit.Poison540 != 2 {
		t.Fatalf("low byte report changed source: packets=%x cache=%d mask=%08x poison=%d", packets, player.Field2172, unit.Field110, unit.Poison540)
	}
	packets = nil
	unit.Field110 = 0xaabbccff
	s.playerReportSelfNative4D9900(unit)
	if len(packets) != 0 || player.Field2172 != 255 {
		t.Fatal("unchanged low byte repeated after only high mask bits changed")
	}
}

func TestPlayerReportPoisonNative4D99A7CachedUpdateAndPostSendPlayerAndValue(t *testing.T) {
	s, unit, entry, before := playerPoisonReportFixture4D99A7(t)
	live, freeLive := alloc.New(server.PlayerUpdateData{})
	after, freeAfter := alloc.New(server.Player{})
	foreign, freeForeign := alloc.New(server.Player{})
	t.Cleanup(freeLive)
	t.Cleanup(freeAfter)
	t.Cleanup(freeForeign)
	live.Player = foreign
	before.PlayerInd, before.Field2172, unit.Field110 = 31, 1, 2
	after.Field2172, foreign.Field2172 = 88, 99
	var packets [][]byte
	s.Server.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		if index != 31 || related != nil || remove != 1 || sequence != 0 {
			t.Fatal("pre-call recipient or transport changed")
		}
		packets = append(packets, append([]byte(nil), packet...))
		unit.UpdateData, entry.Player, unit.Field110 = unsafe.Pointer(live), after, 0xabcdf1
		return -1
	}
	s.playerReportSelfNative4D9900(unit)
	if !reflect.DeepEqual(packets, [][]byte{{91, 2}}) || before.Field2172 != 1 || after.Field2172 != 0xf1 || foreign.Field2172 != 99 {
		t.Fatalf("packets=%x before=%d after=%d foreign=%d", packets, before.Field2172, after.Field2172, foreign.Field2172)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(live), unsafe.Pointer(after), unsafe.Pointer(foreign)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatal("callback-rebound poison pointers must exceed 4 GiB")
			}
		}
	}
}

func TestPlayerReportPoisonNative4D99A7ReloadsAfterGoldCallback(t *testing.T) {
	s, unit, entry, goldPlayer := playerPoisonReportFixture4D99A7(t)
	poisonPlayer, freePoisonPlayer := alloc.New(server.Player{})
	t.Cleanup(freePoisonPlayer)
	goldPlayer.PlayerInd, goldPlayer.GoldVal, goldPlayer.Field2168 = 31, 37, 12
	poisonPlayer.PlayerInd, poisonPlayer.Field2172, unit.Field110 = 128, 3, 2
	var packets [][]byte
	s.Server.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		if related != nil || remove != 1 || sequence != 0 {
			t.Fatal("transport arguments")
		}
		packets = append(packets, append([]byte(nil), packet...))
		switch len(packets) {
		case 1:
			if index != 31 {
				t.Fatal("gold recipient changed")
			}
			entry.Player, unit.Field110 = poisonPlayer, 4

		case 2:
			if index != 128 {
				t.Fatal("poison recipient was not reloaded after gold callback")
			}
		default:
			t.Fatalf("unexpected packet=%x", packet)
		}
		return -1
	}
	s.playerReportSelfNative4D9900(unit)
	if !reflect.DeepEqual(packets, [][]byte{{74, 37, 0, 0, 0}, {91, 4}}) || poisonPlayer.Field2172 != 4 || goldPlayer.Field2172 != 0 {
		t.Fatalf("gold/poison order=%x poison cache=%d entry cache=%d", packets, poisonPlayer.Field2172, goldPlayer.Field2172)
	}
}
