package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Exercise the unchanged Go wrapper and the actual C entrypoint, not a
// substituted export callback. Both the object and its update are C-owned.
func TestUnitHuntRealCWrapper5157A0NativeStack(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	if !srv.Objs.Init(1) {
		t.Fatal("cannot initialize server object allocator")
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
				t.Fatalf("%s pointer=%p, want a native address above 4 GiB", name, ptr)
			}
		}
	}
	t.Logf("real C Hunt: unit=%p update=%p", unit, update)
	Nox_xxx_unitHunt_5157A0(unit)
	if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_HUNT ||
		update.AIStack[0].Args != ([4]uintptr{1, 2, 3, 4}) || update.AIStack[0].Field5 != 0 || !srv.AI.StackChanged {
		t.Fatalf("Hunt stack index=%d head=%+v changed=%t", update.AIStackInd, update.AIStack[0], srv.AI.StackChanged)
	}
	if update.Field2 != 0 || update.Field67 != 0 || update.Field74 != 0 || update.Field91 != 0 ||
		update.Field120_0 != 7 || update.Field120_1 != 0 || update.Field120_2 != 0 || update.Field120_3 != 0 ||
		update.Field124 != 1400 || update.Field137 != 1400 {
		t.Fatal("Hunt did not use the real action-stack reset services")
	}

	// The original gates must return before either action-stack service reads
	// the absent UpdateData or server handle on these C-owned objects.
	blocked, freeBlocked := alloc.New(server.Object{})
	t.Cleanup(freeBlocked)
	Nox_xxx_unitHunt_5157A0(nil)
	blocked.ObjClass = object.ClassPlayer
	Nox_xxx_unitHunt_5157A0(blocked)
	blocked.ObjClass, blocked.ObjFlags = object.ClassMonster, object.FlagDead
	Nox_xxx_unitHunt_5157A0(blocked)
}

func TestUnitHuntCBridge5157A0PreservesPointerAndNull(t *testing.T) {
	unit, free := alloc.New(server.Object{})
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("C-owned object pointer below 4 GiB: %p", unit)
	}
	oldCall := unitHuntCall5157A0
	t.Cleanup(func() { unitHuntCall5157A0 = oldCall })
	var got []*server.Object
	unitHuntCall5157A0 = func(unit *server.Object) { got = append(got, unit) }
	Nox_xxx_unitHunt_5157A0(unit)
	Nox_xxx_unitHunt_5157A0(nil)
	if len(got) != 2 || got[0] != unit || got[1] != nil {
		t.Fatalf("C bridge objects=%v want [%p nil]", got, unit)
	}
}

func TestUnitHuntCEntry5157A0GatesWithoutServerLookup(t *testing.T) {
	oldGetServer := GetServer
	GetServer = func() Server { panic("Hunt queried an active server before its original gates") }
	t.Cleanup(func() { GetServer = oldGetServer })
	unit, free := alloc.New(server.Object{})
	t.Cleanup(free)
	Nox_xxx_unitHunt_5157A0(nil)
	unit.ObjClass = object.ClassPlayer
	Nox_xxx_unitHunt_5157A0(unit)
	unit.ObjClass, unit.ObjFlags = object.ClassMonster, object.FlagDead
	Nox_xxx_unitHunt_5157A0(unit)
}
