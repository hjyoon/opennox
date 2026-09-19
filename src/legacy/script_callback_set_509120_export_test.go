package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestScriptCallbackSetExport509120PreservesNativeObjectEventAndName(t *testing.T) {
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(object)) <= math.MaxUint32 {
		t.Fatalf("object pointer = %p, want native address above 4 GiB", object)
	}

	old := scriptCallbackSetCall509120
	t.Cleanup(func() { scriptCallbackSetCall509120 = old })
	type call struct {
		object *server.Object
		event  int
		name   string
	}
	var calls []call
	scriptCallbackSetCall509120 = func(gotObject *server.Object, event int, name string) {
		calls = append(calls, call{object: gotObject, event: event, name: name})
	}

	scriptCallbackSetExportCall509120(object, math.MinInt32, "OnLostEnemy")
	scriptCallbackSetExportCall509120(object, math.MaxInt32, "")

	if len(calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(calls))
	}
	if calls[0].object != object || calls[0].event != math.MinInt32 || calls[0].name != "OnLostEnemy" {
		t.Fatalf("first call = %#v, want object %p / INT32_MIN / OnLostEnemy", calls[0], object)
	}
	if calls[1].object != object || calls[1].event != math.MaxInt32 || calls[1].name != "" {
		t.Fatalf("second call = %#v, want object %p / INT32_MAX / empty", calls[1], object)
	}
	runtime.KeepAlive(object)
}
