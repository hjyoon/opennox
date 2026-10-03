package opennox

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func playerQuestKeyReportFixture4D9A3F(t *testing.T) (*Server, *server.Object, *server.PlayerUpdateData, *server.Player) {
	t.Helper()
	s, unit, update, player := playerItemEnchantmentReportFixture4D99A7(t)
	noxflags.SetGame(noxflags.GameModeQuest)
	for key, value := range []uint32{0x9234, 0xb678} {
		cache := memmap.PtrT[uint32](0x5D4594, uintptr(1556324+4*key))
		old := *cache
		*cache = value
		t.Cleanup(func() { *cache = old })
	}
	for _, index := range []ntype.PlayerInd{0, 3, 31} {
		recipient := s.Players.ByIndRaw(index)
		recipient.Active, recipient.PlayerInd = 1, 255
		// Key reports only require the player-info record, not a player unit.
		recipient.PlayerUnit = nil
	}
	unit.NetCode = 0xaabbf234
	return s, unit, update, player
}

func TestPlayerReportQuestKeysNative4D9A3FRestoresBothMissingLoops(t *testing.T) {
	s, unit, update, _ := playerQuestKeyReportFixture4D9A3F(t)
	silver, freeSilver := alloc.New(server.Object{})
	gold, freeGold := alloc.New(server.Object{})
	t.Cleanup(freeSilver)
	t.Cleanup(freeGold)
	*silver = server.Object{TypeInd: 0x9234, ObjClass: object.ClassKey}
	*gold = server.Object{TypeInd: 0xb678, ObjClass: object.ClassKey}
	unit.InvFirstItem, silver.InvNextItem = silver, gold
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(silver)) <= math.MaxUint32 || uintptr(unsafe.Pointer(gold)) <= math.MaxUint32) {
		t.Fatal("native key links must exceed 4 GiB")
	}
	if !noxflags.HasGame(noxflags.GameModeQuest) || s.Players.ByInd(0) == nil ||
		*memmap.PtrT[uint32](0x5D4594, 1556324) != uint32(silver.TypeInd) ||
		*memmap.PtrT[uint32](0x5D4594, 1556328) != uint32(gold.TypeInd) {
		t.Fatal("Quest key fixture is not admitted or its type caches/recipient are inconsistent")
	}
	var recipients []int
	var packets [][]byte
	s.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		if related != nil || remove != 1 || sequence != 0 {
			t.Fatal("key report transport arguments")
		}
		recipients = append(recipients, index)
		packets = append(packets, append([]byte(nil), packet...))
		return math.MinInt32 // Failure still acknowledges the computed presence.
	}
	s.playerReportSelfNative4D9900(unit)
	want := [][]byte{{0xf0, 22, 1, 0x34, 0xf2}, {0xf0, 22, 1, 0x34, 0xf2}, {0xf0, 22, 1, 0x34, 0xf2},
		{0xf0, 23, 1, 0x34, 0xf2}, {0xf0, 23, 1, 0x34, 0xf2}, {0xf0, 23, 1, 0x34, 0xf2}}
	if !reflect.DeepEqual(recipients, []int{0, 3, 31, 0, 3, 31}) || !reflect.DeepEqual(packets, want) {
		t.Fatalf("missing original Quest key report: recipients=%v packets=%x quest=%t caches=%x/%x inventory=%p", recipients, packets,
			noxflags.HasGame(noxflags.GameModeQuest), *memmap.PtrT[uint32](0x5D4594, 1556324), *memmap.PtrT[uint32](0x5D4594, 1556328), unit.InvFirstItem)
	}
	for index := range update.QuestPlayerFlagsA {
		value := byte(0)
		if index == 0 || index == 3 || index == 31 {
			value = 1
		}
		if update.QuestPlayerFlagsA[index] != value || update.QuestPlayerFlagsB[index] != value {
			t.Fatalf("recipient %d caches=%d/%d want=%d", index, update.QuestPlayerFlagsA[index], update.QuestPlayerFlagsB[index], value)
		}
	}
	s.playerReportSelfNative4D9900(unit)
	if len(packets) != 6 {
		t.Fatal("unchanged key reports repeated")
	}
	unit.InvFirstItem = nil
	s.playerReportSelfNative4D9900(unit)
	for _, packet := range packets[6:] {
		if packet[2] != 0 {
			t.Fatalf("lost key did not report absence: %x", packet)
		}
	}
	if len(packets) != 12 {
		t.Fatalf("key removal reports=%d want=12", len(packets))
	}
}

func TestPlayerReportQuestKeysNative4D9A3FCachesQuestAdmissionAndEntryUpdate(t *testing.T) {
	s, unit, entry, player := playerQuestKeyReportFixture4D9A3F(t)
	live, freeLive := alloc.New(server.PlayerUpdateData{})
	key, freeKey := alloc.New(server.Object{})
	t.Cleanup(freeLive)
	t.Cleanup(freeKey)
	live.Player, key.TypeInd = player, 0x9234
	recipient := s.Players.ByIndRaw(31)
	recipient.PlayerUnit = unit
	entry.ExtraLives, unit.InvFirstItem = 1, key
	var packets [][]byte
	s.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		packets = append(packets, append([]byte(nil), packet...))
		if len(packets) == 1 {
			if index != 31 || packet[1] != 4 {
				t.Fatalf("original life report must precede key reports: %d/%x", index, packet)
			}
			unit.UpdateData = unsafe.Pointer(live)
			noxflags.UnsetGame(noxflags.GameModeQuest)
		}
		return -1
	}
	s.playerReportSelfNative4D9900(unit)
	if !reflect.DeepEqual(packets, [][]byte{{0xf0, 4, 1, 0x34, 0xf2}, {0xf0, 22, 1, 0x34, 0xf2},
		{0xf0, 22, 1, 0x34, 0xf2}, {0xf0, 22, 1, 0x34, 0xf2}}) ||
		entry.QuestPlayerFlagsA[0] != 1 || entry.QuestPlayerFlagsA[3] != 1 || entry.QuestPlayerFlagsA[31] != 1 ||
		live.QuestPlayerFlagsA != [32]byte{} || unit.UpdateData != unsafe.Pointer(live) {
		t.Fatalf("lost entry update or single Quest admission: packets=%x entry=%v live=%v", packets, entry.QuestPlayerFlagsA, live.QuestPlayerFlagsA)
	}
}

func TestPlayerReportQuestKeysNative4D9A3FSkipsOutsideQuest(t *testing.T) {
	s, unit, update, _ := playerQuestKeyReportFixture4D9A3F(t)
	noxflags.UnsetGame(noxflags.GameModeQuest)
	update.QuestPlayerFlagsA[0], update.QuestPlayerFlagsB[31] = 0x77, 0xaa
	for kind := 0; kind < 2; kind++ {
		*memmap.PtrT[uint32](0x5D4594, uintptr(1556324+4*kind)) = 0
	}
	s.NetSendPacketXxx = func(int, []byte, *server.Object, int, int) int {
		t.Fatal("non-Quest self-report must not report keys")
		return 0
	}
	s.playerReportSelfNative4D9900(unit)
	if update.QuestPlayerFlagsA[0] != 0x77 || update.QuestPlayerFlagsB[31] != 0xaa ||
		*memmap.PtrT[uint32](0x5D4594, 1556324) != 0 || *memmap.PtrT[uint32](0x5D4594, 1556328) != 0 {
		t.Fatal("non-Quest self-report accessed key types or recipient markers")
	}
	// The existing outer gates still precede every update/key access.
	noxflags.SetGame(noxflags.GameModeQuest)
	s.playerReportSelfNative4D9900(nil)
	unit.ObjClass, unit.UpdateData = object.ClassMonster, nil
	s.playerReportSelfNative4D9900(unit)
	if *memmap.PtrT[uint32](0x5D4594, 1556324) != 0 || *memmap.PtrT[uint32](0x5D4594, 1556328) != 0 {
		t.Fatal("outer admission accessed key types")
	}
}
