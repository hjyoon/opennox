package opennox

import (
	"encoding/binary"
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

func playerArmorReportFixture4D992A(t *testing.T) (*Server, *server.Object, *server.PlayerUpdateData, *server.Player) {
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
				t.Fatal("armor report fixture must exceed 4 GiB")
			}
		}
	}
	return &Server{Server: native}, unit, update, player
}

// The original x87 FCOMP tests only C3: equal and unordered both skip the
// report. Classify NaNs by their raw exponent/mantissa independently of the
// production floating-point comparison; signed zero must also compare equal.
func armorReportOrderedChange4D992A(previous, current uint32) bool {
	nan := func(bits uint32) bool { return bits&0x7f800000 == 0x7f800000 && bits&0x007fffff != 0 }
	if nan(previous) || nan(current) {
		return false
	}
	if previous&0x7fffffff == 0 && current&0x7fffffff == 0 {
		return false
	}
	return previous != current
}

func TestPlayerReportArmorNative4D992AFloatingPointPacketAndUnchangedCache(t *testing.T) {
	s, unit, update, player := playerArmorReportFixture4D992A(t)
	values := []uint32{0, 0x80000000, 1, 0x80000001, 0x007fffff, 0x00800000,
		0x3e800000, 0x3f000000, 0x3f800000, 0x3f800001, 0xbf800000,
		0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000,
		0x7fc00000, 0x7fc01234, 0x7f800001, 0xff800001, 0xffffffff}
	var packets [][]byte
	var recipient int
	s.Server.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		if index != recipient || related != nil || remove != 1 || sequence != 0 {
			t.Fatal("armor report transport arguments")
		}
		packets = append(packets, append([]byte(nil), packet...))
		return math.MinInt32 // The cache is acknowledged even when the send fails.
	}
	for _, recipient = range []int{0, 31, 128, 255} {
		player.PlayerInd = byte(recipient)
		for _, previous := range values {
			for _, current := range values {
				packets = nil
				update.Field58, update.Field57 = previous, current
				s.playerReportSelfNative4D9900(unit)
				wantCache := previous
				var wantPackets [][]byte
				if armorReportOrderedChange4D992A(previous, current) {
					wantCache = current
					want := []byte{73, 0, 0, 0, 0}
					binary.LittleEndian.PutUint32(want[1:], current)
					wantPackets = [][]byte{want}
				}
				if update.Field58 != wantCache || !reflect.DeepEqual(packets, wantPackets) {
					t.Fatalf("recipient=%d previous=%08x current=%08x cache=%08x want=%08x packets=%x want=%x", recipient, previous, current, update.Field58, wantCache, packets, wantPackets)
				}
				packets = nil
				s.playerReportSelfNative4D9900(unit)
				if len(packets) != 0 {
					t.Fatalf("unchanged armor repeated packets=%x", packets)
				}
			}
		}
	}
}

func TestPlayerReportArmorNative4D992ACachedUpdateAndLivePostSendValue(t *testing.T) {
	s, unit, entry, player := playerArmorReportFixture4D992A(t)
	live, freeLive := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeLive)
	live.Player, live.Field57, live.Field58 = player, math.Float32bits(0.875), math.Float32bits(0.125)
	player.PlayerInd = 31
	entry.Field57 = math.Float32bits(0.5)
	var packets [][]byte
	s.Server.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		packets = append(packets, append([]byte(nil), packet...))
		unit.UpdateData = unsafe.Pointer(live)
		entry.Field57 = math.Float32bits(0.75)
		return -1
	}
	s.playerReportSelfNative4D9900(unit)
	if !reflect.DeepEqual(packets, [][]byte{{73, 0, 0, 0, 0x3f}}) || entry.Field58 != math.Float32bits(0.75) || live.Field58 != math.Float32bits(0.125) {
		t.Fatalf("packet=%x entry cache=%08x live cache=%08x", packets, entry.Field58, live.Field58)
	}
}

func TestPlayerReportArmorNative4D992ABeforeExistingGoldReport(t *testing.T) {
	s, unit, update, player := playerArmorReportFixture4D992A(t)
	player.GoldVal, player.Field2168, player.PlayerInd = 37, 12, 31
	update.Field57 = math.Float32bits(0.25)
	var packets [][]byte
	s.Server.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		packets = append(packets, append([]byte(nil), packet...))
		return 0
	}
	s.playerReportSelfNative4D9900(unit)
	if !reflect.DeepEqual(packets, [][]byte{{73, 0, 0, 0x80, 0x3e}, {74, 37, 0, 0, 0}}) || player.Field2168 != 37 {
		t.Fatalf("armor/gold order=%x gold cache=%d", packets, player.Field2168)
	}
}
