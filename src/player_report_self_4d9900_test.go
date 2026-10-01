package opennox

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerReportSelfNative4D9900CachesQuestUpdateBeforeGoldCallback(t *testing.T) {
	old := noxflags.GetGame()
	noxflags.SetGame(noxflags.GameModeQuest)
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(old) })
	native := server.New(nil, nil, strman.New())
	t.Cleanup(native.Close)
	s := &Server{Server: native}
	unit, freeUnit := alloc.New(server.Object{})
	entry, freeEntry := alloc.New(server.PlayerUpdateData{})
	live, freeLive := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeUnit)
	t.Cleanup(freeEntry)
	t.Cleanup(freeLive)
	player := native.Players.ByIndRaw(ntype.PlayerInd(31))
	player.Active, player.PlayerInd, player.PlayerUnit = 1, 31, unit
	player.GoldVal, player.Field2168 = 37, 12
	entry.Player, entry.ExtraLives = player, 2
	live.Player, live.ExtraLives = player, 7
	unit.ObjClass, unit.NetCode, unit.UpdateData = object.ClassPlayer, 0x9234, unsafe.Pointer(entry)
	var packets [][]byte
	native.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		if index != 31 || related != nil || remove != 1 || sequence != 0 {
			t.Fatal("transport arguments")
		}
		packets = append(packets, append([]byte(nil), packet...))
		switch len(packets) {
		case 1:
			unit.UpdateData = unsafe.Pointer(live) // The original loop must still compare/acknowledge entry.
		case 2:
			entry.ExtraLives = 0x105 // Original BYTE is reloaded after the report.
		default:
			t.Fatalf("unexpected report=%x", packet)
		}
		return math.MinInt32
	}
	s.playerReportSelfNative4D9900(unit)
	if !reflect.DeepEqual(packets, [][]byte{{74, 37, 0, 0, 0}, {0xf0, 4, 7, 0x34, 0x92}}) || entry.ExtraLives != 0x105 || entry.RespawnMarkers[31] != 5 || live.RespawnMarkers[31] != 0 || player.Field2168 != 37 {
		t.Fatalf("reports=%x entry/lives=%d/%d live marker=%d gold cache=%d", packets, entry.ExtraLives, entry.RespawnMarkers[31], live.RespawnMarkers[31], player.Field2168)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(player)) <= math.MaxUint32 || uintptr(unsafe.Pointer(entry)) <= math.MaxUint32 || uintptr(unsafe.Pointer(live)) <= math.MaxUint32) {
		t.Fatal("native fixture pointers must exceed 4 GiB")
	}
}

func TestPlayerReportSelfNative4D9900RetainsOuterNilAndClassGates(t *testing.T) {
	s := &Server{}
	s.playerReportSelfNative4D9900(nil)
	s.playerReportSelfNative4D9900(&server.Object{ObjClass: object.ClassMonster})
}
