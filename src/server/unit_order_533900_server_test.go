package server

import (
	"math"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func unitOrderTestServer533900(t *testing.T) *Server {
	t.Helper()
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	return s
}

func unitOrderTestMonster533900(s *Server) (*Object, *MonsterUpdateData) {
	update := &MonsterUpdateData{
		MonsterDef:       &MonsterDef{},
		AIStackInd:       0,
		Aggression2:      9.25,
		StatusFlags:      object.MonsterStatus(0xe5),
		WeaponEquipFlags: 0x00000002,
	}
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	unit := &Object{
		TypeInd:      0x1234,
		ObjClass:     object.ClassMonster,
		ObjSubClass:  object.SubClass(object.MonsterNPC),
		PosVec:       types.Ptf(41.25, -82.5),
		Direction1:   Dir16(0x1234),
		SpeedBase:    unitOrderMovingSpeed5339A0,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	return unit, update
}

func TestServerOrderUnit533900NativeActionsAndPointers(t *testing.T) {
	s := unitOrderTestServer533900(t)
	source := &Object{PosVec: types.Ptf(301.5, -602.25), serverHandle: s.handle}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want native address above 4 GiB", source)
	}
	runtimeHooks := UnitOrderRuntime533900{
		MonsterDefByType: func(int) *MonsterDef {
			t.Fatal("cached MonsterDef unexpectedly reloaded")
			return nil
		},
		Banish:  func(*Object) { t.Fatal("unexpected banish") },
		Observe: func(*Object, *Object) { t.Fatal("unexpected observe") },
	}
	tests := []struct {
		name       string
		order      uint32
		action     ai.ActionType
		aggression uint32
		status     uint32
	}{
		{"idle", 2, ai.ACTION_IDLE, 0x3f000000, 0xa5},
		{"guard", 3, ai.ACTION_GUARD, 0x3f000000, 0xe5},
		{"escort", 4, ai.ACTION_ESCORT, 0x3f547ae1, 0xa5},
		{"hunt", 5, ai.ACTION_HUNT, 0x3f547ae1, 0xa5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit, update := unitOrderTestMonster533900(s)
			unit.ObjOwner = source
			s.OrderUnit533900(source, unit, tc.order, runtimeHooks)
			head := update.AIStackHead()
			if head == nil || head.Type() != tc.action {
				t.Fatalf("action = %#v, want %s", head, tc.action)
			}
			if got := math.Float32bits(update.Aggression); got != tc.aggression {
				t.Fatalf("aggression bits = %#08x, want %#08x", got, tc.aggression)
			}
			if update.Aggression2 != 9.25 {
				t.Fatalf("Aggression2 = %g, want untouched 9.25", update.Aggression2)
			}
			if got := uint32(update.StatusFlags); got != tc.status {
				t.Fatalf("status = %#08x, want %#08x", got, tc.status)
			}
			switch tc.action {
			case ai.ACTION_GUARD:
				if got := head.ArgPos(0); got != unit.PosVec {
					t.Fatalf("guard position = %v, want %v", got, unit.PosVec)
				}
				if got := head.ArgU32(2); got != uint32(unit.Direction1) {
					t.Fatalf("guard direction = %#x, want %#x", got, unit.Direction1)
				}
				if got := math.Float32bits(update.SightRange); got != 0x437a0000 {
					t.Fatalf("guard sight bits = %#08x, want 0x437a0000", got)
				}
			case ai.ACTION_ESCORT:
				if got := head.ArgPos(0); got != source.PosVec {
					t.Fatalf("escort position = %v, want %v", got, source.PosVec)
				}
				if got := head.ArgObj(2); got != source {
					t.Fatalf("escort source = %p, want native pointer %p", got, source)
				}
			}
			runtime.KeepAlive(unit)
		})
	}
	runtime.KeepAlive(source)
}

func TestServerOrderUnit533900ExactMovementThreshold(t *testing.T) {
	s := unitOrderTestServer533900(t)
	source := &Object{serverHandle: s.handle}
	unit, update := unitOrderTestMonster533900(s)
	unit.SpeedBase = math.Float32frombits(math.Float32bits(unitOrderMovingSpeed5339A0) - 1)
	s.OrderUnit533900(source, unit, 3, UnitOrderRuntime533900{
		MonsterDefByType: func(int) *MonsterDef { return nil },
	})
	if update.AIStackHead().Type() != ai.ACTION_IDLE || update.Aggression != 0 || update.SightRange != 0 {
		t.Fatalf("below-threshold state changed: action=%s aggression=%g sight=%g",
			update.AIStackHead().Type(), update.Aggression, update.SightRange)
	}
}
