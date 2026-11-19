//go:build amd64 || arm64

package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterActionZeroReferenceRemainsEmpty(t *testing.T) {
	setupStackMetadata238(t)
	oldData := datapath.Data()
	datapath.SetData(t.TempDir())
	t.Cleanup(func() { datapath.SetData(oldData) })
	srv := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	if !srv.Objs.InitHigh(8) {
		t.Fatal("init objects")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "TestActionNullRefs"}); err != nil {
		t.Fatal(err)
	}
	typ := srv.Types.ByID("TestActionNullRefs")
	obj := srv.Objs.NewObject(typ)
	obj.ObjClass = object.ClassMonster
	update, release := alloc.NewHigh(server.MonsterUpdateData{})
	t.Cleanup(release)
	obj.UpdateData = unsafe.Pointer(update)
	decoy := srv.Objs.NewObject(typ)
	decoy.ScriptIDVal = 0
	srv.Objs.List = decoy
	waypoint, release := alloc.NewHigh(server.Waypoint{})
	t.Cleanup(release)
	srv.WPs.List = waypoint
	if srv.ObjectByScriptID4ECF10(0) != decoy || srv.WPs.ByInd(0) != waypoint {
		t.Fatal("zero-ID controls are missing")
	}
	for _, kind := range []uint32{1, 2} {
		index, slot := firstMonsterActionArgumentKind(t, kind)
		*update = server.MonsterUpdateData{AIStackInd: 0}
		update.AIStack[0].Action = index
		applyMonsterXferRefs528DB0(obj, new(server.MonsterActionRefs))
		if got := update.AIStack[0].Args[slot]; got != 0 {
			t.Errorf("empty kind %d at action %d stores address %#x", kind, index, got)
		}
	}
}

func firstMonsterActionArgumentKind(t *testing.T, kind uint32) (uint32, int) {
	t.Helper()
	for action := uint32(0); action < 72; action++ {
		for arg, candidate := range stackKinds238(action) {
			if candidate == kind {
				return action, 2 * arg
			}
		}
	}
	t.Fatalf("no action with argument kind %d", kind)
	return 0, 0
}
