package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type flameCleanseLegacyServer53D510 struct {
	Server
	srv     *server.Server
	deleted *server.Object
}

func (s *flameCleanseLegacyServer53D510) S() *server.Server                { return s.srv }
func (s *flameCleanseLegacyServer53D510) DelayedDelete(obj *server.Object) { s.deleted = obj }

func TestFlameCleanseUpdate53D510NativeDispatch(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("FlameCleanseUpdate")
	if !ok || callback == nil || callback != Get_nox_xxx_updateFlameCleanse_53D510() || size != 0 {
		t.Fatalf("FlameCleanseUpdate registration=%p/%d/%t", callback, size, ok)
	}
	obj, freeObj := alloc.New(server.Object{})
	defer freeObj()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatalf("object=%p, want C-owned pointer above 4 GiB", obj)
	}
	outer := &flameCleanseLegacyServer53D510{srv: new(server.Server)}
	outer.srv.SetFrame(8)
	obj.Field34, obj.Update = 8, callback
	before := *obj
	previous := GetServer
	t.Cleanup(func() { GetServer = previous })
	GetServer = func() Server { return outer }
	// Call the real Object dispatcher and native update implementation; only
	// the outer deletion service is observed, not a supplied update result.
	obj.CallUpdate()
	if outer.deleted != obj || *obj != before {
		t.Fatalf("native delete=%p, want %p; mutated=%t", outer.deleted, obj, *obj != before)
	}
}
