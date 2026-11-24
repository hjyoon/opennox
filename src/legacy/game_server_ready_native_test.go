package legacy

import (
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type gameServerReadyFixture struct {
	Server
	srv      *server.Server
	walls    func(index int)
	gauntlet func() bool
}

func (s *gameServerReadyFixture) S() *server.Server { return s.srv }

func (s *gameServerReadyFixture) Nox_xxx_wall_4DF1E0(index int) {
	s.walls(index)
}

func (s *gameServerReadyFixture) GetFlag3592() bool {
	return s.gauntlet()
}

func TestGameServerReadyNativeCEntry4DD180(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantAllocClassC(nil)
	gameFrameHook = func() uint32 { return 123 }

	oldServer, oldFPS := GetServer, gameFPSHook
	oldOff, oldApply := spellBuffOffExportImpl4FF5B0, buffApplyExportImpl4FF380
	oldLeave, oldStats := Nox_xxx_playerLeaveObserver_0_4E6AA0, Nox_xxx_netStatsMultiplier_4D9C20
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() {
		GetServer, gameFPSHook = oldServer, oldFPS
		spellBuffOffExportImpl4FF5B0, buffApplyExportImpl4FF380 = oldOff, oldApply
		Nox_xxx_playerLeaveObserver_0_4E6AA0, Nox_xxx_netStatsMultiplier_4D9C20 = oldLeave, oldStats
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	gameFPSHook = func() uint32 { return 0x18001 }
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameOnline | noxflags.GameModeCoopTeam | noxflags.GameModeQuest)

	objects, free := alloc.Make([]server.Object{}, 4)
	t.Cleanup(free)
	initial, live, flag, other := &objects[0], &objects[1], &objects[2], &objects[3]
	initial.PosVec, live.PosVec = types.Ptf(17, 19), types.Ptf(43, 47)
	flag.ObjClass, flag.ObjNext = object.ClassFlag, other
	other.ObjClass = object.ClassWeapon
	s.Objs.SetObjects(flag)
	const index = 7
	player := s.Players.ByIndRaw(index)
	player.Active, player.PlayerInd, player.PlayerUnit, player.Field3680 = 1, index, initial, 1
	t.Cleanup(func() { s.Players.Nox_xxx_netUnmarkMinimapObj_417300(index, flag, ^uint32(0)) })

	var trace []string
	fixture := &gameServerReadyFixture{srv: s}
	GetServer = func() Server { return fixture }
	fixture.walls = func(got int) {
		trace = append(trace, "walls")
		if got != index || player.Field3680 != 0x11 || player.Pos3632Vec != live.PosVec ||
			!player.MinimapTracks(flag) || player.MinimapTracks(other) {
			t.Fatal("wall notification preceded the timestamp, observer position, or flag update")
		}
	}
	fixture.gauntlet = func() bool {
		trace = append(trace, "quest rule")
		return true
	}
	spellBuffOffExportImpl4FF5B0 = func(got *server.Object, id int32) int32 {
		trace = append(trace, "remove protection")
		if got != initial || id != int32(server.ENCHANT_INVULNERABLE) {
			t.Fatalf("removed buff %d from %p", id, got)
		}
		player.PlayerUnit = live
		return 11
	}
	buffApplyExportImpl4FF380 = func(got *server.Object, id int32, duration int16, power int8) {
		trace = append(trace, "apply protection")
		if got != live || id != int32(server.ENCHANT_INVULNERABLE) || duration != -32763 || power != 5 {
			t.Fatalf("protected %p with %d for %d at power %d", got, id, duration, power)
		}
	}
	Nox_xxx_playerLeaveObserver_0_4E6AA0 = func(got *server.Player) {
		trace = append(trace, "leave observer")
		if got != player || got.Pos3632Vec != live.PosVec {
			t.Fatal("observer mode received the wrong player or position")
		}
	}
	Nox_xxx_netStatsMultiplier_4D9C20 = func(got *server.Object) int {
		trace = append(trace, "stats")
		if got != live {
			t.Fatalf("stats used unit %p, want %p", got, live)
		}
		return -32771
	}
	if got := callGameServerReadyC(index + 1); got != 0 || len(trace) != 0 {
		t.Fatalf("inactive slot returned %d and ran %v", got, trace)
	}
	if got := callGameServerReadyC(index); got != -32771 {
		t.Fatalf("ready result = %d, want -32771", got)
	}
	if want := []string{"remove protection", "apply protection", "leave observer", "walls", "quest rule", "stats"}; !slices.Equal(trace, want) {
		t.Fatalf("ready operations = %v, want %v", trace, want)
	}
	first, last := importantNativeListC()
	if first == nil || first != last {
		t.Fatal("expected one quest notification")
	}
	packet := (*importantPacketC)(first)
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&packet.payload[0])), int(packet.payload_size))
	if len(payload) != 2 || payload[0] != 240 || payload[1] != 20 || packet.recipient != index {
		t.Fatalf("quest notification = % x, recipient %d", payload, packet.recipient)
	}
}
