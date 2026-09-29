package opennox

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func playerStatsTestPacket48EA70() []byte {
	return []byte{
		byte(netmsg.MSG_REPORT_STATS),
		0x34, 0x92,
		0x56, 0x34,
		0x45, 0x23,
		0x34, 0x12,
		0x78, 0x56,
		0x67, 0x45,
		9,
	}
}

func requirePlayerStatsHighAddress48EA70(t *testing.T, player *server.Player) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", player)
	}
}

func TestDecodePlayerStatsState48EA70(t *testing.T) {
	packet := playerStatsTestPacket48EA70()
	state, ok := decodePlayerStatsState48EA70(packet)
	if !ok {
		t.Fatal("valid player-stats packet was rejected")
	}
	want := playerStatsState48EA70{
		RawNetCode: 0x9234,
		NetCode:    0x1234,
		HealthMax:  0x3456,
		ManaMax:    0x2345,
		Capacity:   0x1234,
		Speed:      0x5678,
		Strength:   0x4567,
		Level:      9,
	}
	if state != want {
		t.Fatalf("decoded state = %+v, want %+v", state, want)
	}
	for size := 0; size < playerStatsPacketSize48EA70; size++ {
		if _, ok := decodePlayerStatsState48EA70(packet[:size]); ok {
			t.Fatalf("%d-byte player-stats packet was accepted", size)
		}
	}
}

func TestHandlePlayerStatsNative48EA70AppliesNativePlayerFields(t *testing.T) {
	player := new(server.Player)
	requirePlayerStatsHighAddress48EA70(t, player)
	player.Info().SetField2235(1)
	player.Info().SetField2239(2)
	player.Info().SetField2243(3)
	player.Info().SetField2247(4)
	player.SetStatsCapacity(5)
	player.Level = 6
	refreshes := 0
	hooks := playerStatsHooks48EA70{
		connected:   func() bool { return true },
		localCode:   func() uint16 { return 0x1234 },
		gameHost:    func() bool { return false },
		localPlayer: func() *server.Player { return player },
		refreshName: func() { refreshes++ },
	}
	packet := playerStatsTestPacket48EA70()
	before := append([]byte(nil), packet...)
	if got := handlePlayerStatsNative48EA70(packet, hooks); got != playerStatsPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d, want %d", got, playerStatsPacketSize48EA70)
	}
	if player.Info().Field2247() != 0x3456 || player.Info().Field2243() != 0x2345 ||
		player.StatsCapacity() != 0x1234 || player.Info().Field2235() != 0x5678 ||
		player.Info().Field2239() != 0x4567 || player.Level != 9 {
		t.Fatalf("player stats = hp:%#x mana:%#x cap:%#x speed:%#x strength:%#x level:%d",
			player.Info().Field2247(), player.Info().Field2243(), player.StatsCapacity(),
			player.Info().Field2235(), player.Info().Field2239(), player.Level)
	}
	if refreshes != 1 {
		t.Fatalf("inventory-name refreshes = %d, want 1", refreshes)
	}
	if !reflect.DeepEqual(packet, before) {
		t.Fatalf("packet mutated: got %x, want %x", packet, before)
	}
}

func TestHandlePlayerStatsNative48EA70GatesUpdates(t *testing.T) {
	packet := playerStatsTestPacket48EA70()
	player := new(server.Player)
	player.Info().SetField2235(77)
	refreshes := 0
	hooks := playerStatsHooks48EA70{
		connected:   func() bool { return false },
		localCode:   func() uint16 { t.Fatal("local code read while disconnected"); return 0 },
		gameHost:    func() bool { t.Fatal("host flag read while disconnected"); return false },
		localPlayer: func() *server.Player { t.Fatal("player read while disconnected"); return nil },
		refreshName: func() { refreshes++ },
	}
	if got := handlePlayerStatsNative48EA70(packet, hooks); got != playerStatsPacketSize48EA70 || refreshes != 0 {
		t.Fatalf("disconnected result/refreshes = %d/%d", got, refreshes)
	}

	hooks.connected = func() bool { return true }
	hooks.localCode = func() uint16 { return 0x9999 }
	if got := handlePlayerStatsNative48EA70(packet, hooks); got != playerStatsPacketSize48EA70 || refreshes != 0 {
		t.Fatalf("foreign result/refreshes = %d/%d", got, refreshes)
	}

	hooks.localCode = func() uint16 { return 0x1234 }
	hooks.gameHost = func() bool { return true }
	hooks.localPlayer = func() *server.Player { t.Fatal("host path read local player"); return nil }
	if got := handlePlayerStatsNative48EA70(packet, hooks); got != playerStatsPacketSize48EA70 || refreshes != 1 {
		t.Fatalf("host result/refreshes = %d/%d", got, refreshes)
	}
	if player.Info().Field2235() != 77 {
		t.Fatalf("host path changed shared stats to %d", player.Info().Field2235())
	}

	hooks.gameHost = func() bool { return false }
	hooks.localPlayer = func() *server.Player { return nil }
	if got := handlePlayerStatsNative48EA70(packet, hooks); got != playerStatsPacketSize48EA70 || refreshes != 2 {
		t.Fatalf("nil-player result/refreshes = %d/%d", got, refreshes)
	}

	if got := handlePlayerStatsNative48EA70(packet[:13], playerStatsHooks48EA70{}); got != -1 {
		t.Fatalf("short packet result = %d, want -1", got)
	}
}

func TestPlayerStatsPacketNative4D8990UsesNamedNativeFields(t *testing.T) {
	player := new(server.Player)
	requirePlayerStatsHighAddress48EA70(t, player)
	player.Info().SetField2235(0x5678)
	player.Info().SetField2239(0x4567)
	player.Level = 9
	update := &server.PlayerUpdateData{Player: player, ManaMax: 0x2345}
	unit := &server.Object{
		ObjClass:      object.ClassPlayer,
		HealthData:    &server.HealthData{Max: 0x3456},
		CarryCapacity: 0x1234,
		UpdateData:    unsafe.Pointer(update),
	}
	packet, ok := playerStatsPacketNative4D8990(unit, func(got *server.Object) int {
		if got != unit {
			t.Fatalf("net-code object = %p, want %p", got, unit)
		}
		return 0x4321
	})
	if !ok {
		t.Fatal("valid player object did not produce a stats packet")
	}
	want := [playerStatsPacketSize48EA70]byte{
		byte(netmsg.MSG_REPORT_STATS),
		0x21, 0x43,
		0x56, 0x34,
		0x45, 0x23,
		0x34, 0x12,
		0x78, 0x56,
		0x67, 0x45,
		9,
	}
	if packet != want {
		t.Fatalf("stats packet = % x, want % x", packet, want)
	}
}

func TestPlayerStatsPacketNative4D8990RejectsInvalidObjects(t *testing.T) {
	netCode := func(*server.Object) int {
		t.Fatal("net code requested for invalid player object")
		return 0
	}
	invalid := []*server.Object{
		nil,
		{ObjClass: object.ClassMonster},
		{ObjClass: object.ClassPlayer},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&server.PlayerUpdateData{})},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&server.PlayerUpdateData{Player: &server.Player{}})},
	}
	for i, unit := range invalid {
		if _, ok := playerStatsPacketNative4D8990(unit, netCode); ok {
			t.Fatalf("invalid object %d produced a stats packet", i)
		}
	}
}

func TestPlayerReportStatsNative4D9900SendsOnceAndClearsDirtyFlag(t *testing.T) {
	player := &server.Player{PlayerInd: 7, StatsReportPending: 1}
	requirePlayerStatsHighAddress48EA70(t, player)
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var calls []string
	hooks := playerStatsReportHooks4D9900{
		totalHealth: func(ind byte, got *server.Object) {
			if ind != 7 || got != unit || player.StatsReportPending != 1 {
				t.Fatalf("health report = ind:%d unit:%p pending:%d", ind, got, player.StatsReportPending)
			}
			calls = append(calls, "health")
		},
		totalMana: func(ind byte, got *server.Object) {
			if ind != 7 || got != unit || player.StatsReportPending != 1 {
				t.Fatalf("mana report = ind:%d unit:%p pending:%d", ind, got, player.StatsReportPending)
			}
			calls = append(calls, "mana")
		},
		stats: func(ind byte, got *server.Object) {
			if ind != 7 || got != unit || player.StatsReportPending != 1 {
				t.Fatalf("stats report = ind:%d unit:%p pending:%d", ind, got, player.StatsReportPending)
			}
			calls = append(calls, "stats")
		},
	}
	playerReportStatsNative4D9900(unit, hooks)
	if want := []string{"health", "mana", "stats"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("report order = %v, want %v", calls, want)
	}
	if player.StatsReportPending != 0 {
		t.Fatalf("pending flag = %d, want 0", player.StatsReportPending)
	}
	playerReportStatsNative4D9900(unit, hooks)
	if len(calls) != 3 {
		t.Fatalf("clean stats were reported again: %v", calls)
	}
}

func TestPlayerReportStatsNative4D9900SkipsInvalidObjects(t *testing.T) {
	hooks := playerStatsReportHooks4D9900{
		totalHealth: func(byte, *server.Object) { t.Fatal("unexpected health report") },
		totalMana:   func(byte, *server.Object) { t.Fatal("unexpected mana report") },
		stats:       func(byte, *server.Object) { t.Fatal("unexpected stats report") },
	}
	playerReportStatsNative4D9900(nil, hooks)
	playerReportStatsNative4D9900(&server.Object{ObjClass: object.ClassMonster}, hooks)
	playerReportStatsNative4D9900(&server.Object{ObjClass: object.ClassPlayer}, hooks)
	playerReportStatsNative4D9900(&server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(&server.PlayerUpdateData{}),
	}, hooks)
	playerReportStatsNative4D9900(&server.Object{
		ObjClass: object.ClassPlayer,
		UpdateData: unsafe.Pointer(&server.PlayerUpdateData{
			Player: &server.Player{},
		}),
	}, hooks)
}
