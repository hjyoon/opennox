package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"reflect"
	"testing"

	playerlib "github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type playerKillStatsEvent425CA0 struct {
	Kind string
	Arg  uint32
}

type playerKillStatsPlayer425CA0 struct {
	Index int32
	Slot  byte
	Team  uint32
	Class byte
	Name  []byte
}

type playerKillStatsState425CA0 struct {
	Count   uint32
	Pairs   uint32
	Records []byte
	Events  []byte
	Players []playerKillStatsPlayer425CA0
}

type playerKillStatsFixture425CA0 struct {
	flags    uint32
	count    uint32
	pairs    uint32
	records  [8192]byte
	events   [32]byte
	players  [2]*Player
	calls    []playerKillStatsEvent425CA0
	flushes  int
	beforeIP func(int)
}

func newPlayerKillStatsFixture425CA0() *playerKillStatsFixture425CA0 {
	f := &playerKillStatsFixture425CA0{flags: 0x2000, count: 20, pairs: 3}
	for i := range f.records {
		f.records[i] = 0xa5
	}
	for i := range f.events {
		f.events[i] = 0xcc
	}
	f.players = [2]*Player{
		{Field4648: 17, PlayerInd: 2, Field2068: 0x12345678},
		{Field4648: 19, PlayerInd: 4, Field2068: 0xaabbccdd},
	}
	f.players[0].SetField2096("First")
	f.players[1].SetField2096("Second")
	f.players[0].Info().SetPlayerClass(playerlib.Warrior)
	f.players[1].Info().SetPlayerClass(playerlib.Wizard)
	return f
}

func (f *playerKillStatsFixture425CA0) runtime() PlayerKillStatsRuntime425CA0 {
	return PlayerKillStatsRuntime425CA0{
		GameFlag: func(flag uint32) bool {
			f.calls = append(f.calls, playerKillStatsEvent425CA0{"flag", flag})
			return f.flags&flag != 0
		},
		PlayerCount: func() uint32 { return f.count },
		SetCount:    func(value uint32) { f.count = value },
		CopyName: func(index uint32, player *Player) {
			name := alloc.GoString(&player.Field2096Buf[0])
			copy(f.records[index*32:], append([]byte(name), 0))
		},
		ConnectionIP: func(connection int) uint32 {
			f.calls = append(f.calls, playerKillStatsEvent425CA0{"ip", uint32(connection)})
			if f.beforeIP != nil {
				f.beforeIP(connection)
			}
			return 0x11223300 + uint32(connection)
		},
		StoreIP: func(index, value uint32) {
			binary.LittleEndian.PutUint32(f.records[index*32+12:], value)
		},
		StoreTeam: func(index, value uint32) {
			binary.LittleEndian.PutUint32(f.records[index*32+16:], value)
		},
		StoreClass: func(index uint32, value byte) { f.records[index*32+20] = value },
		PairCount:  func() uint32 { return f.pairs },
		StorePair: func(index uint32, first, second byte) {
			f.events[index*2], f.events[index*2+1] = first, second
		},
		SetPairCount: func(value uint32) { f.pairs = value },
		Flush: func() {
			f.calls = append(f.calls, playerKillStatsEvent425CA0{"flush", 0})
			f.flushes++
		},
	}
}

func (f *playerKillStatsFixture425CA0) snapshot() playerKillStatsState425CA0 {
	state := playerKillStatsState425CA0{
		Count: f.count, Pairs: f.pairs,
		Records: append([]byte(nil), f.records[:]...),
		Events:  append([]byte(nil), f.events[:]...),
	}
	for _, player := range f.players {
		state.Players = append(state.Players, playerKillStatsPlayer425CA0{
			Index: player.Field4648, Slot: player.PlayerInd, Team: player.Field2068,
			Class: byte(player.Info().PlayerClass()),
			Name:  append([]byte(nil), player.Field2096Buf[:]...),
		})
	}
	return state
}

func TestPlayerKillStats425CA0OriginalMatrix(t *testing.T) {
	cases := []struct {
		name       string
		prepare    func(*playerKillStatsFixture425CA0)
		first      int
		second     int
		wantCount  uint32
		wantPair   [2]byte
		wantIPs    []uint32
		wantFlush  int
		wantReturn bool
	}{
		{name: "offline", prepare: func(f *playerKillStatsFixture425CA0) { f.flags = 0 }, wantReturn: true},
		{name: "quest", prepare: func(f *playerKillStatsFixture425CA0) { f.flags = 0x3000 }, wantReturn: true},
		{name: "nil-first", first: -1, second: 1, wantReturn: true},
		{name: "registered", second: 1, wantCount: 20, wantPair: [2]byte{17, 19}},
		{name: "nil-second", second: -1, wantCount: 20, wantPair: [2]byte{17, 255}},
		{name: "new-first", second: 1, prepare: func(f *playerKillStatsFixture425CA0) { f.players[0].Field4648 = -1 }, wantCount: 21, wantPair: [2]byte{20, 19}, wantIPs: []uint32{3}},
		{name: "new-first-host-second", second: 1, prepare: func(f *playerKillStatsFixture425CA0) { f.players[0].Field4648 = -1; f.players[1].PlayerInd = 31 }, wantCount: 21, wantPair: [2]byte{20, 19}, wantIPs: []uint32{0}},
		{name: "new-both", second: 1, prepare: func(f *playerKillStatsFixture425CA0) { f.players[0].Field4648 = -1; f.players[1].Field4648 = -1 }, wantCount: 22, wantPair: [2]byte{20, 21}, wantIPs: []uint32{3, 5}},
		{name: "new-both-host-second", second: 1, prepare: func(f *playerKillStatsFixture425CA0) {
			f.players[0].Field4648 = -1
			f.players[1].Field4648 = -1
			f.players[1].PlayerInd = 31
		}, wantCount: 22, wantPair: [2]byte{20, 21}, wantIPs: []uint32{0, 0}},
		{name: "new-second", second: 1, prepare: func(f *playerKillStatsFixture425CA0) { f.players[1].Field4648 = -1 }, wantCount: 21, wantPair: [2]byte{17, 20}, wantIPs: []uint32{5}},
		{name: "new-host-second", second: 1, prepare: func(f *playerKillStatsFixture425CA0) { f.players[1].Field4648 = -1; f.players[1].PlayerInd = 31 }, wantCount: 21, wantPair: [2]byte{17, 20}, wantIPs: []uint32{0}},
		{name: "new-aliased", second: 0, prepare: func(f *playerKillStatsFixture425CA0) { f.players[0].Field4648 = -1 }, wantCount: 21, wantPair: [2]byte{20, 20}, wantIPs: []uint32{3}},
		{name: "byte-indices", second: 1, prepare: func(f *playerKillStatsFixture425CA0) { f.players[0].Field4648 = 0x123; f.players[1].Field4648 = 0x456 }, wantCount: 20, wantPair: [2]byte{0x23, 0x56}},
		{name: "below-flush", second: 1, prepare: func(f *playerKillStatsFixture425CA0) { f.count = 254 }, wantCount: 254, wantPair: [2]byte{17, 19}},
		{name: "flush-existing", second: 1, prepare: func(f *playerKillStatsFixture425CA0) { f.count = 255 }, wantCount: 255, wantPair: [2]byte{17, 19}, wantFlush: 1},
		{name: "flush-new-both", second: 1, prepare: func(f *playerKillStatsFixture425CA0) {
			f.count = 254
			f.players[0].Field4648 = -1
			f.players[1].Field4648 = -1
		}, wantCount: 256, wantPair: [2]byte{254, 255}, wantIPs: []uint32{3, 5}, wantFlush: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newPlayerKillStatsFixture425CA0()
			if tc.prepare != nil {
				tc.prepare(f)
			}
			before := f.snapshot()
			var first, second *Player
			if tc.first >= 0 {
				first = f.players[tc.first]
			}
			if tc.second >= 0 {
				second = f.players[tc.second]
			}
			PlayerKillStats425CA0(first, second, f.runtime())
			after := f.snapshot()
			if tc.wantReturn {
				if !reflect.DeepEqual(before, after) || f.flushes != 0 {
					t.Fatal("guard changed statistics or player state")
				}
			} else if f.count != tc.wantCount || f.pairs != 4 || f.flushes != tc.wantFlush || [2]byte{f.events[6], f.events[7]} != tc.wantPair {
				t.Fatalf("count/pairs/flush/pair = %d/%d/%d/%v", f.count, f.pairs, f.flushes, f.events[6:8])
			}
			var ips []uint32
			for _, call := range f.calls {
				if call.Kind == "ip" {
					ips = append(ips, call.Arg)
				}
			}
			if !reflect.DeepEqual(ips, tc.wantIPs) {
				t.Fatalf("IP connections = %v, want %v", ips, tc.wantIPs)
			}
			encoded, err := json.Marshal(struct {
				Case          string
				Flags         uint32
				First, Second int
				Before, After playerKillStatsState425CA0
				Calls         []playerKillStatsEvent425CA0
			}{tc.name, f.flags, tc.first, tc.second, before, after, f.calls})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("PLAYER KILL STATS SNAPSHOT: %s", encoded)
		})
	}
}

func TestPlayerKillStats425CA0LiveFieldsAndCachedIndex(t *testing.T) {
	f := newPlayerKillStatsFixture425CA0()
	f.players[0].Field4648 = -1
	before := f.snapshot()
	f.beforeIP = func(int) {
		f.count = 30
		f.players[0].Field4648 = 123
		f.players[0].Field2068 = 0x87654321
		f.players[0].Info().SetPlayerClass(playerlib.Conjurer)
		f.players[1].Field4648 = -1
		f.players[1].PlayerInd = 31
	}
	PlayerKillStats425CA0(f.players[0], f.players[1], f.runtime())
	// The second IP callback changes the first player's live index again;
	// the final pair still uses the first registration's cached index 20.
	if f.players[0].Field4648 != 123 || f.players[1].Field4648 != 30 || f.count != 30 {
		t.Fatalf("cached/live registration indices changed: %d/%d count=%d", f.players[0].Field4648, f.players[1].Field4648, f.count)
	}
	if got := binary.LittleEndian.Uint32(f.records[20*32+16:]); got != 0x87654321 || f.records[20*32+20] != byte(playerlib.Conjurer) {
		t.Fatal("first team/class were not reloaded after IP callback")
	}
	if got := f.records[20*32 : 20*32+7]; !bytes.Equal(got, []byte("Second\x00")) {
		t.Fatalf("second name destination = %q", got)
	}
	if [2]byte{f.events[6], f.events[7]} != [2]byte{20, 30} {
		t.Fatal("pair indices were not cached across callbacks")
	}
	playerKillStatsLogAux425CA0(t, f, "live-callback-fields", 0, 1, before, true, false)
}

func TestPlayerKillStats425CA0NilSecondFaultOrder(t *testing.T) {
	f := newPlayerKillStatsFixture425CA0()
	f.players[0].Field4648 = -1
	before := f.snapshot()
	defer func() {
		if recover() == nil {
			t.Fatal("new first player with nil second did not fault at the original host-byte read")
		}
		if f.count != 21 || f.pairs != 3 || f.players[0].Field4648 != -1 || string(f.records[20*32:20*32+6]) != "First\x00" {
			t.Fatal("count/name writes or untouched registration/pair differ at original fault")
		}
		if len(f.calls) != 2 || f.flushes != 0 {
			t.Fatal("IP or flush ran before the host-byte fault")
		}
		playerKillStatsLogAux425CA0(t, f, "nil-second-fault", 0, -1, before, false, true)
	}()
	PlayerKillStats425CA0(f.players[0], nil, f.runtime())
}

func playerKillStatsLogAux425CA0(t *testing.T, f *playerKillStatsFixture425CA0, name string, first, second int, before playerKillStatsState425CA0, mutation, fault bool) {
	t.Helper()
	encoded, err := json.Marshal(struct {
		Case            string
		Flags           uint32
		First, Second   int
		Before, After   playerKillStatsState425CA0
		Calls           []playerKillStatsEvent425CA0
		Mutation, Fault bool
	}{name, f.flags, first, second, before, f.snapshot(), f.calls, mutation, fault})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("PLAYER KILL STATS AUX SNAPSHOT: %s", encoded)
}
