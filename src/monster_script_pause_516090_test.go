package opennox

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	ns4 "github.com/opennox/noxscript/ns/v4"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectPauseUsesNativeWidthMonsterRoute516090(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Server.Objs.Init(1) {
		t.Fatal("cannot initialize server object allocator")
	}
	t.Cleanup(s.Server.Objs.FreeObjects)
	oldServer := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = oldServer })

	unit := s.Server.Objs.NewObject(&server.ObjectType{})
	update := new(server.MonsterUpdateData)
	unit.ObjClass = object.ClassMonster
	unit.UpdateData = unsafe.Pointer(update)
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_WAIT)
	s.SetFrame(100)

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("object pointer is not above 4 GiB: %p", unit)
	}
	asObjectS(unit).Pause(ns4.Frames(10))

	if update.AIStackInd != 2 ||
		update.AIStack[0].Type() != ai.ACTION_WAIT ||
		update.AIStack[1].Type() != ai.ACTION_REPORT || update.AIStack[1].ArgU32(0) != uint32(ai.ACTION_WAIT) ||
		update.AIStack[2].Type() != ai.ACTION_WAIT || update.AIStack[2].ArgU32(0) != 110 {
		t.Fatalf("script Pause stack = %#v, index %d", update.AIStack[:3], update.AIStackInd)
	}
	runtime.KeepAlive(unit)
}
