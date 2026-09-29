package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

type questWarpGateFixture4D71F0 struct {
	server       *Server
	units        []Object
	updates      []PlayerUpdateData
	gates        []Object
	destinations []ExitCollideData
}

func newQuestWarpGateFixture4D71F0(count int) *questWarpGateFixture4D71F0 {
	f := &questWarpGateFixture4D71F0{
		server:       new(Server),
		units:        make([]Object, count),
		updates:      make([]PlayerUpdateData, count),
		gates:        make([]Object, count),
		destinations: make([]ExitCollideData, count),
	}
	f.server.Players.list = make([]Player, count)
	for i := 0; i < count; i++ {
		player := &f.server.Players.list[i]
		*player = Player{Active: 1, PlayerInd: uint8(i), Field4792: 1}
		f.destinations[i].DestinationX = float32(100 + i)
		f.destinations[i].DestinationY = float32(200 + i)
		f.gates[i].CollideData = unsafe.Pointer(&f.destinations[i])
		f.updates[i] = PlayerUpdateData{Player: player, QuestWarpGate: &f.gates[i]}
		f.units[i] = Object{
			ObjClass:   object.ClassPlayer,
			NetCode:    uint32(0x12340000 + i),
			UpdateData: unsafe.Pointer(&f.updates[i]),
		}
		player.PlayerUnit = &f.units[i]
	}
	return f
}

func preserveQuestWarpGateFlags4D71F0(t *testing.T) {
	t.Helper()
	game, engine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.ResetGame()
	noxflags.ResetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(game)
		noxflags.ResetEngine()
		noxflags.SetEngine(engine)
	})
}

func assertQuestPointerAbove32Bits4D71F0(t *testing.T, name string, ptr unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(ptr) <= math.MaxUint32 {
		t.Fatalf("%s pointer = %#x, want address above 32-bit range", name, uintptr(ptr))
	}
}

func TestQuestLeaveWarpGate4D7480PreservesNativePointersAndOrder(t *testing.T) {
	f := newQuestWarpGateFixture4D71F0(1)
	unit, update, player := &f.units[0], &f.updates[0], &f.server.Players.list[0]
	gate, destination := &f.gates[0], types.Pointf{X: 100, Y: 200}

	assertQuestPointerAbove32Bits4D71F0(t, "unit", unsafe.Pointer(unit))
	assertQuestPointerAbove32Bits4D71F0(t, "update", unsafe.Pointer(update))
	assertQuestPointerAbove32Bits4D71F0(t, "player", unsafe.Pointer(player))
	assertQuestPointerAbove32Bits4D71F0(t, "gate", unsafe.Pointer(gate))
	assertQuestPointerAbove32Bits4D71F0(t, "collide data", gate.CollideData)

	var events []string
	runtimeHooks := QuestWarpGateRuntime4D71F0{
		LeaveObserver: func(got *Player) {
			if got != player {
				t.Fatalf("observer player = %p, want %p", got, player)
			}
			events = append(events, "observer")
		},
		CameraUnlock: func(got *Object) {
			if got != unit {
				t.Fatalf("camera unit = %p, want %p", got, unit)
			}
			events = append(events, "camera")
		},
		Move: func(got *Object, pos types.Pointf) {
			if got != unit || pos != destination {
				t.Fatalf("move = (%p, %+v), want (%p, %+v)", got, pos, unit, destination)
			}
			if update.QuestWarpGate != nil {
				t.Fatal("gate was not cleared before move")
			}
			events = append(events, "move")
		},
		Audio: func(id sound.ID, got *Object, kind int, code uint32) {
			if id != sound.SoundPlayerEliminated || got != unit || kind != 2 || code != unit.NetCode {
				t.Fatalf("audio = (%d, %p, %d, %#x)", id, got, kind, code)
			}
			events = append(events, "audio")
		},
		PointFX: func(op netmsg.Op, pos types.Pointf) {
			if op != netmsg.MSG_FX_BLUE_SPARKS || pos != destination {
				t.Fatalf("point FX = (%d, %+v), want blue sparks at %+v", op, pos, destination)
			}
			events = append(events, "fx")
		},
	}

	f.server.QuestLeaveWarpGate4D7480(unit, runtimeHooks)
	if want := []string{"observer", "camera", "move", "audio", "fx"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if update.QuestWarpGate != nil {
		t.Fatalf("gate = %p, want nil", update.QuestWarpGate)
	}
	runtime.KeepAlive(f)
}

func TestQuestExitTimeout4D71F0UsesUnsignedFrameWrapAndBroadcasts(t *testing.T) {
	preserveQuestWarpGateFlags4D71F0(t)
	f := newQuestWarpGateFixture4D71F0(2)
	f.updates[0].QuestExit = new(Object)
	f.server.SetFrame(4)
	start := uint32(math.MaxUint32 - 8995)
	stored := start

	var recipient, remove, sequence int
	var packet []byte
	f.server.NetSendPacketXxx = func(gotRecipient int, got []byte, related *Object, gotRemove, gotSequence int) int {
		recipient, remove, sequence = gotRecipient, gotRemove, gotSequence
		packet = append(packet[:0], got...)
		if related != nil {
			t.Fatalf("related object = %p, want nil", related)
		}
		return 0x5a
	}

	got := f.server.QuestExitTimeout4D71F0(QuestWarpGateRuntime4D71F0{
		ExitCountdownStart: func() uint32 { return stored },
		StoreExitCountdownStart: func(value uint32) {
			stored = value
		},
	})
	if got != 0x5a {
		t.Fatalf("result = %#x, want 0x5a", got)
	}
	if stored != 0 {
		t.Fatalf("countdown start = %#x, want 0", stored)
	}
	if !f.server.Doors.Sub_4D72C0() {
		t.Fatal("Quest doors were not opened")
	}
	if recipient != 255 || remove != 1 || sequence != 1 {
		t.Fatalf("send args = (%d, %d, %d), want (255, 1, 1)", recipient, remove, sequence)
	}
	if want := []byte{byte(netmsg.MSG_GAUNTLET), 24, 1}; !reflect.DeepEqual(packet, want) {
		t.Fatalf("packet = %#v, want %#v", packet, want)
	}
	runtime.KeepAlive(f)
}

func TestQuestSetWarpEnabled4D7520CapturesSuccessorBeforeSetOff(t *testing.T) {
	f := newQuestWarpGateFixture4D71F0(1)
	first := &Object{ObjClass: object.ClassExit, ObjSubClass: object.SubClass(2)}
	second := &Object{ObjClass: object.ClassExit, ObjSubClass: object.SubClass(2)}
	first.ObjNext = second
	f.server.Objs.SetObjects(first)
	stored := uint32(1)
	var setOff []*Object

	got := f.server.QuestSetWarpEnabled4D7520(0, QuestWarpGateRuntime4D71F0{
		WarpEnabled:      func() uint32 { return stored },
		StoreWarpEnabled: func(value uint32) { stored = value },
		LeaveObserver:    func(*Player) {},
		CameraUnlock:     func(*Object) {},
		Move:             func(*Object, types.Pointf) {},
		Audio:            func(sound.ID, *Object, int, uint32) {},
		PointFX:          func(netmsg.Op, types.Pointf) {},
		ObjectSetOff: func(obj *Object) uint32 {
			setOff = append(setOff, obj)
			if obj == first {
				obj.ObjNext = nil
				return 0x42
			}
			return 0x99
		},
	})
	if got != 0x99 {
		t.Fatalf("result = %#x, want 0x99", got)
	}
	if stored != 0 {
		t.Fatalf("warp enabled = %#x, want 0", stored)
	}
	if want := []*Object{first, second}; !reflect.DeepEqual(setOff, want) {
		t.Fatalf("set-off objects = %p, want %p", setOff, want)
	}
	if f.updates[0].QuestWarpGate != nil {
		t.Fatalf("player gate = %p, want nil", f.updates[0].QuestWarpGate)
	}
	runtime.KeepAlive(f)
}

func TestQuestCheckWarpGate4D7600FailedWarpReturnsWholeParty(t *testing.T) {
	preserveQuestWarpGateFlags4D71F0(t)
	f := newQuestWarpGateFixture4D71F0(2)
	f.server.SetFrame(4)
	warpFrame := uint32(math.MaxUint32 - 25)
	var events []string
	indexOf := func(obj *Object) int {
		for i := range f.units {
			if obj == &f.units[i] {
				return i
			}
		}
		return -1
	}

	f.server.QuestCheckWarpGate4D7600(QuestWarpGateRuntime4D71F0{
		WarpFrame: func() uint32 { return warpFrame },
		MaybeWarp: func() int32 {
			events = append(events, "warp")
			return 0
		},
		LeaveObserver: func(player *Player) {
			events = append(events, "observer"+string(rune('0'+player.Index())))
		},
		CameraUnlock: func(obj *Object) {
			events = append(events, "camera"+string(rune('0'+indexOf(obj))))
		},
		Move: func(obj *Object, _ types.Pointf) {
			events = append(events, "move"+string(rune('0'+indexOf(obj))))
		},
		Audio: func(_ sound.ID, obj *Object, _ int, _ uint32) {
			events = append(events, "audio"+string(rune('0'+indexOf(obj))))
		},
		PointFX: func(netmsg.Op, types.Pointf) {
			events = append(events, "fx")
		},
		PriMessage: func(obj *Object, id strman.ID, value byte) {
			if id != "Gauntlet.c:WarpRestrictedMulti" || value != 0 {
				t.Fatalf("message = (%q, %d), want multiplayer restriction", id, value)
			}
			if f.updates[indexOf(obj)].QuestWarpGate != nil {
				t.Fatal("message sent before player left gate")
			}
			events = append(events, "message"+string(rune('0'+indexOf(obj))))
		},
	})

	want := []string{
		"warp",
		"observer0", "camera0", "move0", "audio0", "fx", "message0",
		"observer1", "camera1", "move1", "audio1", "fx", "message1",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	for i := range f.updates {
		if f.updates[i].QuestWarpGate != nil {
			t.Fatalf("player %d gate = %p, want nil", i, f.updates[i].QuestWarpGate)
		}
	}
	runtime.KeepAlive(f)
}
