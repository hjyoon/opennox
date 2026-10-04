package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func monsterLookAtFixture5125A0(t *testing.T) (*server.Server, *server.Object, *server.MonsterUpdateData) {
	t.Helper()
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	if !srv.Objs.Init(1) {
		t.Fatal("cannot initialize native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	oldGame := noxflags.GetGame() & (noxflags.GameFlag22 | noxflags.GameFlag23)
	noxflags.UnsetGame(noxflags.GameFlag22 | noxflags.GameFlag23)
	t.Cleanup(func() { noxflags.SetGame(oldGame) })
	oldEngine := noxflags.GetEngine() & noxflags.EngineShowAI
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() { noxflags.SetEngine(oldEngine) })
	unit := srv.Objs.NewObject(&server.ObjectType{})
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	unit.ObjClass = object.ClassMonster
	unit.UpdateData = unsafe.Pointer(update)
	unit.PosVec = types.Ptf(8640, -3120)
	unit.Direction1, unit.Direction2 = 13, 29
	*update = server.MonsterUpdateData{
		Field2: 2, Field67: 67, Field74: 74, Field91: 91,
		Field120_0: 7, Field120_1: 1, Field120_2: 2, Field120_3: 3,
		Field124: 124, Field137: 137, AIStackInd: 0,
	}
	update.AIStack[0] = server.AIStackItem{
		Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{1, 2, 3, 4},
	}
	srv.SetFrame(1400)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{"unit": unsafe.Pointer(unit), "update": unsafe.Pointer(update)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("%s pointer=%p, want address above 4 GiB", name, ptr)
			}
		}
	}
	return srv, unit, update
}

// Both records are C-owned; exercise the real existing Go -> C entrypoint.
// The original must add FACE_LOCATION, not clear WAIT or assign directions.
func TestMonsterLookAtRealCWrapper5125A0NativeStack(t *testing.T) {
	angles := [...]byte{160, 192, 224, 128, 0, 0, 96, 64, 32}
	for direction, angle := range angles {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			srv, unit, update := monsterLookAtFixture5125A0(t)
			previous := update.AIStack[0]
			cosine, sine := server.SinCosDir(angle)
			wantX := math.Float32bits(float32(float64(cosine)*10 + float64(unit.PosVec.X)))
			wantY := math.Float32bits(float32(float64(sine)*10 + float64(unit.PosVec.Y)))
			t.Logf("actual LookAt C wrapper: unit=%p update=%p direction=%d", unit, update, direction)
			Nox_xxx_monsterLookAt_5125A0(unit, direction)
			if update.AIStackInd != 1 || update.AIStack[0] != previous {
				t.Fatalf("previous stack changed: index=%d base=%+v", update.AIStackInd, update.AIStack[0])
			}
			head := &update.AIStack[1]
			if head.Type() != ai.ACTION_FACE_LOCATION || head.Args != [4]uintptr{uintptr(wantX), uintptr(wantY), 0, 0} || head.Field5 != 0 {
				t.Fatalf("face action=%+v want coordinates %08x/%08x", head, wantX, wantY)
			}
			if unit.Direction1 != 13 || unit.Direction2 != 29 || !srv.AI.StackChanged {
				t.Fatalf("direction changed immediately or push missing: %d/%d changed=%t", unit.Direction1, unit.Direction2, srv.AI.StackChanged)
			}
			if update.Field2 != 0 || update.Field67 != 0 || update.Field74 != 0 || update.Field91 != 0 || update.Field120_0 != 7 ||
				update.Field120_1 != 0 || update.Field120_2 != 0 || update.Field120_3 != 0 || update.Field124 != 1400 || update.Field137 != 1400 {
				t.Fatal("LookAt did not use real action-stack reset services")
			}
		})
	}
}

func TestMonsterLookAtCResult5125A0NativeStack(t *testing.T) {
	for _, mode := range []string{"wait", "idle", "full", "dead head"} {
		t.Run(mode, func(t *testing.T) {
			srv, unit, update := monsterLookAtFixture5125A0(t)
			switch mode {
			case "idle":
				update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
			case "full":
				update.AIStackInd = int8(len(update.AIStack) - 1)
			case "dead head":
				update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
			}
			previous := *update
			result := monsterLookAtCResult5125A0(unit, 8)
			if mode == "full" || mode == "dead head" {
				if result != 0 || *update != previous || srv.AI.StackChanged {
					t.Fatal("C entry changed rejected stack or returned a nonzero result")
				}
				return
			}
			index := 1
			if mode == "idle" {
				index = 0
			}
			want := uintptr(unsafe.Pointer(&update.AIStack[index]))
			if result != want || update.AIStackInd != int8(index) || !srv.AI.StackChanged {
				t.Fatalf("C residual action=%x/index=%d want %x/%d", result, update.AIStackInd, want, index)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && result <= math.MaxUint32 {
				t.Fatal("C residual lost native action address")
			}
		})
	}
}

func TestMonsterLookAtCEntry5125A0GatesWithoutServerLookup(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class object.Class
		flags object.Flags
	}{
		{"player", object.ClassPlayer, 0},
		{"nonmonster high class bit", object.ClassImmobile, 0},
		{"dead monster", object.ClassMonster, object.FlagDead},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, free := alloc.New(server.Object{ObjClass: tc.class, ObjFlags: tc.flags})
			t.Cleanup(free)
			// No server handle or update data: original gates need neither.
			angles := [...]uintptr{160, 192, 224, 128, 0, 0, 96, 64, 32}
			for direction, want := range angles {
				if result := monsterLookAtCResult5125A0(unit, int32(direction)); result != want {
					t.Fatalf("C gated direction=%d result=%d want %d", direction, result, want)
				}
			}
		})
	}
}

func TestMonsterLookAtCBridge5125A0PreservesPointerDirectionAndResult(t *testing.T) {
	unit, free := alloc.New(server.Object{})
	t.Cleanup(free)
	previous := monsterLookAtCall5125A0
	t.Cleanup(func() { monsterLookAtCall5125A0 = previous })
	for _, tc := range []struct {
		name      string
		unit      *server.Object
		direction int32
		result    uint64
	}{
		{"native identity", unit, math.MaxInt32, 0x7f3527c00120},
		{"null forwarded", nil, math.MinInt32, 160},
		{"zero result", unit, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			monsterLookAtCall5125A0 = func(got *server.Object, direction int32) uintptr {
				calls++
				if got != tc.unit || direction != tc.direction {
					t.Fatalf("C bridge arguments=%p/%d want %p/%d", got, direction, tc.unit, tc.direction)
				}
				return uintptr(tc.result)
			}
			if result := monsterLookAtCResult5125A0(tc.unit, tc.direction); result != uintptr(tc.result) || calls != 1 {
				t.Fatalf("C residual/calls=%x/%d want %x/1", result, calls, uintptr(tc.result))
			}
		})
	}
}

func TestMonsterLookAtRealCWrapper5125A0FaceActionCompletes(t *testing.T) {
	for _, direction := range []int{1, 7} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			srv, unit, update := monsterLookAtFixture5125A0(t)
			unit.Direction1, unit.Direction2 = 0, 0
			Nox_xxx_monsterLookAt_5125A0(unit, direction)
			for tick := 0; tick < 32 && update.AIStackInd == 1; tick++ {
				oldDirection := unit.Direction2
				srv.MonsterActionFaceLocation545210(unit)
				// The normal unit update commits the pending facing each tick.
				unit.Direction1 = unit.Direction2
				if tick == 0 {
					want := server.RoundDir(int(oldDirection) + 8)
					if direction == 1 {
						want = server.RoundDir(int(oldDirection) - 8)
					}
					if unit.Direction2 != want {
						t.Fatalf("first facing step=%d want %d", unit.Direction2, want)
					}
				}
			}
			if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_WAIT {
				t.Fatalf("facing did not complete back to the previous WAIT: index=%d head=%+v", update.AIStackInd, update.AIStackHead())
			}
		})
	}
}
