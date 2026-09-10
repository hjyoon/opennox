package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestAudioEventZoneExport501C00PreservesNativePointersAndUnsignedResult(t *testing.T) {
	position, freePosition := alloc.New(types.Ptf(123.5, -456.75))
	defer freePosition()
	object, freeObject := alloc.New(server.Object{})
	defer freeObject()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"position": unsafe.Pointer(position),
			"object":   unsafe.Pointer(object),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	oldCall := audioEventZoneCall501C00
	t.Cleanup(func() { audioEventZoneCall501C00 = oldCall })
	type call struct {
		position *types.Pointf
		object   *server.Object
	}
	var calls []call
	audioEventZoneCall501C00 = func(gotPosition *types.Pointf, gotObject *server.Object) uint8 {
		calls = append(calls, call{position: gotPosition, object: gotObject})
		return 0xe7
	}

	if got := audioEventZoneExportCall501C00(position, object); got != 0xe7 {
		t.Fatalf("export result = %#x, want 0xe7", got)
	}
	if got := audioEventZoneExportCall501C00(nil, nil); got != 0xe7 {
		t.Fatalf("nil export result = %#x, want 0xe7", got)
	}
	if len(calls) != 2 || calls[0].position != position || calls[0].object != object ||
		calls[1].position != nil || calls[1].object != nil {
		t.Fatalf("calls = %+v, want [%p/%p nil/nil]", calls, position, object)
	}
	runtime.KeepAlive(position)
	runtime.KeepAlive(object)
}

func TestAudioEventZoneLegacyWrapper501C00UsesGoPath(t *testing.T) {
	oldCall := audioEventZoneCall501C00
	t.Cleanup(func() { audioEventZoneCall501C00 = oldCall })
	position := types.Ptf(7.25, -8.5)
	object := &server.Object{}
	audioEventZoneCall501C00 = func(gotPosition *types.Pointf, gotObject *server.Object) uint8 {
		if gotPosition == nil || *gotPosition != position || gotObject != object {
			t.Fatalf("position/object = %v/%p, want %v/%p", gotPosition, gotObject, position, object)
		}
		return 0xa5
	}
	if got := Sub_501C00(position, object); got != 0xa5 {
		t.Fatalf("wrapper result = %#x, want 0xa5", got)
	}
}
