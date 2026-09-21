package server

import (
	"math"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterActionGet50A020UsesNativeStackHead(t *testing.T) {
	update := new(MonsterUpdateData)
	unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	update.AIStackInd = 2
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	update.AIStack[1].Action = uint32(ai.ACTION_FIGHT)
	update.AIStack[2].Action = uint32(ai.ACTION_FLEE)

	if got := unit.MonsterActionGet50A020(); got != ai.ACTION_FLEE {
		t.Fatalf("current action = %s, want FLEE", got)
	}
	update.AIStackInd = -1
	if got := unit.MonsterActionGet50A020(); got != ai.ACTION_INVALID {
		t.Fatalf("empty-stack action = %s, want INVALID", got)
	}
	update.AIStackInd = int8(len(update.AIStack))
	if got := unit.MonsterActionGet50A020(); got != ai.ACTION_INVALID {
		t.Fatalf("out-of-range action = %s, want INVALID", got)
	}
	if got := (*Object)(nil).MonsterActionGet50A020(); got != ai.ACTION_INVALID {
		t.Fatalf("nil-object action = %s, want INVALID", got)
	}
}

func TestMonsterActionPrevious50A040SkipsEveryPositiveCondition(t *testing.T) {
	update := new(MonsterUpdateData)
	update.AIStackInd = 5
	update.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
	update.AIStack[1].Action = uint32(ai.DEPENDENCY_TIME)
	update.AIStack[2].Action = 72
	update.AIStack[3].Action = math.MaxInt32
	update.AIStack[4].Action = uint32(ai.DEPENDENCY_OR)
	update.AIStack[5].Action = uint32(ai.ACTION_MOVE_TO)

	if got := update.MonsterActionPrevious50A040(); got != ai.ACTION_ESCORT {
		t.Fatalf("previous action = %s, want ESCORT", got)
	}

	// SETG is signed: a sign-bit action is not a condition and terminates the
	// reverse scan even though its unsigned value is above every named action.
	update.AIStack[3].Action = 0x80000000
	if got := update.MonsterActionPrevious50A040(); got != ai.ActionType(0x80000000) {
		t.Fatalf("signed previous action = %#x, want %#x", uint32(got), uint32(0x80000000))
	}

	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.DEPENDENCY_TIME)
	if got := update.MonsterActionPrevious50A040(); got != ai.ACTION_INVALID {
		t.Fatalf("condition-only previous action = %s, want INVALID", got)
	}
	update.AIStackInd = int8(len(update.AIStack))
	if got := update.MonsterActionPrevious50A040(); got != ai.ACTION_INVALID {
		t.Fatalf("out-of-range previous action = %s, want INVALID", got)
	}
}

func TestMonsterActionPushIfChanged50A360(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.Objs.init(s.handle)
	if !s.Objs.Init(1) {
		t.Fatal("object allocator initialization failed")
	}
	t.Cleanup(s.Objs.FreeObjects)

	unit := s.Objs.NewObject(&ObjectType{})
	update := new(MonsterUpdateData)
	var pin runtime.Pinner
	pin.Pin(update)
	t.Cleanup(pin.Unpin)
	unit.ObjClass = object.ClassMonster
	unit.UpdateData = unsafe.Pointer(update)
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)

	if got := unit.MonsterActionPushIfChanged50A360(ai.ACTION_IDLE); got != nil {
		t.Fatalf("same-action push = %p, want nil", got)
	}
	got := unit.MonsterActionPushIfChanged50A360(ai.ACTION_FLEE)
	if got == nil || got != &update.AIStack[0] || got.Type() != ai.ACTION_FLEE {
		t.Fatalf("different-action push = %p/%v, want native FLEE head %p", got, got.Type(), &update.AIStack[0])
	}
	if update.AIStackInd != 0 {
		t.Fatalf("stack index = %d, want 0 after replacing IDLE", update.AIStackInd)
	}

	unit.ObjClass = object.ClassPlayer
	if got := unit.MonsterActionPushIfChanged50A360(ai.ACTION_FIGHT); got != nil {
		t.Fatalf("non-monster push = %p, want nil", got)
	}
}
