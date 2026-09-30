package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMimicCheckMorph534950CEntryDelegatesWholeBody(t *testing.T) {
	obj, freeObj := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeObj)
	t.Cleanup(freeUpdate)
	obj.ObjClass, obj.UpdateData = object.ClassMonster, unsafe.Pointer(update)
	update.AIStack[0].Action = 34
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(obj.CObj()) <= math.MaxUint32 {
		t.Fatalf("object=%p, want above 4 GiB", obj)
	}
	oldCall := mimicCheckMorphCall534950
	t.Cleanup(func() { mimicCheckMorphCall534950 = oldCall })
	var calls []*server.Object
	mimicCheckMorphCall534950 = func(obj *server.Object) { calls = append(calls, obj) }
	// Both the retained C entry and the public Go wrapper must reach the whole
	// native body without reading a PE32 offset or narrowing the argument.
	mimicCheckMorphCEntry534950(obj)
	Nox_xxx_monsterMimicCheckMorph_534950(obj)
	mimicCheckMorphCEntry534950(nil)
	if len(calls) != 3 || calls[0] != obj || calls[1] != obj || calls[2] != nil {
		t.Fatalf("delegation=%v, want exact obj/obj/nil", calls)
	}
}

func TestMimicCheckMorph534950CEntryProductionStack(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Types.Free()
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "MimicMorphCFixture"}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(1) {
		t.Fatal("object allocator initialization failed")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	obj := srv.NewObjectByTypeID("MimicMorphCFixture")
	if obj == nil {
		t.Fatal("native object allocation failed")
	}
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	obj.ObjClass, obj.UpdateData = object.ClassMonster, unsafe.Pointer(update)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, ptr := range []unsafe.Pointer{obj.CObj(), unsafe.Pointer(update), update.AIStack[0].C()} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native pointer=%p, want above 4 GiB", ptr)
			}
		}
	}
	oldGetServer, oldCall := GetServer, mimicCheckMorphCall534950
	GetServer = func() Server { return &questCanJoinLegacyServer4E4100{srv: srv} }
	t.Cleanup(func() { GetServer, mimicCheckMorphCall534950 = oldGetServer, oldCall })
	oldGame, oldEngine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.ResetGame()
	noxflags.ResetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	})
	srv.SetFrame(131)
	srv.SetTickRate(30)
	update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_IDLE)}
	update.Field137 = 100
	mimicCheckMorphCEntry534950(obj)
	if update.AIStackInd != 1 || update.AIStack[0].Type() != ai.DEPENDENCY_UNINTERRUPTABLE ||
		update.AIStack[1].Type() != ai.ACTION_MORPH_INTO_CHEST || update.Field137 != 131 || !srv.AI.StackChanged {
		t.Fatalf("idle production stack=%v timer=%d changed=%v", update.GetAIStack(), update.Field137, srv.AI.StackChanged)
	}
	// No native-call substitution: exercise the active/chest branch too.
	*update = server.MonsterUpdateData{StatusFlags: object.MonsterStatus(0x40000)}
	update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_FIGHT)}
	srv.AI.StackChanged = false
	Nox_xxx_monsterMimicCheckMorph_534950(obj)
	if update.AIStackInd != 2 || update.AIStack[0].Type() != ai.ACTION_FIGHT ||
		update.AIStack[1].Type() != ai.DEPENDENCY_UNINTERRUPTABLE || update.AIStack[2].Type() != ai.ACTION_MORPH_BACK_TO_SELF ||
		update.Field137 != 131 || !srv.AI.StackChanged {
		t.Fatalf("active production stack=%v timer=%d changed=%v", update.GetAIStack(), update.Field137, srv.AI.StackChanged)
	}
}
