package server

import (
	"encoding/binary"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/ntype"
)

type questStatsPlayerFixture4D6540 struct {
	generators uint32
	secrets    uint32
	monsters   uint32
	bonus      uint32
	stage      uint32
	state      uint32
	netCode    uint32
}

func newQuestStatsServer4D6540(fixtures []questStatsPlayerFixture4D6540) (*Server, []PlayerUpdateData, []Object) {
	s := new(Server)
	s.Players.list = make([]Player, len(fixtures))
	updates := make([]PlayerUpdateData, len(fixtures))
	units := make([]Object, len(fixtures))
	for i, fixture := range fixtures {
		player := &s.Players.list[i]
		*player = Player{
			Active:    1,
			PlayerInd: byte(i),
			field4664: fixture.monsters,
			field4668: fixture.generators,
			field4672: fixture.secrets,
			field4680: fixture.bonus,
			field4688: fixture.stage,
			Field4792: fixture.state,
		}
		updates[i].Player = player
		units[i] = Object{
			ObjClass:   object.ClassPlayer,
			NetCode:    fixture.netCode,
			UpdateData: unsafe.Pointer(&updates[i]),
		}
		player.PlayerUnit = &units[i]
	}
	return s, updates, units
}

func TestQuestScore4D66E0PreservesBinary32Rounding(t *testing.T) {
	tests := []struct {
		name                                 string
		generators, secrets, monsters, stage uint32
		exponent                             float64
		want                                 uint32
	}{
		{name: "weighted score", generators: 2, secrets: 3, monsters: 4, stage: 5, exponent: 1, want: 627},
		{name: "half rounds to even zero", monsters: 5, stage: 1, exponent: 1, want: 0},
		{name: "one and half rounds to even two", monsters: 15, stage: 1, exponent: 1, want: 2},
		{name: "x87 indefinite on overflow", generators: math.MaxUint32, stage: math.MaxUint32, exponent: 2, want: 0x80000000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := QuestScore4D66E0(tc.generators, tc.secrets, tc.monsters, tc.stage, tc.exponent)
			if got != tc.want {
				t.Fatalf("score = %#x, want %#x", got, tc.want)
			}
		})
	}
}

func TestQuestPlayerScore4D6540SingleAndMultiplayer(t *testing.T) {
	single, singleUpdates, singleUnits := newQuestStatsServer4D6540([]questStatsPlayerFixture4D6540{
		{generators: 2, secrets: 1, monsters: 3, stage: 2, state: 1, netCode: 0x12345},
	})
	if got := single.QuestPlayerScore4D6540(0, 1); got != 111 {
		t.Fatalf("single-player score = %d, want 111", got)
	}

	multi, multiUpdates, multiUnits := newQuestStatsServer4D6540([]questStatsPlayerFixture4D6540{
		{generators: 2, secrets: 1, monsters: 3, stage: 2, state: 1, netCode: 0x12345},
		{generators: 1, secrets: 2, monsters: 4, stage: 3, state: 1, netCode: 0xabcd},
	})
	if got := multi.QuestPlayerScore4D6540(0, 1); got != 187 {
		t.Fatalf("multiplayer score 0 = %d, want 187", got)
	}
	if got := multi.QuestPlayerScore4D6540(1, 1); got != 407 {
		t.Fatalf("multiplayer score 1 = %d, want 407", got)
	}
	if got := multi.QuestPlayerScore4D6540(ntype.PlayerInd(7), 1); got != 0 {
		t.Fatalf("missing-player score = %d, want 0", got)
	}

	runtime.KeepAlive(singleUpdates)
	runtime.KeepAlive(singleUnits)
	runtime.KeepAlive(multiUpdates)
	runtime.KeepAlive(multiUnits)
}

func TestSendQuestStats4D6770PreservesPointersAndPacket(t *testing.T) {
	s, updates, units := newQuestStatsServer4D6540([]questStatsPlayerFixture4D6540{
		{generators: 2, secrets: 1, monsters: 3, bonus: 0x4567, stage: 2, state: 1, netCode: 0x12345},
		{generators: 1, secrets: 2, monsters: 4, bonus: 0x89ab, stage: 3, state: 1, netCode: 0xabcd},
	})
	var (
		gotRecipient int
		gotPacket    []byte
		gotRelated   *Object
		gotRemove    int
		gotSequence  int
	)
	s.NetSendPacketXxx = func(recipient int, packet []byte, related *Object, remove, sequence int) int {
		gotRecipient = recipient
		gotPacket = append([]byte(nil), packet...)
		gotRelated = related
		gotRemove = remove
		gotSequence = sequence
		return -77
	}
	if got := s.SendQuestStats4D6770(1, 0x7654, 1); got != -77 {
		t.Fatalf("send result = %d, want -77", got)
	}

	want := make([]byte, questStatsPacketSize4D6770)
	want[0], want[1] = 0xf0, 12
	binary.LittleEndian.PutUint16(want[2:4], 0x7654)
	binary.LittleEndian.PutUint16(want[4:6], 3)
	entries := []struct {
		netCode, generators, secrets, bonus, monsters uint16
		score                                         uint32
	}{
		{0x2345, 2, 1, 0x4567, 3, 187},
		{0xabcd, 1, 2, 0x89ab, 4, 407},
	}
	for i, entry := range entries {
		off := 6 + i*14
		binary.LittleEndian.PutUint16(want[off:off+2], entry.netCode)
		binary.LittleEndian.PutUint16(want[off+2:off+4], entry.generators)
		binary.LittleEndian.PutUint16(want[off+4:off+6], entry.secrets)
		binary.LittleEndian.PutUint16(want[off+6:off+8], entry.bonus)
		binary.LittleEndian.PutUint16(want[off+8:off+10], entry.monsters)
		binary.LittleEndian.PutUint32(want[off+10:off+14], entry.score)
	}
	if gotRecipient != 1 || !reflect.DeepEqual(gotPacket, want) || gotRelated != nil || gotRemove != 1 || gotSequence != 1 {
		t.Fatalf("send = recipient %d packet %v related %p remove %d sequence %d", gotRecipient, gotPacket, gotRelated, gotRemove, gotSequence)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"player 0": uintptr(unsafe.Pointer(&s.Players.list[0])),
			"update 0": uintptr(unsafe.Pointer(&updates[0])),
			"unit 0":   uintptr(unsafe.Pointer(&units[0])),
			"player 1": uintptr(unsafe.Pointer(&s.Players.list[1])),
			"update 1": uintptr(unsafe.Pointer(&updates[1])),
			"unit 1":   uintptr(unsafe.Pointer(&units[1])),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(updates)
	runtime.KeepAlive(units)
}

func TestSendQuestStats4D6770CapsPacketAtSixPlayers(t *testing.T) {
	fixtures := make([]questStatsPlayerFixture4D6540, 7)
	for i := range fixtures {
		fixtures[i] = questStatsPlayerFixture4D6540{
			generators: uint32(i + 1),
			stage:      1,
			state:      1,
			netCode:    uint32(0x5000 + i),
		}
	}
	s, updates, units := newQuestStatsServer4D6540(fixtures)
	var packet []byte
	s.NetSendPacketXxx = func(_ int, data []byte, _ *Object, _, _ int) int {
		packet = append([]byte(nil), data...)
		return 1
	}
	if got := s.SendQuestStats4D6770(0, 0, 1); got != 1 {
		t.Fatalf("send result = %d, want 1", got)
	}
	if len(packet) != questStatsPacketSize4D6770 {
		t.Fatalf("packet length = %d, want %d", len(packet), questStatsPacketSize4D6770)
	}
	for i := 0; i < questStatsMaxPlayers4D6770; i++ {
		off := 6 + i*14
		if got, want := binary.LittleEndian.Uint16(packet[off:off+2]), uint16(0x5000+i); got != want {
			t.Fatalf("entry %d net code = %#x, want %#x", i, got, want)
		}
	}
	runtime.KeepAlive(updates)
	runtime.KeepAlive(units)
}
