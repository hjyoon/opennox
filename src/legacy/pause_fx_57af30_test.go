package legacy

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func pauseFXTestRuntime57AF30() pauseFXRuntime57AF30 {
	return pauseFXRuntime57AF30{
		isPaused:        func() bool { return false },
		newObject:       func(string) *server.Object { return nil },
		createAt:        func(*server.Object, *server.Object, types.Pointf) {},
		freeObject:      func(*server.Object) {},
		pointFX:         func(netmsg.Op, types.Pointf) {},
		audio:           func(sound.ID, *server.Object, int, uint32) {},
		loadString:      func(strman.ID, string, int) string { return "" },
		sendLine:        func(*server.Object, string) {},
		setPlayerState:  func(*server.Object, server.PlayerState) bool { return false },
		setAnimation:    func(*server.Object, uint8) {},
		setPause:        func(bool) {},
		ticks:           func() uint64 { return 0 },
		delayedDelete:   func(*server.Object) {},
		isSpellBookOpen: func() bool { return false },
	}
}

func TestPauseFXStart57AF30LevelUpPreservesNativeObjectsAndOrder(t *testing.T) {
	unit := &server.Object{PosVec: types.Pointf{X: 12.5, Y: 34.5}, NetCode: 0xfedcba98}
	effect := new(server.Object)
	state := pauseFXState57AF30{}
	runtime := pauseFXTestRuntime57AF30()
	var events []string
	runtime.isPaused = func() bool {
		events = append(events, "paused")
		return false
	}
	runtime.newObject = func(id string) *server.Object {
		events = append(events, "new")
		if id != pauseFXLevelUpObject57AF30 {
			t.Fatalf("object ID = %q", id)
		}
		return effect
	}
	runtime.createAt = func(gotEffect, owner *server.Object, pos types.Pointf) {
		events = append(events, "create")
		if gotEffect != effect || owner != nil || pos != unit.Pos() {
			t.Fatalf("create = %p/%p/%v, want %p/nil/%v", gotEffect, owner, pos, effect, unit.Pos())
		}
	}
	runtime.pointFX = func(op netmsg.Op, pos types.Pointf) {
		events = append(events, "point")
		if op != pauseFXPointOpcode57AF30 || pos != unit.Pos() {
			t.Fatalf("point FX = %d/%v", op, pos)
		}
	}
	runtime.audio = func(id sound.ID, got *server.Object, kind int, code uint32) {
		events = append(events, "audio")
		if id != pauseFXLevelUpSound57AF30 || got != unit || kind != 2 || code != unit.NetCode {
			t.Fatalf("audio = %d/%p/%d/%08x", id, got, kind, code)
		}
	}
	runtime.loadString = func(id strman.ID, path string, line int) string {
		events = append(events, "string")
		if id != pauseFXLevelUpMessage57AF30 || path != pauseFXLevelUpMessagePath57AF30 || line != pauseFXLevelUpMessageLine57AF30 {
			t.Fatalf("string = %q/%q/%d", id, path, line)
		}
		return "level-up"
	}
	runtime.sendLine = func(got *server.Object, message string) {
		events = append(events, "line")
		if got != unit || message != "level-up" {
			t.Fatalf("line = %p/%q", got, message)
		}
	}
	runtime.setPlayerState = func(got *server.Object, playerState server.PlayerState) bool {
		events = append(events, "state")
		if got != unit || playerState != server.PlayerState30 {
			t.Fatalf("player state = %p/%d", got, playerState)
		}
		return true
	}
	runtime.setAnimation = func(got *server.Object, frame uint8) {
		events = append(events, "animation")
		if got != unit || frame != 4 {
			t.Fatalf("animation = %p/%d", got, frame)
		}
	}
	runtime.setPause = func(value bool) {
		events = append(events, "pause")
		if !value || !state.active {
			t.Fatalf("pause = %t while active = %t", value, state.active)
		}
	}
	runtime.ticks = func() uint64 {
		events = append(events, "ticks")
		return 0x123456789abcdef0
	}

	pauseFXStart57AF30(&state, unit, 0, runtime)

	wantEvents := []string{"paused", "new", "create", "point", "audio", "string", "line", "state", "animation", "pause", "ticks"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	if !state.active || !state.timed || state.mode != 0 || state.started != 0x123456789abcdef0 || state.unit != unit || state.effect != effect {
		t.Fatalf("state = %+v, want active timed level-up state", state)
	}
}

func TestPauseFXStart57AF30ModesAndMissingUnit(t *testing.T) {
	t.Run("oblivion", func(t *testing.T) {
		unit := &server.Object{PosVec: types.Pointf{X: 1, Y: 2}}
		effect := new(server.Object)
		state := pauseFXState57AF30{}
		runtime := pauseFXTestRuntime57AF30()
		pointCalls := 0
		runtime.newObject = func(id string) *server.Object {
			if id != pauseFXOblivionObject57AF30 {
				t.Fatalf("object ID = %q", id)
			}
			return effect
		}
		runtime.pointFX = func(op netmsg.Op, pos types.Pointf) {
			pointCalls++
			if op != pauseFXPointOpcode57AF30 || pos != unit.Pos() {
				t.Fatalf("point FX = %d/%v", op, pos)
			}
		}
		runtime.audio = func(sound.ID, *server.Object, int, uint32) { t.Fatal("audio called") }
		runtime.loadString = func(strman.ID, string, int) string { t.Fatal("string loaded"); return "" }
		runtime.sendLine = func(*server.Object, string) { t.Fatal("line sent") }
		pauseFXStart57AF30(&state, unit, 1, runtime)
		if pointCalls != 1 || !state.active || !state.timed || state.mode != 1 || state.effect != effect {
			t.Fatalf("oblivion state = %+v, point calls = %d", state, pointCalls)
		}
	})

	t.Run("reuse", func(t *testing.T) {
		unit := &server.Object{PosVec: types.Pointf{X: 7, Y: 8}}
		effect := new(server.Object)
		state := pauseFXState57AF30{unit: unit, effect: effect}
		runtime := pauseFXTestRuntime57AF30()
		createCalls := 0
		runtime.newObject = func(string) *server.Object { t.Fatal("object created"); return nil }
		runtime.createAt = func(gotEffect, owner *server.Object, pos types.Pointf) {
			createCalls++
			if gotEffect != effect || owner != nil || pos != unit.Pos() {
				t.Fatalf("create = %p/%p/%v", gotEffect, owner, pos)
			}
		}
		runtime.pointFX = func(netmsg.Op, types.Pointf) { t.Fatal("point FX sent") }
		pauseFXStart57AF30(&state, nil, 2, runtime)
		if createCalls != 1 || !state.active || state.timed || state.mode != 2 || state.unit != unit || state.effect != effect {
			t.Fatalf("reuse state = %+v, create calls = %d", state, createCalls)
		}
	})

	t.Run("missing unit frees effect", func(t *testing.T) {
		effect := new(server.Object)
		state := pauseFXState57AF30{}
		runtime := pauseFXTestRuntime57AF30()
		runtime.newObject = func(string) *server.Object { return effect }
		freeCalls := 0
		runtime.freeObject = func(got *server.Object) {
			freeCalls++
			if got != effect {
				t.Fatalf("freed effect = %p", got)
			}
		}
		runtime.createAt = func(*server.Object, *server.Object, types.Pointf) { t.Fatal("effect created") }
		runtime.pointFX = func(netmsg.Op, types.Pointf) { t.Fatal("point FX sent") }
		runtime.audio = func(sound.ID, *server.Object, int, uint32) { t.Fatal("audio called") }
		runtime.setPlayerState = func(*server.Object, server.PlayerState) bool { t.Fatal("player state set"); return false }
		pauseFXStart57AF30(&state, nil, 0, runtime)
		if freeCalls != 1 || !state.active || !state.timed || state.effect != nil || state.unit != nil {
			t.Fatalf("missing-unit state = %+v, free calls = %d", state, freeCalls)
		}
	})
}

func TestPauseFXStart57AF30HonorsActiveAndGlobalPauseGuards(t *testing.T) {
	for _, test := range []struct {
		name   string
		state  pauseFXState57AF30
		paused bool
	}{
		{name: "active", state: pauseFXState57AF30{active: true}},
		{name: "paused", paused: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			runtime := pauseFXTestRuntime57AF30()
			calls := 0
			runtime.isPaused = func() bool {
				calls++
				return test.paused
			}
			runtime.newObject = func(string) *server.Object {
				calls++
				return nil
			}
			before := test.state
			pauseFXStart57AF30(&test.state, new(server.Object), 0, runtime)
			wantCalls := 1
			if before.active {
				wantCalls = 0
			}
			if calls != wantCalls || test.state != before {
				t.Fatalf("calls/state = %d/%+v, want %d/%+v", calls, test.state, wantCalls, before)
			}
		})
	}
}

func TestPauseFXFinish57B0A0RestoresPlayerAndPause(t *testing.T) {
	unit := &server.Object{PosVec: types.Pointf{X: 45, Y: 67}}
	effect := new(server.Object)
	state := pauseFXState57AF30{active: true, timed: true, mode: 0, started: 99, unit: unit, effect: effect}
	runtime := pauseFXTestRuntime57AF30()
	var events []string
	runtime.pointFX = func(op netmsg.Op, pos types.Pointf) {
		events = append(events, "point")
		if op != pauseFXPointOpcode57AF30 || pos != unit.Pos() {
			t.Fatalf("point FX = %d/%v", op, pos)
		}
	}
	runtime.delayedDelete = func(got *server.Object) {
		events = append(events, "delete")
		if got != effect {
			t.Fatalf("deleted effect = %p", got)
		}
	}
	runtime.setPlayerState = func(got *server.Object, playerState server.PlayerState) bool {
		events = append(events, "state")
		if got != unit || playerState != server.PlayerState13 {
			t.Fatalf("player state = %p/%d", got, playerState)
		}
		return false
	}
	runtime.isSpellBookOpen = func() bool {
		events = append(events, "book")
		return false
	}
	runtime.setPause = func(value bool) {
		events = append(events, "pause")
		if value {
			t.Fatal("pause enabled while finishing")
		}
	}

	pauseFXFinish57B0A0(&state, runtime)

	wantEvents := []string{"point", "delete", "state", "book", "pause"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	if state.active || state.unit != nil || state.effect != nil || !state.timed || state.mode != 0 || state.started != 99 {
		t.Fatalf("finished state = %+v", state)
	}
}

func TestPauseFXFinish57B0A0KeepsPauseWhileSpellBookOpen(t *testing.T) {
	state := pauseFXState57AF30{active: true, mode: 2}
	runtime := pauseFXTestRuntime57AF30()
	runtime.isSpellBookOpen = func() bool { return true }
	runtime.setPause = func(bool) { t.Fatal("pause changed") }
	pauseFXFinish57B0A0(&state, runtime)
	if state.active {
		t.Fatal("state remained active")
	}
}

func TestPauseFXExpired57B140UsesStrictFiveSecondBoundary(t *testing.T) {
	state := pauseFXState57AF30{timed: true, started: 100}
	if pauseFXExpired57B140(&state, 5100) {
		t.Fatal("expired at the exact boundary")
	}
	if !pauseFXExpired57B140(&state, 5101) {
		t.Fatal("did not expire after the boundary")
	}
	state.timed = false
	if pauseFXExpired57B140(&state, math.MaxUint64) {
		t.Fatal("untimed effect expired")
	}
}

func TestPauseFXStartExport57AF30PreservesNativePointerAndMode(t *testing.T) {
	unit := new(server.Object)
	var (
		gotUnit *server.Object
		gotMode int32
		calls   int
	)
	old := pauseFXStartCall57AF30
	pauseFXStartCall57AF30 = func(unit *server.Object, mode int32) {
		gotUnit = unit
		gotMode = mode
		calls++
	}
	t.Cleanup(func() { pauseFXStartCall57AF30 = old })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	pauseFXStartExportCall57AF30(unit, math.MinInt32)
	if calls != 1 || gotUnit != unit || gotMode != math.MinInt32 {
		t.Fatalf("C export = calls %d, unit %p, mode %d; want 1, %p, %d", calls, gotUnit, gotMode, unit, int32(math.MinInt32))
	}
	runtime.KeepAlive(unit)
}
