package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestArachnaphobiaNativeSharedCache53DA60(t *testing.T) {
	cache := memmap.PtrUint32(0x5D4594, 2488708)
	old := *cache
	defer func() { *cache = old }()
	h := arachnaphobiaRawRuntime53DA60()
	*cache = 0
	if h.LoadTypeIndex() != 0 || arachnaphobiaCacheProbe53DA60() != 0 {
		t.Fatal("cache did not share C DWORD")
	}
	h.StoreTypeIndex(0xffffffff)
	if *cache != 0xffffffff || arachnaphobiaCacheProbe53DA60() != 0xffffffff {
		t.Fatal("cache truncated or detached")
	}
	frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
	source := (*server.Object)(unsafe.Pointer(&frame[32]))
	*source = server.Object{Field34: 9, Field32: 1}
	h.Frame = func() uint32 { return 10 }
	h.FPS = func() uint32 { return 30 }
	h.TypeIndexByID = func(string) uint32 { t.Fatal("nonzero cache looked up again"); return 0 }
	h.NewObjectByTypeInd = func(index int32) *server.Object {
		if index != -1 {
			t.Fatal("DWORD index sign changed")
		}
		return nil
	}
	h.RandomInt = func(a, b int) int { return 3 }
	oldRuntime := arachnaphobiaRawRuntime53DA60
	defer func() { arachnaphobiaRawRuntime53DA60 = oldRuntime }()
	arachnaphobiaRawRuntime53DA60 = func() server.ArachnaphobiaRawRuntime53DA60 { return h }
	stateUpdateNativeProbe53DA60(0, source)
	if source.Field34 != 13 || *cache != 0xffffffff {
		t.Fatal("cached allocation/deadline")
	}
	source.Field34 = 9
	arachnaphobiaUpdateNative53DA60(source)
	if source.Field34 != 13 || *cache != 0xffffffff {
		t.Fatal("registered Go path did not share cached-index core")
	}
}

func TestStateUpdateNativeStatusAndListServices53DB30(t *testing.T) {
	for _, kind := range []int{1, 2, 3} {
		t.Run([]string{"", "break", "open", "break-remove"}[kind], func(t *testing.T) {
			frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
			source := (*server.Object)(unsafe.Pointer(&frame[32]))
			*source = server.Object{ObjClass: object.ClassImmobile, ObjFlags: 0x8000, Field5: 2}
			world := new(server.Server)
			world.Objs.AddToUpdatable(source)
			now := uint32(10)
			deletes := 0
			h := stateUpdateRawRuntime53DB30()
			h.Frame = func() uint32 { return now }
			h.FPS = func() uint32 { return 30 }
			h.RemoveFromUpdatable = world.Objs.RemoveFromUpdatable
			h.DelayedDelete = func(o *server.Object) {
				if o != source || o.IsUpdatable != 0 || world.Objs.UpdatableList != nil {
					t.Fatal("delete before list removal")
				}
				deletes++
			}
			old := stateUpdateRawRuntime53DB30
			defer func() { stateUpdateRawRuntime53DB30 = old }()
			stateUpdateRawRuntime53DB30 = func() server.StateUpdateRawRuntime53DB30 { return h }
			stateUpdateNativeProbe53DA60(kind, source)
			if source.Field5 != 4 || source.Field34 != 70 {
				t.Fatal("initial status/deadline")
			}
			for _, v := range source.Field140 {
				if v&0x80000 == 0 {
					t.Fatal("real status service did not mark synchronization")
				}
			}
			now = 70
			stateUpdateNativeProbe53DA60(kind, source)
			if source.Field5 != 4 {
				t.Fatal("strict deadline equality")
			}
			now = 71
			stateUpdateNativeProbe53DA60(kind, source)
			if source.Field5 != 8 {
				t.Fatal("final state")
			}
			stateUpdateNativeProbe53DA60(kind, source)
			wantDeletes := 0
			if kind == 3 {
				wantDeletes = 1
			}
			if source.IsUpdatable != 0 || world.Objs.UpdatableList != nil || deletes != wantDeletes {
				t.Fatal("final removal/deletion")
			}
		})
	}
}
