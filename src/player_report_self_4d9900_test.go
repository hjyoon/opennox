package opennox

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerReportSelfNative4D9900CachesQuestUpdateBeforeGoldCallback(t *testing.T) {
	old := noxflags.GetGame()
	noxflags.SetGame(noxflags.GameModeQuest)
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(old) })
	native := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(native.Close)
	s := &Server{Server: native}
	unit, freeUnit := alloc.NewHigh(server.Object{})
	entry, freeEntry := alloc.NewHigh(server.PlayerUpdateData{})
	live, freeLive := alloc.NewHigh(server.PlayerUpdateData{})
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

func TestPlayerReportSelfNative4D9900SelectsOriginalVitalsMode(t *testing.T) {
	mode := memmap.PtrT[uint32](0x587000, 87404)
	oldMode, oldFlags := *mode, noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		*mode = oldMode
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	native := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(native.Close)
	native.SetTickRate(30)
	native.SetFrame(100)
	s := &Server{Server: native}
	health, freeHealth := alloc.NewHigh(server.HealthData{})
	t.Cleanup(freeHealth)
	*health = server.HealthData{Cur: 5, Max: 100}
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	unit.ObjClass, unit.HealthData = object.ClassPlayer, health
	player := native.Players.ByIndRaw(5)
	player.Active, player.PlayerInd, player.PlayerUnit, player.HealthReportFrame = 1, 5, unit, 100
	defer func() { player.PlayerUnit = nil }()
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	update.Player, update.Field2_1 = player, 4
	unit.UpdateData = unsafe.Pointer(update)
	packets := 0
	native.NetSendPacketXxx = func(index int, payload []byte, related *server.Object, remove, sequence int) int {
		if index != 5 || !reflect.DeepEqual(payload, []byte{0x43, 5, 0}) || related != nil || remove != 1 || sequence != 1 {
			t.Fatal("unexpected health report")
		}
		packets++
		return -1
	}
	*mode = 1
	s.playerReportSelfNative4D9900(unit)
	if packets != 0 || update.Field2_1 != 4 {
		t.Fatalf("mode 1 produced %d packets and cached %d, want zero packets and 4", packets, update.Field2_1)
	}
	*mode = 2
	s.playerReportSelfNative4D9900(unit)
	if packets != 1 || update.Field2_1 != 5 {
		t.Fatalf("startup mode produced %d packets and cached %d, want one packet and 5", packets, update.Field2_1)
	}
}
